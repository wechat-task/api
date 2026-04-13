package repository

import (
	"time"

	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

type SkillSubscriptionRepository struct {
	db *gorm.DB
}

func NewSkillSubscriptionRepository(db *gorm.DB) *SkillSubscriptionRepository {
	return &SkillSubscriptionRepository{db: db}
}

// Create creates a new skill subscription
func (r *SkillSubscriptionRepository) Create(subscription *model.SkillSubscription) error {
	return r.db.Create(subscription).Error
}

// GetByID returns a subscription by its ID
func (r *SkillSubscriptionRepository) GetByID(id uint) (*model.SkillSubscription, error) {
	var subscription model.SkillSubscription
	err := r.db.First(&subscription, id).Error
	return &subscription, err
}

// GetByIDWithSkill returns a subscription by its ID with skill association
func (r *SkillSubscriptionRepository) GetByIDWithSkill(id uint) (*model.SkillSubscription, error) {
	var subscription model.SkillSubscription
	err := r.db.Preload("Skill").First(&subscription, id).Error
	return &subscription, err
}

// Update updates a subscription
func (r *SkillSubscriptionRepository) Update(subscription *model.SkillSubscription) error {
	return r.db.Save(subscription).Error
}

// Delete deletes a subscription by ID
func (r *SkillSubscriptionRepository) Delete(id uint) error {
	return r.db.Delete(&model.SkillSubscription{}, id).Error
}

// GetByUserID returns all subscriptions for a user
func (r *SkillSubscriptionRepository) GetByUserID(userID uint) ([]model.SkillSubscription, error) {
	var subscriptions []model.SkillSubscription
	err := r.db.Preload("Skill").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&subscriptions).Error
	return subscriptions, err
}

// GetBySkillID returns all subscriptions for a skill
func (r *SkillSubscriptionRepository) GetBySkillID(skillID uint) ([]model.SkillSubscription, error) {
	var subscriptions []model.SkillSubscription
	err := r.db.Where("skill_id = ?", skillID).Find(&subscriptions).Error
	return subscriptions, err
}

// GetByUserAndSkill returns a specific subscription by user and skill
func (r *SkillSubscriptionRepository) GetByUserAndSkill(userID, skillID uint) (*model.SkillSubscription, error) {
	var subscription model.SkillSubscription
	err := r.db.Where("user_id = ? AND skill_id = ?", userID, skillID).First(&subscription).Error
	return &subscription, err
}

// GetActiveSubscriptions returns all active subscriptions
func (r *SkillSubscriptionRepository) GetActiveSubscriptions() ([]model.SkillSubscription, error) {
	var subscriptions []model.SkillSubscription
	err := r.db.Preload("Skill").
		Where("status = 'active'").
		Find(&subscriptions).Error
	return subscriptions, err
}

// GetSubscriptionsDueForExecution returns subscriptions that are due for execution
func (r *SkillSubscriptionRepository) GetSubscriptionsDueForExecution(limit int) ([]model.SkillSubscription, error) {
	var subscriptions []model.SkillSubscription
	now := time.Now()
	err := r.db.Preload("Skill").
		Where("status = 'active' AND next_run_at IS NOT NULL AND next_run_at <= ?", now).
		Order("next_run_at ASC").
		Limit(limit).
		Find(&subscriptions).Error
	return subscriptions, err
}

// UpdateNextRunAt updates the next run time for a subscription
func (r *SkillSubscriptionRepository) UpdateNextRunAt(subscriptionID uint, nextRunAt *time.Time) error {
	return r.db.Model(&model.SkillSubscription{}).
		Where("id = ?", subscriptionID).
		Update("next_run_at", nextRunAt).Error
}

// UpdateStatus updates the status of a subscription
func (r *SkillSubscriptionRepository) UpdateStatus(subscriptionID uint, status string) error {
	return r.db.Model(&model.SkillSubscription{}).
		Where("id = ?", subscriptionID).
		Update("status", status).Error
}

// UpdateConfig updates the configuration of a subscription
func (r *SkillSubscriptionRepository) UpdateConfig(subscriptionID uint, config model.SkillExecutionConfig) error {
	return r.db.Model(&model.SkillSubscription{}).
		Where("id = ?", subscriptionID).
		Update("config", config).Error
}

// GetSubscriptionsByBotID returns all subscriptions for a specific bot
func (r *SkillSubscriptionRepository) GetSubscriptionsByBotID(botID uint) ([]model.SkillSubscription, error) {
	var subscriptions []model.SkillSubscription
	err := r.db.Where("bot_id = ?", botID).Find(&subscriptions).Error
	return subscriptions, err
}

// GetSubscriptionsByChannelID returns all subscriptions for a specific channel
func (r *SkillSubscriptionRepository) GetSubscriptionsByChannelID(channelID uint) ([]model.SkillSubscription, error) {
	var subscriptions []model.SkillSubscription
	err := r.db.Where("channel_id = ?", channelID).Find(&subscriptions).Error
	return subscriptions, err
}

// CountBySkillID returns the number of active subscriptions for a skill
func (r *SkillSubscriptionRepository) CountBySkillID(skillID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.SkillSubscription{}).
		Where("skill_id = ? AND status = 'active'", skillID).
		Count(&count).Error
	return count, err
}
