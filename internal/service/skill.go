package service

import (
	"errors"
	"strings"
	"time"

	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
)

// ValidateSkillContent 验证技能内容是否符合Anthropic markdown格式
// 第一阶段要求：
// 1. 内容不能为空或仅包含空白字符
// 2. 必须包含至少一个一级标题（以 "# " 开头的行）
// 3. 后续阶段可能会添加更多验证规则（如参数部分、示例部分等）
func ValidateSkillContent(content string) error {
	// 检查空内容
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return errors.New("skill content cannot be empty")
	}

	// 检查是否有标题（至少一个#开头的行）
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

// ValidateSubscriptionParameters 验证订阅参数是否符合技能参数定义
func ValidateSubscriptionParameters(skill *model.Skill, config model.SkillExecutionConfig) error {
	// 检查必需参数
	for paramName, paramDef := range skill.Parameters {
		if paramDef.Required {
			value, exists := config.Parameters[paramName]
			if !exists {
				return errors.New("missing required parameter: " + paramName)
			}
			// 检查类型
			if err := validateParameterType(paramName, value, paramDef); err != nil {
				return err
			}
		}
	}

	// 检查提供了未定义的参数
	for paramName := range config.Parameters {
		if _, exists := skill.Parameters[paramName]; !exists {
			return errors.New("unknown parameter: " + paramName)
		}
	}

	return nil
}

// validateParameterType 验证参数类型
func validateParameterType(paramName string, value any, paramDef model.SkillParameter) error {
	switch paramDef.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return errors.New("parameter " + paramName + " must be a string")
		}
	case "number":
		// JSON 数字可能是 float64
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
		// 未知类型，跳过验证
	}
	return nil
}

// CanViewSkill 检查用户是否可以查看技能
func CanViewSkill(skill *model.Skill, userID uint) bool {
	switch skill.Visibility {
	case model.SkillVisibilityPublic:
		return true
	case model.SkillVisibilityUnlisted:
		return true // 未列出的技能可以通过链接访问
	case model.SkillVisibilityPrivate:
		return skill.UserID == userID
	default:
		return false
	}
}

// SkillService 提供技能相关的业务逻辑
type SkillService struct {
	skillRepo         *repository.SkillRepository
	subscriptionRepo  *repository.SkillSubscriptionRepository
	executionLogRepo  *repository.SkillExecutionLogRepository
	userLLMConfigRepo *repository.UserLLMConfigRepository
}

// NewSkillService 创建新的SkillService实例
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

// CreateSkill 创建新技能
func (s *SkillService) CreateSkill(userID uint, skillData model.Skill) (*model.Skill, error) {
	// 验证技能内容
	if err := ValidateSkillContent(skillData.Content); err != nil {
		return nil, err
	}

	// 设置默认值
	now := time.Now()
	skill := &model.Skill{
		UserID:      userID,
		Name:        skillData.Name,
		Description: skillData.Description,
		Content:     skillData.Content,
		Visibility:  skillData.Visibility,
		Status:      skillData.Status,
		Category:    skillData.Category,
		Tags:        skillData.Tags,
		Parameters:  skillData.Parameters,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 如果未设置状态，默认为draft
	if skill.Status == "" {
		skill.Status = model.SkillStatusDraft
	}

	// 如果未设置可见性，默认为private
	if skill.Visibility == "" {
		skill.Visibility = model.SkillVisibilityPrivate
	}

	// 保存到数据库
	if err := s.skillRepo.Create(skill); err != nil {
		return nil, err
	}

	return skill, nil
}

// GetSkillByID 根据ID获取技能（不检查可见性，用于内部操作）
func (s *SkillService) GetSkillByID(id uint) (*model.Skill, error) {
	return s.skillRepo.GetByID(id)
}

// GetSkillByIDWithVisibility 根据ID获取技能，检查可见性
func (s *SkillService) GetSkillByIDWithVisibility(userID uint, skillID uint) (*model.Skill, error) {
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	// 检查可见性
	if !CanViewSkill(skill, userID) {
		return nil, errors.New("skill not found")
	}

	return skill, nil
}

// GetUserSkills 获取用户的所有技能
func (s *SkillService) GetUserSkills(userID uint) ([]model.Skill, error) {
	return s.skillRepo.GetByUserID(userID)
}

// UpdateSkill 更新技能
func (s *SkillService) UpdateSkill(userID uint, skillID uint, updates model.Skill) (*model.Skill, error) {
	// 首先获取现有技能
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	// 检查权限（只有创建者可以更新）
	if skill.UserID != userID {
		return nil, errors.New("unauthorized: only skill creator can update")
	}

	// 如果技能已发布，限制可以修改的字段
	if skill.Status == model.SkillStatusPublished {
		// 已发布的技能不能修改内容、名称、描述等关键字段
		// 只能修改可见性、标签、分类等非关键字段
		if updates.Content != "" && updates.Content != skill.Content {
			return nil, errors.New("cannot modify content of published skill")
		}
		if updates.Name != "" && updates.Name != skill.Name {
			return nil, errors.New("cannot modify name of published skill")
		}
		// 可以允许修改其他字段
	}

	// 如果更新了内容，需要验证
	if updates.Content != "" && updates.Content != skill.Content {
		if err := ValidateSkillContent(updates.Content); err != nil {
			return nil, err
		}
		skill.Content = updates.Content
	}

	// 更新其他字段
	if updates.Name != "" {
		skill.Name = updates.Name
	}
	if updates.Description != "" {
		skill.Description = updates.Description
	}
	if updates.Visibility != "" {
		skill.Visibility = updates.Visibility
	}
	if updates.Status != "" {
		skill.Status = updates.Status
	}
	if updates.Category != "" {
		skill.Category = updates.Category
	}
	if updates.Tags != nil {
		skill.Tags = updates.Tags
	}
	// 更新其他字段...
	skill.UpdatedAt = time.Now()

	// 保存更新
	if err := s.skillRepo.Update(skill); err != nil {
		return nil, err
	}

	return skill, nil
}

// DeleteSkill 删除技能
func (s *SkillService) DeleteSkill(userID uint, skillID uint) error {
	// 首先获取现有技能
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return err
	}

	// 检查权限（只有创建者可以删除）
	if skill.UserID != userID {
		return errors.New("unauthorized: only skill creator can delete")
	}

	// 检查是否有活跃订阅
	// 注意：这里简化实现，实际应该检查subscriptionRepo
	if skill.SubscriberCount > 0 {
		return errors.New("cannot delete skill with active subscribers")
	}

	// 如果技能已发布，需要先归档或下架
	if skill.Status == model.SkillStatusPublished {
		return errors.New("cannot delete published skill. Archive it first")
	}

	// 删除技能
	return s.skillRepo.Delete(skillID)
}

// SearchPublicSkills 搜索公开技能
func (s *SkillService) SearchPublicSkills(query string, page, pageSize int) ([]model.Skill, int64, error) {
	// 计算偏移量
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 获取分页结果
	skills, err := s.skillRepo.SearchSkillsPaginated(query, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数
	total, err := s.skillRepo.CountPublicSkills(query)
	if err != nil {
		return nil, 0, err
	}

	return skills, total, nil
}

// SubscribeToSkill 订阅技能
func (s *SkillService) SubscribeToSkill(userID, skillID uint, config model.SkillExecutionConfig) (*model.SkillSubscription, error) {
	// 获取技能
	skill, err := s.skillRepo.GetByID(skillID)
	if err != nil {
		return nil, err
	}

	// 检查技能是否已发布
	if skill.Status != model.SkillStatusPublished {
		return nil, errors.New("skill is not published")
	}

	// 检查可见性（私有技能只能由创建者订阅）
	if skill.Visibility == model.SkillVisibilityPrivate && skill.UserID != userID {
		return nil, errors.New("private skill can only be subscribed by creator")
	}

	// 验证订阅参数
	if err := ValidateSubscriptionParameters(skill, config); err != nil {
		return nil, err
	}

	// 检查是否已订阅
	existing, err := s.subscriptionRepo.GetByUserAndSkill(userID, skillID)
	if err == nil && existing != nil {
		return nil, errors.New("already subscribed to this skill")
	}

	// LLM配置：如果用户提供了自定义LLM配置则使用，否则使用系统LLM
	// 收费规则暂时不实现，所有技能都视为免费并使用系统LLM

	// 创建订阅
	now := time.Now()
	subscription := &model.SkillSubscription{
		UserID:    userID,
		SkillID:   skillID,
		Config:    config,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 设置默认时区
	if subscription.TimeZone == "" {
		subscription.TimeZone = "UTC"
	}

	// 保存订阅
	if err := s.subscriptionRepo.Create(subscription); err != nil {
		return nil, err
	}

	// 增加订阅者计数
	if err := s.skillRepo.IncrementSubscriberCount(skillID); err != nil {
		// 如果增加计数失败，尝试删除订阅（可选）
		return nil, err
	}

	return subscription, nil
}

// UnsubscribeFromSkill 取消订阅技能
func (s *SkillService) UnsubscribeFromSkill(userID, subscriptionID uint) error {
	// 获取订阅
	subscription, err := s.subscriptionRepo.GetByID(subscriptionID)
	if err != nil {
		return err
	}

	// 检查权限
	if subscription.UserID != userID {
		return errors.New("unauthorized: only subscription owner can unsubscribe")
	}

	// 删除订阅
	if err := s.subscriptionRepo.Delete(subscriptionID); err != nil {
		return err
	}

	// 减少订阅者计数
	if err := s.skillRepo.DecrementSubscriberCount(subscription.SkillID); err != nil {
		// 日志记录错误，但不返回错误（订阅已删除）
	}

	return nil
}

// GetUserSubscriptions 获取用户的订阅列表
func (s *SkillService) GetUserSubscriptions(userID uint) ([]model.SkillSubscription, error) {
	return s.subscriptionRepo.GetByUserID(userID)
}

// GetSubscriptionByID 根据ID获取订阅详情
func (s *SkillService) GetSubscriptionByID(userID, subscriptionID uint) (*model.SkillSubscription, error) {
	subscription, err := s.subscriptionRepo.GetByID(subscriptionID)
	if err != nil {
		return nil, err
	}

	// 检查权限（只有订阅者可以查看）
	if subscription.UserID != userID {
		return nil, errors.New("unauthorized: only subscription owner can view")
	}

	return subscription, nil
}

// UpdateSubscription 更新订阅配置
func (s *SkillService) UpdateSubscription(userID, subscriptionID uint, updates model.SkillSubscription) (*model.SkillSubscription, error) {
	// 获取现有订阅
	subscription, err := s.subscriptionRepo.GetByID(subscriptionID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if subscription.UserID != userID {
		return nil, errors.New("unauthorized: only subscription owner can update")
	}

	// 获取关联技能
	skill, err := s.skillRepo.GetByID(subscription.SkillID)
	if err != nil {
		return nil, err
	}

	// 如果更新了配置，验证参数
	if updates.Config.Parameters != nil {
		// 合并现有配置
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

	// 更新其他字段
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

	// 保存更新
	if err := s.subscriptionRepo.Update(subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}
