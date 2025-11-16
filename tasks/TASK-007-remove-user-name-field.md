# TASK-007: Remove User Name Field from Application

## Priority: LOW 🟢

## Status: Not Started

## Category: Code Cleanup / Database Schema

## Description

The User model includes a `Name` field that is not actually used or needed by the application. This field should be removed to simplify the schema and reduce unnecessary data collection.

## Problem

**Current State:**
From `/todo.md:23`:
> Remove user name everywhere from webUI and database as we are not going to use it

**Current Usage:**
The `Name` field exists in:
- User model: `/internal/models/user.go`
- Database schema: `users` table
- Registration forms
- Profile pages
- Email templates (e.g., "Hi {{.Name}}")
- Telegram messages

**Why Remove It:**
1. **Unnecessary data collection** - Minimizes PII (Personally Identifiable Information)
2. **Privacy** - Less personal data to protect
3. **Simplicity** - One less field to validate and maintain
4. **Not used** - Application doesn't need user's real name for functionality
5. **GDPR compliance** - Collect only necessary data

## Proposed Solution

### 1. Update Database Schema

**File to modify:** `/internal/storage/sqlite/migrations.go`

Add migration to remove name column:

```go
{
    Version: 14, // Next version number
    Up: `
        -- Create new table without name column
        CREATE TABLE users_new (
            id TEXT PRIMARY KEY,
            email TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            totp_secret TEXT,
            totp_enabled INTEGER DEFAULT 0,
            telegram_id TEXT,
            telegram_username TEXT,
            github_username TEXT,
            pinging_enabled INTEGER DEFAULT 0,
            ping_frequency INTEGER DEFAULT 3,
            deadline INTEGER DEFAULT 14,
            notification_method TEXT DEFAULT 'email',
            last_activity TIMESTAMP,
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL
        );

        -- Copy data from old table
        INSERT INTO users_new
        SELECT id, email, password_hash, totp_secret, totp_enabled,
               telegram_id, telegram_username, github_username,
               pinging_enabled, ping_frequency, deadline,
               notification_method, last_activity, created_at, updated_at
        FROM users;

        -- Drop old table
        DROP TABLE users;

        -- Rename new table
        ALTER TABLE users_new RENAME TO users;

        -- Recreate indexes
        CREATE UNIQUE INDEX idx_users_email ON users(email);
        CREATE INDEX idx_users_telegram_id ON users(telegram_id);
    `,
    Down: `
        -- Add name column back (for rollback)
        ALTER TABLE users ADD COLUMN name TEXT;
    `,
},
```

### 2. Update User Model

**File to modify:** `/internal/models/user.go`

```go
type User struct {
    ID                 string    `json:"id"`
    Email              string    `json:"email"`
    // Name            string    `json:"name"` // REMOVED
    PasswordHash       []byte    `json:"-"`
    TOTPSecret         string    `json:"-"`
    TOTPEnabled        bool      `json:"totp_enabled"`
    // ... rest of fields
}
```

### 3. Remove from Registration Handler

**File to modify:** `/internal/web/handlers/auth.go`

```go
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
    email := r.FormValue("email")
    password := r.FormValue("password")
    // name := r.FormValue("name") // REMOVED

    user := &models.User{
        ID:    uuid.New().String(),
        Email: email,
        // Name:  name, // REMOVED
        // ...
    }
}
```

### 4. Remove from Templates

**Files to modify:**
- `/web/templates/register.html` - Remove name input field
- `/web/templates/profile.html` - Remove name display/edit
- `/web/templates/settings.html` - Remove name if displayed

**Example removal from register.html:**
```html
<!-- REMOVE THIS:
<div class="form-group">
    <label for="name">Full Name</label>
    <input type="text" name="name" id="name" required>
</div>
-->
```

### 5. Update Email Templates

Since emails currently use "Hi {{.Name}}", we need alternatives:

**Option A:** Use email username
```
Hi {{.EmailUsername}},  // "Hi john" from "john@example.com"
```

**Option B:** Use generic greeting
```
Hello,  // No personalization
```

**Option C:** Use "there"
```
Hi there,
```

**Recommendation:** Option B (generic) for privacy, or Option A if some personalization desired.

**Files to update:**
- All email templates (will be in `/internal/email/templates/` after TASK-006)
- Or update email client if still using code-based templates

### 6. Update Telegram Bot Messages

**File to modify:** `/internal/telegram/bot.go`

```go
// Before:
message := fmt.Sprintf("Hi %s, time for your check-in!", user.Name)

// After:
message := "Hi, time for your check-in!"
// Or use email username if available
```

### 7. Update Repository Methods

**Files to check:**
- `/internal/storage/repository.go` - Update interface if needed
- `/internal/storage/sqlite/sqlite.go` - Update queries

Remove `name` from all INSERT and UPDATE queries:

```go
// Before:
INSERT INTO users (id, email, name, ...) VALUES (?, ?, ?, ...)

// After:
INSERT INTO users (id, email, ...) VALUES (?, ?, ...)
```

### 8. Update Tests

Search for `Name` field usage in tests and update:

```bash
rg "\.Name\s*=" --type go internal/
rg "user\.Name" --type go internal/
```

### 9. Update API Responses

If any API endpoints return user data with name field, update to remove it or mark as deprecated.

## Migration Considerations

**Data Migration:**
- The migration will drop the `name` column
- **No data backup needed** - we're intentionally discarding this data
- Consider logging names before deletion if reversal might be needed

**Rollback Plan:**
- Migration includes `Down` to re-add column
- Column will be empty after rollback

**Communication:**
- If users have entered names, inform them data will be deleted
- Update privacy policy if name was mentioned

## Acceptance Criteria

- [ ] Database migration created to remove `name` column
- [ ] User model updated (field removed)
- [ ] Registration form updated (no name input)
- [ ] Profile page updated (no name display)
- [ ] Email templates updated (no name personalization)
- [ ] Telegram messages updated (no name usage)
- [ ] Repository queries updated
- [ ] All tests updated and passing
- [ ] Migration tested (both up and down)
- [ ] No references to `user.Name` in codebase
- [ ] Documentation updated

## Testing Requirements

1. **Migration Tests:**
   - Test migration up (removes column successfully)
   - Test migration down (adds column back)
   - Test data preservation (other fields intact)

2. **Unit Tests:**
   - Update user creation tests
   - Update user model tests
   - Test email rendering without name

3. **Integration Tests:**
   - Test registration without name
   - Test login flow
   - Test email sending
   - Test Telegram messages

4. **Manual Testing:**
   - Register new user (no name field)
   - Verify emails don't use name
   - Verify Telegram messages don't use name
   - Check profile page (no name)

## Files to Modify

**Database:**
1. `/internal/storage/sqlite/migrations.go` - Add migration

**Models:**
2. `/internal/models/user.go` - Remove Name field

**Handlers:**
3. `/internal/web/handlers/auth.go` - Remove name from registration
4. `/internal/web/handlers/profile.go` - Remove name from profile

**Templates:**
5. `/web/templates/register.html` - Remove name input
6. `/web/templates/profile.html` - Remove name display
7. Email templates (all) - Remove name usage

**Bot:**
8. `/internal/telegram/bot.go` - Remove name from messages

**Storage:**
9. `/internal/storage/sqlite/sqlite.go` - Update queries

**Tests:**
10. Various test files - Update to not use Name field

## Search Commands for Finding All Usages

```bash
# Find all Name field assignments
rg "\.Name\s*=" --type go

# Find all Name field access
rg "user\.Name|User\.Name" --type go

# Find in templates
rg "\.Name" --type html web/templates/

# Find in SQL
rg "name" internal/storage/sqlite/ | grep -i "users"
```

## Alternative: Keep Name as Optional

If there's a reason to keep the name field (e.g., for future use or user preference), consider:

1. Make it optional and nullable in database
2. Don't show it in UI by default
3. Don't use it in communications
4. Add note in privacy policy that it's optional

However, the TODO explicitly says to remove it, so full removal is recommended.

## Privacy Benefits

1. **Minimal PII** - Reduces personal data stored
2. **Reduced attack surface** - Less data to leak if breached
3. **GDPR compliance** - Only collect necessary data
4. **User privacy** - No real name required for functionality

## References

- TODO item: `/todo.md:23`
- User model: `/internal/models/user.go`
- Database schema: `/internal/storage/sqlite/migrations.go`

## Estimated Effort

- **Complexity:** Low-Medium
- **Time:** 3-4 hours
- **Risk:** Low (straightforward field removal)

## Dependencies

None - can be done independently

## Follow-up Tasks

- Consider removing other unnecessary fields
- Audit all PII collection
- Update privacy policy to reflect minimal data collection
- Add data minimization documentation
