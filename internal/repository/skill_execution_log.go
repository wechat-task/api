package repository

import (
	"time"

	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

type SkillExecutionLogRepository struct {
	db *gorm.DB
}

func NewSkillExecutionLogRepository(db *gorm.DB) *SkillExecutionLogRepository {
	return &SkillExecutionLogRepository{db: db}
}

// Create creates a new skill execution log
func (r *SkillExecutionLogRepository) Create(log *model.SkillExecutionLog) error {
	return r.db.Create(log).Error
}

// GetByID returns a log by its ID
func (r *SkillExecutionLogRepository) GetByID(id uint) (*model.SkillExecutionLog, error) {
	var log model.SkillExecutionLog
	err := r.db.First(&log, id).Error
	return &log, err
}

// GetBySubscriptionID returns all logs for a subscription
func (r *SkillExecutionLogRepository) GetBySubscriptionID(subscriptionID uint) ([]model.SkillExecutionLog, error) {
	var logs []model.SkillExecutionLog
	err := r.db.Where("subscription_id = ?", subscriptionID).
		Order("executed_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetBySubscriptionIDWithLimit returns logs for a subscription with limit
func (r *SkillExecutionLogRepository) GetBySubscriptionIDWithLimit(subscriptionID uint, limit int) ([]model.SkillExecutionLog, error) {
	var logs []model.SkillExecutionLog
	err := r.db.Where("subscription_id = ?", subscriptionID).
		Order("executed_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetByUserID returns all logs for a user (through subscription)
func (r *SkillExecutionLogRepository) GetByUserID(userID uint) ([]model.SkillExecutionLog, error) {
	var logs []model.SkillExecutionLog
	err := r.db.Joins("JOIN skill_subscriptions ON skill_subscriptions.id = skill_execution_logs.subscription_id").
		Where("skill_subscriptions.user_id = ?", userID).
		Order("skill_execution_logs.executed_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetBySkillID returns all logs for a skill (through subscription)
func (r *SkillExecutionLogRepository) GetBySkillID(skillID uint) ([]model.SkillExecutionLog, error) {
	var logs []model.SkillExecutionLog
	err := r.db.Joins("JOIN skill_subscriptions ON skill_subscriptions.id = skill_execution_logs.subscription_id").
		Where("skill_subscriptions.skill_id = ?", skillID).
		Order("skill_execution_logs.executed_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetRecentLogs returns recent logs across all subscriptions
func (r *SkillExecutionLogRepository) GetRecentLogs(limit int) ([]model.SkillExecutionLog, error) {
	var logs []model.SkillExecutionLog
	err := r.db.Preload("Subscription").
		Order("executed_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetFailedLogs returns failed execution logs
func (r *SkillExecutionLogRepository) GetFailedLogs(since time.Time) ([]model.SkillExecutionLog, error) {
	var logs []model.SkillExecutionLog
	err := r.db.Where("status = 'failed' AND executed_at >= ?", since).
		Order("executed_at DESC").
		Find(&logs).Error
	return logs, err
}

// CountBySubscriptionID returns the number of logs for a subscription
func (r *SkillExecutionLogRepository) CountBySubscriptionID(subscriptionID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.SkillExecutionLog{}).
		Where("subscription_id = ?", subscriptionID).
		Count(&count).Error
	return count, err
}

// GetAverageDuration returns the average execution duration for a subscription
func (r *SkillExecutionLogRepository) GetAverageDuration(subscriptionID uint) (float64, error) {
	var avgDuration float64
	err := r.db.Model(&model.SkillExecutionLog{}).
		Select("AVG(duration_ms)").
		Where("subscription_id = ? AND status = 'success'", subscriptionID).
		Scan(&avgDuration).Error
	return avgDuration, err
}

// GetTokenUsageSummary returns total token usage for a subscription
func (r *SkillExecutionLogRepository) GetTokenUsageSummary(subscriptionID uint) (int, error) {
	var totalTokens int
	err := r.db.Model(&model.SkillExecutionLog{}).
		Select("COALESCE(SUM(token_usage), 0)").
		Where("subscription_id = ?", subscriptionID).
		Scan(&totalTokens).Error
	return totalTokens, err
}

// DeleteOldLogs deletes logs older than specified date
func (r *SkillExecutionLogRepository) DeleteOldLogs(before time.Time) error {
	return r.db.Where("executed_at < ?", before).
		Delete(&model.SkillExecutionLog{}).Error
}
