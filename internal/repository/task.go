package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) GetByID(id uint) (*model.Task, error) {
	var task model.Task
	err := r.db.First(&task, id).Error
	return &task, err
}

func (r *TaskRepository) List() ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

func (r *TaskRepository) Delete(id uint) error {
	return r.db.Delete(&model.Task{}, id).Error
}
