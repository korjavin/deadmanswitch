package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 parameters
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32

	// Encryption parameters
	nonceSize  = 12
	keySize    = 32
	saltSize   = 16
	tagSize    = 16
	bufferSize = 64 * 1024
)

var (
	// ErrInvalidData is returned when the data to be decrypted is invalid
	ErrInvalidData = errors.New("invalid encrypted data")

	// ErrDecryptionFailed is returned when decryption fails
	ErrDecryptionFailed = errors.New("decryption failed")
)

// DeriveKey derives an encryption key from a password using Argon2id
func DeriveKey(password []byte, salt []byte) ([]byte, error) {
	if salt == nil {
		salt = make([]byte, saltSize)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, fmt.Errorf("failed to generate salt: %w", err)
		}
	}

	key := argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return key, nil
}

// Encrypt encrypts data using AES-GCM with a random nonce
func Encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and seal
	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt decrypts data that was encrypted with Encrypt
func Decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrInvalidData
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// GenerateSalt generates a random salt for key derivation
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// GenerateDataEncryptionKey generates a random key for data encryption
func GenerateDataEncryptionKey() ([]byte, error) {
	dek := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("failed to generate data encryption key: %w", err)
	}
	return dek, nil
}

// EncryptSecret encrypts a secret with a key and returns the complete encrypted package
// Format: base64(salt + encrypted(DEK) + encrypted(secret))
func EncryptSecret(secret []byte, masterKey []byte) (string, error) {
	// Generate a random salt for this operation
	salt, err := GenerateSalt()
	if err != nil {
		return "", err
	}

	// Derive key from master key and salt
	derivedKey, err := DeriveKey(masterKey, salt)
	if err != nil {
		return "", err
	}

	// Generate a data encryption key (DEK) for this secret
	dek, err := GenerateDataEncryptionKey()
	if err != nil {
		return "", err
	}

	// Encrypt the DEK with the derived key
	encryptedDEK, err := Encrypt(dek, derivedKey)
	if err != nil {
		return "", err
	}

	// Encrypt the actual secret with the DEK
	encryptedSecret, err := Encrypt(secret, dek)
	if err != nil {
		return "", err
	}

	// Combine salt + encryptedDEK + encryptedSecret
	dekSize := len(encryptedDEK)
	result := make([]byte, saltSize+4+dekSize+len(encryptedSecret))

	// Copy salt
	copy(result[:saltSize], salt)

	// Copy DEK size (as 4 bytes)
	result[saltSize] = byte(dekSize >> 24)
	result[saltSize+1] = byte(dekSize >> 16)
	result[saltSize+2] = byte(dekSize >> 8)
	result[saltSize+3] = byte(dekSize)

	// Copy encrypted DEK
	copy(result[saltSize+4:saltSize+4+dekSize], encryptedDEK)

	// Copy encrypted secret
	copy(result[saltSize+4+dekSize:], encryptedSecret)

	// Base64 encode everything
	encoded := base64.StdEncoding.EncodeToString(result)
	return encoded, nil
}

// DecryptSecret decrypts a secret that was encrypted with EncryptSecret
func DecryptSecret(encryptedSecret string, masterKey []byte) ([]byte, error) {
	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(encryptedSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Check minimum length
	if len(decoded) < saltSize+4+nonceSize {
		return nil, ErrInvalidData
	}

	// Extract components
	salt := decoded[:saltSize]

	// Read DEK size
	dekSize := int(decoded[saltSize])<<24 | int(decoded[saltSize+1])<<16 | int(decoded[saltSize+2])<<8 | int(decoded[saltSize+3])
	if len(decoded) < saltSize+4+dekSize {
		return nil, ErrInvalidData
	}

	encryptedDEK := decoded[saltSize+4 : saltSize+4+dekSize]
	encryptedData := decoded[saltSize+4+dekSize:]

	// Derive key from master key and salt
	derivedKey, err := DeriveKey(masterKey, salt)
	if err != nil {
		return nil, err
	}

	// Decrypt the DEK
	dek, err := Decrypt(encryptedDEK, derivedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt DEK: %w", err)
	}

	// Decrypt the secret with the DEK
	secret, err := Decrypt(encryptedData, dek)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	return secret, nil
}

// HashPassword creates a secure hash of a password for verification
func HashPassword(password string, salt []byte) ([]byte, error) {
	if salt == nil {
		var err error
		salt, err = GenerateSalt()
		if err != nil {
			return nil, err
		}
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, sha256.Size)
	result := make([]byte, saltSize+sha256.Size)
	copy(result[:saltSize], salt)
	copy(result[saltSize:], hash)

	return result, nil
}

// VerifyPassword checks if a password matches a stored hash
func VerifyPassword(password string, storedHash []byte) (bool, error) {
	if len(storedHash) < saltSize+sha256.Size {
		return false, ErrInvalidData
	}

	salt := storedHash[:saltSize]
	hash := storedHash[saltSize:]

	newHash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, sha256.Size)

	return hmacEqual(hash, newHash), nil
}

// hmacEqual is a constant-time comparison function
func hmacEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}

	return v == 0
}
