package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// WebAuthnConfig holds configuration for WebAuthn
type WebAuthnConfig struct {
	RPDisplayName string // Relying Party display name
	RPID          string // Relying Party ID (domain)
	RPOrigin      string // Relying Party origin (URL)
}

// WebAuthnService handles WebAuthn operations
type WebAuthnService struct {
	webAuthn *webauthn.WebAuthn
	repo     storage.Repository
	sessions map[string]*webauthn.SessionData // In-memory session store
	mutex    sync.Mutex                       // Mutex to protect the sessions map
}

// NewWebAuthnService creates a new WebAuthnService
func NewWebAuthnService(config WebAuthnConfig, repo storage.Repository) (*WebAuthnService, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: config.RPDisplayName,
		RPID:          config.RPID,
		RPOrigins:     []string{config.RPOrigin},
	}

	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn service: %w", err)
	}

	return &WebAuthnService{
		webAuthn: w,
		repo:     repo,
		sessions: make(map[string]*webauthn.SessionData),
	}, nil
}

// BeginRegistration starts the passkey registration process
func (s *WebAuthnService) BeginRegistration(ctx context.Context, user *models.User, response http.ResponseWriter) (*protocol.CredentialCreation, error) {
	// Get existing credentials for the user
	existingCredentials, err := s.getUserCredentials(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials: %w", err)
	}

	// Convert existing credentials to protocol.CredentialDescriptor
	excludeCredentials := make([]protocol.CredentialDescriptor, len(existingCredentials))
	for i, cred := range existingCredentials {
		excludeCredentials[i] = protocol.CredentialDescriptor{
			Type:         protocol.CredentialType("public-key"),
			CredentialID: cred.ID,
		}
	}

	// Create credential creation options
	options, sessionData, err := s.webAuthn.BeginRegistration(user)

	if err != nil {
		return nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	// Generate a unique session ID
	sessionID := fmt.Sprintf("%s-%d", user.ID, time.Now().UnixNano())

	// Store the session data in the sessions map
	s.mutex.Lock()
	s.sessions[sessionID] = sessionData
	s.mutex.Unlock()

	// Set a cookie with the session ID
	http.SetCookie(response, &http.Cookie{
		Name:     "webauthn_session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // Set to true in production with HTTPS
	})

	log.Printf("Stored WebAuthn session with ID: %s", sessionID)

	return options, nil
}

// FinishRegistration completes the passkey registration process
func (s *WebAuthnService) FinishRegistration(ctx context.Context, user *models.User, name string, response *http.Request) (*models.Passkey, error) {
	log.Printf("FinishRegistration called for user %s with name %s", user.Email, name)

	// Get the session ID from the cookie
	cookie, err := response.Cookie("webauthn_session_id")
	if err != nil {
		log.Printf("Error getting webauthn_session_id cookie: %v", err)
		return nil, fmt.Errorf("webauthn session cookie not found: %w", err)
	}
	sessionID := cookie.Value
	log.Printf("Found WebAuthn session ID in cookie: %s", sessionID)

	// Get the session data from the sessions map
	s.mutex.Lock()
	sessionData, ok := s.sessions[sessionID]
	if !ok {
		s.mutex.Unlock()
		log.Printf("Error: webauthn session data not found for ID: %s", sessionID)
		return nil, fmt.Errorf("webauthn session data not found for ID: %s", sessionID)
	}

	// Remove the session data from the map (one-time use)
	delete(s.sessions, sessionID)
	s.mutex.Unlock()

	log.Printf("Session data found with challenge: %v", sessionData.Challenge)

	// Log the request content type and method
	log.Printf("Request content type: %s, method: %s", response.Header.Get("Content-Type"), response.Method)

	// Dump the request body for debugging
	if response.Body != nil {
		bodyBytes, _ := io.ReadAll(response.Body)
		// Restore the body for further processing
		response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		log.Printf("Request body: %s", string(bodyBytes))
	}

	log.Printf("Calling webAuthn.FinishRegistration")
	// Create a new request with the same body for the WebAuthn library to parse
	// This is needed because the WebAuthn library expects a specific format for the request body

	// First, extract the credential data from the JSON request
	var requestData struct {
		Credential json.RawMessage `json:"credential"`
		Name       string          `json:"name"`
	}

	// Create a copy of the request body for parsing
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		return nil, fmt.Errorf("error reading request body: %w", err)
	}
	// Restore the body for further processing
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse the request data
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		log.Printf("Error parsing request data: %v", err)
		return nil, fmt.Errorf("error parsing request data: %w", err)
	}

	// Create a new request with just the credential data
	newRequest, err := http.NewRequest("POST", "/", bytes.NewReader(requestData.Credential))
	if err != nil {
		log.Printf("Error creating new request: %v", err)
		return nil, fmt.Errorf("error creating new request: %w", err)
	}
	newRequest.Header.Set("Content-Type", "application/json")

	// Finish registration with the new request
	// Extract the client data JSON to debug origin validation
	var credentialData struct {
		Response struct {
			ClientDataJSON string `json:"clientDataJSON"`
		} `json:"response"`
	}

	// Create a copy of the request body for parsing
	bodyBytes, readErr := io.ReadAll(newRequest.Body)
	if readErr != nil {
		log.Printf("Error reading request body for debugging: %v", readErr)
	}
	// Restore the body for further processing
	newRequest.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse the credential data for debugging
	if err := json.Unmarshal(bodyBytes, &credentialData); err != nil {
		log.Printf("Error parsing credential data for debugging: %v", err)
	} else {
		// Decode the client data JSON
		clientDataJSONBytes, err := base64.RawURLEncoding.DecodeString(credentialData.Response.ClientDataJSON)
		if err != nil {
			log.Printf("Error decoding client data JSON: %v", err)
		} else {
			// Parse the client data JSON
			var clientData struct {
				Type      string `json:"type"`
				Challenge string `json:"challenge"`
				Origin    string `json:"origin"`
			}
			if err := json.Unmarshal(clientDataJSONBytes, &clientData); err != nil {
				log.Printf("Error parsing client data JSON: %v", err)
			} else {
				log.Printf("Client data origin: %s, expected origin: %s", clientData.Origin, s.webAuthn.Config.RPOrigins)
				log.Printf("Client data challenge: %s, expected challenge: %s", clientData.Challenge, sessionData.Challenge)
			}
		}
	}

	// Finish registration with the parsed credential data
	credential, err := s.webAuthn.FinishRegistration(user, *sessionData, newRequest)
	if err != nil {
		log.Printf("Error in FinishRegistration: %v", err)
		return nil, fmt.Errorf("failed to finish registration: %w", err)
	}
	log.Printf("FinishRegistration successful, got credential ID: %v", credential.ID)

	// Convert transports to string slice
	transports := make([]string, len(credential.Transport))
	for i, t := range credential.Transport {
		transports[i] = string(t)
	}

	// Create a new passkey
	passkey := &models.Passkey{
		UserID:          user.ID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AAGUID:          credential.Authenticator.AAGUID,
		SignCount:       credential.Authenticator.SignCount,
		Name:            name,
		CreatedAt:       time.Now(),
		LastUsedAt:      time.Now(),
		Transports:      transports,
		AttestationType: credential.AttestationType,
	}

	// Save the passkey to the database
	if err := s.repo.CreatePasskey(ctx, passkey); err != nil {
		return nil, fmt.Errorf("failed to save passkey: %w", err)
	}

	return passkey, nil
}

// BeginLogin starts the passkey authentication process
func (s *WebAuthnService) BeginLogin(ctx context.Context, user *models.User, response http.ResponseWriter) (*protocol.CredentialAssertion, error) {
	log.Printf("BeginLogin called for user")

	// Get passkeys directly from the database
	passkeys, err := s.repo.ListPasskeysByUserID(ctx, user.ID)
	if err != nil {
		log.Printf("Error getting passkeys: %v", err)
		return nil, fmt.Errorf("failed to get passkeys: %w", err)
	}

	log.Printf("Found %d passkeys for user", len(passkeys))

	if len(passkeys) == 0 {
		// Let's check if there are any passkeys in the database at all
		allPasskeys, err := s.repo.ListPasskeys(ctx)
		if err != nil {
			log.Printf("Error listing all passkeys: %v", err)
		} else {
			log.Printf("Total passkeys in database: %d", len(allPasskeys))
			// Debug logging of individual passkeys removed for security
		}

		return nil, fmt.Errorf("no passkeys found for user")
	}

	// Create allowed credentials list directly from passkeys
	allowedCredentials := make([]protocol.CredentialDescriptor, len(passkeys))
	for i, passkey := range passkeys {
		// Debug logging removed for security

		allowedCredentials[i] = protocol.CredentialDescriptor{
			Type:         protocol.CredentialType("public-key"),
			CredentialID: passkey.CredentialID,
		}
	}

	// Log only the count of allowed credentials
	log.Printf("Allowed credentials: %d", len(allowedCredentials))

	// Create credential assertion options manually
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		log.Printf("Error generating challenge: %v", err)
		return nil, fmt.Errorf("error generating challenge: %w", err)
	}

	// Create session data
	sessionData := &webauthn.SessionData{
		Challenge:            base64.RawURLEncoding.EncodeToString(challenge),
		UserID:               []byte(user.ID),
		AllowedCredentialIDs: [][]byte{},
		UserVerification:     protocol.VerificationRequired,
	}

	// Add allowed credential IDs to session data
	for _, cred := range allowedCredentials {
		sessionData.AllowedCredentialIDs = append(sessionData.AllowedCredentialIDs, cred.CredentialID)
	}

	// Create credential assertion options
	options := &protocol.CredentialAssertion{
		Response: protocol.PublicKeyCredentialRequestOptions{
			Challenge:          challenge,
			Timeout:            60000, // 60 seconds
			RelyingPartyID:     s.webAuthn.Config.RPID,
			AllowedCredentials: allowedCredentials,
			UserVerification:   protocol.VerificationRequired,
		},
	}

	// Log the options for debugging
	log.Printf("Created credential assertion options with challenge: %v", base64.RawURLEncoding.EncodeToString(challenge))
	log.Printf("RPID: %s", s.webAuthn.Config.RPID)
	log.Printf("AllowCredentials: %d", len(allowedCredentials))

	// Generate a unique session ID
	sessionID := fmt.Sprintf("%s-%d", user.ID, time.Now().UnixNano())

	// Store the session data in the sessions map
	s.mutex.Lock()
	s.sessions[sessionID] = sessionData
	s.mutex.Unlock()

	// Set a cookie with the session ID
	http.SetCookie(response, &http.Cookie{
		Name:     "webauthn_session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // Set to true in production with HTTPS
	})

	log.Printf("Stored WebAuthn login session with ID: %s", sessionID)

	return options, nil
}

// FinishLogin completes the passkey authentication process
func (s *WebAuthnService) FinishLogin(ctx context.Context, user *models.User, response *http.Request) (*models.Passkey, error) {
	log.Printf("FinishLogin called for user")

	// Get the session ID from the cookie
	cookie, err := response.Cookie("webauthn_session_id")
	if err != nil {
		log.Printf("Error getting webauthn_session_id cookie: %v", err)
		return nil, fmt.Errorf("webauthn session cookie not found: %w", err)
	}
	sessionID := cookie.Value
	log.Printf("Found WebAuthn session ID in cookie: %s", sessionID)

	// Get the session data from the sessions map
	s.mutex.Lock()
	sessionData, ok := s.sessions[sessionID]
	if !ok {
		s.mutex.Unlock()
		log.Printf("Error: webauthn session data not found for ID: %s", sessionID)
		return nil, fmt.Errorf("webauthn session data not found for ID: %s", sessionID)
	}

	// Remove the session data from the map (one-time use)
	delete(s.sessions, sessionID)
	s.mutex.Unlock()

	log.Printf("Session data found with challenge: %v", sessionData.Challenge)

	// Create a new request with the same body for the WebAuthn library to parse
	// This is needed because the WebAuthn library expects a specific format for the request body

	// First, extract the credential data from the JSON request
	var requestData struct {
		Credential json.RawMessage `json:"credential"`
		Email      string          `json:"email"`
	}

	// Create a copy of the request body for parsing
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		return nil, fmt.Errorf("error reading request body: %w", err)
	}
	// Restore the body for further processing
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse the request data
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		log.Printf("Error parsing request data: %v", err)
		return nil, fmt.Errorf("error parsing request data: %w", err)
	}

	// Create a new request with just the credential data
	newRequest, err := http.NewRequest("POST", "/", bytes.NewReader(requestData.Credential))
	if err != nil {
		log.Printf("Error creating new request: %v", err)
		return nil, fmt.Errorf("error creating new request: %w", err)
	}
	newRequest.Header.Set("Content-Type", "application/json")

	// Finish login with the new request
	// First, let's try to parse the credential ID from the request
	var parsedCredential struct {
		ID       string          `json:"id"`
		RawID    string          `json:"rawId"`
		Type     string          `json:"type"`
		Response json.RawMessage `json:"response"`
	}

	if err := json.Unmarshal(requestData.Credential, &parsedCredential); err != nil {
		log.Printf("Error parsing credential: %v", err)
		return nil, fmt.Errorf("error parsing credential: %w", err)
	}

	// Debug logging removed for security

	// Try to get the passkey by the credential ID from the request
	credentialID, err := base64.RawURLEncoding.DecodeString(parsedCredential.ID)
	if err != nil {
		log.Printf("Error decoding credential ID: %v", err)
		return nil, fmt.Errorf("error decoding credential ID: %w", err)
	}

	// Debug logging removed for security

	// Try to get the passkey by the credential ID
	passkey, err := s.repo.GetPasskeyByCredentialID(ctx, credentialID)
	if err != nil {
		log.Printf("Error getting passkey by credential ID: %v", err)

		// Try the standard WebAuthn flow
		credential, err := s.webAuthn.FinishLogin(user, *sessionData, newRequest)
		if err != nil {
			log.Printf("Error in FinishLogin: %v", err)
			return nil, fmt.Errorf("failed to finish login: %w", err)
		}

		// Get the passkey from the database
		passkey, err = s.repo.GetPasskeyByCredentialID(ctx, credential.ID)
		if err != nil {
			log.Printf("Error getting passkey by credential ID after FinishLogin: %v", err)
			return nil, fmt.Errorf("failed to get passkey: %w", err)
		}
	}

	// Update the passkey with the new last used time
	passkey.LastUsedAt = time.Now()

	// Save the updated passkey to the database
	if err := s.repo.UpdatePasskey(ctx, passkey); err != nil {
		log.Printf("Failed to update passkey sign count: %v", err)
		// Continue anyway, this is not critical
	}

	return passkey, nil
}

// getUserCredentials gets the WebAuthn credentials for a user
func (s *WebAuthnService) getUserCredentials(ctx context.Context, user *models.User) ([]webauthn.Credential, error) {
	passkeys, err := s.repo.ListPasskeysByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list passkeys: %w", err)
	}

	credentials := make([]webauthn.Credential, len(passkeys))
	for i, passkey := range passkeys {
		// Convert string transports to protocol.AuthenticatorTransport
		transports := make([]protocol.AuthenticatorTransport, len(passkey.Transports))
		for j, t := range passkey.Transports {
			transports[j] = protocol.AuthenticatorTransport(t)
		}

		credentials[i] = webauthn.Credential{
			ID:              passkey.CredentialID,
			PublicKey:       passkey.PublicKey,
			AttestationType: passkey.AttestationType,
			Transport:       transports,
			Authenticator: webauthn.Authenticator{
				AAGUID:    passkey.AAGUID,
				SignCount: passkey.SignCount,
			},
		}
	}

	return credentials, nil
}

// CredentialIDToString converts a credential ID to a base64 string
func CredentialIDToString(credentialID []byte) string {
	return base64.RawURLEncoding.EncodeToString(credentialID)
}

// StringToCredentialID converts a base64 string to a credential ID
func StringToCredentialID(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
