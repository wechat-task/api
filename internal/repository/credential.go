package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

type CredentialRepository struct {
	db *gorm.DB
}

func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{db: db}
}

func (r *CredentialRepository) Create(credential *model.Credential) error {
	return r.db.Create(credential).Error
}

func (r *CredentialRepository) GetByCredentialID(credentialID []byte) (*model.Credential, error) {
	var cred model.Credential
	err := r.db.Where("credential_id = ?", credentialID).First(&cred).Error
	return &cred, err
}

func (r *CredentialRepository) GetByUserID(userID uint) ([]model.Credential, error) {
	var creds []model.Credential
	err := r.db.Where("user_id = ?", userID).Find(&creds).Error
	return creds, err
}

func (r *CredentialRepository) Update(credential *model.Credential) error {
	return r.db.Save(credential).Error
}

func (r *CredentialRepository) UpdateSignCount(credentialID []byte, signCount uint32) error {
	return r.db.Model(&model.Credential{}).Where("credential_id = ?", credentialID).Update("sign_count", signCount).Error
}

func (r *CredentialRepository) Delete(id uint) error {
	return r.db.Delete(&model.Credential{}, id).Error
}
