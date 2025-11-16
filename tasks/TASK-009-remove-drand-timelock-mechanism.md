# TASK-009: Remove drand/Timelock Mechanism and Secret Questions Feature

## Priority: MEDIUM 🟡

## Status: Not Started

## Category: Feature Removal / Code Cleanup

## Description

Remove the entire drand-based timelock encryption and secret questions feature from the codebase. This includes the cryptographic implementation, database layer, web handlers, UI, scheduler tasks, and all related code (~2,500 lines total).

## Problem

**Current State:**
The project includes a sophisticated but unwanted feature for secret questions using:
- **Timelock encryption** - Secrets locked until a specific time (using drand rounds)
- **Shamir secret sharing** - Splitting secrets into shares (k-of-n threshold)
- Combined approach where questions are timelock-encrypted and answers decrypt Shamir shares

**Why Remove It:**
User has lost interest in this feature due to:
- Complexity of implementation and maintenance
- Additional dependency on drand network concepts
- Over-engineering for the core use case
- Incomplete implementation (re-encryption logic partial)
- Adds ~2,500 lines of code with limited benefit

## Scope of Removal

### 1. Cryptography Layer (~600 lines)

**Files to DELETE:**
- `/internal/crypto/timelock.go` (133 lines)
- `/internal/crypto/timelock_test.go` (169 lines)
- `/internal/crypto/shamir.go` (92 lines)
- `/internal/crypto/shamir_test.go` (174 lines)

**Code Removed:**
- `TimelockEncrypt()` - Encrypt with drand round
- `TimelockDecrypt()` - Decrypt after round available
- `CalculateDrandRound()` - Calculate drand round for timestamp
- `EncryptQuestions()` - Combine timelock + Shamir
- `DecryptQuestions()` - Decrypt timelock questions
- `SplitSecret()` - Shamir secret splitting
- `CombineShares()` - Shamir share reconstruction
- `EncryptShare()` - Encrypt share with answer-derived key
- `DecryptShare()` - Decrypt share with answer

### 2. Database Models (~30 lines)

**File to MODIFY:**
- `/internal/models/secret_questions.go` - **DELETE entire file**

**Models Removed:**
```go
type SecretQuestion struct {
    ID                 string
    SecretAssignmentID string
    Question           string
    Salt               []byte
    EncryptedShare     []byte
    CreatedAt          time.Time
    UpdatedAt          time.Time
}

type SecretQuestionSet struct {
    ID                 string
    SecretAssignmentID string
    Threshold          int
    TotalQuestions     int
    TimelockRound      uint64    // drand round
    EncryptedBlob      []byte    // Time-locked blob
    CreatedAt          time.Time
    UpdatedAt          time.Time
}
```

### 3. Database Layer (~570 lines)

**Files to DELETE:**
- `/internal/storage/secret_questions.go` (346 lines)
- `/internal/storage/mock_secret_questions.go` (158 lines)
- `/internal/storage/migrations/add_secret_questions.go` (68 lines)

**File to MODIFY:**
- `/internal/storage/storage.go` - Remove interface methods (lines 54-67):

```go
// REMOVE THESE:
CreateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error
GetSecretQuestion(ctx context.Context, id string) (*models.SecretQuestion, error)
UpdateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error
DeleteSecretQuestion(ctx context.Context, id string) error
ListSecretQuestionsByAssignmentID(ctx context.Context, assignmentID string) ([]*models.SecretQuestion, error)

CreateSecretQuestionSet(ctx context.Context, set *models.SecretQuestionSet) error
GetSecretQuestionSet(ctx context.Context, id string) (*models.SecretQuestionSet, error)
GetSecretQuestionSetByAssignmentID(ctx context.Context, assignmentID string) (*models.SecretQuestionSet, error)
UpdateSecretQuestionSet(ctx context.Context, set *models.SecretQuestionSet) error
DeleteSecretQuestionSet(ctx context.Context, id string) error
ListSecretQuestionSetsNeedingReencryption(ctx context.Context, safeMarginSeconds int64) ([]*models.SecretQuestionSet, error)
```

**Add Migration to Drop Tables:**
Create new migration file to drop the tables for existing installations:

```sql
-- New file: /internal/storage/migrations/drop_secret_questions.go
DROP TABLE IF EXISTS secret_questions;
DROP TABLE IF EXISTS secret_question_sets;
```

### 4. Scheduler (~110 lines)

**File to DELETE:**
- `/internal/scheduler/reencrypt_questions.go` (103 lines)

**File to MODIFY:**
- `/internal/scheduler/scheduler.go` - Remove task registration (lines 149-156):

```go
// REMOVE THIS:
// Task for re-encrypting secret questions
s.AddTask(&Task{
    ID:         uuid.New().String(),
    Name:       "ReencryptQuestionsTask",
    Duration:   1 * time.Hour,
    RunOnStart: true,
    Handler:    s.ReencryptQuestionsTask,
})
```

### 5. Web Layer (~800 lines)

**File to DELETE:**
- `/internal/web/handlers/secret_questions.go` (793 lines)

**File to MODIFY:**
- `/internal/web/server.go` - Remove handler and routes:

```go
// Line 46 - Remove field:
secretQuestions *handlers.SecretQuestionsHandler

// Line 109 - Remove initialization:
server.handlers.secretQuestions = handlers.NewSecretQuestionsHandler(repo, templates.NewTemplateRenderer())

// Lines 284-295 - Remove route handling:
// Handle questions management
if strings.HasSuffix(r.URL.Path, "/questions") {
    // ... remove this block
}
if strings.Contains(r.URL.Path, "/questions/") {
    // ... remove this block
}
```

### 6. Templates

**File to DELETE:**
- `/web/templates/questions.html` (16KB)

### 7. Dependencies

**File to MODIFY:**
- `/go.mod` - Remove shamir dependency:

```go
// REMOVE:
github.com/corvus-ch/shamir v1.0.1
```

Run: `go mod tidy` after removal

### 8. Documentation

**Files to MODIFY:**

**`/todo.md`** - Remove lines 18-22:
```markdown
- [ ] Implement timebased and shamir division for secret questions for a recipient
      - [ ] First let's add interface to add secret questions to a recipient and
          - [ ] Encode them in  timelock‑encrypted JSON , locked as user_active_timestamp+ deadline_time
          - [ ] Let's implement reencoding them once user_active_timestamp+ deadline_time-some_safe_margin reached and user is still active
          - [ ] add some unit tests for this
```

**`/docs/PROJECT_OVERVIEW.md`** - Remove all mentions of:
- Shamir secret sharing
- Time-lock encryption
- Secret questions
- drand

**`/tasks/TASK-004-complete-secret-questions-implementation.md`** - **DELETE entire file**

**`/tasks/README.md`** - Remove TASK-004 from all lists and update statistics

**Files in `/docs/drafts/`** (optional - can keep as historical reference):
- `Idea_EN.md` - Contains timelock/Shamir design
- `Idea_RU.md` - Russian version
- `Implementation_Notes.md` - Implementation steps for timelock
- `Risk_Model.md` - Risk analysis

Decision: Keep drafts folder as-is (historical reference) or delete if unwanted.

### 9. Tests

Remove all test files already listed above:
- `internal/crypto/timelock_test.go`
- `internal/crypto/shamir_test.go`

Check for integration tests mentioning secret questions:
```bash
rg "SecretQuestion|timelock|shamir" tests/ --type go
```

## Implementation Steps

### Phase 1: Remove Web Layer (Prevent UI Access)
1. Delete `/web/templates/questions.html`
2. Delete `/internal/web/handlers/secret_questions.go`
3. Modify `/internal/web/server.go` - remove handler and routes
4. Test: Verify questions page returns 404

### Phase 2: Remove Scheduler (Stop Background Tasks)
1. Delete `/internal/scheduler/reencrypt_questions.go`
2. Modify `/internal/scheduler/scheduler.go` - remove task registration
3. Test: Verify scheduler starts without errors

### Phase 3: Remove Database Layer
1. Create migration to drop tables
2. Delete `/internal/storage/secret_questions.go`
3. Delete `/internal/storage/mock_secret_questions.go`
4. Delete `/internal/storage/migrations/add_secret_questions.go`
5. Modify `/internal/storage/storage.go` - remove interface methods
6. Test: Verify repository compiles

### Phase 4: Remove Crypto Layer
1. Delete `/internal/crypto/timelock.go`
2. Delete `/internal/crypto/timelock_test.go`
3. Delete `/internal/crypto/shamir.go`
4. Delete `/internal/crypto/shamir_test.go`
5. Test: Run `go test ./internal/crypto/...`

### Phase 5: Remove Models
1. Delete `/internal/models/secret_questions.go`
2. Test: Run `go build ./...`

### Phase 6: Clean Up Dependencies
1. Remove `github.com/corvus-ch/shamir` from `go.mod`
2. Run `go mod tidy`
3. Test: `go mod verify`

### Phase 7: Update Documentation
1. Modify `/todo.md` - remove secret questions TODO
2. Modify `/docs/PROJECT_OVERVIEW.md` - remove all mentions
3. Delete `/tasks/TASK-004-complete-secret-questions-implementation.md`
4. Modify `/tasks/README.md` - remove TASK-004 references
5. Decide on `/docs/drafts/` - keep or delete

### Phase 8: Final Testing
1. Run all tests: `go test ./...`
2. Build application: `go build ./cmd/server`
3. Start application and verify no errors
4. Check database migrations run cleanly
5. Verify UI has no broken links to questions

## Migration Strategy

### For Existing Installations

**If database has secret questions data:**

Create migration that:
1. Checks if tables exist
2. If exist, drops them cleanly
3. Logs dropped data (optional backup)

**Migration file:** `/internal/storage/migrations/drop_secret_questions.go`

```go
package migrations

import (
    "database/sql"
    "fmt"
    "log"
)

func DropSecretQuestionsTables(db *sql.DB) error {
    log.Println("Running migration: Dropping secret questions tables")

    // Check if tables exist
    var count int
    err := db.QueryRow(`
        SELECT COUNT(*) FROM sqlite_master
        WHERE type='table' AND name IN ('secret_questions', 'secret_question_sets')
    `).Scan(&count)

    if err != nil {
        return fmt.Errorf("failed to check tables: %w", err)
    }

    if count == 0 {
        log.Println("Secret questions tables don't exist, skipping")
        return nil
    }

    // Optional: Backup data before dropping
    var questionCount, setCount int
    db.QueryRow("SELECT COUNT(*) FROM secret_questions").Scan(&questionCount)
    db.QueryRow("SELECT COUNT(*) FROM secret_question_sets").Scan(&setCount)

    if questionCount > 0 || setCount > 0 {
        log.Printf("WARNING: Dropping %d questions and %d question sets", questionCount, setCount)
    }

    // Drop tables
    _, err = db.Exec(`
        DROP TABLE IF EXISTS secret_questions;
        DROP TABLE IF EXISTS secret_question_sets;
    `)

    if err != nil {
        return fmt.Errorf("failed to drop tables: %w", err)
    }

    log.Println("Successfully dropped secret questions tables")
    return nil
}
```

### Communication

If this is a public project with users:
- Announce feature removal in changelog
- Warn about data loss if questions are in use
- Provide migration window if needed

## Acceptance Criteria

- [ ] All crypto files deleted (timelock.go, shamir.go, tests)
- [ ] All database layer files deleted
- [ ] All web handler files deleted
- [ ] Template file deleted
- [ ] Scheduler task removed
- [ ] Routes removed from server
- [ ] Interface methods removed from storage
- [ ] Models file deleted
- [ ] Shamir dependency removed from go.mod
- [ ] Migration created to drop tables
- [ ] Documentation updated (todo.md, PROJECT_OVERVIEW.md)
- [ ] TASK-004 deleted
- [ ] README.md updated (TASK-004 removed)
- [ ] All tests pass: `go test ./...`
- [ ] Application builds: `go build ./cmd/server`
- [ ] Application starts without errors
- [ ] No import errors or broken references
- [ ] Questions page returns 404
- [ ] No broken links in UI

## Testing Requirements

### 1. Compilation Tests
```bash
# Verify no import errors
go build ./...

# Verify tests compile
go test -c ./...

# Run all tests
go test ./...
```

### 2. Runtime Tests
```bash
# Start application
./deadmanswitch

# Verify scheduler starts without errors (no ReencryptQuestionsTask)
# Check logs for any panics or errors
```

### 3. Database Migration Test
```bash
# On fresh database
# Verify migration doesn't fail when tables don't exist

# On database with existing tables
# Create test secret_questions tables
# Run migration
# Verify tables are dropped
```

### 4. Dependency Check
```bash
# Verify shamir is not in dependencies
go mod graph | grep shamir  # Should return nothing

# Verify no unused dependencies
go mod tidy
git diff go.mod go.sum  # Should show only shamir removal
```

### 5. UI Testing
- Access `/recipients/:id/questions` - should 404
- Check all navigation - no links to questions
- Recipient pages should work normally

## Files Summary

### Files to DELETE (11 files, ~2,500 lines):
1. `/internal/crypto/timelock.go`
2. `/internal/crypto/timelock_test.go`
3. `/internal/crypto/shamir.go`
4. `/internal/crypto/shamir_test.go`
5. `/internal/models/secret_questions.go`
6. `/internal/storage/secret_questions.go`
7. `/internal/storage/mock_secret_questions.go`
8. `/internal/storage/migrations/add_secret_questions.go`
9. `/internal/scheduler/reencrypt_questions.go`
10. `/internal/web/handlers/secret_questions.go`
11. `/web/templates/questions.html`
12. `/tasks/TASK-004-complete-secret-questions-implementation.md`

### Files to MODIFY (6 files):
1. `/internal/storage/storage.go` - Remove 12 interface methods
2. `/internal/scheduler/scheduler.go` - Remove task registration (7 lines)
3. `/internal/web/server.go` - Remove handler field, init, and routes (~15 lines)
4. `/go.mod` - Remove shamir dependency
5. `/todo.md` - Remove secret questions TODO (5 lines)
6. `/docs/PROJECT_OVERVIEW.md` - Remove all mentions of timelock/Shamir
7. `/tasks/README.md` - Remove TASK-004 references, update statistics

### Files to CREATE (1 file):
1. `/internal/storage/migrations/drop_secret_questions.go` - Migration to drop tables

### Optional (Historical Reference):
- `/docs/drafts/` - Keep or delete based on preference

## Benefits of Removal

1. **Reduced Complexity** - ~2,500 fewer lines of code
2. **Simpler Maintenance** - No complex crypto to maintain
3. **Fewer Dependencies** - Remove shamir library
4. **Clearer Focus** - Core dead man's switch functionality only
5. **Easier Testing** - Fewer edge cases and components
6. **Less Cognitive Load** - Simpler mental model
7. **Faster Development** - Less code to understand

## Risks

1. **Breaking Changes** - Users with existing question data will lose it
   - **Mitigation:** Add migration with warning logging

2. **Documentation Drift** - Some docs may still reference feature
   - **Mitigation:** Thorough search and replace

3. **Hidden Dependencies** - Code elsewhere might reference this
   - **Mitigation:** Use compiler to find all references

## Search Commands for Verification

After removal, verify no references remain:

```bash
# Search for model references
rg "SecretQuestion" --type go

# Search for crypto references
rg "timelock|Timelock|TimeLock" --type go -i
rg "shamir|Shamir" --type go -i
rg "drand" --type go -i

# Search for imports
rg "crypto/timelock|crypto/shamir" --type go

# Search in templates
rg "question" web/templates/ -i

# Search in docs
rg "timelock|shamir|drand|secret.*question" docs/ -i
```

All searches should return zero results (except in drafts if kept).

## References

- Timelock implementation: `/internal/crypto/timelock.go`
- Shamir implementation: `/internal/crypto/shamir.go`
- Handler: `/internal/web/handlers/secret_questions.go`
- Scheduler task: `/internal/scheduler/reencrypt_questions.go`
- Original TODO: `/todo.md:18-22`
- Related task: `/tasks/TASK-004-complete-secret-questions-implementation.md`
- Design docs: `/docs/drafts/Implementation_Notes.md`

## Estimated Effort

- **Complexity:** Low-Medium (straightforward deletion, main risk is finding all references)
- **Time:** 3-4 hours
- **Risk:** Low (removal is safer than addition)
- **Lines Removed:** ~2,500
- **Files Deleted:** 12
- **Dependencies Removed:** 1

## Dependencies

**Blocks:**
- This removal invalidates TASK-004 (which should be deleted)

**Blocked by:**
- None - can be done independently

## Follow-up Tasks

After removal:
- Update README to clarify feature scope
- Consider if simpler recipient verification is needed
- Update any architecture diagrams
- Review if any other unwanted features exist

---

**Note:** This is a large deletion. Consider doing it in a dedicated PR for easy review and potential rollback if needed.
