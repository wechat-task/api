package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelContextRepository handles database operations for channel contexts.
type ChannelContextRepository struct {
	db *gorm.DB
}

// NewChannelContextRepository creates a new ChannelContextRepository.
func NewChannelContextRepository(db *gorm.DB) *ChannelContextRepository {
	return &ChannelContextRepository{db: db}
}

// Upsert inserts or updates a channel context (upsert by channel_id + user_id).
func (r *ChannelContextRepository) Upsert(ctx *model.ChannelContext) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "channel_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"context_token", "last_message", "updated_at",
		}),
	}).Create(ctx).Error
}

// GetByChannelAndUser returns the context for a specific channel and user.
func (r *ChannelContextRepository) GetByChannelAndUser(channelID uint, userID string) (*model.ChannelContext, error) {
	var ctx model.ChannelContext
	err := r.db.Where("channel_id = ? AND user_id = ?", channelID, userID).First(&ctx).Error
	return &ctx, err
}

// GetByChannel returns all contexts for a channel.
func (r *ChannelContextRepository) GetByChannel(channelID uint) ([]model.ChannelContext, error) {
	var contexts []model.ChannelContext
	err := r.db.Where("channel_id = ?", channelID).Find(&contexts).Error
	return contexts, err
}
