package crypto

import (
	"bytes"
	"testing"
)

func TestSplitAndCombineSecret(t *testing.T) {
	// Test cases
	testCases := []struct {
		name      string
		secret    []byte
		k         int
		n         int
		expectErr bool
	}{
		{
			name:      "Valid 3-of-5 sharing",
			secret:    []byte("this is a test secret"),
			k:         3,
			n:         5,
			expectErr: false,
		},
		{
			name:      "Valid 2-of-3 sharing",
			secret:    []byte("another secret"),
			k:         2,
			n:         3,
			expectErr: false,
		},
		{
			name:      "Invalid k < 2",
			secret:    []byte("invalid k"),
			k:         1,
			n:         3,
			expectErr: true,
		},
		{
			name:      "Invalid n < k",
			secret:    []byte("invalid n"),
			k:         3,
			n:         2,
			expectErr: true,
		},
		{
			name:      "Empty secret",
			secret:    []byte{},
			k:         2,
			n:         3,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Split the secret
			shares, err := SplitSecret(tc.secret, tc.k, tc.n)

			// Check if error was expected
			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			// If not expecting error, but got one
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify number of shares
			if len(shares) != tc.n {
				t.Errorf("Expected %d shares, got %d", tc.n, len(shares))
			}

			// Test reconstruction with exactly k shares
			kShares := shares[:tc.k]
			reconstructed, err := CombineShares(kShares)
			if err != nil {
				t.Fatalf("Failed to combine shares: %v", err)
			}

			// Verify reconstructed secret
			if !bytes.Equal(reconstructed, tc.secret) {
				t.Errorf("Reconstructed secret doesn't match original. Got %s, expected %s",
					string(reconstructed), string(tc.secret))
			}

			// Test reconstruction with more than k shares
			if tc.k < tc.n {
				moreShares := shares[:tc.k+1]
				reconstructed, err = CombineShares(moreShares)
				if err != nil {
					t.Fatalf("Failed to combine shares with more than k: %v", err)
				}

				// Verify reconstructed secret
				if !bytes.Equal(reconstructed, tc.secret) {
					t.Errorf("Reconstructed secret with more shares doesn't match original. Got %s, expected %s",
						string(reconstructed), string(tc.secret))
				}
			}

			// Test reconstruction with fewer than k shares
			if tc.k > 2 {
				fewerShares := shares[:tc.k-1]
				_, err = CombineShares(fewerShares)
				if err == nil {
					t.Errorf("Expected error when combining fewer than k shares, but got none")
				}
			}
		})
	}
}

func TestEncryptAndDecryptShare(t *testing.T) {
	// Test cases
	testCases := []struct {
		name   string
		share  []byte
		answer string
	}{
		{
			name:   "Simple share and answer",
			share:  []byte("this is a share"),
			answer: "correct answer",
		},
		{
			name:   "Binary share",
			share:  []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			answer: "binary answer",
		},
		{
			name:   "Empty answer",
			share:  []byte("share with empty answer"),
			answer: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt the share
			encryptedShare, salt, err := EncryptShare(tc.share, tc.answer)
			if err != nil {
				t.Fatalf("Failed to encrypt share: %v", err)
			}

			// Verify salt is not empty
			if len(salt) == 0 {
				t.Errorf("Salt should not be empty")
			}

			// Decrypt with correct answer
			decryptedShare, err := DecryptShare(encryptedShare, tc.answer, salt)
			if err != nil {
				t.Fatalf("Failed to decrypt share with correct answer: %v", err)
			}

			// Verify decrypted share
			if !bytes.Equal(decryptedShare, tc.share) {
				t.Errorf("Decrypted share doesn't match original. Got %v, expected %v",
					decryptedShare, tc.share)
			}

			// Try decrypting with wrong answer
			wrongAnswer := tc.answer + "wrong"
			_, err = DecryptShare(encryptedShare, wrongAnswer, salt)
			if err == nil {
				t.Errorf("Expected error when decrypting with wrong answer, but got none")
			}
		})
	}
}
