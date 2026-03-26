package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
	"time"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *model.Session) error {
	return r.db.Create(session).Error
}

func (r *SessionRepository) GetByID(id string) (*model.Session, error) {
	var session model.Session
	err := r.db.Where("id = ? AND expires_at > ?", id, time.Now()).First(&session).Error
	return &session, err
}

func (r *SessionRepository) GetByChallenge(challenge string) (*model.Session, error) {
	var session model.Session
	err := r.db.Where("challenge = ? AND expires_at > ?", challenge, time.Now()).First(&session).Error
	return &session, err
}

func (r *SessionRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Session{}).Error
}

func (r *SessionRepository) DeleteExpired() error {
	return r.db.Where("expires_at <= ?", time.Now()).Delete(&model.Session{}).Error
}

func (r *SessionRepository) CleanupExpired() error {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			r.DeleteExpired()
		}
	}()
	return nil
}
