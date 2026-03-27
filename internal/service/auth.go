package service

import (
	"errors"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
	"net/http"
)

type AuthService struct {
	webAuthn       *webauthn.WebAuthn
	userRepo       *repository.UserRepository
	credentialRepo *repository.CredentialRepository
	sessionService *SessionService
}

func NewAuthService(cfg webauthn.Config, userRepo *repository.UserRepository, credentialRepo *repository.CredentialRepository, sessionService *SessionService) (*AuthService, error) {
	wa, err := webauthn.New(&cfg)
	if err != nil {
		return nil, err
	}

	return &AuthService{
		webAuthn:       wa,
		userRepo:       userRepo,
		credentialRepo: credentialRepo,
		sessionService: sessionService,
	}, nil
}

func (s *AuthService) BeginAuth(username string) (*protocol.CredentialCreation, string, error) {
	webAuthnID := GenerateWebAuthnID()

	user := &model.User{}
	user.SetWebAuthnID(webAuthnID)
	var usernamePtr *string
	if username != "" {
		usernamePtr = &username
		user.Username = usernamePtr
	}

	options, sessionData, err := s.webAuthn.BeginRegistration(
		user,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.Platform,
			UserVerification:        protocol.VerificationPreferred,
		}),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
	)
	if err != nil {
		return nil, "", err
	}

	sessionID, err := s.sessionService.CreateSession(*sessionData, "auth", nil, usernamePtr)
	if err != nil {
		return nil, "", err
	}

	return options, sessionID, nil
}

func (s *AuthService) FinishAuth(sessionID string, r *http.Request) (*model.User, bool, error) {
	session, sessionData, err := s.sessionService.GetSession(sessionID)
	if err != nil {
		return nil, false, errors.New("invalid session")
	}

	user := &model.User{}
	user.SetWebAuthnID(sessionData.UserID)

	credential, err := s.webAuthn.FinishRegistration(user, *sessionData, r)
	if err != nil {
		return nil, false, err
	}

	credentialID := credential.ID

	dbCredential, err := s.credentialRepo.GetByCredentialID(credentialID)
	var existingUser *model.User
	isNewUser := false

	if err == nil {
		existingUser, err = s.userRepo.GetByID(dbCredential.UserID)
		if err != nil {
			return nil, false, err
		}
	} else {
		// Use username from session if provided during registration
		newUser := &model.User{}
		newUser.SetWebAuthnID(sessionData.UserID)
		newUser.Username = session.Username

		if err := s.userRepo.Create(newUser); err != nil {
			return nil, false, err
		}

		dbCredential = &model.Credential{}
		dbCredential.FromWebAuthnCredential(*credential, newUser.ID)

		if err := s.credentialRepo.Create(dbCredential); err != nil {
			return nil, false, err
		}

		existingUser = newUser
		isNewUser = true
	}

	if err := s.credentialRepo.UpdateSignCount(credentialID, credential.Authenticator.SignCount); err != nil {
		return nil, false, err
	}

	s.sessionService.DeleteSession(sessionID)

	return existingUser, isNewUser, nil
}
