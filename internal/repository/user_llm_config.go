package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

type UserLLMConfigRepository struct {
	db *gorm.DB
}

func NewUserLLMConfigRepository(db *gorm.DB) *UserLLMConfigRepository {
	return &UserLLMConfigRepository{db: db}
}

// Create creates a new user LLM configuration
func (r *UserLLMConfigRepository) Create(config *model.UserLLMConfig) error {
	return r.db.Create(config).Error
}

// GetByID returns a configuration by its ID
func (r *UserLLMConfigRepository) GetByID(id uint) (*model.UserLLMConfig, error) {
	var config model.UserLLMConfig
	err := r.db.First(&config, id).Error
	return &config, err
}

// GetByIDWithUser returns a configuration by its ID with user association
func (r *UserLLMConfigRepository) GetByIDWithUser(id uint) (*model.UserLLMConfig, error) {
	var config model.UserLLMConfig
	err := r.db.Preload("User").First(&config, id).Error
	return &config, err
}

// Update updates a configuration
func (r *UserLLMConfigRepository) Update(config *model.UserLLMConfig) error {
	return r.db.Save(config).Error
}

// Delete deletes a configuration by ID
func (r *UserLLMConfigRepository) Delete(id uint) error {
	return r.db.Delete(&model.UserLLMConfig{}, id).Error
}

// GetByUserID returns all LLM configurations for a user
func (r *UserLLMConfigRepository) GetByUserID(userID uint) ([]model.UserLLMConfig, error) {
	var configs []model.UserLLMConfig
	err := r.db.Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&configs).Error
	return configs, err
}

// GetDefaultByUserID returns the default LLM configuration for a user
func (r *UserLLMConfigRepository) GetDefaultByUserID(userID uint) (*model.UserLLMConfig, error) {
	var config model.UserLLMConfig
	err := r.db.Where("user_id = ? AND is_default = true", userID).
		First(&config).Error
	return &config, err
}

// SetAsDefault sets a configuration as default for the user
func (r *UserLLMConfigRepository) SetAsDefault(configID uint) error {
	// First get the config to know the user ID
	var config model.UserLLMConfig
	if err := r.db.First(&config, configID).Error; err != nil {
		return err
	}

	// Start a transaction to update all configs for this user
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Clear default flag from all configs for this user
		if err := tx.Model(&model.UserLLMConfig{}).
			Where("user_id = ?", config.UserID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// Set the specified config as default
		return tx.Model(&model.UserLLMConfig{}).
			Where("id = ?", configID).
			Update("is_default", true).Error
	})
}

// GetByProvider returns configurations by provider for a user
func (r *UserLLMConfigRepository) GetByProvider(userID uint, provider string) ([]model.UserLLMConfig, error) {
	var configs []model.UserLLMConfig
	err := r.db.Where("user_id = ? AND provider = ?", userID, provider).
		Order("is_default DESC, created_at DESC").
		Find(&configs).Error
	return configs, err
}

// GetByProviderAndModel returns configurations by provider and model for a user
func (r *UserLLMConfigRepository) GetByProviderAndModel(userID uint, provider, modelName string) (*model.UserLLMConfig, error) {
	var config model.UserLLMConfig
	err := r.db.Where("user_id = ? AND provider = ? AND model = ?", userID, provider, modelName).
		First(&config).Error
	return &config, err
}

// CountByUserID returns the number of LLM configurations for a user
func (r *UserLLMConfigRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserLLMConfig{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// UpdateAPIKey updates the API key for a configuration
func (r *UserLLMConfigRepository) UpdateAPIKey(configID uint, apiKey string) error {
	return r.db.Model(&model.UserLLMConfig{}).
		Where("id = ?", configID).
		Update("api_key", apiKey).Error
}

// UpdateBaseURL updates the base URL for a configuration
func (r *UserLLMConfigRepository) UpdateBaseURL(configID uint, baseURL *string) error {
	return r.db.Model(&model.UserLLMConfig{}).
		Where("id = ?", configID).
		Update("base_url", baseURL).Error
}

// UpdateModel updates the model for a configuration
func (r *UserLLMConfigRepository) UpdateModel(configID uint, modelName string) error {
	return r.db.Model(&model.UserLLMConfig{}).
		Where("id = ?", configID).
		Update("model", modelName).Error
}
