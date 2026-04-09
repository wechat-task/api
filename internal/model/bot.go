package model

import "time"

// Bot represents a user's bot that can be bound to multiple messaging channels.
type Bot struct {
	ID          uint      `json:"id" gorm:"primaryKey" example:"1"`
	UserID      uint      `json:"user_id" gorm:"not null;index" example:"1"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null" example:"My Bot"`
	Description *string   `json:"description" gorm:"type:text" example:"Work assistant bot"`
	Status      string    `json:"status" gorm:"not null;default:pending" example:"pending"` // pending, active, disconnected, expired
	Channels    []Channel `json:"channels,omitempty" gorm:"foreignKey:BotID"`
	CreatedAt   time.Time `json:"created_at" example:"2026-03-30T10:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2026-03-30T10:00:00Z"`
}
