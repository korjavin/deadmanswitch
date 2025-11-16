# TASK-006: Move Email Templates to Dedicated Templates Folder

## Priority: LOW 🟢

## Status: Not Started

## Category: Code Organization / Maintainability

## Description

Email templates are currently embedded as string literals in Go code. This makes them hard to edit, difficult to preview, and impossible to modify without recompiling the application. They should be moved to dedicated HTML template files.

## Problem

**Current State:**
From `/todo.md:25`:
> Move all the email templates from the code to dedicated templates folder

**Current Implementation:**
Email content is likely defined as strings in:
- `/internal/email/client.go`
- Individual handler files
- Scheduler code

**Example of Current Approach:**
```go
func (c *Client) SendPingEmail(email, name, code string) error {
    body := fmt.Sprintf(`
        <html>
        <body>
            <h1>Check-In Required</h1>
            <p>Hi %s,</p>
            <p>Please verify you're still active.</p>
            <a href="https://example.com/verify?code=%s">Verify Now</a>
        </body>
        </html>
    `, name, code)

    return c.SendEmail(...)
}
```

**Why This is Problematic:**
1. **Hard to edit** - Requires Go code changes for content updates
2. **No syntax highlighting** - HTML in strings lacks editor support
3. **No preview** - Can't preview emails without running app
4. **No hot reload** - Must recompile to see changes
5. **Designer unfriendly** - Non-programmers can't edit templates
6. **Version control noise** - Small HTML changes clutter Go diffs

## Proposed Solution

### 1. Create Email Templates Directory Structure

```
/internal/email/templates/
├── ping/
│   ├── normal.html
│   ├── urgent.html
│   └── final.html
├── delivery/
│   ├── secret_delivery.html
│   └── access_code.html
├── auth/
│   ├── welcome.html
│   ├── password_reset.html
│   └── login_notification.html
├── verification/
│   ├── telegram_connect.html
│   └── recipient_test.html
├── shared/
│   ├── header.html
│   ├── footer.html
│   └── styles.html
└── base.html
```

### 2. Create Base Template

**File:** `/internal/email/templates/base.html`

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Subject}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: #007bff;
            color: white;
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f8f9fa;
            padding: 30px;
            border: 1px solid #dee2e6;
        }
        .footer {
            background-color: #6c757d;
            color: white;
            padding: 20px;
            text-align: center;
            font-size: 12px;
            border-radius: 0 0 5px 5px;
        }
        .button {
            display: inline-block;
            padding: 12px 24px;
            background-color: #007bff;
            color: white !important;
            text-decoration: none;
            border-radius: 5px;
            font-weight: bold;
        }
        .urgent { background-color: #ffc107; color: #000; }
        .critical { background-color: #dc3545; }
    </style>
</head>
<body>
    {{template "header" .}}
    {{template "content" .}}
    {{template "footer" .}}
</body>
</html>
```

### 3. Create Individual Templates

**Example:** `/internal/email/templates/ping/normal.html`

```html
{{define "header"}}
<div class="header">
    <h1>✅ Routine Check-In</h1>
</div>
{{end}}

{{define "content"}}
<div class="content">
    <p>Hi {{.Name}},</p>

    <p>It's time for your regular Dead Man's Switch check-in.</p>

    <p>To confirm you're still active, please click the button below:</p>

    <p style="text-align: center; margin: 30px 0;">
        <a href="{{.VerificationLink}}" class="button">
            I'm Active - Verify Now
        </a>
    </p>

    <p>Or enter this code: <strong>{{.VerificationCode}}</strong></p>

    <p><small>This check-in is required every {{.Frequency}} days.
    Your deadline is {{.Deadline}}.</small></p>
</div>
{{end}}

{{define "footer"}}
<div class="footer">
    <p>Dead Man's Switch - Self Hosted</p>
    <p>This is an automated message. Do not reply to this email.</p>
</div>
{{end}}
```

### 4. Update Email Client to Load Templates

**File to modify:** `/internal/email/client.go`

```go
package email

import (
    "bytes"
    "html/template"
    "path/filepath"
)

type Client struct {
    smtpHost   string
    smtpPort   int
    username   string
    password   string
    from       string
    templates  *template.Template
}

// NewClient creates a new email client with templates
func NewClient(host string, port int, username, password, from, templatesPath string) (*Client, error) {
    // Load all templates
    templates, err := loadTemplates(templatesPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load email templates: %w", err)
    }

    return &Client{
        smtpHost:  host,
        smtpPort:  port,
        username:  username,
        password:  password,
        from:      from,
        templates: templates,
    }, nil
}

// loadTemplates loads all email templates
func loadTemplates(templatesPath string) (*template.Template, error) {
    pattern := filepath.Join(templatesPath, "**", "*.html")
    return template.ParseGlob(pattern)
}

// SendPingEmail sends a ping email with the specified urgency
func (c *Client) SendPingEmail(email, name, code string, urgency models.ReminderUrgency) error {
    templateName := fmt.Sprintf("ping/%s.html", urgency)

    data := map[string]interface{}{
        "Name":             name,
        "VerificationCode": code,
        "VerificationLink": fmt.Sprintf("%s/verify?code=%s", c.baseDomain, code),
        "Frequency":        "3", // From config
        "Deadline":         "14 days", // From config
        "Subject":          c.getSubjectByUrgency(urgency),
    }

    body, err := c.renderTemplate(templateName, data)
    if err != nil {
        return err
    }

    return c.SendEmail(&MessageOptions{
        To:      []string{email},
        Subject: data["Subject"].(string),
        Body:    body,
        IsHTML:  true,
    })
}

// renderTemplate renders a template with the given data
func (c *Client) renderTemplate(name string, data interface{}) (string, error) {
    var buf bytes.Buffer
    if err := c.templates.ExecuteTemplate(&buf, name, data); err != nil {
        return "", fmt.Errorf("failed to render template %s: %w", name, err)
    }
    return buf.String(), nil
}
```

### 5. Update Configuration

**File to modify:** `/internal/config/config.go`

```go
type Config struct {
    // ... existing fields ...

    EmailTemplatesPath string
}

func LoadFromEnv() (*Config, error) {
    // ... existing code ...

    config.EmailTemplatesPath = os.Getenv("EMAIL_TEMPLATES_PATH")
    if config.EmailTemplatesPath == "" {
        config.EmailTemplatesPath = "/app/internal/email/templates"
    }

    return config, nil
}
```

### 6. Update Docker Configuration

**File to modify:** `Dockerfile`

Ensure templates are copied into the Docker image:

```dockerfile
# Copy email templates
COPY internal/email/templates /app/internal/email/templates
```

## Email Types to Convert

Based on codebase analysis, these emails need templates:

1. **Ping Emails** (3 urgency levels)
   - Normal routine check-in
   - Urgent (approaching deadline)
   - Final warning

2. **Secret Delivery**
   - Secret delivery notification
   - Access code email

3. **Authentication**
   - Welcome email after registration
   - Login notification (security alert)

4. **Telegram Verification**
   - Telegram account connection verification
   - Telegram disconnect warning

5. **Recipient Management**
   - Test contact email
   - Recipient confirmation

## Acceptance Criteria

- [ ] Email templates directory structure created
- [ ] Base template created with common layout
- [ ] All ping email templates created (normal, urgent, final)
- [ ] Secret delivery templates created
- [ ] Authentication templates created
- [ ] Template loading implemented in email client
- [ ] All email sending functions updated to use templates
- [ ] Configuration added for templates path
- [ ] Docker configuration updated
- [ ] Documentation updated
- [ ] Preview/testing tool created (optional)

## Testing Requirements

1. **Unit Tests:**
   - Test template loading
   - Test template rendering with data
   - Test error handling for missing templates
   - Test template data escaping (XSS prevention)

2. **Integration Tests:**
   - Send test emails with each template
   - Verify HTML renders correctly
   - Test template inheritance (base + specific)

3. **Manual Testing:**
   - Send actual emails and verify appearance
   - Test in multiple email clients (Gmail, Outlook, etc.)
   - Verify responsive design on mobile
   - Check spam score of generated emails

## Files to Create

**New Template Files:**
1. `/internal/email/templates/base.html`
2. `/internal/email/templates/shared/header.html`
3. `/internal/email/templates/shared/footer.html`
4. `/internal/email/templates/ping/normal.html`
5. `/internal/email/templates/ping/urgent.html`
6. `/internal/email/templates/ping/final.html`
7. `/internal/email/templates/delivery/secret_delivery.html`
8. `/internal/email/templates/auth/welcome.html`
9. `/internal/email/templates/auth/login_notification.html`
10. `/internal/email/templates/verification/telegram_connect.html`

**Modified Files:**
1. `/internal/email/client.go` - Add template loading and rendering
2. `/internal/config/config.go` - Add templates path config
3. `/cmd/server/main.go` - Pass templates path to email client
4. `/Dockerfile` - Copy templates directory

**New Tool (Optional):**
1. `/cmd/preview-email/main.go` - Tool to preview templates locally

## Template Preview Tool (Optional)

Create a simple tool to preview templates:

```go
// cmd/preview-email/main.go
package main

import (
    "flag"
    "fmt"
    "html/template"
    "net/http"
    "path/filepath"
)

func main() {
    templatesPath := flag.String("templates", "./internal/email/templates", "Templates path")
    port := flag.Int("port", 8080, "Preview server port")
    flag.Parse()

    tmpl, err := template.ParseGlob(filepath.Join(*templatesPath, "**", "*.html"))
    if err != nil {
        panic(err)
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        templateName := r.URL.Query().Get("template")
        if templateName == "" {
            // Show list of available templates
            return
        }

        // Render template with sample data
        data := getSampleData(templateName)
        tmpl.ExecuteTemplate(w, templateName, data)
    })

    fmt.Printf("Email preview server running on http://localhost:%d\n", *port)
    http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
```

Usage:
```bash
go run cmd/preview-email/main.go
# Visit http://localhost:8080?template=ping/urgent.html
```

## Migration Strategy

1. **Phase 1:** Create templates directory and base template
2. **Phase 2:** Create one template type (e.g., ping emails)
3. **Phase 3:** Update email client to support both old and new methods
4. **Phase 4:** Migrate remaining email types one by one
5. **Phase 5:** Remove old string-based email generation

This allows gradual migration without breaking existing functionality.

## Benefits

1. **Easier editing** - HTML files with syntax highlighting
2. **Hot reload** - Can reload templates without restart (in dev mode)
3. **Designer friendly** - Non-developers can edit templates
4. **Better version control** - HTML changes in .html files, not .go files
5. **Testability** - Can preview templates without running full app
6. **Reusability** - Shared components (header, footer, styles)
7. **Consistency** - All emails use same base layout

## References

- TODO item: `/todo.md:25`
- Email client: `/internal/email/client.go`
- Scheduler email sending: `/internal/scheduler/scheduler.go`

## Estimated Effort

- **Complexity:** Low-Medium
- **Time:** 4-6 hours
- **Risk:** Low

## Dependencies

- Should be done after TASK-003 (urgency levels) to include all urgency templates

## Follow-up Tasks

- Add email preview in admin interface
- Implement A/B testing for email templates
- Add email analytics (open rates, click rates)
- Create email template versioning system
