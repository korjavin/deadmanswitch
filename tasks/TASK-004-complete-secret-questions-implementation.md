# TASK-004: Complete Secret Questions Implementation (Shamir)

## Priority: MEDIUM 🟡

## Status: Partially Implemented

## Category: Cryptography / Secret Recovery

## Description

The application has a sophisticated secret questions feature using Shamir Secret Sharing combined with timelock encryption. This allows recipients to answer security questions to reconstruct secrets, with questions locked until the deadline approaches. The cryptography is implemented, but the full workflow including re-encryption is incomplete.

## Problem

**Current State:**
- Shamir secret sharing is implemented: `/internal/crypto/shamir.go`
- Frontend UI exists: `/web/templates/questions.html`
- Database model exists: `SecretQuestion` in `/internal/models/secret_question.go`
- Handler partially implemented: `/internal/web/handlers/secret_questions.go`

**Missing Functionality:**
From `/todo.md:18-22`:
- [ ] Encode questions in encrypted JSON

**Why This is Important:**
3. Provides alternative to direct secret delivery
4. Allows threshold-based secret reconstruction (answer 2 of 3 questions)

## Background: How It Works


### Shamir Secret Sharing
The actual secret is split into N shares. Recipient must answer K of N questions correctly to reconstruct the secret (threshold cryptography).


## Proposed Solution

### 1. Implement Question Encoding on Creation

**File to modify:** `internal/web/handlers/secret_questions.go`

When recipient secret questions are saved:

```go
func (h *Handler) HandleSaveSecretQuestions(w http.ResponseWriter, r *http.Request) {
    // ... existing code to parse questions ...

    // Get user's current activity timestamp and deadline
    user, _ := h.repo.GetUserByID(ctx, secret.UserID)
    deadline := user.LastActivity.Add(time.Duration(user.Deadline) * 24 * time.Hour)

    // Prepare question data for encryption
    questionData := make([]crypto.QuestionData, len(questions))
    for i, q := range questions {
        questionData[i] = crypto.QuestionData{
            Question: q.Question,
            Answer:   q.Answer,
        }
    }

    // Encrypt questions with timelock to deadline
    encryptedBlob, timelockRound, err := crypto.EncryptQuestions(
        questionData,
        threshold,
        deadline,
    )
    if err != nil {
        // Handle error
    }

    // Store encrypted blob and timelock round
    secretQuestion := &models.SecretQuestion{
        ID:            uuid.New().String(),
        RecipientID:   recipientID,
        SecretID:      secretID,
        Threshold:     threshold,
        EncryptedData: base64.StdEncoding.EncodeToString(encryptedBlob),
        TimelockRound: timelockRound,
        LockedUntil:   deadline,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    h.repo.CreateSecretQuestion(ctx, secretQuestion)
}
```



## References

- Shamir implementation: `/internal/crypto/shamir.go`
- Handler: `/internal/web/handlers/secret_questions.go`
- Frontend: `/web/templates/questions.html`
