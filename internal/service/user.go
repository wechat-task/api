package service

import (
	"github.com/google/uuid"
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(webAuthnID []byte) (*model.User, error) {
	user := &model.User{}
	user.SetWebAuthnID(webAuthnID)

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) SetUsername(userID uint, username string) error {
	return s.repo.SetUsername(userID, username)
}

func (s *UserService) GetUserByWebAuthnID(webAuthnID []byte) (*model.User, error) {
	return s.repo.GetByWebAuthnID(webAuthnID)
}

func GenerateWebAuthnID() []byte {
	return []byte(uuid.New().String())
}
