package models

import (
	"time"
)

// SecretQuestion represents a question for a recipient to answer to access a secret
type SecretQuestion struct {
	ID                 string    `json:"id"`
	SecretAssignmentID string    `json:"secret_assignment_id"` // Links to the SecretAssignment
	Question           string    `json:"question"`             // The question text
	Salt               []byte    `json:"salt"`                 // Salt for key derivation
	EncryptedShare     []byte    `json:"encrypted_share"`      // Share encrypted with answer-derived key
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// SecretQuestionSet represents a set of questions for a specific secret-recipient pair
type SecretQuestionSet struct {
	ID                 string    `json:"id"`
	SecretAssignmentID string    `json:"secret_assignment_id"` // Links to the SecretAssignment
	Threshold          int       `json:"threshold"`            // Number of correct answers needed (k)
	TotalQuestions     int       `json:"total_questions"`      // Total number of questions (N)
	TimelockRound      uint64    `json:"timelock_round"`       // drand round for time-lock
	EncryptedBlob      []byte    `json:"encrypted_blob"`       // Time-locked encrypted blob containing questions
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
