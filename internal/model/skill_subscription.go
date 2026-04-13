package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// SkillSubscription represents a user's subscription to a skill
type SkillSubscription struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	UserID  uint `json:"user_id" gorm:"not null;index"`
	SkillID uint `json:"skill_id" gorm:"not null;index"`

	// Execution configuration
	Config SkillExecutionConfig `json:"config" gorm:"type:jsonb;default:'{}'"`
	Status string               `json:"status" gorm:"type:varchar(20);default:'active'"`

	// Delivery channel configuration
	BotID     *uint `json:"bot_id,omitempty"`
	ChannelID *uint `json:"channel_id,omitempty"`

	// Schedule overrides (overrides skill's default schedule)
	ScheduleCron *string `json:"schedule_cron,omitempty" gorm:"type:varchar(100)"`
	TimeZone     string  `json:"time_zone" gorm:"type:varchar(50);default:'UTC'"`

	// Next execution time (for scheduler)
	NextRunAt *time.Time `json:"next_run_at" gorm:"index"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Associations
	Skill   Skill    `json:"skill,omitempty" gorm:"foreignKey:SkillID"`
	Bot     *Bot     `json:"bot,omitempty" gorm:"foreignKey:BotID"`
	Channel *Channel `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
}

// SkillExecutionConfig holds execution configuration for a subscription
type SkillExecutionConfig struct {
	Parameters map[string]any `json:"parameters"`           // Parameter values for this subscription
	LLMConfig  *LLMConfig     `json:"llm_config,omitempty"` // User-specified LLM configuration
}

// LLMConfig holds configuration for LLM provider
type LLMConfig struct {
	Provider    string  `json:"provider"`           // openai, anthropic, azure, local
	APIKey      string  `json:"api_key"`            // Encrypted API key
	Model       string  `json:"model"`              // Model name (e.g., gpt-4, claude-3)
	BaseURL     string  `json:"base_url,omitempty"` // Custom base URL for self-hosted
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

// Scan implements sql.Scanner for SkillExecutionConfig
func (c *SkillExecutionConfig) Scan(value any) error {
	if value == nil {
		*c = SkillExecutionConfig{Parameters: make(map[string]any)}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer for SkillExecutionConfig
func (c SkillExecutionConfig) Value() (driver.Value, error) {
	if c.Parameters == nil {
		c.Parameters = make(map[string]any)
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// Scan implements sql.Scanner for LLMConfig
func (c *LLMConfig) Scan(value any) error {
	if value == nil {
		*c = LLMConfig{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer for LLMConfig
func (c LLMConfig) Value() (driver.Value, error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}
