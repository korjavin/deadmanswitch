package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/korjavin/deadmanswitch/internal/crypto"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/utils"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// SecretQuestionsHandler handles secret questions-related requests
type SecretQuestionsHandler struct {
	repo      storage.Repository
	templates *templates.TemplateRenderer
}

// NewSecretQuestionsHandler creates a new SecretQuestionsHandler
func NewSecretQuestionsHandler(repo storage.Repository, templates *templates.TemplateRenderer) *SecretQuestionsHandler {
	return &SecretQuestionsHandler{
		repo:      repo,
		templates: templates,
	}
}

// RegisterRoutes registers the routes for the SecretQuestionsHandler
// Note: This function is not used in the current implementation
// Routes are registered in server.go instead
func (h *SecretQuestionsHandler) RegisterRoutes(r *http.ServeMux, auth func(http.HandlerFunc) http.HandlerFunc) {
	// This function is kept for reference but is not used
	// Routes for managing secret questions are registered in server.go
}

// ShowQuestionsPage shows the page for managing secret questions for a recipient
func (h *SecretQuestionsHandler) ShowQuestionsPage(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipient ID from the context
	recipientID, ok := r.Context().Value("recipientID").(string)
	if !ok || recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Get the recipient
	recipient, err := h.repo.GetRecipientByID(r.Context(), recipientID)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Recipient not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching recipient", http.StatusInternalServerError)
			log.Printf("Error fetching recipient: %v", err)
		}
		return
	}

	// Check if the recipient belongs to the user
	if recipient.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the secret assignments for the recipient
	assignments, err := h.repo.ListSecretAssignmentsByRecipientID(r.Context(), recipientID)
	if err != nil {
		http.Error(w, "Error fetching secret assignments", http.StatusInternalServerError)
		log.Printf("Error fetching secret assignments: %v", err)
		return
	}

	// Get the secrets for each assignment
	type SecretWithQuestions struct {
		Secret           *models.Secret
		Assignment       *models.SecretAssignment
		QuestionSet      *models.SecretQuestionSet
		Questions        []*models.SecretQuestion
		TimelockDeadline time.Time
	}

	secretsWithQuestions := make([]*SecretWithQuestions, 0)

	for _, assignment := range assignments {
		// Get the secret
		secret, err := h.repo.GetSecretByID(r.Context(), assignment.SecretID)
		if err != nil {
			log.Printf("Error fetching secret %s: %v", assignment.SecretID, err)
			continue
		}

		// Get the question set for this assignment
		questionSet, err := h.repo.GetSecretQuestionSetByAssignmentID(r.Context(), assignment.ID)
		if err != nil && err != storage.ErrNotFound {
			log.Printf("Error fetching question set for assignment %s: %v", assignment.ID, err)
			continue
		}

		// Get the questions for this assignment
		questions, err := h.repo.ListSecretQuestionsByAssignmentID(r.Context(), assignment.ID)
		if err != nil {
			log.Printf("Error fetching questions for assignment %s: %v", assignment.ID, err)
			continue
		}

		// Calculate the timelock deadline if a question set exists
		var timelockDeadline time.Time
		if questionSet != nil {
			// Get the drand client info to calculate the time for the round
			// This is a simplified version - in a real implementation, we would use the drand client
			// For now, we'll just use a placeholder time
			timelockDeadline = time.Now().Add(time.Duration(user.PingDeadline) * 24 * time.Hour)
		}

		secretWithQuestions := &SecretWithQuestions{
			Secret:           secret,
			Assignment:       assignment,
			QuestionSet:      questionSet,
			Questions:        questions,
			TimelockDeadline: timelockDeadline,
		}

		secretsWithQuestions = append(secretsWithQuestions, secretWithQuestions)
	}

	// Render the template
	data := map[string]interface{}{
		"User":                 user,
		"Recipient":            recipient,
		"SecretsWithQuestions": secretsWithQuestions,
	}

	if err := h.templates.Render(w, "questions.html", data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v", err)
	}
}

// CreateQuestions creates new secret questions for a recipient
func (h *SecretQuestionsHandler) CreateQuestions(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipient ID from the context
	recipientID, ok := r.Context().Value("recipientID").(string)
	if !ok || recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Get the recipient
	recipient, err := h.repo.GetRecipientByID(r.Context(), recipientID)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Recipient not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching recipient", http.StatusInternalServerError)
			log.Printf("Error fetching recipient: %v", err)
		}
		return
	}

	// Check if the recipient belongs to the user
	if recipient.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		log.Printf("Error parsing form: %v", err)
		return
	}

	// Get the assignment ID
	assignmentID := r.Form.Get("assignment_id")
	if assignmentID == "" {
		http.Error(w, "Assignment ID is required", http.StatusBadRequest)
		return
	}

	// Get the assignment
	assignment, err := h.repo.GetSecretAssignmentByID(r.Context(), assignmentID)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Assignment not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching assignment", http.StatusInternalServerError)
			log.Printf("Error fetching assignment: %v", err)
		}
		return
	}

	// Check if the assignment belongs to the user
	if assignment.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the secret
	secret, err := h.repo.GetSecretByID(r.Context(), assignment.SecretID)
	if err != nil {
		http.Error(w, "Error fetching secret", http.StatusInternalServerError)
		log.Printf("Error fetching secret: %v", err)
		return
	}

	// Get the questions and answers from the form
	questions := r.Form["question"]
	answers := r.Form["answer"]

	if len(questions) != len(answers) {
		http.Error(w, "Number of questions and answers must match", http.StatusBadRequest)
		return
	}

	if len(questions) < 3 {
		http.Error(w, "At least 3 questions are required", http.StatusBadRequest)
		return
	}

	// Get the threshold from the form
	thresholdStr := r.Form.Get("threshold")
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil || threshold < 2 || threshold > len(questions) {
		http.Error(w, "Invalid threshold", http.StatusBadRequest)
		return
	}

	// In a real implementation, we would get the master key from the user's session
	// For now, we'll use a dummy master key for demonstration
	masterKey := []byte("this-is-a-dummy-master-key-for-demo-only")

	// Decrypt the secret to get the DEK
	secretContent, err := crypto.DecryptSecret(secret.EncryptedData, masterKey)
	if err != nil {
		http.Error(w, "Error decrypting secret", http.StatusInternalServerError)
		log.Printf("Error decrypting secret: %v", err)
		return
	}

	// Split the DEK using Shamir's Secret Sharing
	shares, err := crypto.SplitSecret(secretContent, threshold, len(questions))
	if err != nil {
		http.Error(w, "Error splitting secret", http.StatusInternalServerError)
		log.Printf("Error splitting secret: %v", err)
		return
	}

	// Create the questions and encrypt the shares
	var questionData []crypto.QuestionData
	var secretQuestions []*models.SecretQuestion

	for i, question := range questions {
		// Encrypt the share with the answer
		encryptedShare, salt, err := crypto.EncryptShare(shares[i], answers[i])
		if err != nil {
			http.Error(w, "Error encrypting share", http.StatusInternalServerError)
			log.Printf("Error encrypting share: %v", err)
			return
		}

		// Create the question data for timelock encryption
		questionData = append(questionData, crypto.QuestionData{
			Question:       question,
			Salt:           salt,
			EncryptedShare: encryptedShare,
		})

		// Create the secret question
		secretQuestion := &models.SecretQuestion{
			ID:                 utils.GenerateID(),
			SecretAssignmentID: assignmentID,
			Question:           question,
			Salt:               salt,
			EncryptedShare:     encryptedShare,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		secretQuestions = append(secretQuestions, secretQuestion)
	}

	// Calculate the deadline for the timelock
	deadline := time.Now().Add(time.Duration(user.PingDeadline) * 24 * time.Hour)

	// Encrypt the questions with timelock
	encryptedBlob, timelockRound, err := crypto.EncryptQuestions(questionData, threshold, deadline)
	if err != nil {
		http.Error(w, "Error encrypting questions with timelock", http.StatusInternalServerError)
		log.Printf("Error encrypting questions with timelock: %v", err)
		return
	}

	// Create the question set
	questionSet := &models.SecretQuestionSet{
		ID:                 utils.GenerateID(),
		SecretAssignmentID: assignmentID,
		Threshold:          threshold,
		TotalQuestions:     len(questions),
		TimelockRound:      timelockRound,
		EncryptedBlob:      encryptedBlob,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Start a transaction
	tx, err := h.repo.BeginTx(r.Context())
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		log.Printf("Error starting transaction: %v", err)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("Failed to rollback transaction: %v", err)
		}
	}()

	// Create the question set
	if err := tx.CreateSecretQuestionSet(r.Context(), questionSet); err != nil {
		http.Error(w, "Error creating question set", http.StatusInternalServerError)
		log.Printf("Error creating question set: %v", err)
		return
	}

	// Create the questions
	for _, question := range secretQuestions {
		if err := tx.CreateSecretQuestion(r.Context(), question); err != nil {
			http.Error(w, "Error creating question", http.StatusInternalServerError)
			log.Printf("Error creating question: %v", err)
			return
		}
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Action:    "create_secret_questions",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Created %d secret questions for recipient %s", len(questions), recipient.Name),
	}

	if err := tx.CreateAuditLog(r.Context(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		log.Printf("Error committing transaction: %v", err)
		return
	}

	// Redirect back to the questions page
	http.Redirect(w, r, fmt.Sprintf("/recipients/%s/questions", recipientID), http.StatusSeeOther)
}

// UpdateQuestion updates a secret question
func (h *SecretQuestionsHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipient ID from the context
	recipientID, ok := r.Context().Value("recipientID").(string)
	if !ok || recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Extract the question ID from the URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var questionID string
	for i, part := range parts {
		if part == "questions" && i+1 < len(parts) {
			questionID = parts[i+1]
			break
		}
	}

	if questionID == "" {
		http.Error(w, "Question ID is required", http.StatusBadRequest)
		return
	}

	// Get the recipient
	recipient, err := h.repo.GetRecipientByID(r.Context(), recipientID)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Recipient not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching recipient", http.StatusInternalServerError)
			log.Printf("Error fetching recipient: %v", err)
		}
		return
	}

	// Check if the recipient belongs to the user
	if recipient.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the question
	question, err := h.repo.GetSecretQuestion(r.Context(), questionID)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Question not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching question", http.StatusInternalServerError)
			log.Printf("Error fetching question: %v", err)
		}
		return
	}

	// Get the assignment
	assignment, err := h.repo.GetSecretAssignmentByID(r.Context(), question.SecretAssignmentID)
	if err != nil {
		http.Error(w, "Error fetching assignment", http.StatusInternalServerError)
		log.Printf("Error fetching assignment: %v", err)
		return
	}

	// Check if the assignment belongs to the user
	if assignment.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		log.Printf("Error parsing form: %v", err)
		return
	}

	// Get the new question text and answer
	newQuestion := r.Form.Get("question")
	newAnswer := r.Form.Get("answer")

	if newQuestion == "" || newAnswer == "" {
		http.Error(w, "Question and answer are required", http.StatusBadRequest)
		return
	}

	// Get the secret
	secret, err := h.repo.GetSecretByID(r.Context(), assignment.SecretID)
	if err != nil {
		http.Error(w, "Error fetching secret", http.StatusInternalServerError)
		log.Printf("Error fetching secret: %v", err)
		return
	}

	// In a real implementation, we would get the master key from the user's session
	// For now, we'll use a dummy master key for demonstration
	masterKey := []byte("this-is-a-dummy-master-key-for-demo-only")

	// Decrypt the secret to get the DEK
	secretContent, err := crypto.DecryptSecret(secret.EncryptedData, masterKey)
	if err != nil {
		http.Error(w, "Error decrypting secret", http.StatusInternalServerError)
		log.Printf("Error decrypting secret: %v", err)
		return
	}

	// Get the question set
	questionSet, err := h.repo.GetSecretQuestionSetByAssignmentID(r.Context(), question.SecretAssignmentID)
	if err != nil {
		http.Error(w, "Error fetching question set", http.StatusInternalServerError)
		log.Printf("Error fetching question set: %v", err)
		return
	}

	// Get all questions for this assignment
	questions, err := h.repo.ListSecretQuestionsByAssignmentID(r.Context(), question.SecretAssignmentID)
	if err != nil {
		http.Error(w, "Error fetching questions", http.StatusInternalServerError)
		log.Printf("Error fetching questions: %v", err)
		return
	}

	// Create a new share for this question
	shares, err := crypto.SplitSecret(secretContent, questionSet.Threshold, questionSet.TotalQuestions)
	if err != nil {
		http.Error(w, "Error splitting secret", http.StatusInternalServerError)
		log.Printf("Error splitting secret: %v", err)
		return
	}

	// Find the index of the current question
	var questionIndex int
	for i, q := range questions {
		if q.ID == questionID {
			questionIndex = i
			break
		}
	}

	// Encrypt the share with the new answer
	encryptedShare, salt, err := crypto.EncryptShare(shares[questionIndex], newAnswer)
	if err != nil {
		http.Error(w, "Error encrypting share", http.StatusInternalServerError)
		log.Printf("Error encrypting share: %v", err)
		return
	}

	// Update the question
	question.Question = newQuestion
	question.Salt = salt
	question.EncryptedShare = encryptedShare
	question.UpdatedAt = time.Now()

	// Update all questions with new shares
	var questionData []crypto.QuestionData
	for i, q := range questions {
		if q.ID == questionID {
			// Use the updated question
			questionData = append(questionData, crypto.QuestionData{
				Question:       newQuestion,
				Salt:           salt,
				EncryptedShare: encryptedShare,
			})
		} else {
			// Re-encrypt the share for this question
			answer := r.Form.Get("answer_" + q.ID)
			if answer == "" {
				// If no answer provided, we can't update this question
				// In a real implementation, we would handle this differently
				http.Error(w, "All answers are required to update a question", http.StatusBadRequest)
				return
			}

			encShare, qSalt, err := crypto.EncryptShare(shares[i], answer)
			if err != nil {
				http.Error(w, "Error encrypting share", http.StatusInternalServerError)
				log.Printf("Error encrypting share: %v", err)
				return
			}

			// Update the question
			q.Salt = qSalt
			q.EncryptedShare = encShare
			q.UpdatedAt = time.Now()

			questionData = append(questionData, crypto.QuestionData{
				Question:       q.Question,
				Salt:           qSalt,
				EncryptedShare: encShare,
			})
		}
	}

	// Calculate the deadline for the timelock
	deadline := time.Now().Add(time.Duration(user.PingDeadline) * 24 * time.Hour)

	// Encrypt the questions with timelock
	encryptedBlob, timelockRound, err := crypto.EncryptQuestions(questionData, questionSet.Threshold, deadline)
	if err != nil {
		http.Error(w, "Error encrypting questions with timelock", http.StatusInternalServerError)
		log.Printf("Error encrypting questions with timelock: %v", err)
		return
	}

	// Update the question set
	questionSet.TimelockRound = timelockRound
	questionSet.EncryptedBlob = encryptedBlob
	questionSet.UpdatedAt = time.Now()

	// Start a transaction
	tx, err := h.repo.BeginTx(r.Context())
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		log.Printf("Error starting transaction: %v", err)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("Failed to rollback transaction: %v", err)
		}
	}()

	// Update the question set
	if err := tx.UpdateSecretQuestionSet(r.Context(), questionSet); err != nil {
		http.Error(w, "Error updating question set", http.StatusInternalServerError)
		log.Printf("Error updating question set: %v", err)
		return
	}

	// Update the question
	if err := tx.UpdateSecretQuestion(r.Context(), question); err != nil {
		http.Error(w, "Error updating question", http.StatusInternalServerError)
		log.Printf("Error updating question: %v", err)
		return
	}

	// Update all other questions
	for _, q := range questions {
		if q.ID != questionID {
			if err := tx.UpdateSecretQuestion(r.Context(), q); err != nil {
				http.Error(w, "Error updating question", http.StatusInternalServerError)
				log.Printf("Error updating question: %v", err)
				return
			}
		}
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Action:    "update_secret_question",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Updated secret question for recipient %s", recipient.Name),
	}

	if err := tx.CreateAuditLog(r.Context(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		log.Printf("Error committing transaction: %v", err)
		return
	}

	// Redirect back to the questions page
	http.Redirect(w, r, fmt.Sprintf("/recipients/%s/questions", recipientID), http.StatusSeeOther)
}

// DeleteQuestion deletes a secret question
func (h *SecretQuestionsHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipient ID from the context
	recipientID, ok := r.Context().Value("recipientID").(string)
	if !ok || recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Extract the question ID from the URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var questionID string
	for i, part := range parts {
		if part == "questions" && i+1 < len(parts) {
			questionID = parts[i+1]
			break
		}
	}

	if questionID == "" {
		http.Error(w, "Question ID is required", http.StatusBadRequest)
		return
	}

	// Get the recipient
	recipient, err := h.repo.GetRecipientByID(r.Context(), recipientID)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Recipient not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching recipient", http.StatusInternalServerError)
			log.Printf("Error fetching recipient: %v", err)
		}
		return
	}

	// Check if the recipient belongs to the user
	if recipient.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the question
	question, err := h.repo.GetSecretQuestion(r.Context(), questionID)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Question not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching question", http.StatusInternalServerError)
			log.Printf("Error fetching question: %v", err)
		}
		return
	}

	// Get the assignment
	assignment, err := h.repo.GetSecretAssignmentByID(r.Context(), question.SecretAssignmentID)
	if err != nil {
		http.Error(w, "Error fetching assignment", http.StatusInternalServerError)
		log.Printf("Error fetching assignment: %v", err)
		return
	}

	// Check if the assignment belongs to the user
	if assignment.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the question set
	questionSet, err := h.repo.GetSecretQuestionSetByAssignmentID(r.Context(), question.SecretAssignmentID)
	if err != nil {
		http.Error(w, "Error fetching question set", http.StatusInternalServerError)
		log.Printf("Error fetching question set: %v", err)
		return
	}

	// Get all questions for this assignment
	questions, err := h.repo.ListSecretQuestionsByAssignmentID(r.Context(), question.SecretAssignmentID)
	if err != nil {
		http.Error(w, "Error fetching questions", http.StatusInternalServerError)
		log.Printf("Error fetching questions: %v", err)
		return
	}

	// Check if we have enough questions left after deletion
	if len(questions) <= questionSet.Threshold {
		http.Error(w, "Cannot delete question: not enough questions would remain to meet the threshold", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := h.repo.BeginTx(r.Context())
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		log.Printf("Error starting transaction: %v", err)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("Failed to rollback transaction: %v", err)
		}
	}()

	// Delete the question
	if err := tx.DeleteSecretQuestion(r.Context(), questionID); err != nil {
		http.Error(w, "Error deleting question", http.StatusInternalServerError)
		log.Printf("Error deleting question: %v", err)
		return
	}

	// Update the question set
	questionSet.TotalQuestions--
	questionSet.UpdatedAt = time.Now()

	if err := tx.UpdateSecretQuestionSet(r.Context(), questionSet); err != nil {
		http.Error(w, "Error updating question set", http.StatusInternalServerError)
		log.Printf("Error updating question set: %v", err)
		return
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Action:    "delete_secret_question",
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Deleted secret question for recipient %s", recipient.Name),
	}

	if err := tx.CreateAuditLog(r.Context(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		log.Printf("Error committing transaction: %v", err)
		return
	}

	// Redirect back to the questions page
	http.Redirect(w, r, fmt.Sprintf("/recipients/%s/questions", recipientID), http.StatusSeeOther)
}
