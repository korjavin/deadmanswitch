package email

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSMTPServer is a simple mock SMTP server for testing
type mockSMTPServer struct {
	addr     string
	messages []string
	listener net.Listener
	quit     chan bool
}

// newMockSMTPServer creates a new mock SMTP server
func newMockSMTPServer() (*mockSMTPServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	server := &mockSMTPServer{
		addr:     listener.Addr().String(),
		messages: []string{},
		listener: listener,
		quit:     make(chan bool),
	}

	go server.start()
	return server, nil
}

// start starts the mock SMTP server
func (s *mockSMTPServer) start() {
	for {
		select {
		case <-s.quit:
			return
		default:
			// listener.Accept() will block until a new connection or an error (e.g. listener closed).
			conn, err := s.listener.Accept()
			if err != nil {
				// When listener is closed, Accept returns an error. Check if it's due to closing.
				select {
				case <-s.quit: // If quit channel is closed, this is an expected error during shutdown.
					return
				default: // Otherwise, it's an unexpected error.
					// Optional: log unexpected errors, e.g., fmt.Printf("Mock SMTP server accept error: %v\n", err)
				}
				return // Exit goroutine on accept error if not explicitly quitting.
			}
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a new SMTP connection
func (s *mockSMTPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Send greeting
	conn.Write([]byte("220 mock.smtp.server ESMTP\r\n"))

	// Buffer to store the message
	var message strings.Builder

	// Read commands
	buf := make([]byte, 1024)
	dataMode := false

	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		cmd := string(buf[:n])

		if dataMode {
			if strings.HasSuffix(cmd, "\r\n.\r\n") {
				// End of data
				message.WriteString(strings.TrimSuffix(cmd, "\r\n.\r\n"))
				s.messages = append(s.messages, message.String())
				conn.Write([]byte("250 OK: message accepted\r\n"))
				dataMode = false
				message.Reset() // Reset for next message
			} else {
				message.WriteString(cmd)
			}
			continue
		}

		// Handle commands
		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			conn.Write([]byte("250-mock.smtp.server\r\n250-SIZE 35882577\r\n250-AUTH LOGIN PLAIN\r\n250-STARTTLS\r\n250 OK\r\n"))
		case strings.HasPrefix(cmd, "HELO"):
			conn.Write([]byte("250 Hello\r\n"))
		case strings.HasPrefix(cmd, "MAIL FROM:"):
			conn.Write([]byte("250 OK\r\n"))
		case strings.HasPrefix(cmd, "RCPT TO:"):
			conn.Write([]byte("250 OK\r\n"))
		case strings.HasPrefix(cmd, "DATA"):
			conn.Write([]byte("354 End data with <CR><LF>.<CR><LF>\r\n"))
			dataMode = true
			message.Reset()
		case strings.HasPrefix(cmd, "STARTTLS"):
			conn.Write([]byte("220 Ready to start TLS\r\n"))
			// In a real server, TLS handshake would happen here.
			// Mock just acknowledges. Client will proceed on same connection.
		case strings.HasPrefix(cmd, "AUTH"):
			// Could add logic to check auth details if needed for a test
			conn.Write([]byte("235 Authentication successful\r\n"))
		case strings.HasPrefix(cmd, "QUIT"):
			conn.Write([]byte("221 Bye\r\n"))
			return
		default:
			// Log unrecognized command for debugging if necessary
			// fmt.Printf("Mock SMTP Unrecognized command: %s\n", cmd)
			conn.Write([]byte("500 Unrecognized command\r\n"))
		}
	}
}

// stop stops the mock SMTP server
func (s *mockSMTPServer) stop() {
	close(s.quit)
	s.listener.Close() // This will interrupt listener.Accept()
}

// getHost returns the host part of the server address
func (s *mockSMTPServer) getHost() string {
	host, _, _ := net.SplitHostPort(s.addr)
	return host
}

// getPort returns the port part of the server address
func (s *mockSMTPServer) getPort() int {
	_, portStr, _ := net.SplitHostPort(s.addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return port
}

// resetMessages clears the captured messages
func (s *mockSMTPServer) resetMessages() {
	s.messages = []string{}
}

func TestNewClient(t *testing.T) {
	// Test with valid config
	cfg := &config.Config{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "password",
		SMTPFrom:     "noreply@example.com",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err) // Corrected from if err != nil
	require.NotNil(t, client) // Corrected from if client == nil
	assert.Equal(t, cfg, client.config) // Corrected from if client.config != cfg
	assert.NotNil(t, client.auth) // Corrected from if client.auth == nil

	// Test with invalid config (missing host)
	invalidCfgHost := &config.Config{ // Renamed to avoid conflict
		SMTPUsername: "user@example.com",
		SMTPPassword: "password",
	}
	client, err = NewClient(invalidCfgHost)
	require.Error(t, err) // Corrected from if err == nil
	require.Nil(t, client) // Corrected from if client != nil

	// Test with invalid config (missing username)
	invalidCfgUser := &config.Config{ // Renamed
		SMTPHost:     "smtp.example.com",
		SMTPPassword: "password",
	}
	client, err = NewClient(invalidCfgUser)
	require.Error(t, err)
	require.Nil(t, client)

	// Test with invalid config (missing password)
	invalidCfgPass := &config.Config{ // Renamed
		SMTPHost:     "smtp.example.com",
		SMTPUsername: "user@example.com",
	}
	client, err = NewClient(invalidCfgPass)
	require.Error(t, err)
	require.Nil(t, client)
}

func TestSendEmailSimple_Validation(t *testing.T) {
	// Create a client
	cfg := &config.Config{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "password",
		SMTPFrom:     "noreply@example.com",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err) // Corrected

	// Test with empty recipients
	err = client.SendEmailSimple([]string{}, "Subject", "Body", false)
	require.Error(t, err) // Corrected
	assert.Contains(t, err.Error(), "no recipients specified") // Corrected
}

func TestSendEmailSimple_WithMockServer(t *testing.T) {
	mockServer, err := newMockSMTPServer()
	require.NoError(t, err)
	defer mockServer.stop()

	cfg := &config.Config{
		SMTPHost:     mockServer.getHost(),
		SMTPPort:     mockServer.getPort(),
		SMTPUsername: "testuser",
		SMTPPassword: "testpassword",
		SMTPFrom:     "sender@example.com",
		SMTPNoTLS:    true, // Disable STARTTLS for this mock server test
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	recipient := "recipient@example.net"
	subject := "Hello Test"
	body := "This is a test email body."

	err = client.SendEmailSimple([]string{recipient}, subject, body, false)
	require.NoError(t, err)

	require.Len(t, mockServer.messages, 1, "Expected one message to be captured")
	capturedMessage := mockServer.messages[0]

	assert.Contains(t, capturedMessage, fmt.Sprintf("From: %s", cfg.SMTPFrom))
	assert.Contains(t, capturedMessage, fmt.Sprintf("To: %s", recipient))
	assert.Contains(t, capturedMessage, fmt.Sprintf("Subject: %s", subject))
	assert.Contains(t, capturedMessage, "Content-Type: text/plain; charset=utf-8")
	assert.Contains(t, capturedMessage, "\r\n\r\n"+body) // Ensure body is after headers
}

func TestSendPingEmail_WithMockServer(t *testing.T) {
	mockServer, err := newMockSMTPServer()
	require.NoError(t, err)
	defer mockServer.stop()

	cfg := &config.Config{
		BaseDomain:   "test.deadmanswitch.com",
		SMTPHost:     mockServer.getHost(),
		SMTPPort:     mockServer.getPort(),
		SMTPUsername: "testuser",
		SMTPPassword: "testpassword",
		SMTPFrom:     "noreply@test.deadmanswitch.com",
		SMTPNoTLS:    true,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	userEmail := "ping.user@example.org"
	userName := "Ping User Name"
	verificationCode := "ping123abc"
	expectedVerificationURL := fmt.Sprintf("https://%s/verify/%s", cfg.BaseDomain, verificationCode)

	err = client.SendPingEmail(userEmail, userName, verificationCode)
	require.NoError(t, err)

	require.Len(t, mockServer.messages, 1, "Expected one message to be captured")
	capturedMessage := mockServer.messages[0]

	assert.Contains(t, capturedMessage, fmt.Sprintf("From: %s", cfg.SMTPFrom))
	assert.Contains(t, capturedMessage, fmt.Sprintf("To: %s", userEmail))
	assert.Contains(t, capturedMessage, "Subject: Action Required: Dead Man's Switch Check-In")
	assert.Contains(t, capturedMessage, "Content-Type: text/html; charset=utf-8")
	assert.Contains(t, capturedMessage, fmt.Sprintf("Hello %s,", userName))
	assert.Contains(t, capturedMessage, fmt.Sprintf("href=\"%s\"", expectedVerificationURL))
}

func TestSendSecretDeliveryEmail_WithMockServer(t *testing.T) {
	mockServer, err := newMockSMTPServer()
	require.NoError(t, err)
	defer mockServer.stop()

	cfg := &config.Config{
		BaseDomain:   "secret.delivery.com",
		SMTPHost:     mockServer.getHost(),
		SMTPPort:     mockServer.getPort(),
		SMTPUsername: "testuser",
		SMTPPassword: "testpassword",
		SMTPFrom:     "delivery@secret.delivery.com",
		SMTPNoTLS:    true,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	recipientEmail := "recipient.secret@example.net"
	recipientName := "Valued Recipient"
	customMessage := "Here is the secret information you were promised."
	accessCode := "accessCodeXYZ789"
	expectedAccessURL := fmt.Sprintf("https://%s/access/%s", cfg.BaseDomain, accessCode)

	err = client.SendSecretDeliveryEmail(recipientEmail, recipientName, customMessage, accessCode)
	require.NoError(t, err)

	require.Len(t, mockServer.messages, 1, "Expected one message to be captured")
	capturedMessage := mockServer.messages[0]

	assert.Contains(t, capturedMessage, fmt.Sprintf("From: %s", cfg.SMTPFrom))
	assert.Contains(t, capturedMessage, fmt.Sprintf("To: %s", recipientEmail))
	assert.Contains(t, capturedMessage, "Subject: Important: Confidential Information Access")
	assert.Contains(t, capturedMessage, "Content-Type: text/html; charset=utf-8")
	assert.Contains(t, capturedMessage, fmt.Sprintf("Hello %s,", recipientName))
	assert.Contains(t, capturedMessage, customMessage)
	assert.Contains(t, capturedMessage, fmt.Sprintf("href=\"%s\"", expectedAccessURL))
}


// TestSendEmail, TestSendPingEmail, TestSendSecretDeliveryEmail (original connection-failing tests)
// can be kept or removed. For this exercise, I'll assume they are not needed as the new mock tests
// cover the sending logic more effectively. If they are to be kept, they should be clearly marked
// as integration tests that require a real SMTP server or a more advanced mock.
// For now, I will comment them out to avoid confusion and ensure only mock-based tests run.

/*
func TestSendEmail(t *testing.T) {
	// Create a client
	cfg := &config.Config{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "password",
		SMTPFrom:     "noreply@example.com",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err) // Was: if err != nil { t.Fatalf(...) }

	// Test SendEmail method
	options := &MessageOptions{
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Test Body",
		IsHTML:  false,
	}

	// This will fail because we're not actually connecting to an SMTP server
	// but we can verify that it attempts to send the email
	err = client.SendEmail(options)
	require.Error(t, err) // Was: if err == nil { t.Fatal(...) }
	// The error should be about connecting to the SMTP server
	assert.Contains(t, err.Error(), "failed to connect to SMTP server") // Was: if !strings.Contains(...)
}

func TestSendPingEmail(t *testing.T) {
	// Create a client
	cfg := &config.Config{
		BaseDomain:   "example.com",
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "password",
		SMTPFrom:     "noreply@example.com",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Test SendPingEmail method
	email := "user@example.com"
	name := "Test User"
	verificationCode := "abc123"

	err = client.SendPingEmail(email, name, verificationCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to SMTP server")
}

func TestSendSecretDeliveryEmail(t *testing.T) {
	// Create a client
	cfg := &config.Config{
		BaseDomain:   "example.com",
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "password",
		SMTPFrom:     "noreply@example.com",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Test SendSecretDeliveryEmail method
	recipientEmail := "recipient@example.com"
	recipientName := "Test Recipient"
	message := "Here are my secrets"
	accessCode := "xyz789"

	err = client.SendSecretDeliveryEmail(recipientEmail, recipientName, message, accessCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to SMTP server")
}
*/
