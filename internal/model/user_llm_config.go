package model

import (
	"time"
)

// UserLLMConfig stores a user's LLM provider configuration
type UserLLMConfig struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	UserID    uint    `json:"user_id" gorm:"not null;index"`
	Name      string  `json:"name" gorm:"type:varchar(255);not null"`      // Configuration name (e.g., "My OpenAI")
	Provider  string  `json:"provider" gorm:"type:varchar(50);not null"`   // openai, anthropic, azure, local
	APIKey    string  `json:"-" gorm:"type:text;not null"`                 // Encrypted API key (not exposed in JSON)
	Model     string  `json:"model" gorm:"type:varchar(100);not null"`     // Model name
	BaseURL   *string `json:"base_url,omitempty" gorm:"type:varchar(255)"` // Custom base URL for self-hosted
	IsDefault bool    `json:"is_default" gorm:"default:false"`             // Default configuration for user

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
