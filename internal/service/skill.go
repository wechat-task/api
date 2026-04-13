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
		UserID:        userID,
		Name:          skillData.Name,
		Description:   skillData.Description,
		Content:       skillData.Content,
		Visibility:    skillData.Visibility,
		Status:        skillData.Status,
		Category:      skillData.Category,
		Tags:          skillData.Tags,
		IsFree:        skillData.IsFree,
		UsesSystemLLM: skillData.UsesSystemLLM,
		MaxTokens:     skillData.MaxTokens,
		Parameters:    skillData.Parameters,
		ScheduleCron:  skillData.ScheduleCron,
		CreatedAt:     now,
		UpdatedAt:     now,
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

// GetSkillByID 根据ID获取技能
func (s *SkillService) GetSkillByID(id uint) (*model.Skill, error) {
	return s.skillRepo.GetByID(id)
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
	// repository中没有SearchPublicSkills和CountPublicSkills方法
	// 使用现有的SearchSkills方法
	limit := pageSize
	skills, err := s.skillRepo.SearchSkills(query, limit)
	if err != nil {
		return nil, 0, err
	}

	// 计算总数（简化实现）
	total := int64(len(skills))
	// 实际实现中应该有一个Count方法，但这里先简化

	return skills, total, nil
}
