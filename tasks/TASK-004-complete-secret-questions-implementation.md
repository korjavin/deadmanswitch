# TASK-004: Complete Secret Questions Implementation (Shamir + Timelock)

## Priority: MEDIUM 🟡

## Status: Partially Implemented

## Category: Cryptography / Secret Recovery

## Description

The application has a sophisticated secret questions feature using Shamir Secret Sharing combined with timelock encryption. This allows recipients to answer security questions to reconstruct secrets, with questions locked until the deadline approaches. The cryptography is implemented, but the full workflow including re-encryption is incomplete.

## Problem

**Current State:**
- Shamir secret sharing is implemented: `/internal/crypto/shamir.go`
- Timelock encryption is implemented: `/internal/crypto/timelock.go`
- Frontend UI exists: `/web/templates/questions.html`
- Database model exists: `SecretQuestion` in `/internal/models/secret_question.go`
- Handler partially implemented: `/internal/web/handlers/secret_questions.go`

**Missing Functionality:**
From `/todo.md:18-22`:
- [ ] Encode questions in timelock-encrypted JSON, locked as `user_active_timestamp + deadline_time`
- [ ] Implement re-encoding when `user_active_timestamp + deadline_time - safe_margin` reached and user is still active
- [ ] Add unit tests for re-encryption mechanism

**Why This is Important:**
1. Prevents premature access to questions before deadline
2. Automatically extends protection if user remains active
3. Provides alternative to direct secret delivery
4. Allows threshold-based secret reconstruction (answer 2 of 3 questions)

## Background: How It Works

### Timelock Encryption
Questions are encrypted to a specific Unix timestamp (deadline). The encryption uses time-based rounds that can only be computed after that time passes.

### Shamir Secret Sharing
The actual secret is split into N shares. Recipient must answer K of N questions correctly to reconstruct the secret (threshold cryptography).

### Re-encryption Challenge
If the user is still active, the deadline extends. Questions must be re-encrypted to the new deadline to prevent early access.

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

### 2. Implement Re-encryption Scheduler Task

**File to modify:** `internal/scheduler/scheduler.go`

Add a new scheduled task to re-encrypt questions when deadline approaches but user is still active:

```go
func (s *Scheduler) Start(ctx context.Context) {
    // ... existing tasks ...

    // Re-encrypt secret questions for active users
    go s.runReencryptionTask(ctx)
}

func (s *Scheduler) runReencryptionTask(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Hour) // Check hourly
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            s.reencryptSecretQuestions(ctx)
        case <-ctx.Done():
            return
        }
    }
}

func (s *Scheduler) reencryptSecretQuestions(ctx context.Context) {
    // Define safety margin (e.g., 7 days before deadline)
    safetyMargin := 7 * 24 * time.Hour

    // Find secret questions that need re-encryption:
    // - LockedUntil - now < safetyMargin
    // - User is still active (LastActivity + Deadline > now)
    questions, err := s.repo.GetSecretQuestionsNeedingReencryption(ctx, safetyMargin)
    if err != nil {
        log.Printf("Failed to get questions for re-encryption: %v", err)
        return
    }

    for _, question := range questions {
        if err := s.reencryptQuestion(ctx, question); err != nil {
            log.Printf("Failed to re-encrypt question %s: %v", question.ID, err)
            continue
        }

        log.Printf("Successfully re-encrypted secret question %s", question.ID)

        // Create audit log
        s.repo.CreateAuditLog(ctx, &models.AuditLog{
            ID:        uuid.New().String(),
            UserID:    question.UserID,
            Action:    "secret_question_reencrypted",
            Details:   fmt.Sprintf("Question ID: %s", question.ID),
            Timestamp: time.Now(),
            IPAddress: "scheduler",
        })
    }
}

func (s *Scheduler) reencryptQuestion(ctx context.Context, question *models.SecretQuestion) error {
    // Get user to determine new deadline
    user, err := s.repo.GetUserByID(ctx, question.UserID)
    if err != nil {
        return err
    }

    // Calculate new deadline
    newDeadline := user.LastActivity.Add(time.Duration(user.Deadline) * 24 * time.Hour)

    // Decrypt current questions (timelock should be open since we're near deadline)
    encryptedBlob, err := base64.StdEncoding.DecodeString(question.EncryptedData)
    if err != nil {
        return err
    }

    timelockData, err := crypto.DecryptQuestions(encryptedBlob)
    if err != nil {
        // If we can't decrypt yet, timelock not open - skip for now
        return fmt.Errorf("timelock not yet open: %w", err)
    }

    // Re-encrypt with new deadline
    newEncryptedBlob, newTimelockRound, err := crypto.EncryptQuestions(
        timelockData.Questions,
        timelockData.Threshold,
        newDeadline,
    )
    if err != nil {
        return err
    }

    // Update database
    question.EncryptedData = base64.StdEncoding.EncodeToString(newEncryptedBlob)
    question.TimelockRound = newTimelockRound
    question.LockedUntil = newDeadline
    question.UpdatedAt = time.Now()

    return s.repo.UpdateSecretQuestion(ctx, question)
}
```

### 3. Add Repository Methods

**File to modify:** `internal/storage/repository.go`

```go
type Repository interface {
    // ... existing methods ...

    GetSecretQuestionsNeedingReencryption(ctx context.Context, safetyMargin time.Duration) ([]*models.SecretQuestion, error)
    UpdateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error
}
```

**File to modify:** `internal/storage/sqlite/sqlite.go`

```go
func (r *SQLiteRepository) GetSecretQuestionsNeedingReencryption(ctx context.Context, safetyMargin time.Duration) ([]*models.SecretQuestion, error) {
    query := `
        SELECT sq.* FROM secret_questions sq
        JOIN secrets s ON sq.secret_id = s.id
        JOIN users u ON s.user_id = u.id
        WHERE
            -- Questions approaching deadline
            datetime(sq.locked_until) <= datetime('now', '+' || ? || ' seconds')
            -- User is still active (not past deadline)
            AND datetime(u.last_activity, '+' || u.deadline || ' days') > datetime('now')
            -- Questions haven't been accessed yet
            AND sq.accessed_at IS NULL
    `

    rows, err := r.db.QueryContext(ctx, query, int(safetyMargin.Seconds()))
    // ... implementation ...
}

func (r *SQLiteRepository) UpdateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error {
    query := `
        UPDATE secret_questions
        SET encrypted_data = ?, timelock_round = ?, locked_until = ?, updated_at = ?
        WHERE id = ?
    `

    _, err := r.db.ExecContext(ctx, query,
        question.EncryptedData,
        question.TimelockRound,
        question.LockedUntil,
        question.UpdatedAt,
        question.ID,
    )
    return err
}
```

### 4. Add Configuration

**File to modify:** `internal/config/config.go`

```go
type Config struct {
    // ... existing fields ...

    // Secret question re-encryption safety margin
    QuestionReencryptMargin time.Duration
}

func LoadFromEnv() (*Config, error) {
    // ... existing code ...

    marginStr := os.Getenv("QUESTION_REENCRYPT_MARGIN_DAYS")
    if marginStr == "" {
        config.QuestionReencryptMargin = 7 * 24 * time.Hour // 7 days default
    } else {
        days, err := strconv.Atoi(marginStr)
        if err != nil {
            return nil, fmt.Errorf("invalid QUESTION_REENCRYPT_MARGIN_DAYS: %w", err)
        }
        config.QuestionReencryptMargin = time.Duration(days) * 24 * time.Hour
    }

    return config, nil
}
```

### 5. Update Database Schema

**File to modify:** `internal/storage/sqlite/migrations.go`

Ensure `secret_questions` table has required fields:

```sql
CREATE TABLE IF NOT EXISTS secret_questions (
    id TEXT PRIMARY KEY,
    recipient_id TEXT NOT NULL,
    secret_id TEXT NOT NULL,
    threshold INTEGER NOT NULL,
    encrypted_data TEXT NOT NULL,
    timelock_round INTEGER NOT NULL,
    locked_until TIMESTAMP NOT NULL,
    accessed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (recipient_id) REFERENCES recipients(id) ON DELETE CASCADE,
    FOREIGN KEY (secret_id) REFERENCES secrets(id) ON DELETE CASCADE
);

CREATE INDEX idx_secret_questions_locked_until ON secret_questions(locked_until);
CREATE INDEX idx_secret_questions_secret_id ON secret_questions(secret_id);
```

## Acceptance Criteria

- [ ] Questions encoded with timelock encryption on creation
- [ ] Deadline calculated as `user.LastActivity + user.Deadline`
- [ ] Re-encryption scheduler task implemented
- [ ] Re-encryption runs hourly
- [ ] Safety margin configurable (default 7 days)
- [ ] Questions re-encrypted when: `deadline - now < safety_margin` AND user active
- [ ] Database methods implemented for re-encryption queries
- [ ] Audit log entries created for re-encryption events
- [ ] Unit tests for re-encryption logic
- [ ] Integration tests for full workflow
- [ ] Documentation updated

## Testing Requirements

1. **Unit Tests:**
   - Test timelock encryption with specific deadline
   - Test question re-encryption logic
   - Test safety margin calculation
   - Test query for questions needing re-encryption

2. **Integration Tests:**
   - Create secret questions for recipient
   - Verify questions locked until deadline
   - Simulate user activity extending deadline
   - Verify questions re-encrypted to new deadline
   - Verify old timelock becomes invalid after re-encryption

3. **Manual Testing:**
   - Create secret with questions (threshold 2 of 3)
   - Verify cannot decrypt before deadline
   - Keep user active past safety margin
   - Verify re-encryption occurs
   - Verify new deadline set correctly

## Files to Create/Modify

**Files to Modify:**
1. `/internal/web/handlers/secret_questions.go` - Implement timelock encoding on save
2. `/internal/scheduler/scheduler.go` - Add re-encryption task
3. `/internal/storage/repository.go` - Add re-encryption query methods
4. `/internal/storage/sqlite/sqlite.go` - Implement re-encryption queries
5. `/internal/config/config.go` - Add safety margin config
6. `/internal/storage/sqlite/migrations.go` - Ensure schema correct
7. `/.env.example` - Document new variable

**Test Files:**
1. `/internal/scheduler/reencryption_test.go` - New test file
2. `/tests/integration/secret_questions_test.go` - Integration tests

## Security Considerations

1. **Timelock Security:**
   - Ensure timelock rounds calculated correctly
   - Verify deadlines use UTC timestamps
   - Prevent timelock manipulation

2. **Re-encryption Timing:**
   - Safety margin prevents race conditions
   - Hourly checks ensure timely re-encryption
   - Handle concurrent re-encryption attempts

3. **Audit Trail:**
   - Log all re-encryption events
   - Track timelock round changes
   - Monitor failed re-encryption attempts

## Performance Considerations

- Re-encryption is computationally expensive (Argon2id + AES)
- Limit concurrent re-encryption operations
- Consider batching if many questions need re-encryption
- Add metrics for re-encryption performance

## References

- TODO items: `/todo.md:18-22`
- Shamir implementation: `/internal/crypto/shamir.go`
- Timelock implementation: `/internal/crypto/timelock.go`
- Handler: `/internal/web/handlers/secret_questions.go`
- Frontend: `/web/templates/questions.html`

## Estimated Effort

- **Complexity:** High
- **Time:** 8-10 hours
- **Risk:** Medium-High (cryptography complexity)

## Dependencies

- Should complete after TASK-001 (master key management)
- Requires understanding of timelock and Shamir cryptography

## Follow-up Tasks

- Add recipient interface to answer questions
- Implement secret reconstruction from shares
- Add progress indicator for timelock rounds
- Consider pre-computing timelock rounds for performance
