package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/korjavin/deadmanswitch/internal/config"
)

// Client provides methods for sending emails
type Client struct {
	config *config.Config
	auth   smtp.Auth
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

	return &Client{
		config: config,
		auth:   auth,
	}, nil
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
	if err := smtpClient.StartTLS(&tls.Config{ServerName: c.config.SMTPHost}); err != nil {
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
func (c *Client) SendPingEmail(email string, name string, verificationCode string) error {
	baseURL := fmt.Sprintf("https://%s", c.config.BaseDomain)
	verificationURL := fmt.Sprintf("%s/verify/%s", baseURL, verificationCode)

	subject := "Action Required: Dead Man's Switch Check-In"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Action Required: Dead Man's Switch Check-In</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin-bottom: 20px;">
        <h2 style="color: #343a40;">Dead Man's Switch: Action Required</h2>
    </div>

    <p>Hello %s,</p>

    <p>This is an automated check-in request from your Dead Man's Switch service.</p>

    <p><strong>Action Required:</strong> Please confirm you're okay by clicking the button below within your configured deadline.</p>

    <div style="text-align: center; margin: 30px 0;">
        <a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; font-weight: bold;">I'm OK - Confirm</a>
    </div>

    <p><strong>Important:</strong> If you don't respond, your pre-configured secrets will be automatically sent to your designated recipients.</p>

    <p>If you can't click the button, copy and paste this URL into your browser:</p>
    <p style="word-break: break-all; background-color: #f8f9fa; padding: 10px; border-radius: 4px;">%s</p>

    <div style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #6c757d;">
        <p>This is an automated message from your self-hosted Dead Man's Switch service.</p>
        <p>If you did not set up this service, please disregard this email.</p>
    </div>
</body>
</html>
`, name, verificationURL, verificationURL)

	return c.SendEmail(&MessageOptions{
		To:      []string{email},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// SendSecretDeliveryEmail sends an email with access to a user's secrets
func (c *Client) SendSecretDeliveryEmail(recipientEmail, recipientName, message string, accessCode string) error {
	baseURL := fmt.Sprintf("https://%s", c.config.BaseDomain)
	accessURL := fmt.Sprintf("%s/access/%s", baseURL, accessCode)

	subject := "Important: Confidential Information Access"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Important: Confidential Information Access</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin-bottom: 20px;">
        <h2 style="color: #343a40;">Confidential Information Access</h2>
    </div>

    <p>Hello %s,</p>

    <p>Someone has designated you as a recipient of confidential information through their Dead Man's Switch service.</p>

    <p>They have left you the following message:</p>

    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0; font-style: italic;">
        %s
    </div>

    <p>To access the confidential information they've shared with you, please click the button below:</p>

    <div style="text-align: center; margin: 30px 0;">
        <a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; font-weight: bold;">Access Confidential Information</a>
    </div>

    <p>If you can't click the button, copy and paste this URL into your browser:</p>
    <p style="word-break: break-all; background-color: #f8f9fa; padding: 10px; border-radius: 4px;">%s</p>

    <p><strong>Important:</strong> This link will expire after a limited time for security reasons.</p>

    <div style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #6c757d;">
        <p>This is a one-time notification. No further action is required if you choose not to access the information.</p>
    </div>
</body>
</html>
`, recipientName, message, accessURL, accessURL)

	return c.SendEmail(&MessageOptions{
		To:      []string{recipientEmail},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}
