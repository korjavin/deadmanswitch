package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func TestDefaultTOTPConfig(t *testing.T) {
	config := DefaultTOTPConfig()

	if config.Issuer != "DeadMansSwitch" {
		t.Errorf("Expected Issuer to be 'DeadMansSwitch', got '%s'", config.Issuer)
	}
	if config.Period != 30 {
		t.Errorf("Expected Period to be 30, got %d", config.Period)
	}
	if config.Digits != otp.DigitsSix {
		t.Errorf("Expected Digits to be DigitsSix, got %v", config.Digits)
	}
	if config.Algorithm != otp.AlgorithmSHA1 {
		t.Errorf("Expected Algorithm to be AlgorithmSHA1, got %v", config.Algorithm)
	}
}

func TestGenerateTOTPSecret(t *testing.T) {
	email := "test@example.com"
	config := DefaultTOTPConfig()

	secret, qrCode, err := GenerateTOTPSecret(email, config)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret failed: %v", err)
	}

	// Check that secret is not empty and is base32 encoded
	if secret == "" {
		t.Error("Expected non-empty secret")
	}
	// Base32 characters are A-Z, 2-7, and may end with = padding
	validBase32 := true
	for _, c := range strings.ToUpper(secret) {
		if !((c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7') || c == '=') {
			validBase32 = false
			break
		}
	}
	if !validBase32 {
		t.Errorf("Secret '%s' is not valid base32", secret)
	}

	// Check that QR code is not empty and is base64 encoded
	if qrCode == "" {
		t.Error("Expected non-empty QR code")
	}
	// Simple check that it starts with a base64 PNG header
	if !strings.HasPrefix(qrCode, "iVBOR") {
		t.Errorf("QR code doesn't appear to be a valid base64-encoded PNG")
	}
}

func TestValidateTOTP(t *testing.T) {
	// Generate a secret
	email := "test@example.com"
	config := DefaultTOTPConfig()

	secret, _, err := GenerateTOTPSecret(email, config)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret failed: %v", err)
	}

	// Generate a valid code
	validCode, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("Failed to generate TOTP code: %v", err)
	}

	// Test with valid code
	if !ValidateTOTP(secret, validCode, config) {
		t.Errorf("ValidateTOTP failed with valid code")
	}

	// Test with invalid code
	invalidCode := "000000"
	if validCode == invalidCode {
		invalidCode = "999999" // Ensure we have a different code
	}
	if ValidateTOTP(secret, invalidCode, config) {
		t.Errorf("ValidateTOTP succeeded with invalid code")
	}
}

func TestGenerateTOTPCode(t *testing.T) {
	// Generate a secret
	email := "test@example.com"
	config := DefaultTOTPConfig()

	secret, _, err := GenerateTOTPSecret(email, config)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret failed: %v", err)
	}

	// Generate a code
	code, err := GenerateTOTPCode(secret, config)
	if err != nil {
		t.Fatalf("GenerateTOTPCode failed: %v", err)
	}

	// Check that code is not empty and has the correct length
	if code == "" {
		t.Error("Expected non-empty code")
	}
	if len(code) != int(config.Digits) {
		t.Errorf("Expected code length to be %d, got %d", config.Digits, len(code))
	}

	// Validate the generated code
	if !ValidateTOTP(secret, code, config) {
		t.Errorf("Generated code failed validation")
	}
}
