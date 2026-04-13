package model

import (
	"time"
)

// SkillExecutionLog records the execution of a skill subscription
type SkillExecutionLog struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	SubscriptionID uint      `json:"subscription_id" gorm:"not null;index"`
	Status         string    `json:"status" gorm:"type:varchar(20);not null"` // success, failed, cancelled
	Input          string    `json:"input" gorm:"type:text"`                  // Input sent to LLM
	Output         string    `json:"output" gorm:"type:text"`                 // Output from LLM
	Error          *string   `json:"error,omitempty" gorm:"type:text"`        // Error message if failed
	TokenUsage     int       `json:"token_usage" gorm:"default:0"`            // Tokens used
	DurationMs     int64     `json:"duration_ms" gorm:"default:0"`            // Execution duration in milliseconds
	ExecutedAt     time.Time `json:"executed_at" gorm:"not null"`             // When execution started
	CompletedAt    time.Time `json:"completed_at" gorm:"not null"`            // When execution completed
}
