package models

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// Passkey represents a WebAuthn credential for a user
type Passkey struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	CredentialID   []byte    `json:"credential_id"`
	PublicKey      []byte    `json:"public_key"`
	AAGUID         []byte    `json:"aaguid"`
	SignCount      uint32    `json:"sign_count"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
	LastUsedAt     time.Time `json:"last_used_at"`
	Transports     []string  `json:"transports,omitempty"`
	AttestationType string    `json:"attestation_type"`
}

// WebAuthnID implements webauthn.User interface
func (u *User) WebAuthnID() []byte {
	return []byte(u.ID)
}

// WebAuthnName implements webauthn.User interface
func (u *User) WebAuthnName() string {
	return u.Email
}

// WebAuthnDisplayName implements webauthn.User interface
func (u *User) WebAuthnDisplayName() string {
	return u.Email
}

// WebAuthnIcon implements webauthn.User interface
func (u *User) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials implements webauthn.User interface
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return []webauthn.Credential{} // This will be populated from the database
}
