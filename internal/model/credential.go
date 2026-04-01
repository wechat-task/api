package model

import (
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type Credential struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          uint      `json:"user_id" gorm:"not null;index"`
	CredentialID    []byte    `json:"-" gorm:"not null;uniqueIndex"`
	PublicKey       []byte    `json:"-" gorm:"not null"`
	AttestationType string    `json:"attestation_type"`
	Transport       string    `json:"transport"`
	Flags           []byte    `json:"flags"`
	SignCount       uint32    `json:"sign_count"`
	CloneWarning    bool      `json:"clone_warning"`
	AAGUID          []byte    `json:"aaguid" gorm:"column:aaguid"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (c *Credential) ToWebAuthnCredential() webauthn.Credential {
	cred := webauthn.Credential{
		ID:              c.CredentialID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       parseTransports(c.Transport),
		Authenticator: webauthn.Authenticator{
			SignCount:    c.SignCount,
			CloneWarning: c.CloneWarning,
		},
	}
	if len(c.Flags) > 0 {
		var flags webauthn.CredentialFlags
		if err := json.Unmarshal(c.Flags, &flags); err == nil {
			cred.Flags = flags
		}
	}
	return cred
}

func (c *Credential) FromWebAuthnCredential(cred webauthn.Credential, userID uint) {
	c.CredentialID = cred.ID
	c.PublicKey = cred.PublicKey
	c.AttestationType = cred.AttestationType
	c.Transport = serializeTransports(cred.Transport)
	c.SignCount = cred.Authenticator.SignCount
	c.CloneWarning = cred.Authenticator.CloneWarning
	c.AAGUID = cred.Authenticator.AAGUID
	c.UserID = userID
	if flagsBytes, err := json.Marshal(cred.Flags); err == nil {
		c.Flags = flagsBytes
	}
}

func parseTransports(transport string) []protocol.AuthenticatorTransport {
	if transport == "" {
		return []protocol.AuthenticatorTransport{}
	}
	return []protocol.AuthenticatorTransport{protocol.AuthenticatorTransport(transport)}
}

func serializeTransports(transports []protocol.AuthenticatorTransport) string {
	if len(transports) == 0 {
		return ""
	}
	return string(transports[0])
}
