package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	password := []byte("test-password")

	// Test with nil salt (should generate a random salt)
	key1, err := DeriveKey(password, nil)
	if err != nil {
		t.Fatalf("DeriveKey with nil salt failed: %v", err)
	}
	if len(key1) != argonKeyLen {
		t.Errorf("Expected key length %d, got %d", argonKeyLen, len(key1))
	}

	// Test with provided salt
	salt := []byte("0123456789abcdef") // 16 bytes
	key2, err := DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("DeriveKey with provided salt failed: %v", err)
	}
	if len(key2) != argonKeyLen {
		t.Errorf("Expected key length %d, got %d", argonKeyLen, len(key2))
	}

	// Derive key again with same salt - should get same result
	key2Again, err := DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("DeriveKey with same salt failed: %v", err)
	}
	if !bytes.Equal(key2, key2Again) {
		t.Errorf("Keys derived with same password and salt should be identical")
	}

	// Different password should produce different key
	differentPassword := []byte("different-password")
	key3, err := DeriveKey(differentPassword, salt)
	if err != nil {
		t.Fatalf("DeriveKey with different password failed: %v", err)
	}
	if bytes.Equal(key2, key3) {
		t.Errorf("Keys derived with different passwords should be different")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, keySize)
	for i := range key {
		key[i] = byte(i)
	}

	testData := []byte("This is a test message to encrypt and decrypt")

	// Encrypt the data
	encrypted, err := Encrypt(testData, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Encrypted data should be different from original
	if bytes.Equal(encrypted, testData) {
		t.Errorf("Encrypted data should be different from original")
	}

	// Encrypted data should be longer than original (nonce + ciphertext + tag)
	if len(encrypted) <= len(testData) {
		t.Errorf("Encrypted data should be longer than original")
	}

	// Decrypt the data
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	// Decrypted data should match original
	if !bytes.Equal(decrypted, testData) {
		t.Errorf("Decrypted data doesn't match original")
	}

	// Test decryption with wrong key
	wrongKey := make([]byte, keySize)
	_, err = Decrypt(encrypted, wrongKey)
	if err == nil {
		t.Errorf("Decryption with wrong key should fail")
	}

	// Test decryption with invalid data
	invalidData := []byte("too short")
	_, err = Decrypt(invalidData, key)
	if err == nil {
		t.Errorf("Decryption with invalid data should fail")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}
	if len(salt1) != saltSize {
		t.Errorf("Expected salt length %d, got %d", saltSize, len(salt1))
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	// Two generated salts should be different
	if bytes.Equal(salt1, salt2) {
		t.Errorf("Two generated salts should be different")
	}
}

func TestGenerateDataEncryptionKey(t *testing.T) {
	dek1, err := GenerateDataEncryptionKey()
	if err != nil {
		t.Fatalf("GenerateDataEncryptionKey failed: %v", err)
	}
	if len(dek1) != keySize {
		t.Errorf("Expected DEK length %d, got %d", keySize, len(dek1))
	}

	dek2, err := GenerateDataEncryptionKey()
	if err != nil {
		t.Fatalf("GenerateDataEncryptionKey failed: %v", err)
	}

	// Two generated DEKs should be different
	if bytes.Equal(dek1, dek2) {
		t.Errorf("Two generated DEKs should be different")
	}
}

func TestEncryptDecryptSecret(t *testing.T) {
	masterKey := []byte("master-password-for-testing-purposes")
	secret := []byte("This is a secret message that needs to be encrypted")

	// Encrypt the secret
	encryptedSecret, err := EncryptSecret(secret, masterKey)
	if err != nil {
		t.Fatalf("EncryptSecret failed: %v", err)
	}

	// Encrypted secret should be a base64 string
	_, err = base64.StdEncoding.DecodeString(encryptedSecret)
	if err != nil {
		t.Errorf("Encrypted secret is not valid base64: %v", err)
	}

	// Decrypt the secret
	decryptedSecret, err := DecryptSecret(encryptedSecret, masterKey)
	if err != nil {
		t.Fatalf("DecryptSecret failed: %v", err)
	}

	// Decrypted secret should match original
	if !bytes.Equal(decryptedSecret, secret) {
		t.Errorf("Decrypted secret doesn't match original")
	}

	// Test decryption with wrong master key
	wrongMasterKey := []byte("wrong-master-password")
	_, err = DecryptSecret(encryptedSecret, wrongMasterKey)
	if err == nil {
		t.Errorf("Decryption with wrong master key should fail")
	}

	// Test decryption with invalid encrypted secret
	invalidEncryptedSecret := "invalid-base64"
	_, err = DecryptSecret(invalidEncryptedSecret, masterKey)
	if err == nil {
		t.Errorf("Decryption with invalid encrypted secret should fail")
	}
}

func TestHashVerifyPassword(t *testing.T) {
	password := "secure-password-123"

	// Hash the password
	hashedPassword, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify the password
	valid, err := VerifyPassword(password, hashedPassword)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Errorf("Password verification should succeed with correct password")
	}

	// Verify with wrong password
	wrongPassword := "wrong-password"
	valid, err = VerifyPassword(wrongPassword, hashedPassword)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if valid {
		t.Errorf("Password verification should fail with wrong password")
	}

	// Test with invalid hash
	invalidHash := []byte("too-short")
	_, err = VerifyPassword(password, invalidHash)
	if err == nil {
		t.Errorf("VerifyPassword should fail with invalid hash")
	}

	// Test with provided salt
	salt := []byte("0123456789abcdef") // 16 bytes
	hashedPassword2, err := HashPassword(password, salt)
	if err != nil {
		t.Fatalf("HashPassword with provided salt failed: %v", err)
	}

	// Verify the password
	valid, err = VerifyPassword(password, hashedPassword2)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Errorf("Password verification should succeed with correct password")
	}
}

func TestHmacEqual(t *testing.T) {
	// Test equal slices
	a := []byte{1, 2, 3, 4, 5}
	b := []byte{1, 2, 3, 4, 5}
	if !hmacEqual(a, b) {
		t.Errorf("hmacEqual should return true for equal slices")
	}

	// Test different length slices
	c := []byte{1, 2, 3, 4}
	if hmacEqual(a, c) {
		t.Errorf("hmacEqual should return false for different length slices")
	}

	// Test different content slices
	d := []byte{1, 2, 3, 4, 6}
	if hmacEqual(a, d) {
		t.Errorf("hmacEqual should return false for different content slices")
	}
}
