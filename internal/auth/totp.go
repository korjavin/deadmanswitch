package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPConfig holds configuration for TOTP
type TOTPConfig struct {
	Issuer    string        // The issuer name (usually the app name)
	Period    uint          // The period in seconds (default: 30)
	Digits    otp.Digits    // The number of digits (default: 6)
	Algorithm otp.Algorithm // The algorithm (default: SHA1)
}

// DefaultTOTPConfig returns the default TOTP configuration
func DefaultTOTPConfig() TOTPConfig {
	return TOTPConfig{
		Issuer:    "DeadMansSwitch",
		Period:    30,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}
}

// GenerateTOTPSecret generates a new TOTP secret for a user
func GenerateTOTPSecret(email string, config TOTPConfig) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.Issuer,
		AccountName: email,
		Period:      config.Period,
		Digits:      config.Digits,
		Algorithm:   config.Algorithm,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Get the secret in base32 format
	secret := key.Secret()

	// Generate QR code
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	if err := png.Encode(&buf, img); err != nil {
		return "", "", fmt.Errorf("failed to encode QR code: %w", err)
	}

	qrCode := base64.StdEncoding.EncodeToString(buf.Bytes())
	return secret, qrCode, nil
}

// ValidateTOTP validates a TOTP code against a secret
func ValidateTOTP(secret, code string, config TOTPConfig) bool {
	return totp.Validate(code, secret)
}

// GenerateTOTPCode generates a TOTP code for a given secret (useful for testing)
func GenerateTOTPCode(secret string, config TOTPConfig) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}
