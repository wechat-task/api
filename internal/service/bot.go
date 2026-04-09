package service

import (
	"errors"
	"fmt"

	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
)

// BotService handles bot management business logic.
type BotService struct {
	repo *repository.BotRepository
}

// NewBotService creates a new BotService.
func NewBotService(repo *repository.BotRepository) *BotService {
	return &BotService{repo: repo}
}

// CreateBotRequest holds fields for creating a bot.
type CreateBotRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

// CreateBot creates a new bot.
func (s *BotService) CreateBot(userID uint, req *CreateBotRequest) (*model.Bot, error) {
	bot := &model.Bot{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Status:      "pending",
	}

	if err := s.repo.Create(bot); err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	return bot, nil
}

// GetBot returns a bot by ID, verifying ownership.
func (s *BotService) GetBot(userID, botID uint) (*model.Bot, error) {
	bot, err := s.repo.GetByID(botID)
	if err != nil {
		return nil, err
	}
	if bot.UserID != userID {
		return nil, errors.New("bot not found")
	}
	return bot, nil
}

// ListBots returns all bots for a user.
func (s *BotService) ListBots(userID uint) ([]model.Bot, error) {
	return s.repo.GetByUserID(userID)
}

// UpdateBotRequest holds optional fields for updating a bot.
type UpdateBotRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// UpdateBot updates a bot's name and description.
func (s *BotService) UpdateBot(userID, botID uint, req *UpdateBotRequest) (*model.Bot, error) {
	bot, err := s.GetBot(userID, botID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		bot.Name = *req.Name
	}
	if req.Description != nil {
		bot.Description = req.Description
	}

	if err := s.repo.Update(bot); err != nil {
		return nil, fmt.Errorf("update bot: %w", err)
	}

	return bot, nil
}

// DeleteBot removes a bot, verifying ownership.
func (s *BotService) DeleteBot(userID, botID uint) error {
	bot, err := s.GetBot(userID, botID)
	if err != nil {
		return err
	}
	return s.repo.Delete(bot.ID)
}
