package models

import (
	"testing"
	"time"
)

func TestUserWebAuthnID(t *testing.T) {
	user := &User{
		ID: "test-user-id",
	}

	id := user.WebAuthnID()
	if string(id) != "test-user-id" {
		t.Errorf("Expected WebAuthnID to be 'test-user-id', got '%s'", string(id))
	}
}

func TestUserWebAuthnName(t *testing.T) {
	user := &User{
		Email: "test@example.com",
	}

	name := user.WebAuthnName()
	if name != "test@example.com" {
		t.Errorf("Expected WebAuthnName to be 'test@example.com', got '%s'", name)
	}
}

func TestUserWebAuthnDisplayName(t *testing.T) {
	user := &User{
		Email: "test@example.com",
	}

	displayName := user.WebAuthnDisplayName()
	if displayName != "test@example.com" {
		t.Errorf("Expected WebAuthnDisplayName to be 'test@example.com', got '%s'", displayName)
	}
}

func TestUserWebAuthnIcon(t *testing.T) {
	user := &User{
		Email: "test@example.com",
	}

	icon := user.WebAuthnIcon()
	if icon != "" {
		t.Errorf("Expected WebAuthnIcon to be empty, got '%s'", icon)
	}
}

func TestUserWebAuthnCredentials(t *testing.T) {
	user := &User{
		Email: "test@example.com",
	}

	credentials := user.WebAuthnCredentials()
	if len(credentials) != 0 {
		t.Errorf("Expected WebAuthnCredentials to be empty, got %d credentials", len(credentials))
	}
}

func TestUserModel(t *testing.T) {
	now := time.Now()
	user := User{
		ID:                "user123",
		Email:             "user@example.com",
		PasswordHash:      []byte("hashed_password"),
		TelegramID:        "tg123",
		TelegramUsername:  "tguser",
		LastActivity:      now,
		CreatedAt:         now,
		UpdatedAt:         now,
		PingFrequency:     3,
		PingDeadline:      14,
		PingingEnabled:    true,
		PingMethod:        "both",
		NextScheduledPing: now.Add(3 * 24 * time.Hour),
		TOTPSecret:        "totp_secret",
		TOTPEnabled:       true,
		TOTPVerified:      true,
	}

	// Test field values
	if user.ID != "user123" {
		t.Errorf("Expected ID to be 'user123', got '%s'", user.ID)
	}
	if user.Email != "user@example.com" {
		t.Errorf("Expected Email to be 'user@example.com', got '%s'", user.Email)
	}
	if string(user.PasswordHash) != "hashed_password" {
		t.Errorf("Expected PasswordHash to be 'hashed_password', got '%s'", string(user.PasswordHash))
	}
	if user.TelegramID != "tg123" {
		t.Errorf("Expected TelegramID to be 'tg123', got '%s'", user.TelegramID)
	}
	if user.TelegramUsername != "tguser" {
		t.Errorf("Expected TelegramUsername to be 'tguser', got '%s'", user.TelegramUsername)
	}
	if !user.LastActivity.Equal(now) {
		t.Errorf("Expected LastActivity to be %v, got %v", now, user.LastActivity)
	}
	if !user.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt to be %v, got %v", now, user.CreatedAt)
	}
	if !user.UpdatedAt.Equal(now) {
		t.Errorf("Expected UpdatedAt to be %v, got %v", now, user.UpdatedAt)
	}
	if user.PingFrequency != 3 {
		t.Errorf("Expected PingFrequency to be 3, got %d", user.PingFrequency)
	}
	if user.PingDeadline != 14 {
		t.Errorf("Expected PingDeadline to be 14, got %d", user.PingDeadline)
	}
	if !user.PingingEnabled {
		t.Errorf("Expected PingingEnabled to be true")
	}
	if user.PingMethod != "both" {
		t.Errorf("Expected PingMethod to be 'both', got '%s'", user.PingMethod)
	}
	if !user.NextScheduledPing.Equal(now.Add(3 * 24 * time.Hour)) {
		t.Errorf("Expected NextScheduledPing to be %v, got %v", now.Add(3*24*time.Hour), user.NextScheduledPing)
	}
	if user.TOTPSecret != "totp_secret" {
		t.Errorf("Expected TOTPSecret to be 'totp_secret', got '%s'", user.TOTPSecret)
	}
	if !user.TOTPEnabled {
		t.Errorf("Expected TOTPEnabled to be true")
	}
	if !user.TOTPVerified {
		t.Errorf("Expected TOTPVerified to be true")
	}
}

func TestPasskeyModel(t *testing.T) {
	now := time.Now()
	passkey := Passkey{
		ID:             "passkey123",
		UserID:         "user123",
		CredentialID:   []byte("credential_id"),
		PublicKey:      []byte("public_key"),
		AAGUID:         []byte("aaguid"),
		SignCount:      42,
		Name:           "My Passkey",
		CreatedAt:      now,
		LastUsedAt:     now,
		Transports:     []string{"internal", "usb"},
		AttestationType: "none",
	}

	// Test field values
	if passkey.ID != "passkey123" {
		t.Errorf("Expected ID to be 'passkey123', got '%s'", passkey.ID)
	}
	if passkey.UserID != "user123" {
		t.Errorf("Expected UserID to be 'user123', got '%s'", passkey.UserID)
	}
	if string(passkey.CredentialID) != "credential_id" {
		t.Errorf("Expected CredentialID to be 'credential_id', got '%s'", string(passkey.CredentialID))
	}
	if string(passkey.PublicKey) != "public_key" {
		t.Errorf("Expected PublicKey to be 'public_key', got '%s'", string(passkey.PublicKey))
	}
	if string(passkey.AAGUID) != "aaguid" {
		t.Errorf("Expected AAGUID to be 'aaguid', got '%s'", string(passkey.AAGUID))
	}
	if passkey.SignCount != 42 {
		t.Errorf("Expected SignCount to be 42, got %d", passkey.SignCount)
	}
	if passkey.Name != "My Passkey" {
		t.Errorf("Expected Name to be 'My Passkey', got '%s'", passkey.Name)
	}
	if !passkey.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt to be %v, got %v", now, passkey.CreatedAt)
	}
	if !passkey.LastUsedAt.Equal(now) {
		t.Errorf("Expected LastUsedAt to be %v, got %v", now, passkey.LastUsedAt)
	}
	if len(passkey.Transports) != 2 || passkey.Transports[0] != "internal" || passkey.Transports[1] != "usb" {
		t.Errorf("Expected Transports to be ['internal', 'usb'], got %v", passkey.Transports)
	}
	if passkey.AttestationType != "none" {
		t.Errorf("Expected AttestationType to be 'none', got '%s'", passkey.AttestationType)
	}
}

func TestSecretModel(t *testing.T) {
	now := time.Now()
	secret := Secret{
		ID:             "secret123",
		UserID:         "user123",
		Name:           "My Secret",
		EncryptedData:  "encrypted_data",
		CreatedAt:      now,
		UpdatedAt:      now,
		EncryptionType: "aes-256-gcm",
	}

	// Test field values
	if secret.ID != "secret123" {
		t.Errorf("Expected ID to be 'secret123', got '%s'", secret.ID)
	}
	if secret.UserID != "user123" {
		t.Errorf("Expected UserID to be 'user123', got '%s'", secret.UserID)
	}
	if secret.Name != "My Secret" {
		t.Errorf("Expected Name to be 'My Secret', got '%s'", secret.Name)
	}
	if secret.EncryptedData != "encrypted_data" {
		t.Errorf("Expected EncryptedData to be 'encrypted_data', got '%s'", secret.EncryptedData)
	}
	if !secret.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt to be %v, got %v", now, secret.CreatedAt)
	}
	if !secret.UpdatedAt.Equal(now) {
		t.Errorf("Expected UpdatedAt to be %v, got %v", now, secret.UpdatedAt)
	}
	if secret.EncryptionType != "aes-256-gcm" {
		t.Errorf("Expected EncryptionType to be 'aes-256-gcm', got '%s'", secret.EncryptionType)
	}
}

func TestRecipientModel(t *testing.T) {
	now := time.Now()
	confirmedAt := now.Add(-1 * time.Hour)
	confirmationSentAt := now.Add(-2 * time.Hour)
	
	recipient := Recipient{
		ID:                 "recipient123",
		UserID:             "user123",
		Email:              "recipient@example.com",
		Name:               "Test Recipient",
		Message:            "Here are my secrets",
		CreatedAt:          now,
		UpdatedAt:          now,
		PhoneNumber:        "+1234567890",
		IsConfirmed:        true,
		ConfirmedAt:        &confirmedAt,
		ConfirmationCode:   "abc123",
		ConfirmationSentAt: &confirmationSentAt,
	}

	// Test field values
	if recipient.ID != "recipient123" {
		t.Errorf("Expected ID to be 'recipient123', got '%s'", recipient.ID)
	}
	if recipient.UserID != "user123" {
		t.Errorf("Expected UserID to be 'user123', got '%s'", recipient.UserID)
	}
	if recipient.Email != "recipient@example.com" {
		t.Errorf("Expected Email to be 'recipient@example.com', got '%s'", recipient.Email)
	}
	if recipient.Name != "Test Recipient" {
		t.Errorf("Expected Name to be 'Test Recipient', got '%s'", recipient.Name)
	}
	if recipient.Message != "Here are my secrets" {
		t.Errorf("Expected Message to be 'Here are my secrets', got '%s'", recipient.Message)
	}
	if !recipient.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt to be %v, got %v", now, recipient.CreatedAt)
	}
	if !recipient.UpdatedAt.Equal(now) {
		t.Errorf("Expected UpdatedAt to be %v, got %v", now, recipient.UpdatedAt)
	}
	if recipient.PhoneNumber != "+1234567890" {
		t.Errorf("Expected PhoneNumber to be '+1234567890', got '%s'", recipient.PhoneNumber)
	}
	if !recipient.IsConfirmed {
		t.Errorf("Expected IsConfirmed to be true")
	}
	if !recipient.ConfirmedAt.Equal(confirmedAt) {
		t.Errorf("Expected ConfirmedAt to be %v, got %v", confirmedAt, *recipient.ConfirmedAt)
	}
	if recipient.ConfirmationCode != "abc123" {
		t.Errorf("Expected ConfirmationCode to be 'abc123', got '%s'", recipient.ConfirmationCode)
	}
	if !recipient.ConfirmationSentAt.Equal(confirmationSentAt) {
		t.Errorf("Expected ConfirmationSentAt to be %v, got %v", confirmationSentAt, *recipient.ConfirmationSentAt)
	}
}

func TestSessionModel(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	
	session := Session{
		ID:           "session123",
		UserID:       "user123",
		Token:        "token123",
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		LastActivity: now,
		IPAddress:    "127.0.0.1",
		UserAgent:    "Mozilla/5.0",
	}

	// Test field values
	if session.ID != "session123" {
		t.Errorf("Expected ID to be 'session123', got '%s'", session.ID)
	}
	if session.UserID != "user123" {
		t.Errorf("Expected UserID to be 'user123', got '%s'", session.UserID)
	}
	if session.Token != "token123" {
		t.Errorf("Expected Token to be 'token123', got '%s'", session.Token)
	}
	if !session.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt to be %v, got %v", now, session.CreatedAt)
	}
	if !session.ExpiresAt.Equal(expiresAt) {
		t.Errorf("Expected ExpiresAt to be %v, got %v", expiresAt, session.ExpiresAt)
	}
	if !session.LastActivity.Equal(now) {
		t.Errorf("Expected LastActivity to be %v, got %v", now, session.LastActivity)
	}
	if session.IPAddress != "127.0.0.1" {
		t.Errorf("Expected IPAddress to be '127.0.0.1', got '%s'", session.IPAddress)
	}
	if session.UserAgent != "Mozilla/5.0" {
		t.Errorf("Expected UserAgent to be 'Mozilla/5.0', got '%s'", session.UserAgent)
	}
}
