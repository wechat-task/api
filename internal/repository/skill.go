package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

type SkillRepository struct {
	db *gorm.DB
}

func NewSkillRepository(db *gorm.DB) *SkillRepository {
	return &SkillRepository{db: db}
}

// Create creates a new skill
func (r *SkillRepository) Create(skill *model.Skill) error {
	return r.db.Create(skill).Error
}

// GetByID returns a skill by its ID
func (r *SkillRepository) GetByID(id uint) (*model.Skill, error) {
	var skill model.Skill
	err := r.db.First(&skill, id).Error
	return &skill, err
}

// GetByIDWithUser returns a skill by its ID with user association
func (r *SkillRepository) GetByIDWithUser(id uint) (*model.Skill, error) {
	var skill model.Skill
	err := r.db.Preload("User").First(&skill, id).Error
	return &skill, err
}

// Update updates a skill
func (r *SkillRepository) Update(skill *model.Skill) error {
	return r.db.Save(skill).Error
}

// Delete deletes a skill by ID
func (r *SkillRepository) Delete(id uint) error {
	return r.db.Delete(&model.Skill{}, id).Error
}

// GetByUserID returns all skills created by a user
func (r *SkillRepository) GetByUserID(userID uint) ([]model.Skill, error) {
	var skills []model.Skill
	err := r.db.Where("user_id = ?", userID).Find(&skills).Error
	return skills, err
}

// GetPublicSkills returns all public published skills
func (r *SkillRepository) GetPublicSkills() ([]model.Skill, error) {
	var skills []model.Skill
	err := r.db.Where("visibility = ? AND status = ?", model.SkillVisibilityPublic, model.SkillStatusPublished).
		Order("created_at DESC").
		Find(&skills).Error
	return skills, err
}

// GetByCategory returns public skills by category
func (r *SkillRepository) GetByCategory(category string) ([]model.Skill, error) {
	var skills []model.Skill
	err := r.db.Where("category = ? AND visibility = ? AND status = ?",
		category, model.SkillVisibilityPublic, model.SkillStatusPublished).
		Order("created_at DESC").
		Find(&skills).Error
	return skills, err
}

// SearchSkills searches skills by name or description
func (r *SkillRepository) SearchSkills(query string, limit int) ([]model.Skill, error) {
	var skills []model.Skill
	searchQuery := "%" + query + "%"
	err := r.db.Where("(name ILIKE ? OR description ILIKE ?) AND visibility = ? AND status = ?",
		searchQuery, searchQuery, model.SkillVisibilityPublic, model.SkillStatusPublished).
		Limit(limit).
		Order("created_at DESC").
		Find(&skills).Error
	return skills, err
}

// IncrementSubscriberCount increments the subscriber count for a skill
func (r *SkillRepository) IncrementSubscriberCount(skillID uint) error {
	return r.db.Model(&model.Skill{}).
		Where("id = ?", skillID).
		Update("subscriber_count", gorm.Expr("subscriber_count + 1")).Error
}

// DecrementSubscriberCount decrements the subscriber count for a skill
func (r *SkillRepository) DecrementSubscriberCount(skillID uint) error {
	return r.db.Model(&model.Skill{}).
		Where("id = ?", skillID).
		Update("subscriber_count", gorm.Expr("GREATEST(subscriber_count - 1, 0)")).Error
}

// IncrementExecutionCount increments the execution count for a skill
func (r *SkillRepository) IncrementExecutionCount(skillID uint) error {
	return r.db.Model(&model.Skill{}).
		Where("id = ?", skillID).
		Update("execution_count", gorm.Expr("execution_count + 1")).Error
}

// UpdateStatus updates the status of a skill
func (r *SkillRepository) UpdateStatus(skillID uint, status model.SkillStatus) error {
	return r.db.Model(&model.Skill{}).
		Where("id = ?", skillID).
		Update("status", status).Error
}

// GetSkillsByStatus returns skills by status
func (r *SkillRepository) GetSkillsByStatus(status model.SkillStatus) ([]model.Skill, error) {
	var skills []model.Skill
	err := r.db.Where("status = ?", status).Find(&skills).Error
	return skills, err
}

// SearchSkillsPaginated searches public published skills with pagination
func (r *SkillRepository) SearchSkillsPaginated(query string, offset, limit int) ([]model.Skill, error) {
	var skills []model.Skill
	searchQuery := "%" + query + "%"
	err := r.db.Where("(name ILIKE ? OR description ILIKE ?) AND visibility = ? AND status = ?",
		searchQuery, searchQuery, model.SkillVisibilityPublic, model.SkillStatusPublished).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&skills).Error
	return skills, err
}

// CountPublicSkills counts public published skills matching query
func (r *SkillRepository) CountPublicSkills(query string) (int64, error) {
	var count int64
	searchQuery := "%" + query + "%"
	err := r.db.Model(&model.Skill{}).
		Where("(name ILIKE ? OR description ILIKE ?) AND visibility = ? AND status = ?",
			searchQuery, searchQuery, model.SkillVisibilityPublic, model.SkillStatusPublished).
		Count(&count).Error
	return count, err
}
