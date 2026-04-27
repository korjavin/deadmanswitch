package auth

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// TestNewWebAuthnService tests creating a new WebAuthnService
func TestNewWebAuthnService(t *testing.T) {
	repo := storage.NewMockRepository()

	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}

	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	if service == nil {
		t.Fatal("Expected non-nil WebAuthnService")
	}

	if service.webAuthn == nil {
		t.Error("Expected non-nil webAuthn in service")
	}

	if service.repo != repo {
		t.Error("Expected repo in service to match input repo")
	}

	if service.sessions == nil {
		t.Error("Expected non-nil sessions map in service")
	}
}

// TestCredentialIDConversions tests the credential ID conversion functions
func TestCredentialIDConversions(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"Empty", []byte{}},
		{"Simple", []byte("test-credential")},
		{"Complex", []byte{0, 1, 2, 3, 4, 5, 255, 254, 253}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Convert to string
			str := CredentialIDToString(test.input)

			// Convert back to byte slice
			result, err := StringToCredentialID(str)
			if err != nil {
				t.Fatalf("StringToCredentialID failed: %v", err)
			}

			// Check that the result matches the input
			if len(result) != len(test.input) {
				t.Errorf("Expected result length %d, got %d", len(test.input), len(result))
			}

			match := true
			for i, b := range test.input {
				if i >= len(result) || result[i] != b {
					match = false
					break
				}
			}

			if !match {
				t.Errorf("Result does not match input. Input: %v, Result: %v", test.input, result)
			}
		})
	}
}

// TestStringToCredentialIDError tests error cases for StringToCredentialID
func TestStringToCredentialIDError(t *testing.T) {
	// Invalid base64 strings
	invalidInputs := []string{
		"!@#$%", // Not base64
		"a===",  // Invalid padding
	}

	for _, input := range invalidInputs {
		_, err := StringToCredentialID(input)
		if err == nil {
			t.Errorf("Expected error for invalid input: %s", input)
		}
	}
}

// TestGetUserCredentials tests the getUserCredentials method
func TestGetUserCredentials(t *testing.T) {
	// Create mock repository
	repo := storage.NewMockRepository()

	// Create a WebAuthnService
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}

	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	// Create a test user
	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
	}

	// Test with no passkeys
	ctx := context.Background()
	credentials, err := service.getUserCredentials(ctx, user)
	if err != nil {
		t.Fatalf("getUserCredentials failed with no passkeys: %v", err)
	}
	if len(credentials) != 0 {
		t.Errorf("Expected 0 credentials with no passkeys, got %d", len(credentials))
	}

	// Add some passkeys to the repository
	passkey1 := &models.Passkey{
		ID:              "passkey1",
		UserID:          user.ID,
		CredentialID:    []byte("credential1"),
		PublicKey:       []byte("publickey1"),
		AAGUID:          []byte("aaguid1"),
		SignCount:       1,
		AttestationType: "none",
		Transports:      []string{"internal", "usb"},
		CreatedAt:       time.Now(),
		LastUsedAt:      time.Now(),
	}

	passkey2 := &models.Passkey{
		ID:              "passkey2",
		UserID:          user.ID,
		CredentialID:    []byte("credential2"),
		PublicKey:       []byte("publickey2"),
		AAGUID:          []byte("aaguid2"),
		SignCount:       2,
		AttestationType: "direct",
		Transports:      []string{"ble"},
		CreatedAt:       time.Now(),
		LastUsedAt:      time.Now(),
	}

	repo.Passkeys = append(repo.Passkeys, passkey1, passkey2)

	// Test with passkeys
	credentials, err = service.getUserCredentials(ctx, user)
	if err != nil {
		t.Fatalf("getUserCredentials failed with passkeys: %v", err)
	}
	if len(credentials) != 2 {
		t.Errorf("Expected 2 credentials with passkeys, got %d", len(credentials))
	}

	// Check the first credential
	if !byteSliceEqual(credentials[0].ID, passkey1.CredentialID) {
		t.Errorf("Expected credential ID to match passkey1")
	}
	if !byteSliceEqual(credentials[0].PublicKey, passkey1.PublicKey) {
		t.Errorf("Expected public key to match passkey1")
	}
	if credentials[0].AttestationType != passkey1.AttestationType {
		t.Errorf("Expected attestation type to match passkey1")
	}
	if len(credentials[0].Transport) != len(passkey1.Transports) {
		t.Errorf("Expected transport count to match passkey1")
	}
	if !byteSliceEqual(credentials[0].Authenticator.AAGUID, passkey1.AAGUID) {
		t.Errorf("Expected AAGUID to match passkey1")
	}
	if credentials[0].Authenticator.SignCount != passkey1.SignCount {
		t.Errorf("Expected sign count to match passkey1")
	}
}

// Helper function to compare byte slices
func byteSliceEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// TestBeginRegistrationResidentKeyPreferred ensures new passkey enrollments
// request a discoverable (resident-key) credential so the username-less login
// flow can find them via userHandle.
func TestBeginRegistrationResidentKeyPreferred(t *testing.T) {
	repo := storage.NewMockRepository()
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}
	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
	}

	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	options, err := service.BeginRegistration(context.Background(), user, req, rw)
	if err != nil {
		t.Fatalf("BeginRegistration failed: %v", err)
	}
	if options == nil {
		t.Fatal("Expected non-nil options")
	}

	got := options.Response.AuthenticatorSelection.ResidentKey
	if got != protocol.ResidentKeyRequirementPreferred {
		t.Errorf("Expected ResidentKey %q, got %q",
			protocol.ResidentKeyRequirementPreferred, got)
	}
	if options.Response.AuthenticatorSelection.UserVerification != protocol.VerificationRequired {
		t.Errorf("Expected UserVerification %q, got %q",
			protocol.VerificationRequired,
			options.Response.AuthenticatorSelection.UserVerification)
	}
}

// TestBeginRegistrationNilUser ensures BeginRegistration returns an error
// rather than panicking when called with a nil user.
func TestBeginRegistrationNilUser(t *testing.T) {
	repo := storage.NewMockRepository()
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}
	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	if _, err := service.BeginRegistration(context.Background(), nil, req, rw); err == nil {
		t.Fatal("Expected error for nil user, got nil")
	}
}

// TestBeginDiscoverableLogin asserts the returned options have an empty
// allowCredentials list, a non-empty challenge, the configured RPID, and that
// UserVerification is required for this passwordless flow so a credential that
// only proves user presence cannot complete login.
func TestBeginDiscoverableLogin(t *testing.T) {
	repo := storage.NewMockRepository()
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}
	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	rw := httptest.NewRecorder()
	options, err := service.BeginDiscoverableLogin(httptest.NewRequest("POST", "/login/passkey/discover/begin", nil), rw)
	if err != nil {
		t.Fatalf("BeginDiscoverableLogin failed: %v", err)
	}
	if options == nil {
		t.Fatal("Expected non-nil options")
	}

	if len(options.Response.AllowedCredentials) != 0 {
		t.Errorf("Expected empty AllowedCredentials, got %d entries",
			len(options.Response.AllowedCredentials))
	}
	if len(options.Response.Challenge) == 0 {
		t.Error("Expected non-empty challenge")
	}
	if options.Response.RelyingPartyID != config.RPID {
		t.Errorf("Expected RPID %q, got %q", config.RPID, options.Response.RelyingPartyID)
	}
	if options.Response.UserVerification != protocol.VerificationRequired {
		t.Errorf("Expected UserVerification %q, got %q",
			protocol.VerificationRequired, options.Response.UserVerification)
	}
}

// TestBeginDiscoverableLoginSessionUVRequired asserts the SessionData stored
// in the in-memory map carries UserVerification=required, which is what the
// library checks during FinishDiscoverableLogin to enforce the UV bit. Without
// this, a credential that only proves user presence would complete a
// passwordless login.
func TestBeginDiscoverableLoginSessionUVRequired(t *testing.T) {
	service, sessionID := newTestServiceWithSession(t)

	service.mutex.Lock()
	entry, ok := service.sessions[sessionID]
	service.mutex.Unlock()
	if !ok {
		t.Fatalf("expected session %q stored", sessionID)
	}
	if entry.data.UserVerification != protocol.VerificationRequired {
		t.Errorf("session UserVerification = %q, want %q",
			entry.data.UserVerification, protocol.VerificationRequired)
	}
}

// TestCleanupExpiredSessions verifies the sweeper drops entries whose expiry
// has passed and leaves fresh entries alone. This is what keeps abandoned
// conditional-UI begin calls from leaking memory: the login page now fires
// /discover/begin on every load, so without expiry-based cleanup the
// in-memory sessions map would grow without bound.
func TestCleanupExpiredSessions(t *testing.T) {
	repo := storage.NewMockRepository()
	service, err := NewWebAuthnService(WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	service.mutex.Lock()
	service.sessions["fresh"] = sessionEntry{
		data:    &webauthn.SessionData{Challenge: "fresh"},
		expires: time.Now().Add(time.Minute),
	}
	service.sessions["stale"] = sessionEntry{
		data:    &webauthn.SessionData{Challenge: "stale"},
		expires: time.Now().Add(-time.Minute),
	}
	service.mutex.Unlock()

	service.cleanupExpiredSessions()

	service.mutex.Lock()
	_, freshOK := service.sessions["fresh"]
	_, staleOK := service.sessions["stale"]
	service.mutex.Unlock()

	if !freshOK {
		t.Error("expected fresh session to remain after cleanup")
	}
	if staleOK {
		t.Error("expected stale session to be evicted by cleanup")
	}
}

// TestTakeSessionExpired ensures takeSession refuses to return data for an
// entry past its expiry, even before the sweeper has run. This keeps an
// attacker from replaying a long-abandoned cookie if the cleanup goroutine is
// slow.
func TestTakeSessionExpired(t *testing.T) {
	repo := storage.NewMockRepository()
	service, err := NewWebAuthnService(WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	service.mutex.Lock()
	service.sessions["expired"] = sessionEntry{
		data:    &webauthn.SessionData{Challenge: "expired"},
		expires: time.Now().Add(-time.Second),
	}
	service.mutex.Unlock()

	if got := service.takeSession("expired"); got != nil {
		t.Errorf("expected nil for expired session, got %+v", got)
	}
	service.mutex.Lock()
	_, stillThere := service.sessions["expired"]
	service.mutex.Unlock()
	if stillThere {
		t.Error("expected expired session to be removed by takeSession")
	}
}

// TestGetUserCredentialsRestoresFlags verifies that BackupEligible/BackupState
// stored on the Passkey are propagated into the webauthn.Credential struct.
// go-webauthn rejects login when the stored BackupEligible disagrees with the
// assertion, so cloud-synced passkeys (iCloud Keychain, Google Password
// Manager, 1Password) require these flags to be persisted and restored.
func TestGetUserCredentialsRestoresFlags(t *testing.T) {
	repo := storage.NewMockRepository()
	service, err := NewWebAuthnService(WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	user := &models.User{ID: "user-flags", Email: "flags@example.com"}
	repo.Passkeys = append(repo.Passkeys, &models.Passkey{
		ID:             "pk-flags",
		UserID:         user.ID,
		CredentialID:   []byte("cred-flags"),
		PublicKey:      []byte("pk"),
		BackupEligible: true,
		BackupState:    true,
	})

	creds, err := service.getUserCredentials(context.Background(), user)
	if err != nil {
		t.Fatalf("getUserCredentials failed: %v", err)
	}
	if len(creds) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(creds))
	}
	if !creds[0].Flags.BackupEligible {
		t.Error("expected restored BackupEligible=true")
	}
	if !creds[0].Flags.BackupState {
		t.Error("expected restored BackupState=true")
	}
}

// TestBeginDiscoverableLoginSetsSessionCookie ensures BeginDiscoverableLogin
// sets the webauthn_session_id cookie and stores matching session data in the
// in-memory map.
func TestBeginDiscoverableLoginSetsSessionCookie(t *testing.T) {
	repo := storage.NewMockRepository()
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}
	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	rw := httptest.NewRecorder()
	if _, err := service.BeginDiscoverableLogin(httptest.NewRequest("POST", "/login/passkey/discover/begin", nil), rw); err != nil {
		t.Fatalf("BeginDiscoverableLogin failed: %v", err)
	}

	var sessionCookie *http.Cookie
	for _, c := range rw.Result().Cookies() {
		if c.Name == "webauthn_session_id" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("webauthn_session_id cookie not set")
	}
	if sessionCookie.Value == "" {
		t.Error("Expected non-empty session cookie value")
	}
	if !sessionCookie.HttpOnly {
		t.Error("Expected HttpOnly cookie")
	}
	if sessionCookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("Expected SameSite=Strict, got %v", sessionCookie.SameSite)
	}
	if sessionCookie.MaxAge != 300 {
		t.Errorf("Expected MaxAge=300, got %d", sessionCookie.MaxAge)
	}
	if sessionCookie.Path != "/" {
		t.Errorf("Expected Path=/, got %q", sessionCookie.Path)
	}

	service.mutex.Lock()
	stored, ok := service.sessions[sessionCookie.Value]
	service.mutex.Unlock()
	if !ok {
		t.Fatalf("Expected session %q to be stored in sessions map", sessionCookie.Value)
	}
	if stored.data == nil {
		t.Fatal("Expected non-nil stored session data")
	}
	if len(stored.data.AllowedCredentialIDs) != 0 {
		t.Errorf("Expected empty AllowedCredentialIDs in session, got %d",
			len(stored.data.AllowedCredentialIDs))
	}
}

// newTestServiceWithSession creates a WebAuthnService and stores a fresh
// discoverable session so tests can drive FinishDiscoverableLogin past the
// cookie/session lookup.
func newTestServiceWithSession(t *testing.T) (*WebAuthnService, string) {
	t.Helper()
	repo := storage.NewMockRepository()
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}
	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	rw := httptest.NewRecorder()
	if _, err := service.BeginDiscoverableLogin(httptest.NewRequest("POST", "/login/passkey/discover/begin", nil), rw); err != nil {
		t.Fatalf("BeginDiscoverableLogin failed: %v", err)
	}
	var sessionID string
	for _, c := range rw.Result().Cookies() {
		if c.Name == "webauthn_session_id" {
			sessionID = c.Value
			break
		}
	}
	if sessionID == "" {
		t.Fatal("expected webauthn_session_id cookie to be set")
	}
	return service, sessionID
}

// TestFinishDiscoverableLoginMissingCookie verifies a request with no session
// cookie returns an error and does not panic.
func TestFinishDiscoverableLoginMissingCookie(t *testing.T) {
	repo := storage.NewMockRepository()
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}
	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	req := httptest.NewRequest("POST", "/login/passkey/discover/finish",
		strings.NewReader(`{"credential":{}}`))
	user, passkey, err := service.FinishDiscoverableLogin(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing session cookie, got nil")
	}
	if user != nil || passkey != nil {
		t.Errorf("expected nil user and passkey on error, got user=%v passkey=%v", user, passkey)
	}
}

// TestFinishDiscoverableLoginMissingSession verifies a request with a cookie
// that does not match any in-memory session returns an error.
func TestFinishDiscoverableLoginMissingSession(t *testing.T) {
	repo := storage.NewMockRepository()
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}
	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	req := httptest.NewRequest("POST", "/login/passkey/discover/finish",
		strings.NewReader(`{"credential":{}}`))
	req.AddCookie(&http.Cookie{Name: "webauthn_session_id", Value: "does-not-exist"})

	if _, _, err := service.FinishDiscoverableLogin(context.Background(), req); err == nil {
		t.Fatal("expected error for missing session, got nil")
	}
}

// TestFinishDiscoverableLoginConsumesSession verifies the session is removed
// from the in-memory map after a single call so a stale cookie cannot be
// replayed (even when the underlying assertion fails to parse).
func TestFinishDiscoverableLoginConsumesSession(t *testing.T) {
	service, sessionID := newTestServiceWithSession(t)

	req := httptest.NewRequest("POST", "/login/passkey/discover/finish",
		strings.NewReader(`{"credential":{}}`))
	req.AddCookie(&http.Cookie{Name: "webauthn_session_id", Value: sessionID})

	// First call consumes the session even though the assertion is bogus.
	if _, _, err := service.FinishDiscoverableLogin(context.Background(), req); err == nil {
		t.Fatal("expected error from bogus assertion, got nil")
	}

	service.mutex.Lock()
	_, stillThere := service.sessions[sessionID]
	service.mutex.Unlock()
	if stillThere {
		t.Errorf("expected session %q to be removed after FinishDiscoverableLogin", sessionID)
	}

	// Second call must fail because the session has been consumed.
	req2 := httptest.NewRequest("POST", "/login/passkey/discover/finish",
		strings.NewReader(`{"credential":{}}`))
	req2.AddCookie(&http.Cookie{Name: "webauthn_session_id", Value: sessionID})
	if _, _, err := service.FinishDiscoverableLogin(context.Background(), req2); err == nil {
		t.Fatal("expected error on session replay, got nil")
	}
}

// errReader is an io.ReadCloser that always fails on Read; used to drive the
// io.ReadAll error branch in FinishDiscoverableLogin.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) { return 0, errReadFailed }
func (errReader) Close() error               { return nil }

var errReadFailed = &readErr{msg: "synthetic read failure"}

type readErr struct{ msg string }

func (e *readErr) Error() string { return e.msg }

// TestFinishDiscoverableLoginBodyReadError verifies that when the request body
// fails on read, the function returns an error rather than panicking.
func TestFinishDiscoverableLoginBodyReadError(t *testing.T) {
	service, sessionID := newTestServiceWithSession(t)

	req := httptest.NewRequest("POST", "/login/passkey/discover/finish", nil)
	req.Body = errReader{}
	req.AddCookie(&http.Cookie{Name: "webauthn_session_id", Value: sessionID})

	if _, _, err := service.FinishDiscoverableLogin(context.Background(), req); err == nil {
		t.Fatal("expected error for failing body read, got nil")
	}
}

// TestFinishDiscoverableLoginMalformedBody verifies a request body that is not
// valid JSON produces an error rather than a panic.
func TestFinishDiscoverableLoginMalformedBody(t *testing.T) {
	service, sessionID := newTestServiceWithSession(t)

	req := httptest.NewRequest("POST", "/login/passkey/discover/finish",
		strings.NewReader("not-json"))
	req.AddCookie(&http.Cookie{Name: "webauthn_session_id", Value: sessionID})

	if _, _, err := service.FinishDiscoverableLogin(context.Background(), req); err == nil {
		t.Fatal("expected error for malformed JSON body, got nil")
	}
}

// TestFinishDiscoverableLoginMissingCredentialField verifies a request body
// missing the `credential` wrapper field returns an error.
func TestFinishDiscoverableLoginMissingCredentialField(t *testing.T) {
	service, sessionID := newTestServiceWithSession(t)

	req := httptest.NewRequest("POST", "/login/passkey/discover/finish",
		strings.NewReader(`{}`))
	req.AddCookie(&http.Cookie{Name: "webauthn_session_id", Value: sessionID})

	if _, _, err := service.FinishDiscoverableLogin(context.Background(), req); err == nil {
		t.Fatal("expected error for missing credential field, got nil")
	}
}

// TestFinishDiscoverableLoginInvalidAssertion verifies that when the wrapped
// credential JSON cannot be parsed as a WebAuthn assertion (e.g. arbitrary
// JSON), the library returns an error and the function surfaces it. This
// exercises the same code path a tampered or unsigned-by-the-authenticator
// assertion would hit before reaching signature verification.
func TestFinishDiscoverableLoginInvalidAssertion(t *testing.T) {
	service, sessionID := newTestServiceWithSession(t)

	// The wrapper field is present but the inner credential is empty/invalid;
	// the library's ParseCredentialRequestResponse must reject this.
	body := `{"credential":{"id":"","type":"public-key"}}`
	req := httptest.NewRequest("POST", "/login/passkey/discover/finish",
		strings.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "webauthn_session_id", Value: sessionID})

	user, passkey, err := service.FinishDiscoverableLogin(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid assertion, got nil")
	}
	if user != nil || passkey != nil {
		t.Errorf("expected nil user and passkey on error, got user=%v passkey=%v", user, passkey)
	}
}

// TestDiscoverableUserAdapter verifies the discoverableUser wrapper passes
// through identity methods to the embedded *models.User and exposes the
// preloaded credentials slice. Without these direct assertions the wrapper is
// only exercised indirectly by paths that need a real authenticator response.
func TestDiscoverableUserAdapter(t *testing.T) {
	user := &models.User{ID: "user-42", Email: "alice@example.com"}
	creds := []webauthn.Credential{
		{ID: []byte("cred-a"), PublicKey: []byte("pk-a")},
		{ID: []byte("cred-b"), PublicKey: []byte("pk-b")},
	}
	d := &discoverableUser{User: user, credentials: creds}

	if got := d.WebAuthnID(); string(got) != user.ID {
		t.Errorf("WebAuthnID = %q, want %q", string(got), user.ID)
	}
	if got := d.WebAuthnName(); got != user.Email {
		t.Errorf("WebAuthnName = %q, want %q", got, user.Email)
	}
	if got := d.WebAuthnDisplayName(); got != user.Email {
		t.Errorf("WebAuthnDisplayName = %q, want %q", got, user.Email)
	}
	if got := d.WebAuthnIcon(); got != "" {
		t.Errorf("WebAuthnIcon = %q, want empty", got)
	}
	got := d.WebAuthnCredentials()
	if len(got) != len(creds) {
		t.Fatalf("WebAuthnCredentials returned %d creds, want %d", len(got), len(creds))
	}
	for i := range creds {
		if !byteSliceEqual(got[i].ID, creds[i].ID) {
			t.Errorf("credential[%d].ID = %v, want %v", i, got[i].ID, creds[i].ID)
		}
	}
}

// TestBeginDiscoverableLoginCookieSecureWithTLS asserts the webauthn_session_id
// cookie is marked Secure when the request was received over TLS.
func TestBeginDiscoverableLoginCookieSecureWithTLS(t *testing.T) {
	repo := storage.NewMockRepository()
	service, err := NewWebAuthnService(WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "https://localhost",
	}, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	req := httptest.NewRequest("POST", "https://localhost/login/passkey/discover/begin", nil)
	req.TLS = &tls.ConnectionState{}
	rw := httptest.NewRecorder()
	if _, err := service.BeginDiscoverableLogin(req, rw); err != nil {
		t.Fatalf("BeginDiscoverableLogin failed: %v", err)
	}

	var cookie *http.Cookie
	for _, c := range rw.Result().Cookies() {
		if c.Name == "webauthn_session_id" {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatal("webauthn_session_id cookie not set")
	}
	if !cookie.Secure {
		t.Error("expected Secure flag when request has TLS state")
	}
}

// TestWebAuthnSessionHandling tests session creation and cookie setting
func TestWebAuthnSessionHandling(t *testing.T) {
	// Create mock repository
	repo := storage.NewMockRepository()

	// Create a WebAuthnService
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}

	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	// Create a test response writer
	rw := httptest.NewRecorder()

	// Create a session ID and store a mock session
	sessionID := "test-session"
	sessionData := &webauthn.SessionData{
		Challenge:            base64.RawURLEncoding.EncodeToString([]byte("test-challenge")),
		UserID:               []byte("user123"),
		AllowedCredentialIDs: [][]byte{},
	}

	// Store the session
	service.storeSession(sessionID, sessionData)

	// Set a cookie with the session ID
	http.SetCookie(rw, &http.Cookie{
		Name:     "webauthn_session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false,
	})

	// Check that the cookie was set
	cookies := rw.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "webauthn_session_id" {
			found = true
			if cookie.Value != sessionID {
				t.Errorf("Expected cookie value %s, got %s", sessionID, cookie.Value)
			}
		}
	}

	if !found {
		t.Error("webauthn_session_id cookie not found")
	}
}
