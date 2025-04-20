package crypto

import (
	"fmt"

	"github.com/corvus-ch/shamir"
)

// SplitSecret splits a secret into n shares, where k shares are required to reconstruct the secret
func SplitSecret(secret []byte, k, n int) ([][]byte, error) {
	if k < 2 {
		return nil, fmt.Errorf("threshold (k) must be at least 2")
	}
	if n < k {
		return nil, fmt.Errorf("total shares (n) must be at least equal to threshold (k)")
	}
	if len(secret) == 0 {
		return nil, fmt.Errorf("secret cannot be empty")
	}

	// shamir.Split returns a map[byte][]byte, we need to convert it to [][]byte
	shamirShares, err := shamir.Split(secret, n, k)
	if err != nil {
		return nil, fmt.Errorf("failed to split secret: %w", err)
	}

	shares := make([][]byte, n)
	for i := 0; i < n; i++ {
		// The key is the share ID (1-based)
		shares[i] = shamirShares[byte(i+1)]
	}

	return shares, nil
}

// CombineShares reconstructs a secret from k shares
func CombineShares(shares [][]byte) ([]byte, error) {
	if len(shares) < 2 {
		return nil, fmt.Errorf("at least 2 shares are required")
	}

	// Convert [][]byte to map[byte][]byte for shamir.Combine
	shamirShares := make(map[byte][]byte)
	for i, share := range shares {
		if share != nil {
			// Use 1-based index as key
			shamirShares[byte(i+1)] = share
		}
	}

	return shamir.Combine(shamirShares)
}

// EncryptShare encrypts a share with a key derived from an answer
func EncryptShare(share []byte, answer string) ([]byte, []byte, error) {
	// Generate a random salt
	salt, err := GenerateSalt()
	if err != nil {
		return nil, nil, err
	}

	// Derive key from answer and salt
	key, err := DeriveKey([]byte(answer), salt)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt the share with the derived key
	encryptedShare, err := Encrypt(share, key)
	if err != nil {
		return nil, nil, err
	}

	return encryptedShare, salt, nil
}

// DecryptShare decrypts a share with a key derived from an answer
func DecryptShare(encryptedShare []byte, answer string, salt []byte) ([]byte, error) {
	// Derive key from answer and salt
	key, err := DeriveKey([]byte(answer), salt)
	if err != nil {
		return nil, err
	}

	// Decrypt the share with the derived key
	share, err := Decrypt(encryptedShare, key)
	if err != nil {
		return nil, err
	}

	return share, nil
}
