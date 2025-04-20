package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/korjavin/deadmanswitch/internal/crypto"
	"github.com/korjavin/deadmanswitch/internal/models"
)

// ReencryptQuestionsTask re-encrypts question sets that are approaching their timelock deadline
func (s *Scheduler) ReencryptQuestionsTask(ctx context.Context) error {
	log.Println("Running reencryptQuestionsTask")

	// Define the safe margin (e.g., 24 hours before the deadline)
	safeMarginSeconds := int64(24 * 60 * 60)

	// Get all question sets that need re-encryption
	questionSets, err := s.repo.ListSecretQuestionSetsNeedingReencryption(ctx, safeMarginSeconds)
	if err != nil {
		return fmt.Errorf("failed to list question sets needing re-encryption: %w", err)
	}

	log.Printf("Found %d question sets needing re-encryption", len(questionSets))

	for _, set := range questionSets {
		// Get the assignment
		assignment, err := s.repo.GetSecretAssignmentByID(ctx, set.SecretAssignmentID)
		if err != nil {
			log.Printf("Error fetching assignment %s: %v", set.SecretAssignmentID, err)
			continue
		}

		// Get the user
		user, err := s.repo.GetUserByID(ctx, assignment.UserID)
		if err != nil {
			log.Printf("Error fetching user %s: %v", assignment.UserID, err)
			continue
		}

		// Check if the user is still active
		if time.Since(user.LastActivity) > time.Duration(user.PingDeadline)*24*time.Hour {
			log.Printf("User %s is inactive, not re-encrypting question set %s", user.ID, set.ID)
			continue
		}

		// Get the questions
		questions, err := s.repo.ListSecretQuestionsByAssignmentID(ctx, set.SecretAssignmentID)
		if err != nil {
			log.Printf("Error fetching questions for assignment %s: %v", set.SecretAssignmentID, err)
			continue
		}

		// Prepare question data for re-encryption
		var questionData []crypto.QuestionData
		for _, q := range questions {
			questionData = append(questionData, crypto.QuestionData{
				Question:       q.Question,
				Salt:           q.Salt,
				EncryptedShare: q.EncryptedShare,
			})
		}

		// Calculate the new deadline
		deadline := time.Now().Add(time.Duration(user.PingDeadline) * 24 * time.Hour)

		// Re-encrypt the questions with a new timelock
		encryptedBlob, timelockRound, err := crypto.EncryptQuestions(questionData, set.Threshold, deadline)
		if err != nil {
			log.Printf("Error re-encrypting questions for set %s: %v", set.ID, err)
			continue
		}

		// Update the question set
		set.TimelockRound = timelockRound
		set.EncryptedBlob = encryptedBlob
		set.UpdatedAt = time.Now()

		if err := s.repo.UpdateSecretQuestionSet(ctx, set); err != nil {
			log.Printf("Error updating question set %s: %v", set.ID, err)
			continue
		}

		log.Printf("Successfully re-encrypted question set %s for user %s", set.ID, user.ID)

		// Create an audit log entry
		auditLog := &models.AuditLog{
			UserID:    user.ID,
			Action:    "reencrypt_questions",
			Timestamp: time.Now(),
			Details:   fmt.Sprintf("Re-encrypted questions for secret assignment %s", set.SecretAssignmentID),
		}

		if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
			log.Printf("Error creating audit log: %v", err)
			// Continue anyway, don't fail the whole operation
		}
	}

	return nil
}
