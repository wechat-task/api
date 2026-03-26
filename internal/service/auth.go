package service

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
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

func (s *AuthService) BeginRegistration(displayName, icon string) (*protocol.CredentialCreation, string, error) {
	webAuthnID := GenerateWebAuthnID()

	user := &model.User{
		DisplayName: displayName,
		Icon:        icon,
	}
	user.SetWebAuthnID(webAuthnID)

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

	sessionID, err := s.sessionService.CreateSession(*sessionData, "registration", nil)
	if err != nil {
		return nil, "", err
	}

	return options, sessionID, nil
}

func (s *AuthService) BeginLogin() (*protocol.CredentialAssertion, string, error) {
	options, sessionData, err := s.webAuthn.BeginDiscoverableLogin()
	if err != nil {
		return nil, "", err
	}

	sessionID, err := s.sessionService.CreateSession(*sessionData, "authentication", nil)
	if err != nil {
		return nil, "", err
	}

	return options, sessionID, nil
}

func (s *AuthService) GetWebAuthn() *webauthn.WebAuthn {
	return s.webAuthn
}

func (s *AuthService) GetUserRepository() *repository.UserRepository {
	return s.userRepo
}

func (s *AuthService) GetCredentialRepository() *repository.CredentialRepository {
	return s.credentialRepo
}

func (s *AuthService) GetSessionService() *SessionService {
	return s.sessionService
}
