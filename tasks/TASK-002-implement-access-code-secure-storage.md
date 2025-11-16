# TASK-002: Implement Secure Access Code Storage with TTL

## Priority: HIGH 🟠

## Status: Not Started

## Category: Security / Secret Delivery

## Description

When the dead man's switch is triggered and secrets are delivered to recipients, the system generates access codes for recipients to retrieve their secrets. Currently, these access codes are generated but not stored securely, and there's no TTL (Time To Live) mechanism.

## Problem

**Current State:**
- Access codes are generated in `internal/scheduler/scheduler.go:490`
- A TODO comment exists at line 492: `// TODO: Store access code securely with TTL`
- Access codes are sent via email but not persisted
- No expiration mechanism
- No way to verify access codes later
- No protection against brute force attempts

**Current Code:**
```go
// internal/scheduler/scheduler.go:490-494
// Generate access code for the secrets
accessCode := generateAccessCode()

// TODO: Store access code securely with TTL
// For now, this is a simplified version - in a production system,
// we would store this code securely and set an expiration
```

**Why This is Important:**
1. Recipients need a way to access their delivered secrets securely
2. Access codes should expire to limit exposure window
3. Need to track which codes have been used
4. Need rate limiting to prevent brute force attacks
5. Audit trail for security compliance

## Proposed Solution

### 1. Create AccessCode Database Model

**New file:** `internal/models/access_code.go`

```go
package models

import "time"

// AccessCode represents a time-limited access code for secret delivery
type AccessCode struct {
    ID              string    `json:"id"`
    Code            string    `json:"code"`                // Hashed access code
    RecipientID     string    `json:"recipient_id"`
    UserID          string    `json:"user_id"`
    DeliveryEventID string    `json:"delivery_event_id"`
    CreatedAt       time.Time `json:"created_at"`
    ExpiresAt       time.Time `json:"expires_at"`
    UsedAt          *time.Time `json:"used_at,omitempty"`  // NULL if not used yet
    AttemptCount    int       `json:"attempt_count"`       // Track failed attempts
    MaxAttempts     int       `json:"max_attempts"`        // Default: 5
}
```

### 2. Add Database Migration

**File to modify:** `internal/storage/sqlite/migrations.go`

```sql
CREATE TABLE IF NOT EXISTS access_codes (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL,  -- Store hashed version
    recipient_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    delivery_event_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    attempt_count INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 5,
    FOREIGN KEY (recipient_id) REFERENCES recipients(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (delivery_event_id) REFERENCES delivery_events(id) ON DELETE CASCADE
);

CREATE INDEX idx_access_codes_code ON access_codes(code);
CREATE INDEX idx_access_codes_expires_at ON access_codes(expires_at);
CREATE INDEX idx_access_codes_recipient_id ON access_codes(recipient_id);
```

### 3. Add Repository Methods

**File to modify:** `internal/storage/repository.go`

Add interface methods:
```go
type Repository interface {
    // ... existing methods ...

    // Access code methods
    CreateAccessCode(ctx context.Context, code *models.AccessCode) error
    GetAccessCodeByCode(ctx context.Context, code string) (*models.AccessCode, error)
    VerifyAccessCode(ctx context.Context, code string) (*models.AccessCode, error)
    MarkAccessCodeAsUsed(ctx context.Context, id string) error
    IncrementAccessCodeAttempts(ctx context.Context, id string) error
    DeleteExpiredAccessCodes(ctx context.Context) error
}
```

**File to modify:** `internal/storage/sqlite/sqlite.go`

Implement the methods with:
- Hash storage (not plaintext codes)
- Expiration checks
- Attempt counting
- Thread-safe operations

### 4. Update Scheduler to Store Access Codes

**File to modify:** `internal/scheduler/scheduler.go`

Replace the TODO section (lines 492-494) with:

```go
// Generate and store access code securely
accessCode := generateAccessCode()
accessCodeHash := hashAccessCode(accessCode) // Use Argon2id

codeModel := &models.AccessCode{
    ID:              uuid.New().String(),
    Code:            accessCodeHash,
    RecipientID:     recipient.ID,
    UserID:          user.ID,
    DeliveryEventID: deliveryEvent.ID,
    CreatedAt:       time.Now(),
    ExpiresAt:       time.Now().Add(30 * 24 * time.Hour), // 30 days expiration
    MaxAttempts:     5,
}

if err := s.repo.CreateAccessCode(ctx, codeModel); err != nil {
    log.Printf("Failed to store access code for recipient %s: %v", recipient.ID, err)
    continue
}
```

### 5. Create Access Code Verification Handler

**New file:** `internal/web/handlers/access.go`

```go
package handlers

// HandleAccessSecrets verifies access code and shows secrets
func (h *Handler) HandleAccessSecrets(w http.ResponseWriter, r *http.Request) {
    code := r.FormValue("access_code")

    // Verify code (checks expiration, attempts, etc.)
    accessCode, err := h.repo.VerifyAccessCode(r.Context(), code)
    if err != nil {
        // Increment failed attempts
        // Show error message
        return
    }

    // Mark as used
    h.repo.MarkAccessCodeAsUsed(r.Context(), accessCode.ID)

    // Retrieve and display secrets
    // ...
}
```

### 6. Add Cleanup Task to Scheduler

**File to modify:** `internal/scheduler/scheduler.go`

Add a new scheduled task to clean up expired access codes:

```go
// Add to Start() method
go func() {
    ticker := time.NewTicker(24 * time.Hour) // Daily cleanup
    defer ticker.Stop()

    for range ticker.C {
        if err := s.cleanupExpiredAccessCodes(ctx); err != nil {
            log.Printf("Failed to cleanup expired access codes: %v", err)
        }
    }
}()

// New method
func (s *Scheduler) cleanupExpiredAccessCodes(ctx context.Context) error {
    return s.repo.DeleteExpiredAccessCodes(ctx)
}
```

## Configuration

Add new environment variables:

```bash
# Access code settings
ACCESS_CODE_EXPIRATION_DAYS=30  # How long codes remain valid
ACCESS_CODE_MAX_ATTEMPTS=5      # Maximum failed attempts before lockout
```

**File to modify:** `internal/config/config.go`

## Acceptance Criteria

- [ ] `AccessCode` model created with all required fields
- [ ] Database migration added for `access_codes` table
- [ ] Repository interface extended with access code methods
- [ ] Repository implementation completed with hash storage
- [ ] Scheduler updated to store access codes (TODO removed)
- [ ] Access code verification handler created
- [ ] Rate limiting implemented (max attempts)
- [ ] Expiration mechanism implemented
- [ ] Cleanup task added to scheduler
- [ ] Configuration options added
- [ ] Unit tests for all repository methods
- [ ] Integration tests for access code flow
- [ ] Documentation updated

## Testing Requirements

1. **Unit Tests:**
   - Test access code creation
   - Test access code verification (valid)
   - Test access code verification (expired)
   - Test access code verification (max attempts exceeded)
   - Test access code hashing
   - Test cleanup of expired codes

2. **Integration Tests:**
   - Test full delivery flow with access code
   - Test recipient accessing secrets with code
   - Test expired code rejection
   - Test brute force protection

3. **Manual Testing:**
   - Trigger dead man's switch
   - Verify email contains access code
   - Access secrets with code
   - Verify code expires after configured time
   - Test max attempts protection

## Files to Create/Modify

**New Files:**
1. `/internal/models/access_code.go` - AccessCode model
2. `/internal/web/handlers/access.go` - Access code verification handler
3. `/web/templates/access.html` - Access code entry page
4. `/web/templates/access-secrets.html` - Secrets display for recipients

**Files to Modify:**
1. `/internal/storage/repository.go` - Add interface methods
2. `/internal/storage/sqlite/sqlite.go` - Implement methods
3. `/internal/storage/sqlite/migrations.go` - Add migration
4. `/internal/scheduler/scheduler.go` - Remove TODO, implement storage (line 492)
5. `/internal/config/config.go` - Add access code config options
6. `/internal/web/server.go` - Add access routes
7. `/.env.example` - Document new variables

**Test Files:**
1. `/internal/storage/sqlite/access_code_test.go`
2. `/tests/integration/access_code_test.go`

## Security Considerations

1. **Hash Access Codes:**
   - Never store plaintext codes
   - Use Argon2id with same parameters as passwords
   - Store salt with hash

2. **Rate Limiting:**
   - Limit failed verification attempts
   - Consider IP-based rate limiting
   - Lock codes after max attempts

3. **Expiration:**
   - Default 30 days (configurable)
   - Clearly communicate expiration in email
   - Warning before expiration (optional enhancement)

4. **Audit Logging:**
   - Log all access code verifications (success/failure)
   - Log when codes are used
   - Log cleanup operations

## References

- Current TODO: `/internal/scheduler/scheduler.go:492`
- Access code generation: `/internal/scheduler/scheduler.go:490`
- Delivery event model: `/internal/models/delivery_event.go`
- Related security docs: `/docs/security.md`

## Estimated Effort

- **Complexity:** Medium-High
- **Time:** 6-8 hours
- **Risk:** Medium

## Dependencies

- Must complete after TASK-001 (master key management)
- Should coordinate with email template updates

## Follow-up Tasks

- Add access code expiration warnings via email
- Implement code resend functionality
- Add admin interface to revoke codes
- Consider 2FA for high-security access codes
