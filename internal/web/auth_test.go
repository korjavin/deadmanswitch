package web

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"testing"

	tpl "github.com/korjavin/deadmanswitch/internal/web/templates"
)

// renderAuthTemplate parses layout.html + the named auth template and
// returns the rendered HTML. Auth templates render in either the
// authenticated or unauthenticated shell depending on the data passed.
func renderAuthTemplate(t *testing.T, name string, data tpl.TemplateData) string {
	t.Helper()
	root := projectRoot(t)
	layoutPath := filepath.Join(root, "web", "templates", "layout.html")
	pagePath := filepath.Join(root, "web", "templates", name)

	tmpl, err := template.New("layout.html").Funcs(tpl.FuncMap()).
		ParseFiles(layoutPath, pagePath)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		t.Fatalf("execute %s: %v", name, err)
	}
	return buf.String()
}

func TestRenderLogin_Heartbeat(t *testing.T) {
	out := renderAuthTemplate(t, "login.html", tpl.TemplateData{
		Title:           "Login",
		ActivePage:      "login",
		IsAuthenticated: false,
	})

	mustContain := []string{
		`data-test="auth-login"`,
		`· WELCOME BACK`,
		`Sign in to Heartbeat`,
		// Form action + field names preserved (handlers depend on these)
		`action="/login"`,
		`method="POST"`,
		`name="email"`,
		`name="password"`,
		`name="remember"`,
		// Mono labels with dot prefix
		`· EMAIL`,
		`· PASSWORD`,
		// Passkey path is offered alongside password
		`Sign in with passkey`,
		`id="passkey-login-button"`,
		`OR PASSWORD`,
		// Footer link to register
		`href="/register"`,
		`Create your account`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("login.html missing snippet %q", s)
		}
	}

	// Old Bootstrap-ish copy / classes must be gone.
	mustNotContain := []string{
		`Login with Passkey`,
		`Login with Password`,
		`form-control`,
		`form-check-input`,
		`btn-block`,
		`fa-key`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("login.html still contains old markup %q", s)
		}
	}
}

func TestRenderRegister_Heartbeat(t *testing.T) {
	out := renderAuthTemplate(t, "register.html", tpl.TemplateData{
		Title:           "Register",
		ActivePage:      "register",
		IsAuthenticated: false,
	})

	mustContain := []string{
		`data-test="auth-register"`,
		`· STEP 1 · ABOUT YOU`,
		`Create your account`,
		// Form preserved
		`action="/register"`,
		`method="POST"`,
		`id="registerForm"`,
		`name="email"`,
		`name="password"`,
		`name="confirmPassword"`,
		// Mono labels
		`· EMAIL ADDRESS`,
		`· PASSWORD`,
		`· CONFIRM PASSWORD`,
		// Soft password requirements (match new short labels)
		`id="req-length"`,
		`id="req-uppercase"`,
		`id="req-lowercase"`,
		`id="req-number"`,
		`id="req-special"`,
		// Footer link to login
		`href="/login"`,
		`Sign in`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("register.html missing snippet %q", s)
		}
	}

	mustNotContain := []string{
		`form-control`,
		`Don't have an account`,
		// Old verbose requirement copy
		`At least 8 characters long`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("register.html still contains old markup %q", s)
		}
	}
}

func TestRenderLogin2FA_Heartbeat(t *testing.T) {
	out := renderAuthTemplate(t, "login-2fa.html", tpl.TemplateData{
		Title:      "Two-Factor Authentication",
		ActivePage: "login",
		Data: map[string]interface{}{
			"Email":      "alex@example.com",
			"Password":   "supersecret",
			"RememberMe": true,
		},
	})

	mustContain := []string{
		`data-test="auth-login-2fa"`,
		`· ONE MORE STEP`,
		`A second key, please`,
		// Form preserved with hidden carry-over fields
		`action="/login"`,
		`method="POST"`,
		`name="email" value="alex@example.com"`,
		`name="password"`,
		`name="remember" value="on"`,
		`name="totp_code"`,
		`pattern="[0-9]{6}"`,
		`maxlength="6"`,
		`autocomplete="one-time-code"`,
		`· VERIFICATION CODE`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("login-2fa.html missing snippet %q", s)
		}
	}

	if strings.Contains(out, `Two-Factor Authentication</h1>`) {
		t.Errorf("login-2fa.html still has old h1 'Two-Factor Authentication'")
	}
}

func TestRenderLogin2FA_RendersErrorAlert(t *testing.T) {
	out := renderAuthTemplate(t, "login-2fa.html", tpl.TemplateData{
		Data: map[string]interface{}{
			"Email":    "alex@example.com",
			"Password": "x",
			"Error":    "Invalid verification code. Please try again.",
		},
	})
	if !strings.Contains(out, `class="alert alert-danger"`) {
		t.Error("login-2fa.html error alert markup missing")
	}
	if !strings.Contains(out, "Invalid verification code") {
		t.Error("login-2fa.html error message not rendered")
	}
}

func TestRender2FASetup_Heartbeat(t *testing.T) {
	out := renderAuthTemplate(t, "2fa-setup.html", tpl.TemplateData{
		Title:           "Set Up Two-Factor Authentication",
		ActivePage:      "profile",
		IsAuthenticated: true,
		User:            map[string]interface{}{"Email": "alex@example.com"},
		Data: map[string]interface{}{
			"QRCode":     "iVBORw0KGgoAAAANSUhEUg==",
			"TOTPSecret": "JBSWY3DPEHPK3PXP",
		},
	})

	mustContain := []string{
		`data-test="auth-2fa-setup"`,
		`· ACCOUNT · A SECOND LOCK`,
		`Add a second key`,
		// Form preserved
		`action="/2fa/verify"`,
		`method="POST"`,
		`name="code"`,
		`pattern="[0-9]{6}"`,
		`maxlength="6"`,
		// QR + secret rendered
		`data:image/png;base64,iVBORw0KGgoAAAANSUhEUg==`,
		`JBSWY3DPEHPK3PXP`,
		// Cancel link kept
		`href="/profile"`,
		// Step list and instructions
		`Install an authenticator app`,
		`Scan the QR code`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("2fa-setup.html missing snippet %q", s)
		}
	}

	// Old Bootstrap-ish copy gone
	mustNotContain := []string{
		`Set Up Two-Factor Authentication</h1>`,
		`form-control`,
		`btn-secondary`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("2fa-setup.html still contains old markup %q", s)
		}
	}
}

func TestRenderPasskeys_WithEntries(t *testing.T) {
	out := renderAuthTemplate(t, "passkeys.html", tpl.TemplateData{
		Title:           "Manage Passkeys",
		ActivePage:      "profile",
		IsAuthenticated: true,
		User:            map[string]interface{}{"Email": "alex@example.com"},
		Data: map[string]interface{}{
			"Passkeys": []map[string]interface{}{
				{
					"ID":         "pk-1",
					"Name":       "MacBook",
					"CreatedAt":  "January 2, 2026",
					"LastUsedAt": "April 1, 2026 at 9:00 AM",
				},
				{
					"ID":         "pk-2",
					"Name":       "iPhone",
					"CreatedAt":  "February 14, 2026",
					"LastUsedAt": "April 27, 2026 at 8:00 AM",
				},
			},
		},
	})

	mustContain := []string{
		`data-test="auth-passkeys"`,
		`· ACCOUNT · KEYS`,
		`Your passkeys`,
		`data-test="passkey-list"`,
		`MacBook`,
		`iPhone`,
		`· REGISTERED`,
		// DELETE form action preserved
		`action="/profile/passkeys/pk-1"`,
		`action="/profile/passkeys/pk-2"`,
		`name="_method" value="DELETE"`,
		// Add form
		`id="register-passkey-form"`,
		`name="name"`,
		`Register passkey`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("passkeys.html (with entries) missing snippet %q", s)
		}
	}

	// Empty-state should not appear when there are passkeys.
	if strings.Contains(out, `data-test="passkey-empty"`) {
		t.Error("passkeys.html shows empty-state even though passkeys are present")
	}
}

func TestRenderPasskeys_EmptyState(t *testing.T) {
	out := renderAuthTemplate(t, "passkeys.html", tpl.TemplateData{
		Title:           "Manage Passkeys",
		ActivePage:      "profile",
		IsAuthenticated: true,
		User:            map[string]interface{}{"Email": "alex@example.com"},
		Data: map[string]interface{}{
			"Passkeys": []map[string]interface{}{},
		},
	})

	mustContain := []string{
		`data-test="passkey-empty"`,
		`No passkeys yet`,
		// Add-passkey form is still available
		`id="register-passkey-form"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("passkeys.html (empty) missing snippet %q", s)
		}
	}
	if strings.Contains(out, `data-test="passkey-list"`) {
		t.Error("passkeys.html shows passkey-list when empty")
	}
}

func TestRenderVerifySuccess_Heartbeat(t *testing.T) {
	out := renderAuthTemplate(t, "verify-success.html", tpl.TemplateData{
		Title:           "Email Verified",
		IsAuthenticated: false,
	})

	mustContain := []string{
		`data-test="auth-verify-success"`,
		`· EMAIL VERIFIED`,
		`You're all set.`,
		`href="/login"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("verify-success.html missing snippet %q", s)
		}
	}

	// Old Dead Man's Switch copy gone
	if strings.Contains(out, "Dead Man's Switch") {
		t.Error("verify-success.html still mentions Dead Man's Switch")
	}
	if strings.Contains(out, "Email Verified!") {
		t.Error("verify-success.html still has old exclamation heading")
	}
}

func TestRenderConfirmation_Heartbeat(t *testing.T) {
	out := renderAuthTemplate(t, "confirmation.html", tpl.TemplateData{
		Title:           "Confirmation Successful",
		IsAuthenticated: false,
		Data: map[string]interface{}{
			"Message": "Thank you for confirming your contact information. The user has been notified.",
		},
	})

	mustContain := []string{
		`data-test="auth-confirmation"`,
		`· CONFIRMED`,
		`Thank you for confirming.`,
		// The handler's Message is interpolated
		`The user has been notified`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("confirmation.html missing snippet %q", s)
		}
	}

	if strings.Contains(out, "Dead Man's Switch") {
		t.Error("confirmation.html still mentions Dead Man's Switch")
	}
	if strings.Contains(out, "Confirmation Successful</h2>") {
		t.Error("confirmation.html still has old h2 heading")
	}
}
