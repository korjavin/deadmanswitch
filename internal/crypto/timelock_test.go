package crypto

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimelockData(t *testing.T) {
	// Test serialization and deserialization of TimelockData
	questions := []QuestionData{
		{
			Question:       "What is your favorite color?",
			Salt:           []byte{0x01, 0x02, 0x03},
			EncryptedShare: []byte{0x04, 0x05, 0x06},
		},
		{
			Question:       "What is your pet's name?",
			Salt:           []byte{0x07, 0x08, 0x09},
			EncryptedShare: []byte{0x0A, 0x0B, 0x0C},
		},
	}

	data := TimelockData{
		Questions: questions,
		Threshold: 2,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal TimelockData: %v", err)
	}

	// Deserialize from JSON
	var deserializedData TimelockData
	if err := json.Unmarshal(jsonData, &deserializedData); err != nil {
		t.Fatalf("Failed to unmarshal TimelockData: %v", err)
	}

	// Verify deserialized data
	if deserializedData.Threshold != data.Threshold {
		t.Errorf("Threshold doesn't match. Got %d, expected %d", deserializedData.Threshold, data.Threshold)
	}

	if len(deserializedData.Questions) != len(data.Questions) {
		t.Fatalf("Questions length doesn't match. Got %d, expected %d", len(deserializedData.Questions), len(data.Questions))
	}

	for i, q := range deserializedData.Questions {
		if q.Question != data.Questions[i].Question {
			t.Errorf("Question %d text doesn't match. Got %s, expected %s", i, q.Question, data.Questions[i].Question)
		}
	}
}

func TestCalculateDrandRound(t *testing.T) {
	// Skip this test in automated environments
	t.Skip("Skipping test that requires network access to drand")

	// Test calculating drand round for a future time
	futureTime := time.Now().Add(24 * time.Hour)
	round, err := CalculateDrandRound(futureTime)
	if err != nil {
		t.Fatalf("Failed to calculate drand round: %v", err)
	}

	// Verify round is greater than 0
	if round <= 0 {
		t.Errorf("Expected round to be greater than 0, got %d", round)
	}

	// Test calculating drand round for current time
	currentTime := time.Now()
	currentRound, err := CalculateDrandRound(currentTime)
	if err != nil {
		t.Fatalf("Failed to calculate drand round for current time: %v", err)
	}

	// Verify future round is greater than current round
	if round <= currentRound {
		t.Errorf("Expected future round (%d) to be greater than current round (%d)", round, currentRound)
	}
}

func TestTimelockEncryptDecrypt(t *testing.T) {
	// Skip this test in automated environments
	t.Skip("Skipping test that requires network access to drand")

	// Test data
	testData := []byte("This is a test message for timelock encryption")

	// Get a round in the past to ensure it's already available
	pastTime := time.Now().Add(-1 * time.Hour)
	round, err := CalculateDrandRound(pastTime)
	if err != nil {
		t.Fatalf("Failed to calculate drand round: %v", err)
	}

	// Encrypt with timelock
	encrypted, err := TimelockEncrypt(testData, round)
	if err != nil {
		t.Fatalf("Failed to encrypt with timelock: %v", err)
	}

	// Decrypt with timelock
	decrypted, err := TimelockDecrypt(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt with timelock: %v", err)
	}

	// Verify decrypted data
	if string(decrypted) != string(testData) {
		t.Errorf("Decrypted data doesn't match original. Got %s, expected %s", string(decrypted), string(testData))
	}
}

func TestEncryptDecryptQuestions(t *testing.T) {
	// Skip this test in automated environments
	t.Skip("Skipping test that requires network access to drand")

	// Test data
	questions := []QuestionData{
		{
			Question:       "What is your favorite color?",
			Salt:           []byte{0x01, 0x02, 0x03},
			EncryptedShare: []byte{0x04, 0x05, 0x06},
		},
		{
			Question:       "What is your pet's name?",
			Salt:           []byte{0x07, 0x08, 0x09},
			EncryptedShare: []byte{0x0A, 0x0B, 0x0C},
		},
	}
	threshold := 2
	deadline := time.Now().Add(-1 * time.Hour) // Use past time to ensure round is available

	// Encrypt questions
	encryptedBlob, round, err := EncryptQuestions(questions, threshold, deadline)
	if err != nil {
		t.Fatalf("Failed to encrypt questions: %v", err)
	}

	// Verify round is greater than 0
	if round <= 0 {
		t.Errorf("Expected round to be greater than 0, got %d", round)
	}

	// Decrypt questions
	decryptedData, err := DecryptQuestions(encryptedBlob)
	if err != nil {
		t.Fatalf("Failed to decrypt questions: %v", err)
	}

	// Verify decrypted data
	if decryptedData.Threshold != threshold {
		t.Errorf("Threshold doesn't match. Got %d, expected %d", decryptedData.Threshold, threshold)
	}

	if len(decryptedData.Questions) != len(questions) {
		t.Fatalf("Questions length doesn't match. Got %d, expected %d", len(decryptedData.Questions), len(questions))
	}

	for i, q := range decryptedData.Questions {
		if q.Question != questions[i].Question {
			t.Errorf("Question %d text doesn't match. Got %s, expected %s", i, q.Question, questions[i].Question)
		}
	}
}
