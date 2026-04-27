package models

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// Passkey represents a WebAuthn credential for a user
//
// BackupEligible/BackupState come from the authenticator's flags at registration
// time (BE/BS bits, see WebAuthn §6.1.4). go-webauthn rejects login when the
// stored BackupEligible value disagrees with the assertion, which is why these
// must be persisted — cloud-synced passkeys (iCloud Keychain, Google Password
// Manager, 1Password) report BackupEligible=true, and an unset stored value
// would fail validation.
type Passkey struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	CredentialID    []byte    `json:"credential_id"`
	PublicKey       []byte    `json:"public_key"`
	AAGUID          []byte    `json:"aaguid"`
	SignCount       uint32    `json:"sign_count"`
	Name            string    `json:"name"`
	CreatedAt       time.Time `json:"created_at"`
	LastUsedAt      time.Time `json:"last_used_at"`
	Transports      []string  `json:"transports,omitempty"`
	AttestationType string    `json:"attestation_type"`
	BackupEligible  bool      `json:"backup_eligible"`
	BackupState     bool      `json:"backup_state"`
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
