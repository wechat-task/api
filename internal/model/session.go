package model

import (
	"encoding/json"
	"github.com/go-webauthn/webauthn/webauthn"
	"time"
)

type Session struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Challenge   string    `json:"challenge" gorm:"not null;index"`
	SessionData []byte    `json:"session_data" gorm:"not null"`
	ExpiresAt   time.Time `json:"expires_at" gorm:"not null;index"`
	UserID      *uint     `json:"user_id" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Session) SetSessionData(data webauthn.SessionData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	s.SessionData = jsonData
	s.Challenge = data.Challenge
	return nil
}

func (s *Session) GetSessionData() (*webauthn.SessionData, error) {
	var data webauthn.SessionData
	err := json.Unmarshal(s.SessionData, &data)
	return &data, err
}
