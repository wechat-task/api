package service

import (
	"errors"
	"strings"
	"time"

	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
)

// ValidateSkillContent validates skill content follows Anthropic markdown format
// Phase 1 requirements:
// 1. Content must not be empty or whitespace only
// 2. Must contain at least one H1 title (line starting with "# ")
// 3. Future phases may add more validation rules (e.g. parameters section, examples section)
func ValidateSkillContent(content string) error {
	// Check empty content
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return errors.New("skill content cannot be empty")
	}

	// Check for at least one H1 title
	hasTitle := false
	lines := strings.Split(trimmed, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "# ") {
			hasTitle = true
			break
		}
	}

	if !hasTitle {
		return errors.New("skill content must have a title (line starting with '# ')")
	}

	return nil
}

// ValidateSubscriptionParameters validates subscription parameters match skill parameter definitions
func ValidateSubscriptionParameters(skill *model.Skill, config model.SkillExecutionConfig) error {
	// Check required parameters
	for paramName, paramDef := range skill.Parameters {
		if paramDef.Required {
			value, exists := config.Parameters[paramName]
			if !exists {
				return errors.New("missing required parameter: " + paramName)
			}
			// Check type
			if err := validateParameterType(paramName, value, paramDef); err != nil {
				return err
			}
		}
	}

	// Check for undefined parameters
	for paramName := range config.Parameters {
		if _, exists := skill.Parameters[paramName]; !exists {
			return errors.New("unknown parameter: " + paramName)
		}
	}

	return nil
}

// validateParameterType validates parameter type
func validateParameterType(paramName string, value any, paramDef model.SkillParameter) error {
	switch paramDef.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return errors.New("parameter " + paramName + " must be a string")
		}
	case "number":
		// JSON numbers are parsed as float64
		if _, ok := value.(float64); !ok {
			return errors.New("parameter " + paramName + " must be a number")
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return errors.New("parameter " + paramName + " must be a boolean")
		}
	case "enum":
		strVal, ok := value.(string)
		if !ok {
			return errors.New("parameter " + paramName + " must be a string for enum")
		}
		found := false
		for _, enumVal := range paramDef.EnumValues {
			if enumVal == strVal {
				found = true
				break
			}
		}
		if !found {
			return errors.New("parameter " + paramName + " must be one of: " + strings.Join(paramDef.EnumValues, ", "))
		}
	default:
		// Unknown type, skip validation
	}
	return nil
}

// CanViewSkill checks if a user can view the skill
func CanViewSkill(skill *model.Skill, userID uint) bool {
	switch skill.Visibility {
	case model.SkillVisibilityPublic:
		return true
	case model.SkillVisibilityUnlisted:
		return true // Unlisted skills are accessible via link
	case model.SkillVisibilityPrivate:
		return skill.UserID == userID
	default:
		return false
	}
}

// SkillService provides skill business logic
type SkillService struct {
	skillRepo         *repository.SkillRepository
	subscriptionRepo  *repository.SkillSubscriptionRepository
	executionLogRepo  *repository.SkillExecutionLogRepository
	userLLMConfigRepo *repository.UserLLMConfigRepository
}

// NewSkillService creates a new SkillService instance
func NewSkillService(
	skillRepo *repository.SkillRepository,
	subscriptionRepo *repository.SkillSubscriptionRepository,
	executionLogRepo *repository.SkillExecutionLogRepository,
	userLLMConfigRepo *repository.UserLLMConfigRepository,
) *SkillService {
	return &SkillService{
		skillRepo:         skillRepo,
		subscriptionRepo:  subscriptionRepo,
		executionLogRepo:  executionLogRepo,
		userLLMConfigRepo: userLLMConfigRepo,
	}
}

// CreateSkill creates a new skill
func (s *SkillService) CreateSkill(userID uint, skillData model.Skill) (*model.Skill, error) {
	// Validate skill content
	if err := ValidateSkillContent(skillData.Content); err != nil {
		return nil, err
	}

	// Set defaults
	now := time.Now()
	skill := &model.Skill{
		UserID:      userID,
		Name:        skillData.Name,
		Description: skillData.Description,
		Content:     skillData.Content,
		Visibility:  skillData.Visibility,
		Status:      skillData.Status,
		Version:     skillData.Version,
		Category:    skillData.Category,
		Tags:        skillData.Tags,
		Parameters:  skillData.Parameters,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Default status to draft if not set
	if skill.Status == "" {
		skill.Status = model.SkillStatusDraft
	}

	// Default version to 1.0.0 if not set
	if skill.Version == "" {
		skill.Version = "1.0.0"
	}

	// Default visibility to private if not set
	if skill.Visibility == "" {
		skill.Visibility = model.SkillVisibilityPrivate
	}

	// Save to database
	if err := s.skillRepo.Create(skill); err != nil {
		return nil, err
	}

	return skill, nil
}

// GetSkillByID returns a skill by ID without visibility check (for internal use)
func (s *SkillService) GetSkillByID(id uint) (*model.Skill, error) {
	return s.skillRepo.GetByID(id)
}

// GetSkillByIDWithVisibility returns a skill by ID with visibility check
func (s *SkillService) GetSkillByIDWithVisibility(userID uint, skillID uint) (*model.Skill, error) {
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	// Check visibility
	if !CanViewSkill(skill, userID) {
		return nil, errors.New("skill not found")
	}

	return skill, nil
}

// GetUserSkills returns all skills for a user
func (s *SkillService) GetUserSkills(userID uint) ([]model.Skill, error) {
	return s.skillRepo.GetByUserID(userID)
}

// UpdateSkill updates a draft skill
func (s *SkillService) UpdateSkill(userID uint, skillID uint, updates model.Skill) (*model.Skill, error) {
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	if skill.UserID != userID {
		return nil, errors.New("unauthorized: only skill creator can update")
	}

	if skill.Status != model.SkillStatusDraft {
		return nil, errors.New("only draft skills can be modified")
	}

	// Validate content if updating
	if updates.Content != "" && updates.Content != skill.Content {
		if err := ValidateSkillContent(updates.Content); err != nil {
			return nil, err
		}
		skill.Content = updates.Content
	}

	if updates.Name != "" {
		skill.Name = updates.Name
	}
	if updates.Description != "" {
		skill.Description = updates.Description
	}
	if updates.Visibility != "" {
		skill.Visibility = updates.Visibility
	}
	if updates.Version != "" {
		skill.Version = updates.Version
	}
	if updates.Category != "" {
		skill.Category = updates.Category
	}
	if updates.Tags != nil {
		skill.Tags = updates.Tags
	}
	if updates.Parameters != nil {
		skill.Parameters = updates.Parameters
	}

	skill.UpdatedAt = time.Now()

	if err := s.skillRepo.Update(skill); err != nil {
		return nil, err
	}

	return skill, nil
}

// PublishSkill publishes a draft skill
func (s *SkillService) PublishSkill(userID uint, skillID uint) (*model.Skill, error) {
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	if skill.UserID != userID {
		return nil, errors.New("unauthorized: only skill creator can publish")
	}

	if skill.Status == model.SkillStatusPublished {
		return nil, errors.New("skill is already published")
	}

	if skill.Status == model.SkillStatusArchived {
		return nil, errors.New("cannot publish archived skill")
	}

	// Only draft skills can be published
	if skill.Status != model.SkillStatusDraft {
		return nil, errors.New("only draft skills can be published")
	}

	skill.Status = model.SkillStatusPublished
	skill.UpdatedAt = time.Now()

	if err := s.skillRepo.Update(skill); err != nil {
		return nil, err
	}

	return skill, nil
}

// ArchiveSkill archives a published skill
func (s *SkillService) ArchiveSkill(userID uint, skillID uint) (*model.Skill, error) {
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	if skill.UserID != userID {
		return nil, errors.New("unauthorized: only skill creator can archive")
	}

	if skill.Status == model.SkillStatusArchived {
		return nil, errors.New("skill is already archived")
	}

	if skill.Status == model.SkillStatusDraft {
		return nil, errors.New("cannot archive draft skill, delete it instead")
	}

	if skill.Status != model.SkillStatusPublished {
		return nil, errors.New("only published skills can be archived")
	}

	skill.Status = model.SkillStatusArchived
	skill.UpdatedAt = time.Now()

	if err := s.skillRepo.Update(skill); err != nil {
		return nil, err
	}

	return skill, nil
}

// DeleteSkill deletes a skill
func (s *SkillService) DeleteSkill(userID uint, skillID uint) error {
	// Fetch existing skill
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return err
	}

	// Check permission (only creator can delete)
	if skill.UserID != userID {
		return errors.New("unauthorized: only skill creator can delete")
	}

	// Check for active subscribers
	// Note: simplified check, should verify via subscriptionRepo
	if skill.SubscriberCount > 0 {
		return errors.New("cannot delete skill with active subscribers")
	}

	// Published skills must be archived first
	if skill.Status == model.SkillStatusPublished {
		return errors.New("cannot delete published skill. Archive it first")
	}

	// Delete the skill
	return s.skillRepo.Delete(skillID)
}

// SearchPublicSkills searches public skills
func (s *SkillService) SearchPublicSkills(query string, page, pageSize int) ([]model.Skill, int64, error) {
	// Calculate offset
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Fetch paginated results
	skills, err := s.skillRepo.SearchSkillsPaginated(query, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// Fetch total count
	total, err := s.skillRepo.CountPublicSkills(query)
	if err != nil {
		return nil, 0, err
	}

	return skills, total, nil
}

// SubscribeToSkill subscribes a user to a skill
func (s *SkillService) SubscribeToSkill(userID, skillID uint, config model.SkillExecutionConfig) (*model.SkillSubscription, error) {
	// Fetch the skill
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	// Check if skill is published
	if skill.Status != model.SkillStatusPublished {
		return nil, errors.New("skill is not published")
	}

	// Check visibility (private skills only subscribable by creator)
	if skill.Visibility == model.SkillVisibilityPrivate && skill.UserID != userID {
		return nil, errors.New("private skill can only be subscribed by creator")
	}

	// Validate subscription parameters
	if err := ValidateSubscriptionParameters(skill, config); err != nil {
		return nil, err
	}

	// Check if already subscribed
	existing, err := s.subscriptionRepo.GetByUserAndSkill(userID, skillID)
	if err == nil && existing != nil {
		return nil, errors.New("already subscribed to this skill")
	}

	// LLM config: use user-provided config if available, otherwise use system LLM
	// Billing not implemented yet; all skills are treated as free using system LLM

	// Create subscription
	now := time.Now()
	subscription := &model.SkillSubscription{
		UserID:    userID,
		SkillID:   skillID,
		Config:    config,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set default timezone
	if subscription.TimeZone == "" {
		subscription.TimeZone = "UTC"
	}

	// Save subscription
	if err := s.subscriptionRepo.Create(subscription); err != nil {
		return nil, err
	}

	// Increment subscriber count
	if err := s.skillRepo.IncrementSubscriberCount(skillID); err != nil {
		// If count increment fails, subscription creation fails
		return nil, err
	}

	return subscription, nil
}

// UnsubscribeFromSkill unsubscribes from a skill by subscription ID
func (s *SkillService) UnsubscribeFromSkill(userID, subscriptionID uint) error {
	// Fetch subscription
	subscription, err := s.subscriptionRepo.GetByID(subscriptionID)
	if err != nil {
		return err
	}

	// Check permission
	if subscription.UserID != userID {
		return errors.New("unauthorized: only subscription owner can unsubscribe")
	}

	// Delete subscription
	if err := s.subscriptionRepo.Delete(subscriptionID); err != nil {
		return err
	}

	// Decrement subscriber count
	if err := s.skillRepo.DecrementSubscriberCount(subscription.SkillID); err != nil {
		// Log error but don't return it (subscription already deleted)
	}

	return nil
}

// UnsubscribeFromSkillBySkillID unsubscribes from a skill by skill ID
func (s *SkillService) UnsubscribeFromSkillBySkillID(userID, skillID uint) error {
	// Fetch subscription by user and skill
	subscription, err := s.subscriptionRepo.GetByUserAndSkill(userID, skillID)
	if err != nil {
		return errors.New("subscription not found")
	}

	// Delete subscription
	if err := s.subscriptionRepo.Delete(subscription.ID); err != nil {
		return err
	}

	// Decrement subscriber count
	if err := s.skillRepo.DecrementSubscriberCount(skillID); err != nil {
		// Log error but don't return it (subscription already deleted)
	}

	return nil
}

// GetUserSubscriptions returns all subscriptions for a user
func (s *SkillService) GetUserSubscriptions(userID uint) ([]model.SkillSubscription, error) {
	return s.subscriptionRepo.GetByUserID(userID)
}

// GetSubscriptionByID returns a subscription by ID
func (s *SkillService) GetSubscriptionByID(userID, subscriptionID uint) (*model.SkillSubscription, error) {
	subscription, err := s.subscriptionRepo.GetByID(subscriptionID)
	if err != nil {
		return nil, err
	}

	// Check permission (only subscriber can view)
	if subscription.UserID != userID {
		return nil, errors.New("unauthorized: only subscription owner can view")
	}

	return subscription, nil
}

// UpdateSubscription updates a subscription
func (s *SkillService) UpdateSubscription(userID, subscriptionID uint, updates model.SkillSubscription) (*model.SkillSubscription, error) {
	// Fetch existing subscription
	subscription, err := s.subscriptionRepo.GetByID(subscriptionID)
	if err != nil {
		return nil, err
	}

	// Check permission
	if subscription.UserID != userID {
		return nil, errors.New("unauthorized: only subscription owner can update")
	}

	// Fetch associated skill
	skill, err := s.skillRepo.GetByID(subscription.SkillID)
	if err != nil {
		return nil, err
	}

	// Validate parameters if config updated
	if updates.Config.Parameters != nil {
		// Merge with existing config
		newConfig := subscription.Config
		newConfig.Parameters = updates.Config.Parameters
		if updates.Config.LLMConfig != nil {
			newConfig.LLMConfig = updates.Config.LLMConfig
		}
		if err := ValidateSubscriptionParameters(skill, newConfig); err != nil {
			return nil, err
		}
		subscription.Config = newConfig
	}

	// Update other fields
	if updates.Status != "" {
		subscription.Status = updates.Status
	}
	if updates.ScheduleCron != nil {
		subscription.ScheduleCron = updates.ScheduleCron
	}
	if updates.TimeZone != "" {
		subscription.TimeZone = updates.TimeZone
	}
	if updates.BotID != nil {
		subscription.BotID = updates.BotID
	}
	if updates.ChannelID != nil {
		subscription.ChannelID = updates.ChannelID
	}
	if updates.NextRunAt != nil {
		subscription.NextRunAt = updates.NextRunAt
	}

	subscription.UpdatedAt = time.Now()

	// Save updates
	if err := s.subscriptionRepo.Update(subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}
