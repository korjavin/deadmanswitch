package web

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"testing"

	tpl "github.com/korjavin/deadmanswitch/internal/web/templates"
)

// renderProfile parses layout.html + profile.html together and returns the
// rendered HTML in the authenticated state.
func renderProfile(t *testing.T, data tpl.TemplateData) string {
	t.Helper()
	root := projectRoot(t)
	layoutPath := filepath.Join(root, "web", "templates", "layout.html")
	pagePath := filepath.Join(root, "web", "templates", "profile.html")

	tmpl, err := template.New("layout.html").Funcs(tpl.FuncMap()).
		ParseFiles(layoutPath, pagePath)
	if err != nil {
		t.Fatalf("parse profile templates: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		t.Fatalf("execute profile: %v", err)
	}
	return buf.String()
}

func profileData(githubConnected, telegramConnected, twoFA bool) tpl.TemplateData {
	user := map[string]interface{}{
		"Email":     "alex@example.com",
		"Name":      "alex@example.com",
		"CreatedAt": "January 2, 2026",
		"LastLogin": "April 27, 2026 at 9:00 AM",
	}
	github := map[string]interface{}{"Connected": githubConnected}
	if githubConnected {
		github["Username"] = "alexgh"
	}
	telegram := map[string]interface{}{
		"Connected":   telegramConnected,
		"BotUsername": "@HeartbeatBot",
	}
	if telegramConnected {
		telegram["Username"] = "alex_tg"
		telegram["ID"] = "1234567"
	}
	return tpl.TemplateData{
		Title:           "My Profile",
		ActivePage:      "profile",
		IsAuthenticated: true,
		User:            user,
		Data: map[string]interface{}{
			"User":     user,
			"GitHub":   github,
			"Telegram": telegram,
			"TwoFA":    map[string]interface{}{"Enabled": twoFA},
		},
	}
}

func TestRenderProfile_HeartbeatPageHead(t *testing.T) {
	out := renderProfile(t, profileData(false, false, false))

	mustContain := []string{
		`data-test="profile-page"`,
		`· ABOUT YOU · PROFILE`,
		`Your account`,
		// Form preserved
		`action="/profile"`,
		`method="POST"`,
		`name="email"`,
		`name="currentPassword"`,
		`name="newPassword"`,
		`name="confirmPassword"`,
		// Mono labels
		`· EMAIL`,
		`· CURRENT PASSWORD`,
		`· NEW PASSWORD`,
		`· CONFIRM NEW PASSWORD`,
		// Section heads
		`Identity`,
		`Connected channels`,
		`Authentication`,
		// Email is rendered into the form
		`alex@example.com`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("profile.html missing snippet %q", s)
		}
	}

	// Old Bootstrap-ish copy / classes must be gone.
	mustNotContain := []string{
		`Dead Man's Switch`,
		`form-control`,
		`<h1>My Profile</h1>`,
		`fa-shield-alt`,
		`fa-github`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("profile.html still contains old markup %q", s)
		}
	}
}

func TestRenderProfile_GitHubDisconnected(t *testing.T) {
	out := renderProfile(t, profileData(false, false, false))

	gh := sliceAround(out, `data-test="channel-github"`, 1200)
	if !strings.Contains(gh, `· NOT YET`) {
		t.Errorf("github row missing NOT YET badge when disconnected: %s", gh)
	}
	if strings.Contains(gh, `· CONNECTED`) {
		t.Errorf("github row must not show CONNECTED when disconnected: %s", gh)
	}
	if !strings.Contains(gh, `name="github_username"`) {
		t.Errorf("github row should expose connect form when disconnected: %s", gh)
	}
}

func TestRenderProfile_GitHubConnected(t *testing.T) {
	out := renderProfile(t, profileData(true, false, false))

	gh := sliceAround(out, `data-test="channel-github"`, 1200)
	if !strings.Contains(gh, `· CONNECTED`) {
		t.Errorf("github row missing CONNECTED badge when connected: %s", gh)
	}
	if !strings.Contains(gh, `@alexgh`) {
		t.Errorf("github row missing handle @alexgh: %s", gh)
	}
	if !strings.Contains(gh, `action="/profile/github/disconnect"`) {
		t.Errorf("github row missing disconnect form action: %s", gh)
	}
}

func TestRenderProfile_TelegramDisconnectedShowsInstructions(t *testing.T) {
	out := renderProfile(t, profileData(false, false, false))

	tg := sliceAround(out, `data-test="channel-telegram"`, 1200)
	if !strings.Contains(tg, `· NOT YET`) {
		t.Errorf("telegram row missing NOT YET badge: %s", tg)
	}
	if !strings.Contains(out, `data-test="telegram-instructions"`) {
		t.Error("profile.html should show telegram instructions when not connected")
	}
	if !strings.Contains(out, `/connect alex@example.com`) {
		t.Error("profile.html telegram instructions should include /connect <email>")
	}
}

func TestRenderProfile_TelegramConnectedHidesInstructions(t *testing.T) {
	out := renderProfile(t, profileData(false, true, false))

	tg := sliceAround(out, `data-test="channel-telegram"`, 1200)
	if !strings.Contains(tg, `· CONNECTED`) {
		t.Errorf("telegram row missing CONNECTED badge: %s", tg)
	}
	if !strings.Contains(tg, `@alex_tg`) {
		t.Errorf("telegram row missing handle when connected: %s", tg)
	}
	if strings.Contains(out, `data-test="telegram-instructions"`) {
		t.Error("profile.html should NOT show telegram instructions when connected")
	}
}

func TestRenderProfile_TwoFAEnabled(t *testing.T) {
	out := renderProfile(t, profileData(false, false, true))

	row := sliceAround(out, `data-test="auth-2fa-row"`, 800)
	if !strings.Contains(row, `· ENABLED`) {
		t.Errorf("2FA row missing ENABLED badge when twoFA on: %s", row)
	}
	if !strings.Contains(row, `action="/2fa/disable"`) {
		t.Errorf("2FA row missing disable form: %s", row)
	}
	if !strings.Contains(row, `name="code"`) {
		t.Errorf("2FA disable form missing name=code: %s", row)
	}
}

func TestRenderProfile_TwoFADisabled(t *testing.T) {
	out := renderProfile(t, profileData(false, false, false))

	row := sliceAround(out, `data-test="auth-2fa-row"`, 800)
	if !strings.Contains(row, `· NOT YET`) {
		t.Errorf("2FA row missing NOT YET badge when twoFA off: %s", row)
	}
	if !strings.Contains(row, `href="/2fa/setup"`) {
		t.Errorf("2FA row missing setup link: %s", row)
	}
}

func TestRenderProfile_NoSMSOrPhoneField(t *testing.T) {
	out := renderProfile(t, profileData(true, true, true))

	mustNotContain := []string{
		`name="phoneNumber"`,
		`name="phone_number"`,
		`name="phone"`,
		`name="sms"`,
		`SMS`,
		`Phone Number`,
		`Phone number`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("profile.html still contains forbidden phone/SMS reference %q", s)
		}
	}
}

func TestRenderProfile_PasskeysLink(t *testing.T) {
	out := renderProfile(t, profileData(false, false, false))
	if !strings.Contains(out, `data-test="auth-passkey-row"`) {
		t.Fatal("profile.html missing passkey row")
	}
	row := sliceAround(out, `data-test="auth-passkey-row"`, 1500)
	if !strings.Contains(row, `href="/profile/passkeys"`) {
		t.Errorf("passkey row missing link to /profile/passkeys: %s", row)
	}
}
