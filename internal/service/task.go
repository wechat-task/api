package service

import (
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
	"time"
)

type TaskService struct {
	repo *repository.TaskRepository
}

func NewTaskService(repo *repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(title, description, priority string, dueDate *time.Time) (*model.Task, error) {
	task := &model.Task{
		Title:       title,
		Description: description,
		Status:      "pending",
		Priority:    priority,
		DueDate:     dueDate,
	}
	if err := s.repo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) GetTask(id uint) (*model.Task, error) {
	return s.repo.GetByID(id)
}

func (s *TaskService) ListTasks() ([]model.Task, error) {
	return s.repo.List()
}

func (s *TaskService) UpdateTask(id uint, title, description, status, priority string, dueDate *time.Time) (*model.Task, error) {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	task.Title = title
	task.Description = description
	task.Status = status
	task.Priority = priority
	task.DueDate = dueDate
	if err := s.repo.Update(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) DeleteTask(id uint) error {
	return s.repo.Delete(id)
}

func (s *TaskService) CompleteTask(id uint) (*model.Task, error) {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	task.Status = "completed"
	if err := s.repo.Update(task); err != nil {
		return nil, err
	}
	return task, nil
}
