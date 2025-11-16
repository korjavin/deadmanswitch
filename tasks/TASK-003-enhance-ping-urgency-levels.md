# TASK-003: Enhance Ping Messages with Urgency Levels

## Priority: MEDIUM 🟡

## Status: Not Started

## Category: Notifications / User Experience

## Description

The scheduler has a reminder escalation system that categorizes ping reminders based on proximity to deadline (normal, urgent, final warning). However, the email and Telegram interfaces don't support passing urgency levels, so all pings appear the same to users regardless of how close they are to the deadline.

## Problem

**Current State:**
- Scheduler calculates urgency levels in `internal/scheduler/scheduler.go`
- Two TODO comments indicate missing functionality:
  - Line 690: `// TODO: Enhance the SendPingMessage interface to include urgency level`
  - Line 729: `// TODO: Enhance the SendPingEmail interface to include urgency level`
- Emails and Telegram messages are identical for all urgency levels
- Users have no visual indication of approaching deadline urgency

**Current Code:**
```go
// internal/scheduler/scheduler.go:686-691
if user.NotificationMethod == models.NotificationMethodTelegram ||
    user.NotificationMethod == models.NotificationMethodBoth {
    if err := s.telegramBot.SendPingMessage(ctx, user, ping.ID); err != nil {
        log.Printf("Failed to send Telegram ping to user %s: %v", user.ID, err)
    }
    // TODO: Enhance the SendPingMessage interface to include urgency level
}

// internal/scheduler/scheduler.go:725-730
if user.NotificationMethod == models.NotificationMethodEmail ||
    user.NotificationMethod == models.NotificationMethodBoth {
    if err := s.emailClient.SendPingEmail(user.Email, user.Name, verificationCode); err != nil {
        log.Printf("Failed to send ping email to user %s: %v", user.ID, err)
    }
    // TODO: Enhance the SendPingEmail interface to include urgency level
}
```

**Current Urgency Calculation:**
The scheduler already calculates urgency in `getReminderUrgency()`:
- **Normal**: > 3 days from deadline
- **Urgent**: 1-3 days from deadline
- **FinalWarning**: < 1 day from deadline

## Proposed Solution

### 1. Define Urgency Level Type

**File to modify:** `internal/models/notification.go` (create if doesn't exist)

```go
package models

// ReminderUrgency represents the urgency level of a ping reminder
type ReminderUrgency string

const (
    ReminderNormal       ReminderUrgency = "normal"
    ReminderUrgent       ReminderUrgency = "urgent"
    ReminderFinalWarning ReminderUrgency = "final_warning"
)
```

### 2. Update Email Client Interface

**File to modify:** `internal/email/client.go`

Update method signature:
```go
// Old
SendPingEmail(email, name, verificationCode string) error

// New
SendPingEmail(email, name, verificationCode string, urgency models.ReminderUrgency) error
```

Implement urgency-based templates:
```go
func (c *Client) SendPingEmail(email, name, code string, urgency models.ReminderUrgency) error {
    subject := c.getSubjectByUrgency(urgency)
    body := c.getBodyByUrgency(name, code, urgency)

    return c.SendEmail(&MessageOptions{
        To:      []string{email},
        Subject: subject,
        Body:    body,
        IsHTML:  true,
    })
}

func (c *Client) getSubjectByUrgency(urgency models.ReminderUrgency) string {
    switch urgency {
    case models.ReminderFinalWarning:
        return "🚨 URGENT: Final Check-In Required - Dead Man's Switch"
    case models.ReminderUrgent:
        return "⚠️ IMPORTANT: Check-In Required Soon - Dead Man's Switch"
    default:
        return "✅ Routine Check-In - Dead Man's Switch"
    }
}

func (c *Client) getBodyByUrgency(name, code string, urgency models.ReminderUrgency) string {
    // Load different templates based on urgency
    // Include urgency-specific messaging and formatting
}
```

### 3. Update Telegram Bot Interface

**File to modify:** `internal/telegram/bot.go`

Update method signature:
```go
// Old
SendPingMessage(ctx context.Context, user *models.User, pingID string) error

// New
SendPingMessage(ctx context.Context, user *models.User, pingID string, urgency models.ReminderUrgency) error
```

Implement urgency-based messages:
```go
func (b *Bot) SendPingMessage(ctx context.Context, user *models.User, pingID string, urgency models.ReminderUrgency) error {
    message := b.getMessageByUrgency(user.Name, urgency)

    // Add urgency-specific emoji and formatting
    keyboard := b.getPingKeyboard(pingID)

    msg := tgbotapi.NewMessage(user.TelegramID, message)
    msg.ReplyMarkup = keyboard
    msg.ParseMode = "Markdown"

    _, err := b.api.Send(msg)
    return err
}

func (b *Bot) getMessageByUrgency(name string, urgency models.ReminderUrgency) string {
    switch urgency {
    case models.ReminderFinalWarning:
        return fmt.Sprintf(
            "🚨 *FINAL WARNING* 🚨\n\n" +
            "Hi %s,\n\n" +
            "This is your LAST check-in reminder before your Dead Man's Switch triggers!\n" +
            "You have less than 24 hours to respond.\n\n" +
            "Click 'I'm OK' to confirm you're active.",
            name,
        )
    case models.ReminderUrgent:
        return fmt.Sprintf(
            "⚠️ *URGENT CHECK-IN REQUIRED* ⚠️\n\n" +
            "Hi %s,\n\n" +
            "Your deadline is approaching soon (1-3 days).\n" +
            "Please confirm you're still active.\n\n" +
            "Click 'I'm OK' to check in.",
            name,
        )
    default:
        return fmt.Sprintf(
            "✅ *Routine Check-In*\n\n" +
            "Hi %s,\n\n" +
            "Time for your regular check-in!\n\n" +
            "Click 'I'm OK' to confirm you're active.",
            name,
        )
    }
}
```

### 4. Update Scheduler to Pass Urgency

**File to modify:** `internal/scheduler/scheduler.go`

Update lines 686-730 to pass urgency level:

```go
// Calculate urgency level
urgency := s.getReminderUrgency(user)

// Send Telegram ping with urgency
if user.NotificationMethod == models.NotificationMethodTelegram ||
    user.NotificationMethod == models.NotificationMethodBoth {
    if err := s.telegramBot.SendPingMessage(ctx, user, ping.ID, urgency); err != nil {
        log.Printf("Failed to send Telegram ping to user %s: %v", user.ID, err)
    }
}

// Send Email ping with urgency
if user.NotificationMethod == models.NotificationMethodEmail ||
    user.NotificationMethod == models.NotificationMethodBoth {
    if err := s.emailClient.SendPingEmail(user.Email, user.Name, verificationCode, urgency); err != nil {
        log.Printf("Failed to send ping email to user %s: %v", user.ID, err)
    }
}
```

### 5. Create Urgency-Specific Email Templates

**New files:**
- `/internal/email/templates/ping_normal.html`
- `/internal/email/templates/ping_urgent.html`
- `/internal/email/templates/ping_final.html`

Example final warning template:
```html
<!DOCTYPE html>
<html>
<head>
    <style>
        .urgent-banner {
            background-color: #dc3545;
            color: white;
            padding: 20px;
            text-align: center;
            font-size: 24px;
            font-weight: bold;
        }
        .warning-box {
            border: 3px solid #dc3545;
            padding: 20px;
            margin: 20px 0;
        }
        .cta-button {
            background-color: #dc3545;
            color: white;
            padding: 15px 30px;
            font-size: 18px;
            border-radius: 5px;
        }
    </style>
</head>
<body>
    <div class="urgent-banner">
        🚨 FINAL WARNING - IMMEDIATE ACTION REQUIRED 🚨
    </div>
    <div class="warning-box">
        <h2>Your Dead Man's Switch Will Trigger Soon!</h2>
        <p>You have <strong>less than 24 hours</strong> to respond.</p>
        <p>If you do not check in, your secrets will be automatically delivered to your designated recipients.</p>
    </div>
    <a href="{{.VerificationLink}}" class="cta-button">
        CHECK IN NOW
    </a>
</body>
</html>
```

## Acceptance Criteria

- [ ] `ReminderUrgency` type defined in models
- [ ] `SendPingEmail` signature updated with urgency parameter
- [ ] `SendPingMessage` signature updated with urgency parameter
- [ ] Scheduler passes urgency to both email and Telegram
- [ ] Three email templates created (normal, urgent, final)
- [ ] Email subjects vary by urgency level
- [ ] Telegram messages vary by urgency level
- [ ] Urgency-specific emoji and formatting applied
- [ ] All TODO comments removed (lines 690, 729)
- [ ] Tests updated for new signatures
- [ ] Documentation updated

## Testing Requirements

1. **Unit Tests:**
   - Test urgency calculation
   - Test email template selection by urgency
   - Test Telegram message formatting by urgency
   - Test subject line generation

2. **Integration Tests:**
   - Test ping sending with normal urgency
   - Test ping sending with urgent urgency
   - Test ping sending with final warning urgency
   - Verify correct template/message used for each

3. **Manual Testing:**
   - Trigger ping at different urgency levels
   - Verify email formatting and content
   - Verify Telegram message formatting
   - Check that urgency is visually clear

## Files to Create/Modify

**New Files:**
1. `/internal/models/notification.go` - ReminderUrgency type
2. `/internal/email/templates/ping_normal.html`
3. `/internal/email/templates/ping_urgent.html`
4. `/internal/email/templates/ping_final.html`

**Files to Modify:**
1. `/internal/email/client.go` - Update SendPingEmail signature
2. `/internal/telegram/bot.go` - Update SendPingMessage signature
3. `/internal/scheduler/scheduler.go` - Pass urgency (lines 686-730, remove TODOs)
4. `/internal/scheduler/scheduler_test.go` - Update tests for new signatures

## Design Mockups

### Email Subject Lines
- ✅ Normal: "✅ Routine Check-In - Dead Man's Switch"
- ⚠️ Urgent: "⚠️ IMPORTANT: Check-In Required Soon - Dead Man's Switch"
- 🚨 Final: "🚨 URGENT: Final Check-In Required - Dead Man's Switch"

### Telegram Messages
```
Normal:
✅ Routine Check-In
Hi John,
Time for your regular check-in!
[I'm OK button]

Urgent:
⚠️ URGENT CHECK-IN REQUIRED ⚠️
Hi John,
Your deadline is approaching soon (1-3 days).
[I'm OK button]

Final:
🚨 FINAL WARNING 🚨
Hi John,
This is your LAST check-in reminder!
Less than 24 hours remaining.
[I'm OK NOW button]
```

## User Experience Benefits

1. **Clear urgency communication** - Users immediately understand severity
2. **Reduced false triggers** - Users more likely to respond to urgent reminders
3. **Better engagement** - Varied messaging prevents notification fatigue
4. **Improved reliability** - Escalating warnings reduce missed check-ins

## References

- Current TODOs: `/internal/scheduler/scheduler.go:690`, `/internal/scheduler/scheduler.go:729`
- Urgency calculation: `/internal/scheduler/scheduler.go` (getReminderUrgency method)
- Email client: `/internal/email/client.go`
- Telegram bot: `/internal/telegram/bot.go`

## Estimated Effort

- **Complexity:** Low-Medium
- **Time:** 3-4 hours
- **Risk:** Low

## Dependencies

None - can be implemented independently

## Follow-up Tasks

- Add urgency level to audit logs
- Track response rates by urgency level
- A/B test urgency messaging effectiveness
- Add user preference for urgency sensitivity
