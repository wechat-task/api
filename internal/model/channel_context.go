package model

import "time"

// ChannelContext stores a user's contextToken for a channel, used for replying.
type ChannelContext struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ChannelID    uint      `json:"channel_id" gorm:"not null;uniqueIndex:idx_channel_user"`
	UserID       string    `json:"user_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_channel_user"`
	ContextToken string    `json:"context_token" gorm:"type:text;not null"`
	LastMessage  *string   `json:"last_message,omitempty" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
