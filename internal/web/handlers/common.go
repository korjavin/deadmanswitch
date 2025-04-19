package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// updateEntityAssignments is a shared helper function for updating associations between
// entities (secrets and recipients) in the database.
//
// Parameters:
// - ctx: The context for database operations
// - repo: Repository interface for database operations
// - userID: ID of the user performing the operation
// - entityID: ID of the entity being updated (either secret or recipient)
// - entityType: Type of entity ("secret" or "recipient")
// - entityName: Name of the entity for audit logging
// - selectedIDs: IDs of entities to assign (recipient IDs or secret IDs)
//
// This function handles both updating secret recipients and recipient secrets.
func updateEntityAssignments(
	ctx context.Context,
	repo storage.Repository,
	userID string,
	entityID string,
	entityType string,
	entityName string,
	selectedIDs []string,
) error {
	var err error
	var currentAssignments []*models.SecretAssignment

	// Get current assignments based on entity type
	if entityType == "secret" {
		currentAssignments, err = repo.ListSecretAssignmentsBySecretID(ctx, entityID)
	} else {
		currentAssignments, err = repo.ListSecretAssignmentsByRecipientID(ctx, entityID)
	}

	if err != nil {
		return fmt.Errorf("error fetching current assignments: %w", err)
	}

	// Create a map of current assignments for quick lookup
	currentMap := make(map[string]string) // key: recipientID/secretID, value: assignmentID
	for _, assignment := range currentAssignments {
		var lookupID string
		if entityType == "secret" {
			lookupID = assignment.RecipientID
		} else {
			lookupID = assignment.SecretID
		}
		currentMap[lookupID] = assignment.ID
	}

	// Create a map of selected IDs for quick lookup
	selectedMap := make(map[string]bool)
	for _, id := range selectedIDs {
		selectedMap[id] = true
	}

	// Process removals - delete assignments that are no longer selected
	for lookupID, assignmentID := range currentMap {
		if !selectedMap[lookupID] {
			if err := repo.DeleteSecretAssignment(ctx, assignmentID); err != nil {
				return fmt.Errorf("error removing assignment: %w", err)
			}
		}
	}

	// Process additions - create new assignments for newly selected entities
	for _, selectedID := range selectedIDs {
		if _, exists := currentMap[selectedID]; !exists {
			// Create new assignment
			assignment := &models.SecretAssignment{
				ID:        generateID(),
				UserID:    userID,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			if entityType == "secret" {
				assignment.SecretID = entityID
				assignment.RecipientID = selectedID
			} else {
				assignment.SecretID = selectedID
				assignment.RecipientID = entityID
			}

			if err := repo.CreateSecretAssignment(ctx, assignment); err != nil {
				return fmt.Errorf("error creating assignment: %w", err)
			}
		}
	}

	return nil
}
