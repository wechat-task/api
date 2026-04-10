package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

// ChannelRepository handles database operations for channels.
type ChannelRepository struct {
	db *gorm.DB
}

// NewChannelRepository creates a new ChannelRepository.
func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

// Create inserts a new channel record.
func (r *ChannelRepository) Create(channel *model.Channel) error {
	return r.db.Create(channel).Error
}

// GetByID returns a channel by its ID.
func (r *ChannelRepository) GetByID(id uint) (*model.Channel, error) {
	var channel model.Channel
	err := r.db.First(&channel, id).Error
	return &channel, err
}

// GetByBotID returns all channels for a bot.
func (r *ChannelRepository) GetByBotID(botID uint) ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.Where("bot_id = ?", botID).Find(&channels).Error
	return channels, err
}

// Update saves channel changes.
func (r *ChannelRepository) Update(channel *model.Channel) error {
	return r.db.Save(channel).Error
}

// Delete removes a channel by ID.
func (r *ChannelRepository) Delete(id uint) error {
	return r.db.Delete(&model.Channel{}, id).Error
}

// GetByStatus returns all channels with the given status.
func (r *ChannelRepository) GetByStatus(status string) ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.Where("status = ?", status).Find(&channels).Error
	return channels, err
}

// GetByType returns all channels of a given type and status.
func (r *ChannelRepository) GetByType(channelType model.ChannelType, status string) ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.Where("type = ? AND status = ?", channelType, status).Find(&channels).Error
	return channels, err
}
