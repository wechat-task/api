package model

import "time"

// Bot represents a user's iLink bot binding.
type Bot struct {
	ID          uint      `json:"id" gorm:"primaryKey" example:"1"`
	UserID      uint      `json:"user_id" gorm:"not null;index" example:"1"`
	Name        *string   `json:"name" gorm:"type:varchar(255)" example:"My Bot"`
	Description *string   `json:"description" gorm:"type:text" example:"Work assistant bot"`
	Status      string    `json:"status" gorm:"not null;default:pending" example:"pending"` // pending, active, disconnected, expired
	BotToken    *string   `json:"-" gorm:"column:bot_token"`
	BaseURL     *string   `json:"base_url" gorm:"column:base_url"`
	ILinkBotID  *string   `json:"ilink_bot_id" gorm:"column:ilink_bot_id"`
	ILinkUserID *string   `json:"ilink_user_id" gorm:"column:ilink_user_id"`
	LastCursor  *string   `json:"last_cursor" gorm:"column:last_cursor"`
	QRCodeID    *string   `json:"qrcode_id" gorm:"column:qrcode_id"`
	QRCodeImage *string   `json:"qrcode_image" gorm:"column:qrcode_image;type:text"`
	CreatedAt   time.Time `json:"created_at" example:"2026-03-30T10:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2026-03-30T10:00:00Z"`
}
