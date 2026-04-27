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

// sessionEntry wraps a *webauthn.SessionData with a server-enforced expiry so
// abandoned begin-only flows cannot leak memory in the in-memory session map.
// The browser cookie's MaxAge does not delete server-side state on its own, so
// this struct is what the cleanup sweeper checks.
type sessionEntry struct {
	data    *webauthn.SessionData
	expires time.Time
}

// sessionTTL bounds how long a server-side WebAuthn session can sit in the
// in-memory map before it's swept. It matches the cookie MaxAge of 5 minutes.
const sessionTTL = 5 * time.Minute

// WebAuthnService handles WebAuthn operations
type WebAuthnService struct {
	webAuthn *webauthn.WebAuthn
	repo     storage.Repository
	sessions map[string]sessionEntry // In-memory session store
	mutex    sync.Mutex              // Mutex to protect the sessions map
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
		sessions: make(map[string]sessionEntry),
	}, nil
}

// storeSession records a WebAuthn session under the given ID with an expiry,
// which the next cleanupExpiredSessions sweep will use to evict abandoned
// entries.
func (s *WebAuthnService) storeSession(id string, data *webauthn.SessionData) {
	s.mutex.Lock()
	s.sessions[id] = sessionEntry{data: data, expires: time.Now().Add(sessionTTL)}
	s.mutex.Unlock()
}

// takeSession atomically removes and returns the session data for id, mirroring
// the original "pop on consume" semantics. Returns nil if the entry is missing
// or has expired.
func (s *WebAuthnService) takeSession(id string) *webauthn.SessionData {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	entry, ok := s.sessions[id]
	if !ok {
		return nil
	}
	delete(s.sessions, id)
	if time.Now().After(entry.expires) {
		return nil
	}
	return entry.data
}

// cleanupExpiredSessions removes session entries whose expiry has passed. Each
// Begin* call invokes this so the conditional-UI auto-fetch (which posts on
// every login page load, including from bots) cannot grow s.sessions
// unboundedly.
func (s *WebAuthnService) cleanupExpiredSessions() {
	now := time.Now()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for id, entry := range s.sessions {
		if now.After(entry.expires) {
			delete(s.sessions, id)
		}
	}
}

// BeginRegistration starts the passkey registration process. The request is
// used to derive the Secure cookie flag from r.TLS so the cookie is only
// marked Secure when served over HTTPS.
func (s *WebAuthnService) BeginRegistration(ctx context.Context, user *models.User, r *http.Request, response http.ResponseWriter) (*protocol.CredentialCreation, error) {
	if user == nil {
		return nil, fmt.Errorf("user is required")
	}

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

	// Request a discoverable (resident-key) credential when the authenticator
	// supports it so the username-less login flow can find it via userHandle.
	// UV is required at registration to match the login policy — both legacy
	// BeginLogin and BeginDiscoverableLogin require UV. Registering without UV
	// would silently produce a passkey that can never authenticate.
	authenticatorSelection := protocol.AuthenticatorSelection{
		ResidentKey:      protocol.ResidentKeyRequirementPreferred,
		UserVerification: protocol.VerificationRequired,
	}

	// Create credential creation options
	options, sessionData, err := s.webAuthn.BeginRegistration(
		user,
		webauthn.WithAuthenticatorSelection(authenticatorSelection),
		webauthn.WithExclusions(excludeCredentials),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	// Generate a unique session ID
	sessionID := fmt.Sprintf("%s-%d", user.ID, time.Now().UnixNano())

	s.cleanupExpiredSessions()
	s.storeSession(sessionID, sessionData)

	// Set a cookie with the session ID
	http.SetCookie(response, &http.Cookie{
		Name:     "webauthn_session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r != nil && r.TLS != nil,
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

	sessionData := s.takeSession(sessionID)
	if sessionData == nil {
		log.Printf("Error: webauthn session data not found for ID: %s", sessionID)
		return nil, fmt.Errorf("webauthn session data not found for ID: %s", sessionID)
	}

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
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
	}

	// Save the passkey to the database
	if err := s.repo.CreatePasskey(ctx, passkey); err != nil {
		return nil, fmt.Errorf("failed to save passkey: %w", err)
	}

	return passkey, nil
}

// BeginDiscoverableLogin starts a username-less passkey authentication using
// client-side discoverable credentials (resident keys). No AllowedCredentials
// are populated — the authenticator presents a chooser of every passkey scoped
// to the relying party and returns the userHandle so the server can resolve
// the user during FinishDiscoverableLogin. The request is used to derive the
// Secure cookie flag from r.TLS. UserVerification is required because this is
// a passwordless flow — a credential that only proves user presence must not
// complete login.
func (s *WebAuthnService) BeginDiscoverableLogin(r *http.Request, w http.ResponseWriter) (*protocol.CredentialAssertion, error) {
	s.cleanupExpiredSessions()

	options, sessionData, err := s.webAuthn.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to begin discoverable login: %w", err)
	}

	// Use a cryptographically-random session ID. Unlike the legacy login/register
	// flows which prefix with user.ID, the discoverable flow has no user yet, so
	// a UnixNano-only ID would let two simultaneous calls (the conditional-UI
	// auto-fetch fires on every login page load) collide and overwrite each
	// other's session data.
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("failed to generate session id: %w", err)
	}
	sessionID := "discover-" + base64.RawURLEncoding.EncodeToString(idBytes)

	s.storeSession(sessionID, sessionData)

	http.SetCookie(w, &http.Cookie{
		Name:     "webauthn_session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r != nil && r.TLS != nil,
	})

	return options, nil
}

// discoverableUser wraps a *models.User together with the user's WebAuthn
// credentials so it satisfies the webauthn.User interface during validation.
// The base *models.User intentionally returns no credentials from
// WebAuthnCredentials(); the discoverable login flow loads them on demand from
// the repository and attaches them here so the library can verify the
// signature. WebAuthnID/Name/DisplayName/Icon are inherited from the embedded
// user; only WebAuthnCredentials is overridden.
type discoverableUser struct {
	*models.User
	credentials []webauthn.Credential
}

func (d *discoverableUser) WebAuthnCredentials() []webauthn.Credential { return d.credentials }

// FinishDiscoverableLogin completes a username-less passkey authentication.
// The userHandle returned by the authenticator (== user.ID bytes) is used to
// resolve the user; the library then verifies the signature against that
// user's stored credentials. The session cookie is consumed regardless of
// outcome so a stale cookie cannot be replayed.
func (s *WebAuthnService) FinishDiscoverableLogin(ctx context.Context, r *http.Request) (*models.User, *models.Passkey, error) {
	cookie, err := r.Cookie("webauthn_session_id")
	if err != nil {
		return nil, nil, fmt.Errorf("webauthn session cookie not found: %w", err)
	}
	sessionID := cookie.Value

	sessionData := s.takeSession(sessionID)
	if sessionData == nil {
		return nil, nil, fmt.Errorf("webauthn session data not found for ID: %s", sessionID)
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading request body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var requestData struct {
		Credential json.RawMessage `json:"credential"`
	}
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		return nil, nil, fmt.Errorf("error parsing request data: %w", err)
	}
	if len(requestData.Credential) == 0 {
		return nil, nil, fmt.Errorf("missing credential field in request body")
	}

	newRequest, err := http.NewRequest("POST", "/", bytes.NewReader(requestData.Credential))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating new request: %w", err)
	}
	newRequest.Header.Set("Content-Type", "application/json")

	var resolvedUser *models.User
	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		userID := string(userHandle)
		u, err := s.repo.GetUserByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("user not found for handle %q: %w", userID, err)
		}
		creds, err := s.getUserCredentials(ctx, u)
		if err != nil {
			return nil, fmt.Errorf("failed to load credentials for user %q: %w", userID, err)
		}
		resolvedUser = u
		return &discoverableUser{User: u, credentials: creds}, nil
	}

	credential, err := s.webAuthn.FinishDiscoverableLogin(handler, *sessionData, newRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to finish discoverable login: %w", err)
	}

	passkey, err := s.repo.GetPasskeyByCredentialID(ctx, credential.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to look up passkey by credential ID: %w", err)
	}

	// Defense in depth: the library validates the assertion against credentials
	// loaded from resolvedUser, but the post-success lookup above fetches the
	// passkey by credential_id from a non-unique index. If a duplicate
	// credential_id ever existed for a different user, we would otherwise issue
	// a session for resolvedUser while attributing the audit log to a different
	// passkey owner.
	if passkey.UserID != resolvedUser.ID {
		return nil, nil, fmt.Errorf("passkey owner mismatch: credential_id resolves to passkey owned by another user")
	}

	passkey.LastUsedAt = time.Now()
	passkey.SignCount = credential.Authenticator.SignCount
	passkey.BackupState = credential.Flags.BackupState
	if err := s.repo.UpdatePasskey(ctx, passkey); err != nil {
		log.Printf("failed to update passkey after discoverable login: %v", err)
	}

	return resolvedUser, passkey, nil
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
			Flags: webauthn.CredentialFlags{
				BackupEligible: passkey.BackupEligible,
				BackupState:    passkey.BackupState,
			},
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
