# TASK-005: Consolidate Hardcoded Time and Duration Constants

## Priority: LOW 🟢

## Status: Not Started

## Category: Code Quality / Configuration

## Description

There are multiple time and duration-related constants hardcoded throughout the codebase. These should be either made configurable via environment variables or at minimum defined once as named constants for maintainability and clarity.

## Problem

**Current State:**
From `/todo.md:24`:
> There are multiple time/duration related constants hardcoded, make them configurable, and if not necessary to have it configurable at least, make it defined once as a constant with a good name.

**Examples of Hardcoded Durations:**
Common patterns to search for:
- `24 * time.Hour`, `7 * 24 * time.Hour`, `30 * 24 * time.Hour`
- `time.Minute * 15`, `time.Hour * 2`
- Magic numbers in context of time: `86400`, `3600`, `60`
- Session timeouts, cleanup intervals, retry delays, etc.

**Why This is Important:**
1. **Maintainability** - Changing durations requires finding all occurrences
2. **Consistency** - Same durations may be hardcoded differently in different places
3. **Testability** - Hard to test time-dependent code with fixed durations
4. **Configuration** - Some durations should be user-configurable
5. **Documentation** - Named constants are self-documenting

## Proposed Solution

### Step 1: Audit All Time Constants

Search the codebase for hardcoded time values and categorize them:

**A. Should be configurable** (via environment variables)
**B. Should be constants** (fixed but named)
**C. Are acceptable as-is** (truly one-off or test values)

### Step 2: Create Time Constants File

**New file:** `internal/constants/time.go`

```go
package constants

import "time"

// Session and Authentication
const (
    SessionDuration       = 30 * 24 * time.Hour // 30 days
    SessionCleanupInterval = 1 * time.Hour       // Cleanup old sessions hourly
    TwoFactorCodeValidity  = 5 * time.Minute     // TOTP code valid window
    PasskeyTimeout         = 5 * time.Minute     // WebAuthn ceremony timeout
)

// Ping and Deadline System
const (
    // These are defaults; actual values come from config
    DefaultPingFrequency = 3 * 24 * time.Hour  // 3 days
    DefaultPingDeadline  = 14 * 24 * time.Hour // 14 days

    // Reminder escalation thresholds
    UrgentReminderThreshold      = 3 * 24 * time.Hour // "Urgent" if < 3 days to deadline
    FinalWarningThreshold        = 1 * 24 * time.Hour // "Final" if < 1 day to deadline
)

// Activity Monitoring
const (
    ActivityCheckInterval    = 1 * time.Hour      // Check external activity hourly
    GitHubActivityWindow     = 7 * 24 * time.Hour // Look back 7 days for activity
    ActivityProviderTimeout  = 30 * time.Second   // HTTP timeout for API calls
)

// Secret Questions and Timelock
const (
    QuestionReencryptMargin  = 7 * 24 * time.Hour  // Re-encrypt 7 days before deadline
    QuestionReencryptInterval = 1 * time.Hour       // Check for re-encryption hourly
    TimelockComputeTimeout    = 10 * time.Second    // Max time for timelock computation
)

// Access Codes and Delivery
const (
    AccessCodeValidity       = 30 * 24 * time.Hour // 30 days
    AccessCodeCleanupInterval = 24 * time.Hour      // Daily cleanup
    DeliveryRetryDelay        = 5 * time.Minute     // Retry failed deliveries after 5min
    DeliveryMaxRetries        = 3                   // Max retry attempts
)

// Telegram and Notifications
const (
    TelegramPingTimeout      = 30 * time.Second    // Timeout for Telegram API calls
    TelegramVerifyExpiry     = 24 * time.Hour      // Verification code expiry
    EmailSendTimeout          = 30 * time.Second    // SMTP timeout
)

// Scheduler Intervals
const (
    PingTaskInterval         = 15 * time.Minute    // Check for pending pings
    DeadlineCheckInterval    = 5 * time.Minute     // Check for expired deadlines
    CleanupTaskInterval      = 24 * time.Hour      // Daily cleanup tasks
)

// Rate Limiting and Security
const (
    LoginAttemptWindow       = 15 * time.Minute    // Rate limit window
    MaxLoginAttempts         = 5                   // Max attempts per window
    IPBlockDuration          = 1 * time.Hour       // Block duration after max attempts
    AuditLogRetention        = 90 * 24 * time.Hour // Keep audit logs for 90 days
)
```

### Step 3: Update Configuration for Configurable Durations

**File to modify:** `internal/config/config.go`

```go
type Config struct {
    // ... existing fields ...

    // Session configuration
    SessionDuration       time.Duration
    SessionCleanupInterval time.Duration

    // Access code configuration
    AccessCodeExpiration  time.Duration
    AccessCodeMaxAttempts int

    // Activity monitoring
    ActivityCheckInterval time.Duration
    GitHubActivityWindow  time.Duration

    // Question re-encryption
    QuestionReencryptMargin time.Duration

    // Scheduler intervals
    PingTaskInterval      time.Duration
    DeadlineCheckInterval time.Duration
}

func LoadFromEnv() (*Config, error) {
    config := &Config{}

    // Load with defaults from constants
    config.SessionDuration = getDurationEnv("SESSION_DURATION", constants.SessionDuration)
    config.AccessCodeExpiration = getDurationEnv("ACCESS_CODE_EXPIRATION", constants.AccessCodeValidity)
    config.ActivityCheckInterval = getDurationEnv("ACTIVITY_CHECK_INTERVAL", constants.ActivityCheckInterval)
    // ... etc

    return config, nil
}

// Helper to get duration from env with default
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }

    duration, err := time.ParseDuration(value)
    if err != nil {
        log.Printf("Invalid duration for %s: %v, using default", key, err)
        return defaultValue
    }

    return duration
}
```

### Step 4: Replace Hardcoded Values Throughout Codebase

Search and replace hardcoded durations with constants or config values.

**Example Replacements:**

```go
// Before:
ticker := time.NewTicker(15 * time.Minute)

// After:
ticker := time.NewTicker(constants.PingTaskInterval)
```

```go
// Before:
expiresAt := time.Now().Add(30 * 24 * time.Hour)

// After:
expiresAt := time.Now().Add(h.config.AccessCodeExpiration)
```

```go
// Before:
if timeSinceActivity < 7 * 24 * time.Hour {

// After:
if timeSinceActivity < constants.GitHubActivityWindow {
```

## Files Likely to Contain Hardcoded Durations

Based on codebase analysis:

1. `/internal/scheduler/scheduler.go` - Intervals, timeouts
2. `/internal/web/middleware/auth.go` - Session durations
3. `/internal/telegram/bot.go` - Telegram timeouts
4. `/internal/email/client.go` - SMTP timeouts
5. `/internal/activity/github.go` - Activity window
6. `/internal/web/handlers/*.go` - Various timeouts
7. `/internal/storage/sqlite/sqlite.go` - Cleanup intervals

## Acceptance Criteria

- [ ] Constants file created: `/internal/constants/time.go`
- [ ] All hardcoded time durations identified and categorized
- [ ] Configurable durations added to `Config` struct
- [ ] Environment variable loading implemented with defaults
- [ ] All hardcoded durations in production code replaced
- [ ] `.env.example` updated with new time configuration options
- [ ] Documentation updated explaining time configuration
- [ ] Tests updated to use constants (or mock time)
- [ ] No magic number durations remain in production code

## Testing Requirements

1. **Unit Tests:**
   - Test config loading with custom durations
   - Test config loading with invalid durations (use defaults)
   - Test config loading with missing env vars (use defaults)

2. **Regression Tests:**
   - Ensure all existing functionality works with constants
   - Verify default values match previous hardcoded values
   - Test that configurable values can be changed

3. **Manual Testing:**
   - Test with custom configuration values
   - Verify scheduler intervals respect config
   - Verify session timeouts respect config

## Implementation Plan

### Phase 1: Discovery (1-2 hours)
1. Search codebase for time-related patterns:
   ```bash
   rg "time\.(Hour|Minute|Second|Day)" --type go
   rg "\d+\s*\*\s*time\." --type go
   rg "time\.Duration\(\d+\)" --type go
   ```

2. Create spreadsheet of all findings with:
   - File and line number
   - Current hardcoded value
   - Context/purpose
   - Recommendation (config/constant/keep)

### Phase 2: Create Constants (1 hour)
1. Create `/internal/constants/time.go`
2. Define all identified constants
3. Add comprehensive documentation

### Phase 3: Update Configuration (1 hour)
1. Add fields to `Config` struct
2. Implement environment variable loading
3. Update `.env.example`

### Phase 4: Replace Usage (2-3 hours)
1. Replace hardcoded values file by file
2. Update imports to include constants package
3. Update tests

### Phase 5: Documentation (1 hour)
1. Update README with configuration options
2. Document default values
3. Add migration guide for existing deployments

## Documentation Requirements

### README.md
Add section on time configuration:
```markdown
## Time Configuration

The application uses configurable time intervals for various operations:

| Variable | Default | Description |
|----------|---------|-------------|
| SESSION_DURATION | 720h (30 days) | User session validity |
| ACCESS_CODE_EXPIRATION | 720h (30 days) | Access code validity |
| ACTIVITY_CHECK_INTERVAL | 1h | External activity check frequency |
| PING_TASK_INTERVAL | 15m | Ping reminder check frequency |

All duration values use Go duration format: `30s`, `5m`, `2h`, `24h`, etc.
```

### .env.example
```bash
# Time Configuration (all values use Go duration format: 30s, 5m, 2h, 24h)
SESSION_DURATION=720h
ACCESS_CODE_EXPIRATION=720h
ACTIVITY_CHECK_INTERVAL=1h
PING_TASK_INTERVAL=15m
DEADLINE_CHECK_INTERVAL=5m
```

## Migration Guide

For existing deployments, no action is required unless you want to customize time values. All defaults match previous hardcoded values.

To customize:
1. Add desired duration variables to `.env`
2. Use Go duration format: `1h30m`, `24h`, `7d` (if supported)
3. Restart application

## Benefits

1. **Flexibility** - Operators can tune timing without code changes
2. **Testing** - Easier to test time-dependent code with configurable durations
3. **Clarity** - Named constants are self-documenting
4. **Maintenance** - Single source of truth for each duration
5. **Consistency** - Same duration used everywhere it's needed

## References

- TODO item: `/todo.md:24`
- Scheduler code: `/internal/scheduler/scheduler.go`
- Configuration: `/internal/config/config.go`

## Estimated Effort

- **Complexity:** Low-Medium
- **Time:** 6-8 hours
- **Risk:** Low (mostly refactoring)

## Dependencies

None - can be done independently

## Follow-up Tasks

- Consider adding duration validation ranges
- Add metrics for configured vs actual durations
- Create admin UI for duration configuration
- Consider dynamic duration adjustment based on system load
