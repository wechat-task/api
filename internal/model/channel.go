package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ChannelType enumerates supported channel types.
type ChannelType string

const (
	ChannelTypeWechatClawbot ChannelType = "wechat_clawbot"
	ChannelTypeLark          ChannelType = "lark"
)

// Channel represents a messaging channel bound to a bot.
type Channel struct {
	ID         uint          `json:"id" gorm:"primaryKey" example:"1"`
	BotID      uint          `json:"bot_id" gorm:"not null;index" example:"1"`
	Type       ChannelType   `json:"type" gorm:"type:varchar(50);not null" example:"wechat_clawbot"`
	Status     string        `json:"status" gorm:"not null;default:pending" example:"pending"` // pending, active, disconnected, expired
	Config     ChannelConfig `json:"config" gorm:"type:jsonb;default:'{}'"`
	LastCursor *string       `json:"last_cursor,omitempty" gorm:"column:last_cursor"`
	CreatedAt  time.Time     `json:"created_at" example:"2026-03-30T10:00:00Z"`
	UpdatedAt  time.Time     `json:"updated_at" example:"2026-03-30T10:00:00Z"`
}

// ChannelConfig stores type-specific configuration as JSONB.
type ChannelConfig map[string]interface{}

// Scan implements sql.Scanner for ChannelConfig.
func (c *ChannelConfig) Scan(value interface{}) error {
	if value == nil {
		*c = ChannelConfig{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer for ChannelConfig.
func (c ChannelConfig) Value() (driver.Value, error) {
	if c == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// GetString retrieves a string value from config.
func (c ChannelConfig) GetString(key string) string {
	if v, ok := c[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Set stores a key-value pair in config.
func (c ChannelConfig) Set(key string, value interface{}) {
	c[key] = value
}
