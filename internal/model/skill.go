package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// SkillVisibility defines who can see and access a skill
type SkillVisibility string

const (
	SkillVisibilityPrivate  SkillVisibility = "private"  // Only visible to creator
	SkillVisibilityPublic   SkillVisibility = "public"   // Visible to all users
	SkillVisibilityUnlisted SkillVisibility = "unlisted" // Accessible via link
)

// SkillStatus represents the publication state of a skill
type SkillStatus string

const (
	SkillStatusDraft     SkillStatus = "draft"     // Draft, not executable
	SkillStatusPublished SkillStatus = "published" // Published and executable
	SkillStatusArchived  SkillStatus = "archived"  // Archived, not executable
)

// Skill represents an AI skill that can be executed by LLMs
type Skill struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	UserID      uint   `json:"user_id" gorm:"not null;index"`
	Name        string `json:"name" gorm:"not null;type:varchar(255)"`
	Description string `json:"description" gorm:"type:text"`
	Version     string `json:"version" gorm:"type:varchar(50);default:'1.0.0'"`

	// Markdown content following Anthropic skill format
	Content string `json:"content" gorm:"type:text;not null"`

	// Metadata
	Visibility SkillVisibility `json:"visibility" gorm:"type:varchar(20);default:'private'"`
	Status     SkillStatus     `json:"status" gorm:"type:varchar(20);default:'draft'"`
	Category   string          `json:"category" gorm:"type:varchar(100)"`
	Tags       StringArray     `json:"tags" gorm:"type:jsonb;default:'[]'"`

	// Parameter definitions
	Parameters SkillParameters `json:"parameters" gorm:"type:jsonb;default:'{}'"`

	// Statistics
	SubscriberCount int `json:"subscriber_count" gorm:"default:0"`
	ExecutionCount  int `json:"execution_count" gorm:"default:0"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SkillParameters defines the parameters that a skill accepts
type SkillParameters map[string]SkillParameter

// SkillParameter defines a single parameter for a skill
type SkillParameter struct {
	Type        string   `json:"type"`                  // string, number, boolean, enum
	Description string   `json:"description"`           // Parameter description
	Required    bool     `json:"required"`              // Whether parameter is required
	Default     any      `json:"default,omitempty"`     // Default value
	EnumValues  []string `json:"enum_values,omitempty"` // Valid enum values
}

// StringArray is a custom type for storing string arrays in JSONB
type StringArray []string

// Scan implements sql.Scanner for StringArray
func (s *StringArray) Scan(value any) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for StringArray
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// Scan implements sql.Scanner for SkillParameters
func (s *SkillParameters) Scan(value any) error {
	if value == nil {
		*s = SkillParameters{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for SkillParameters
func (s SkillParameters) Value() (driver.Value, error) {
	if s == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}
