package service

import (
	"errors"
	"net/http"

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

// BeginRegistration generates passkey registration options without creating the user yet.
// Username is saved in the session; user is created upon successful verification.
func (s *AuthService) BeginRegistration(username string) (*protocol.CredentialCreation, string, error) {
	webAuthnID := GenerateWebAuthnID()

	user := &model.User{}
	user.SetWebAuthnID(webAuthnID)
	user.Username = &username

	options, sessionData, err := s.webAuthn.BeginMediatedRegistration(
		user,
		protocol.MediationDefault,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.Platform,
			ResidentKey:             protocol.ResidentKeyRequirementRequired,
			UserVerification:        protocol.VerificationPreferred,
		}),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
	)
	if err != nil {
		return nil, "", err
	}

	sessionID, err := s.sessionService.CreateSession(*sessionData, "register", nil, &username)
	if err != nil {
		return nil, "", err
	}

	return options, sessionID, nil
}

// FinishRegistration verifies the passkey registration, creates the user and stores the credential.
func (s *AuthService) FinishRegistration(sessionID string, r *http.Request) (*model.User, error) {
	session, sessionData, err := s.sessionService.GetSession(sessionID)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	// Reconstruct the transient user from session data (webAuthnID from SessionData.UserID)
	user := &model.User{}
	user.SetWebAuthnID(sessionData.UserID)
	user.Username = session.Username

	credential, err := s.webAuthn.FinishRegistration(user, *sessionData, r)
	if err != nil {
		return nil, err
	}

	// Now create the user in DB
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	dbCredential := &model.Credential{}
	dbCredential.FromWebAuthnCredential(*credential, user.ID)

	if err := s.credentialRepo.Create(dbCredential); err != nil {
		return nil, err
	}

	s.sessionService.DeleteSession(sessionID)

	return user, nil
}

// BeginLogin generates passkey login options using discoverable credentials.
func (s *AuthService) BeginLogin() (*protocol.CredentialAssertion, string, error) {
	options, sessionData, err := s.webAuthn.BeginDiscoverableMediatedLogin(protocol.MediationDefault)
	if err != nil {
		return nil, "", err
	}

	sessionID, err := s.sessionService.CreateSession(*sessionData, "login", nil, nil)
	if err != nil {
		return nil, "", err
	}

	return options, sessionID, nil
}

// FinishLogin verifies the passkey login response and returns the matching user.
func (s *AuthService) FinishLogin(sessionID string, r *http.Request) (*model.User, error) {
	_, sessionData, err := s.sessionService.GetSession(sessionID)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	validatedUser, credential, err := s.webAuthn.FinishPasskeyLogin(
		s.discoverableUserHandler,
		*sessionData,
		r,
	)
	if err != nil {
		return nil, err
	}

	user, ok := validatedUser.(*model.User)
	if !ok {
		return nil, errors.New("invalid user type")
	}

	if err := s.credentialRepo.UpdateSignCount(credential.ID, credential.Authenticator.SignCount); err != nil {
		return nil, err
	}

	s.sessionService.DeleteSession(sessionID)

	return user, nil
}

// discoverableUserHandler looks up a user by their WebAuthn userHandle (which is the user's web_authn_id).
func (s *AuthService) discoverableUserHandler(rawID, userHandle []byte) (webauthn.User, error) {
	user, err := s.userRepo.GetByWebAuthnID(userHandle)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}
