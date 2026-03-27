package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Credentials").First(&user, id).Error
	return &user, err
}

func (r *UserRepository) GetByWebAuthnID(webAuthnID []byte) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Credentials").Where("web_authn_id = ?", webAuthnID).First(&user).Error
	return &user, err
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Credentials").Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) SetUsername(userID uint, username string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("username", &username).Error
}
