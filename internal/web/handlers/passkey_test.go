package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/korjavin/deadmanswitch/internal/auth"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// newTestPasskeyHandler builds a PasskeyHandler backed by a real
// WebAuthnService and the in-memory MockRepository, mirroring the setup used
// in internal/auth/webauthn_test.go.
func newTestPasskeyHandler(t *testing.T) (*PasskeyHandler, *auth.WebAuthnService, *storage.MockRepository) {
	t.Helper()

	repo := storage.NewMockRepository()
	service, err := auth.NewWebAuthnService(auth.WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	return NewPasskeyHandler(repo, service), service, repo
}

// TestHandleBeginDiscoverableLogin verifies the begin endpoint returns 200
// with a JSON body that has a non-empty challenge and an empty (or absent)
// allowCredentials list, and sets the webauthn_session_id cookie.
func TestHandleBeginDiscoverableLogin(t *testing.T) {
	handler, _, _ := newTestPasskeyHandler(t)

	req := httptest.NewRequest("POST", "/login/passkey/discover/begin", nil)
	rr := httptest.NewRecorder()

	handler.HandleBeginDiscoverableLogin(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var resp struct {
		PublicKey struct {
			Challenge          string        `json:"challenge"`
			AllowedCredentials []interface{} `json:"allowCredentials"`
			RPID               string        `json:"rpId"`
		} `json:"publicKey"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not decode response: %v (body: %s)", err, rr.Body.String())
	}

	if resp.PublicKey.Challenge == "" {
		t.Error("expected non-empty publicKey.challenge")
	}
	if len(resp.PublicKey.AllowedCredentials) != 0 {
		t.Errorf("expected empty publicKey.allowCredentials, got %d entries", len(resp.PublicKey.AllowedCredentials))
	}
	if resp.PublicKey.RPID != "localhost" {
		t.Errorf("expected publicKey.rpId=localhost, got %q", resp.PublicKey.RPID)
	}

	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "webauthn_session_id" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected webauthn_session_id cookie to be set")
	}
	if sessionCookie.Value == "" {
		t.Error("expected non-empty session cookie value")
	}
}

// TestHandleFinishDiscoverableLoginMissingCookie verifies the failure path:
// the WebAuthnService surfaces an error when no session cookie is present, the
// handler responds 4xx, and no session_token cookie is issued.
func TestHandleFinishDiscoverableLoginMissingCookie(t *testing.T) {
	handler, _, repo := newTestPasskeyHandler(t)

	req := httptest.NewRequest("POST", "/login/passkey/discover/finish",
		strings.NewReader(`{"credential":{}}`))
	rr := httptest.NewRecorder()

	handler.HandleFinishDiscoverableLogin(rr, req)

	if rr.Code < 400 || rr.Code >= 500 {
		t.Errorf("expected 4xx response, got %d", rr.Code)
	}
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session_token" {
			t.Errorf("session_token cookie should not be set on failure, got value %q", c.Value)
		}
	}
	if len(repo.Sessions) != 0 {
		t.Errorf("expected no sessions to be created on failure, got %d", len(repo.Sessions))
	}
	if len(repo.AuditLogs) != 0 {
		t.Errorf("expected no audit logs on failure, got %d", len(repo.AuditLogs))
	}
}

// TestHandleFinishDiscoverableLoginInvalidCredential verifies the failure path
// when a session cookie is present but the wrapped credential cannot be parsed
// as a valid WebAuthn assertion.
func TestHandleFinishDiscoverableLoginInvalidCredential(t *testing.T) {
	handler, service, repo := newTestPasskeyHandler(t)

	// Prime a discoverable session so the cookie/session lookup succeeds and
	// the failure happens during assertion parsing.
	beginRR := httptest.NewRecorder()
	beginReq := httptest.NewRequest("POST", "/login/passkey/discover/begin", nil)
	handler.HandleBeginDiscoverableLogin(beginRR, beginReq)
	if beginRR.Code != http.StatusOK {
		t.Fatalf("priming begin call failed: status %d", beginRR.Code)
	}
	var sessionCookie *http.Cookie
	for _, c := range beginRR.Result().Cookies() {
		if c.Name == "webauthn_session_id" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected webauthn_session_id cookie from priming begin")
	}

	finishReq := httptest.NewRequest("POST", "/login/passkey/discover/finish",
		strings.NewReader(`{"credential":{"id":"","type":"public-key"}}`))
	finishReq.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	handler.HandleFinishDiscoverableLogin(rr, finishReq)

	if rr.Code < 400 || rr.Code >= 500 {
		t.Errorf("expected 4xx response, got %d", rr.Code)
	}
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session_token" {
			t.Errorf("session_token cookie should not be set on failure, got value %q", c.Value)
		}
	}
	if len(repo.Sessions) != 0 {
		t.Errorf("expected no sessions to be created on failure, got %d", len(repo.Sessions))
	}
	if len(repo.AuditLogs) != 0 {
		t.Errorf("expected no audit logs on failure, got %d", len(repo.AuditLogs))
	}

	// The session must have been consumed even though the assertion was bogus.
	// Build a fresh request (the previous finishReq's body has been drained) so
	// the failure is unambiguously the deleted session rather than EOF on the
	// body read.
	replayReq := httptest.NewRequest("POST", "/login/passkey/discover/finish",
		strings.NewReader(`{"credential":{"id":"","type":"public-key"}}`))
	replayReq.AddCookie(sessionCookie)
	_, _, err := service.FinishDiscoverableLogin(replayReq.Context(), replayReq)
	if err == nil {
		t.Fatal("expected session to be consumed after first FinishDiscoverableLogin call")
	}
	if !strings.Contains(err.Error(), "session data not found") {
		t.Errorf("expected error to mention session-not-found, got %v", err)
	}
}
