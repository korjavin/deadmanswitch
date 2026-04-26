package web

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"testing"

	tpl "github.com/korjavin/deadmanswitch/internal/web/templates"
)

// renderSettings parses layout.html + settings.html together and returns the
// rendered HTML in the authenticated state.
func renderSettings(t *testing.T, data tpl.TemplateData) string {
	t.Helper()
	root := projectRoot(t)
	layoutPath := filepath.Join(root, "web", "templates", "layout.html")
	pagePath := filepath.Join(root, "web", "templates", "settings.html")

	tmpl, err := template.New("layout.html").Funcs(tpl.FuncMap()).
		ParseFiles(layoutPath, pagePath)
	if err != nil {
		t.Fatalf("parse settings templates: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		t.Fatalf("execute settings: %v", err)
	}
	return buf.String()
}

func settingsData(pingFreq, pingDeadline int, pingMethod string, pingingEnabled bool, telegramConnected bool, twoFA bool) tpl.TemplateData {
	user := map[string]interface{}{
		"Email":            "alex@example.com",
		"PingFrequency":    pingFreq,
		"PingDeadline":     pingDeadline,
		"PingMethod":       pingMethod,
		"PingingEnabled":   pingingEnabled,
		"TelegramID":       "",
		"TelegramUsername": "",
	}
	if telegramConnected {
		user["TelegramID"] = "1234567"
		user["TelegramUsername"] = "alex_tg"
	}
	return tpl.TemplateData{
		Title:           "Account Settings",
		ActivePage:      "settings",
		IsAuthenticated: true,
		User:            user,
		Data: map[string]interface{}{
			"User": user,
			"Settings": map[string]interface{}{
				"EmailCheckIn":     true,
				"EmailWarning":     true,
				"TwoFactorEnabled": twoFA,
			},
		},
	}
}

func TestRenderSettings_HeartbeatPageHead(t *testing.T) {
	out := renderSettings(t, settingsData(30, 60, "email", true, false, false))

	mustContain := []string{
		`data-test="settings-page"`,
		`· ACCOUNT · CADENCE`,
		`How often we check in`,
		// Form action + field names preserved
		`action="/settings/deadmanswitch"`,
		`method="POST"`,
		`name="pingFrequency"`,
		`name="pingDeadline"`,
		`name="pingMethod"`,
		`name="pingingEnabled"`,
		`name="notifications[]"`,
		`value="email_checkin"`,
		`value="email_warning"`,
		// Mono labels
		`· WAIT BEFORE NUDGING`,
		`· GRACE PERIOD`,
		// Section heads
		`Cadence`,
		`Channels we reach out on`,
		`How we'll nudge you`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("settings.html missing snippet %q", s)
		}
	}

	// Old Bootstrap-ish copy / classes must be gone.
	mustNotContain := []string{
		`Dead Man's Switch`,
		`form-control`,
		`form-check-input`,
		`<h1>Account Settings</h1>`,
		`Save Dead Man's Switch Settings`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("settings.html still contains old markup %q", s)
		}
	}
}

func TestRenderSettings_CadencePreviewIncludesValues(t *testing.T) {
	out := renderSettings(t, settingsData(30, 60, "email", true, false, false))

	if !strings.Contains(out, `data-test="cadence-preview"`) {
		t.Fatal("settings.html missing cadence-preview marker")
	}
	// The mono-pre should include both numeric values formatted with the `d` suffix.
	for _, s := range []string{`30d`, `60d`, `today ──30d──`, `nudge ──60d──● delivered`} {
		if !strings.Contains(out, s) {
			t.Errorf("cadence-preview missing %q in: %s", s, extractPreview(out))
		}
	}

	// The number inputs reflect the same values.
	if !strings.Contains(out, `name="pingFrequency" id="pingFrequency"`) ||
		!strings.Contains(out, `value="30"`) {
		t.Error("settings.html cadence number input missing or wrong value")
	}
	if !strings.Contains(out, `name="pingDeadline" id="pingDeadline"`) ||
		!strings.Contains(out, `value="60"`) {
		t.Error("settings.html grace number input missing or wrong value")
	}
}

func TestRenderSettings_CadencePreviewVariesWithValues(t *testing.T) {
	out := renderSettings(t, settingsData(7, 14, "email", true, false, false))
	for _, s := range []string{`today ──7d──`, `nudge ──14d──● delivered`} {
		if !strings.Contains(out, s) {
			t.Errorf("settings.html cadence-preview missing %q for 7/14 case", s)
		}
	}
	// The 30/60 strings should not appear when values change.
	if strings.Contains(out, `today ──30d──`) {
		t.Error("settings.html cadence-preview still has 30d placeholder when value is 7")
	}
}

func TestRenderSettings_TelegramChannelDisconnected(t *testing.T) {
	out := renderSettings(t, settingsData(30, 60, "email", true, false, false))

	// Telegram row shows "NOT YET" dashed badge when disconnected.
	if !strings.Contains(out, `data-test="channel-telegram"`) {
		t.Fatal("settings.html missing telegram channel row")
	}
	tgRow := sliceAround(out, `data-test="channel-telegram"`, 600)
	if !strings.Contains(tgRow, `· NOT YET`) {
		t.Errorf("telegram row missing NOT YET badge: %s", tgRow)
	}
	if strings.Contains(tgRow, `· CONNECTED`) {
		t.Errorf("telegram row should not show CONNECTED when disconnected: %s", tgRow)
	}
}

func TestRenderSettings_TelegramChannelConnected(t *testing.T) {
	out := renderSettings(t, settingsData(30, 60, "both", true, true, false))

	tgRow := sliceAround(out, `data-test="channel-telegram"`, 600)
	if !strings.Contains(tgRow, `· CONNECTED`) {
		t.Errorf("telegram row missing CONNECTED badge: %s", tgRow)
	}
	if !strings.Contains(tgRow, `@alex_tg`) {
		t.Errorf("telegram row missing handle when connected: %s", tgRow)
	}
}

func TestRenderSettings_MethodSelected(t *testing.T) {
	out := renderSettings(t, settingsData(30, 60, "telegram", true, true, false))

	// The selected hb-method-card has class `selected` and the radio is checked.
	if !strings.Contains(out, `value="telegram" checked`) {
		t.Error("settings.html telegram method radio should be checked when PingMethod=telegram")
	}
	// Walk forward from the telegram radio to verify it sits inside a label
	// that carries the .selected class. Whitespace between attrs makes a
	// single-string match brittle; instead grep the surrounding window.
	telegramRadio := `name="pingMethod" value="telegram" checked`
	idx := strings.Index(out, telegramRadio)
	if idx == -1 {
		t.Fatalf("settings.html missing telegram radio with checked attr")
	}
	// Find the label tag that opens the card containing this radio.
	prefix := out[:idx]
	labelOpen := strings.LastIndex(prefix, `<label`)
	if labelOpen == -1 {
		t.Fatalf("could not locate enclosing <label> for telegram method card")
	}
	labelTag := prefix[labelOpen:]
	if !strings.Contains(labelTag, `hb-method-card`) || !strings.Contains(labelTag, `selected`) {
		t.Errorf("telegram method card should carry hb-method-card + selected; label tag was: %q", labelTag)
	}

	// And the email card should not be selected when PingMethod=telegram.
	emailRadio := `name="pingMethod" value="email"`
	eIdx := strings.Index(out, emailRadio)
	if eIdx == -1 {
		t.Fatal("settings.html missing email radio")
	}
	emailPrefix := out[:eIdx]
	emailLabelOpen := strings.LastIndex(emailPrefix, `<label`)
	if emailLabelOpen == -1 {
		t.Fatal("could not locate enclosing <label> for email method card")
	}
	emailLabel := emailPrefix[emailLabelOpen:]
	if strings.Contains(emailLabel, `selected`) {
		t.Errorf("email method card should not be .selected when PingMethod=telegram; got %q", emailLabel)
	}
}

func TestRenderSettings_AlwaysOnEmailAndWeb(t *testing.T) {
	out := renderSettings(t, settingsData(30, 60, "email", true, false, false))
	if !strings.Contains(out, `data-test="channel-email"`) {
		t.Error("settings.html missing email channel row")
	}
	if !strings.Contains(out, `data-test="channel-web"`) {
		t.Error("settings.html missing web channel row")
	}
	// Both email and web are always-on.
	if strings.Count(out, `· ALWAYS ON`) < 2 {
		t.Error("settings.html should show ALWAYS ON for email and web channels")
	}
}

func TestRenderSettings_DangerZone(t *testing.T) {
	out := renderSettings(t, settingsData(30, 60, "email", true, false, false))
	if !strings.Contains(out, `data-test="danger-zone"`) {
		t.Error("settings.html missing danger zone marker")
	}
	if !strings.Contains(out, `· DANGER ZONE`) {
		t.Error("settings.html missing DANGER ZONE label")
	}
	if !strings.Contains(out, `Delete my account`) {
		t.Error("settings.html missing delete account button")
	}
}

// extractPreview returns the cadence-preview block for debug output.
func extractPreview(html string) string {
	start := strings.Index(html, `data-test="cadence-preview"`)
	if start == -1 {
		return ""
	}
	end := start + 500
	if end > len(html) {
		end = len(html)
	}
	return html[start:end]
}

// sliceAround returns at most `n` characters of `html` starting at marker.
func sliceAround(html, marker string, n int) string {
	start := strings.Index(html, marker)
	if start == -1 {
		return ""
	}
	end := start + n
	if end > len(html) {
		end = len(html)
	}
	return html[start:end]
}
