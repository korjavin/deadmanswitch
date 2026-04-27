package web

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"testing"

	tpl "github.com/korjavin/deadmanswitch/internal/web/templates"
)

// renderDashboard parses layout.html + dashboard.html together and
// returns the rendered HTML in the authenticated state.
func renderDashboard(t *testing.T, data tpl.TemplateData) string {
	t.Helper()
	root := projectRoot(t)
	layoutPath := filepath.Join(root, "web", "templates", "layout.html")
	pagePath := filepath.Join(root, "web", "templates", "dashboard.html")

	tmpl, err := template.New("layout.html").Funcs(tpl.FuncMap()).
		ParseFiles(layoutPath, pagePath)
	if err != nil {
		t.Fatalf("parse dashboard templates: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		t.Fatalf("execute dashboard: %v", err)
	}
	return buf.String()
}

// dashboardData builds a complete TemplateData payload mirroring what the
// real handler now constructs. Tests vary one slice (status, circle, etc.)
// per case.
func dashboardData(variant string, circle []map[string]interface{}, githubConnected, telegramConnected, hasPasskey bool) tpl.TemplateData {
	return tpl.TemplateData{
		Title:           "Dashboard",
		ActivePage:      "dashboard",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email":             "alex@example.com",
			"Name":              "alex@example.com",
			"FirstName":         "Alex",
			"GitHubUsername":    "alex",
			"GitHubConnected":   githubConnected,
			"TelegramUsername":  "alex_tg",
			"TelegramConnected": telegramConnected,
			"HasPasskey":        hasPasskey,
		},
		Data: map[string]interface{}{
			"Status":        statusForVariant(variant),
			"StatusVariant": variant,
			"TodayLabel":    "MON, APR 27",
			"PingFrequency": 30,
			"PingDeadline":  60,
			"Activities": []map[string]string{
				{"Time": "Apr 27, 2026 09:00", "Description": "Logged in"},
				{"Time": "Apr 26, 2026 18:00", "Description": "Checked in"},
			},
			"Circle": circle,
			"Timeline": map[string]interface{}{
				"LastSeen":  map[string]string{"Value": "2 hours ago", "Sub": "Apr 27, 2026"},
				"NextNudge": map[string]string{"Value": "5 days", "Sub": "May 2, 2026"},
				"Delivery":  map[string]string{"Value": "8 days", "Sub": "May 5, 2026"},
			},
		},
	}
}

func statusForVariant(v string) string {
	switch v {
	case "ok":
		return "active"
	case "warn":
		return "caution"
	default:
		return "danger"
	}
}

func TestRenderDashboard_OkVariant(t *testing.T) {
	out := renderDashboard(t, dashboardData("ok", nil, true, true, true))

	mustContain := []string{
		`data-test="dashboard"`,
		`TODAY · MON, APR 27`,
		`You're good, Alex.`,
		`Next gentle nudge in 5 days`,
		`heartbeat ok`,
		`data-test="status-strip"`,
		// Timeline strip values
		`· LAST SEEN`,
		`· NEXT NUDGE`,
		`· DELIVERY (IF SILENT)`,
		`2 hours ago`,
		`5 days`,
		`8 days`,
		// Big check-in card
		`· ONE TAP · CONFIRMS YOU'RE HERE`,
		`id="checkInButton"`,
		`I'm here`,
		// All six channel cards present
		`data-test="channel-signin"`,
		`data-test="channel-github"`,
		`data-test="channel-telegram"`,
		`data-test="channel-email"`,
		`data-test="channel-passkey"`,
		`data-test="channel-anywhere"`,
		`How we know you're alive`,
		`ALWAYS ON`,
		// GitHub connected → WATCHING badge + handle
		`WATCHING`,
		`@alex`,
		// Telegram connected
		`@alex_tg`,
		// Passkey registered
		`REGISTERED`,
		// Bottom row
		`data-test="recent-activity"`,
		`Recent activity`,
		`VIEW ALL →`,
		`data-test="your-circle"`,
		`Your circle`,
		`MANAGE →`,
		// How-it-works modal markup
		`HOW HEARTBEAT WORKS`,
		`id="howItWorksModal"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("dashboard (ok) missing snippet %q", s)
		}
	}
}

func TestRenderDashboard_WarnVariant(t *testing.T) {
	out := renderDashboard(t, dashboardData("warn", nil, false, false, false))

	mustContain := []string{
		`Just checking in.`,
		`It's been a quiet stretch`,
		`heartbeat warn`,
		// Disconnected channels
		`NOT CONNECTED`,
		// Passkey missing → ADD ONE prompt
		`ADD ONE`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("dashboard (warn) missing snippet %q", s)
		}
	}
	if strings.Contains(out, "You're good, Alex.") {
		t.Error("dashboard (warn) should not show ok-variant title")
	}
	if strings.Contains(out, "Are you there?") {
		t.Error("dashboard (warn) should not show crit-variant title")
	}
	if strings.Contains(out, "Pushing code counts") {
		t.Error("dashboard (warn) should not mention pushing code when GitHub disconnected")
	}
}

func TestRenderDashboard_CritVariant(t *testing.T) {
	out := renderDashboard(t, dashboardData("crit", nil, false, false, false))

	mustContain := []string{
		`Are you there?`,
		`If we don't hear from you soon, your letters will start to go out`,
		`heartbeat crit`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("dashboard (crit) missing snippet %q", s)
		}
	}
}

func TestRenderDashboard_EmptyCircle(t *testing.T) {
	out := renderDashboard(t, dashboardData("ok", nil, true, true, true))

	if !strings.Contains(out, `data-test="circle-empty"`) {
		t.Error("empty circle should render circle-empty marker")
	}
	if !strings.Contains(out, "No one in your circle yet.") {
		t.Error("empty circle should render the soft prompt copy")
	}
	if !strings.Contains(out, `href="/recipients"`) {
		t.Error("empty circle prompt should link to /recipients")
	}
}

func TestRenderDashboard_PopulatedCircle(t *testing.T) {
	circle := []map[string]interface{}{
		{"Name": "Sam Carter", "Email": "sam@example.com", "LetterCount": 2, "Verified": true},
		{"Name": "Jamie Lee", "Email": "jamie@example.com", "LetterCount": 0, "Verified": false},
	}
	out := renderDashboard(t, dashboardData("ok", circle, true, true, true))

	mustContain := []string{
		`Sam Carter`,
		`Jamie Lee`,
		`2 letters`,
		`0 letters`,
		`✓ confirmed`,
		`awaiting`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("populated circle missing snippet %q", s)
		}
	}
	if strings.Contains(out, `data-test="circle-empty"`) {
		t.Error("populated circle should not show empty marker")
	}
}

func TestRenderDashboard_SingularLetter(t *testing.T) {
	circle := []map[string]interface{}{
		{"Name": "Sam Carter", "LetterCount": 1, "Verified": true},
	}
	out := renderDashboard(t, dashboardData("ok", circle, true, true, true))
	if !strings.Contains(out, `1 letter`) {
		t.Error("singular letter count should render '1 letter'")
	}
	if strings.Contains(out, `1 letters`) {
		t.Error("singular letter count must not render '1 letters'")
	}
}

func TestRenderDashboard_RemovesOldAlarmCopy(t *testing.T) {
	out := renderDashboard(t, dashboardData("ok", nil, true, true, true))

	mustNotContain := []string{
		"System Active",
		"Critical Action Required",
		"Action Required",
		"Check In Now",
		`class="status-indicator`,
		`class="check-in-box`,
		`Welcome back!`,
		`Total Secrets`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("dashboard still contains old alarm copy %q", s)
		}
	}
}
