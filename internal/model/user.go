package model

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"time"
)

type User struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	authnID     []byte       `json:"-" gorm:"column:web_authn_id;uniqueIndex;not null"`
	Username    *string      `json:"username" gorm:"uniqueIndex"`
	Icon        string       `json:"icon"`
	Credentials []Credential `json:"-" gorm:"foreignKey:UserID"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (u *User) WebAuthnID() []byte {
	return u.authnID
}

func (u *User) SetWebAuthnID(id []byte) {
	u.authnID = id
}

func (u *User) WebAuthnName() string {
	if u.Username != nil {
		return *u.Username
	}
	return "User"
}

func (u *User) WebAuthnDisplayName() string {
	if u.Username != nil {
		return *u.Username
	}
	return "User"
}

func (u *User) WebAuthnIcon() string {
	return u.Icon
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))
	for i, cred := range u.Credentials {
		creds[i] = cred.ToWebAuthnCredential()
	}
	return creds
}
