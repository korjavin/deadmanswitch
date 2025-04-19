package models

import (
	"time"
)

// User represents a registered user of the application
type User struct {
	ID                string    `json:"id"`
	Email             string    `json:"email"`
	PasswordHash      []byte    `json:"-"`
	TelegramID        string    `json:"telegram_id,omitempty"`
	TelegramUsername  string    `json:"telegram_username,omitempty"`
	GitHubUsername    string    `json:"github_username,omitempty"`
	LastActivity      time.Time `json:"last_activity"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	PingFrequency     int       `json:"ping_frequency"` // Days
	PingDeadline      int       `json:"ping_deadline"`  // Days
	PingingEnabled    bool      `json:"pinging_enabled"`
	PingMethod        string    `json:"ping_method"` // "telegram", "email", or "both"
	NextScheduledPing time.Time `json:"next_scheduled_ping"`
	// 2FA fields
	TOTPSecret   string `json:"totp_secret,omitempty"` // Secret for TOTP-based 2FA
	TOTPEnabled  bool   `json:"totp_enabled"`          // Whether 2FA is enabled
	TOTPVerified bool   `json:"totp_verified"`         // Whether 2FA has been verified
}

// Secret represents an encrypted secret note
type Secret struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Name           string    `json:"name"`
	EncryptedData  string    `json:"encrypted_data"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	EncryptionType string    `json:"encryption_type"` // e.g., "aes-256-gcm"
}

// Recipient represents someone who will receive secrets
type Recipient struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"` // The user who created this recipient
	Email              string     `json:"email"`
	Name               string     `json:"name"`
	Message            string     `json:"message"` // Custom message to send with the secrets
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	PhoneNumber        string     `json:"phone_number,omitempty"`
	IsConfirmed        bool       `json:"is_confirmed"`
	ConfirmedAt        *time.Time `json:"confirmed_at,omitempty"`
	ConfirmationCode   string     `json:"confirmation_code,omitempty"`
	ConfirmationSentAt *time.Time `json:"confirmation_sent_at,omitempty"`
}

// SecretAssignment links secrets to recipients
type SecretAssignment struct {
	ID          string    `json:"id"`
	SecretID    string    `json:"secret_id"`
	RecipientID string    `json:"recipient_id"`
	UserID      string    `json:"user_id"` // The user who created this assignment
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PingHistory records all pings sent to a user
type PingHistory struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	SentAt      time.Time  `json:"sent_at"`
	Method      string     `json:"method"` // "email" or "telegram"
	Status      string     `json:"status"` // "sent", "delivered", "responded"
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// PingVerification stores verification codes for email pings
type PingVerification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// DeliveryEvent records when secrets are delivered to recipients
type DeliveryEvent struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	RecipientID  string    `json:"recipient_id"`
	SentAt       time.Time `json:"sent_at"`
	Status       string    `json:"status"` // "sent", "delivered", "failed"
	ErrorMessage string    `json:"error_message,omitempty"`
}

// AuditLog stores important security events
type AuditLog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Details   string    `json:"details,omitempty"`
}

// Session represents a user login session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Token        string    `json:"token"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
}
