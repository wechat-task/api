package service

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
	"time"
)

type SessionService struct {
	repo *repository.SessionRepository
}

func NewSessionService(repo *repository.SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) CreateSession(data webauthn.SessionData, sessionType string, userID *uint) (string, error) {
	session := &model.Session{
		ID:        generateSessionID(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
		UserID:    userID,
	}

	if err := session.SetSessionData(data); err != nil {
		return "", err
	}

	if err := s.repo.Create(session); err != nil {
		return "", err
	}

	return session.ID, nil
}

func (s *SessionService) GetSession(id string) (*model.Session, *webauthn.SessionData, error) {
	session, err := s.repo.GetByID(id)
	if err != nil {
		return nil, nil, err
	}

	data, err := session.GetSessionData()
	if err != nil {
		return nil, nil, err
	}

	return session, data, nil
}

func (s *SessionService) DeleteSession(id string) error {
	return s.repo.Delete(id)
}

func (s *SessionService) CleanupExpired() error {
	return s.repo.CleanupExpired()
}

func generateSessionID() string {
	return generateRandomString(32)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
