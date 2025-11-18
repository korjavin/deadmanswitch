package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"

	"github.com/korjavin/deadmanswitch/internal/config"
)

// Client provides methods for sending emails
type Client struct {
	config    *config.Config
	auth      smtp.Auth
	templates *template.Template
}

// MessageOptions defines options for an email message
type MessageOptions struct {
	From    string
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// NewClient creates a new email client
func NewClient(config *config.Config) (*Client, error) {
	if config.SMTPHost == "" || config.SMTPUsername == "" || config.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP configuration is incomplete")
	}

	// Create SMTP auth
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)

	// Load email templates
	templates, err := loadTemplates(config.EmailTemplatesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	return &Client{
		config:    config,
		auth:      auth,
		templates: templates,
	}, nil
}

// loadTemplates loads all email templates from the specified path
func loadTemplates(templatesPath string) (*template.Template, error) {
	// Load all HTML files from the templates directory
	pattern := filepath.Join(templatesPath, "**", "*.html")
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		// Try alternative pattern for direct subdirectories
		pattern = filepath.Join(templatesPath, "*", "*.html")
		tmpl, err = template.ParseGlob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to parse templates: %w", err)
		}
	}
	return tmpl, nil
}

// renderTemplate renders a template with the given data
func (c *Client) renderTemplate(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := c.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateName, err)
	}
	return buf.String(), nil
}

// SendEmail sends an email with the specified options
func (c *Client) SendEmail(options *MessageOptions) error {
	return c.SendEmailSimple(options.To, options.Subject, options.Body, options.IsHTML)
}

// SendEmailSimple sends an email with basic parameters
func (c *Client) SendEmailSimple(to []string, subject, body string, isHTML bool) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	// Use configured From address
	from := c.config.SMTPFrom

	// Build headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(to, ", ")
	headers["Subject"] = subject
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["MIME-Version"] = "1.0"

	// Set content type based on IsHTML flag
	contentType := "text/plain; charset=utf-8"
	if isHTML {
		contentType = "text/html; charset=utf-8"
	}
	headers["Content-Type"] = contentType

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the server and enable TLS
	smtpClient, err := smtp.Dial(fmt.Sprintf("%s:%d", c.config.SMTPHost, c.config.SMTPPort))
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer smtpClient.Close()

	// Enable TLS if the server supports it
	if err := smtpClient.StartTLS(&tls.Config{
		ServerName: c.config.SMTPHost,
		MinVersion: tls.VersionTLS12, // Require TLS 1.2 or higher for security
	}); err != nil {
		// Some servers might not support TLS, continue without it
		// but log a warning
		fmt.Printf("Warning: TLS not supported by SMTP server: %s\n", err)
	}

	// Authenticate
	if err := smtpClient.Auth(c.auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set the sender and recipients
	if err := smtpClient.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	for _, addr := range to {
		if err := smtpClient.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", addr, err)
		}
	}

	// Send the email body
	w, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write email data: %w", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return smtpClient.Quit()
}

// SendPingEmail sends a ping email to a user with a verification link
func (c *Client) SendPingEmail(email string, verificationCode string, urgency string) error {
	baseURL := fmt.Sprintf("https://%s", c.config.BaseDomain)
	verificationURL := fmt.Sprintf("%s/verify/%s", baseURL, verificationCode)

	// Determine template name based on urgency
	templateName := fmt.Sprintf("%s.html", urgency)
	if urgency == "" || urgency == "normal" {
		templateName = "normal.html"
	}

	// Prepare template data
	data := map[string]interface{}{
		"VerificationURL": verificationURL,
	}

	// Render template
	body, err := c.renderTemplate(templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	subject := c.getSubjectByUrgency(urgency)

	return c.SendEmail(&MessageOptions{
		To:      []string{email},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// getSubjectByUrgency returns an urgency-appropriate subject line
func (c *Client) getSubjectByUrgency(urgency string) string {
	switch urgency {
	case "final_warning":
		return "🚨 URGENT: Final Check-In Required - Dead Man's Switch"
	case "urgent":
		return "⚠️ IMPORTANT: Check-In Required Soon - Dead Man's Switch"
	default:
		return "✅ Routine Check-In - Dead Man's Switch"
	}
}

// SendSecretDeliveryEmail sends an email with access to a user's secrets
func (c *Client) SendSecretDeliveryEmail(recipientEmail, recipientName, message string, accessCode string) error {
	baseURL := fmt.Sprintf("https://%s", c.config.BaseDomain)
	accessURL := fmt.Sprintf("%s/access/%s", baseURL, accessCode)

	// Prepare template data
	data := map[string]interface{}{
		"RecipientName": recipientName,
		"Message":       message,
		"AccessURL":     accessURL,
	}

	// Render template
	body, err := c.renderTemplate("secret_delivery.html", data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	subject := "Important: Confidential Information Access"

	return c.SendEmail(&MessageOptions{
		To:      []string{recipientEmail},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}
