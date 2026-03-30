package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

// BotRepository handles database operations for bots.
type BotRepository struct {
	db *gorm.DB
}

// NewBotRepository creates a new BotRepository.
func NewBotRepository(db *gorm.DB) *BotRepository {
	return &BotRepository{db: db}
}

// Create inserts a new bot record.
func (r *BotRepository) Create(bot *model.Bot) error {
	return r.db.Create(bot).Error
}

// GetByID returns a bot by its ID.
func (r *BotRepository) GetByID(id uint) (*model.Bot, error) {
	var bot model.Bot
	err := r.db.First(&bot, id).Error
	return &bot, err
}

// GetByUserID returns all bots belonging to a user.
func (r *BotRepository) GetByUserID(userID uint) ([]model.Bot, error) {
	var bots []model.Bot
	err := r.db.Where("user_id = ?", userID).Find(&bots).Error
	return bots, err
}

// Update saves bot changes.
func (r *BotRepository) Update(bot *model.Bot) error {
	return r.db.Save(bot).Error
}

// Delete removes a bot by ID.
func (r *BotRepository) Delete(id uint) error {
	return r.db.Delete(&model.Bot{}, id).Error
}

// GetByQRCodeID returns a pending bot by its QR code ID.
func (r *BotRepository) GetByQRCodeID(qrcodeID string) (*model.Bot, error) {
	var bot model.Bot
	err := r.db.Where("qrcode_id = ? AND status = ?", qrcodeID, "pending").First(&bot).Error
	return &bot, err
}

// GetByStatus returns all bots with the given status.
func (r *BotRepository) GetByStatus(status string) ([]model.Bot, error) {
	var bots []model.Bot
	err := r.db.Where("status = ?", status).Find(&bots).Error
	return bots, err
}
