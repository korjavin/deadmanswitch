package web

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"testing"

	tpl "github.com/korjavin/deadmanswitch/internal/web/templates"
)

// renderHistory parses layout.html + history.html together and returns the
// rendered HTML in the authenticated state.
func renderHistory(t *testing.T, data tpl.TemplateData) string {
	t.Helper()
	root := projectRoot(t)
	layoutPath := filepath.Join(root, "web", "templates", "layout.html")
	pagePath := filepath.Join(root, "web", "templates", "history.html")

	tmpl, err := template.New("layout.html").Funcs(tpl.FuncMap()).
		ParseFiles(layoutPath, pagePath)
	if err != nil {
		t.Fatalf("parse history templates: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		t.Fatalf("execute history: %v", err)
	}
	return buf.String()
}

func historyData(activities []map[string]interface{}, pagination map[string]interface{}) tpl.TemplateData {
	data := map[string]interface{}{
		"User":       map[string]interface{}{"Email": "alex@example.com"},
		"Activities": activities,
	}
	if pagination != nil {
		data["Pagination"] = pagination
	}
	return tpl.TemplateData{
		Title:           "Activity History",
		ActivePage:      "history",
		IsAuthenticated: true,
		User:            map[string]interface{}{"Email": "alex@example.com"},
		Data:            data,
	}
}

// TestRenderHistory_HeartbeatPageHead verifies the new design copy lands.
func TestRenderHistory_HeartbeatPageHead(t *testing.T) {
	out := renderHistory(t, historyData([]map[string]interface{}{}, nil))

	mustContain := []string{
		`data-test="history-page"`,
		`· AUDIT TRAIL`,
		`Everything we've recorded`,
		`A complete log of check-ins, nudges, and changes.`,
	}
	for _, want := range mustContain {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q\noutput: %s", want, out)
		}
	}
}

// TestRenderHistory_AllBadges renders one entry per supported type and asserts
// each badge label appears, dates render in mono with tabular-nums, and the
// timeline list is present.
func TestRenderHistory_AllBadges(t *testing.T) {
	activities := []map[string]interface{}{
		{
			"Type":        "security",
			"Title":       "Login",
			"Description": "Signed in via web",
			"Timestamp":   "Apr 27, 2026 at 9:00 AM",
		},
		{
			"Type":        "activity",
			"Title":       "GitHub Activity",
			"Description": "Pushed 4 commits to alex/notes",
			"Timestamp":   "Apr 27, 2026 at 6:14 AM",
		},
		{
			"Type":        "checkin",
			"Title":       "Manual Check-in",
			"Description": "Confirmed via web",
			"Timestamp":   "Apr 26, 2026 at 9:30 PM",
		},
		{
			"Type":        "nudge",
			"Title":       "Nudge sent",
			"Description": "Email reminder dispatched",
			"Timestamp":   "Apr 25, 2026 at 8:00 AM",
		},
		{
			"Type":        "secret",
			"Title":       "Letter created",
			"Description": "Added new letter",
			"Timestamp":   "Apr 24, 2026 at 4:12 PM",
		},
		{
			"Type":        "recipient",
			"Title":       "Recipient invited",
			"Description": "Added Sam Reyes",
			"Timestamp":   "Apr 23, 2026 at 11:00 AM",
		},
		{
			"Type":        "settings",
			"Title":       "Settings updated",
			"Description": "Changed nudge frequency",
			"Timestamp":   "Apr 22, 2026 at 2:30 PM",
		},
	}

	out := renderHistory(t, historyData(activities, nil))

	for _, badge := range []string{"· LOGIN", "· GITHUB", "· CHECKIN", "· NUDGE", "· LETTER", "· RECIPIENT", "· SETTINGS"} {
		if !strings.Contains(out, badge) {
			t.Errorf("expected badge %q in output", badge)
		}
	}

	// Each row marker should appear once per entry.
	if got := strings.Count(out, `data-test="history-row"`); got != len(activities) {
		t.Errorf("expected %d history-row markers, got %d", len(activities), got)
	}

	// Tabular-nums + mono date class drives the date column visual.
	if !strings.Contains(out, `font-variant-numeric: tabular-nums`) {
		t.Error("expected history dates to declare tabular-nums for aligned mono columns")
	}
	if !strings.Contains(out, `class="hb-history-date"`) {
		t.Error("expected hb-history-date class on date cells")
	}

	// Each entry's title and description must be rendered.
	for _, a := range activities {
		title, _ := a["Title"].(string)
		desc, _ := a["Description"].(string)
		if !strings.Contains(out, title) {
			t.Errorf("expected title %q in output", title)
		}
		if !strings.Contains(out, desc) {
			t.Errorf("expected description %q in output", desc)
		}
	}
}

// TestRenderHistory_EmptyState renders the page with no activities and asserts
// the envelope-style empty block appears, no list markers do, and the
// pagination container is absent.
func TestRenderHistory_EmptyState(t *testing.T) {
	out := renderHistory(t, historyData(nil, nil))

	if !strings.Contains(out, `data-test="history-empty"`) {
		t.Error("expected history-empty marker on empty state")
	}
	if !strings.Contains(out, `No activity recorded yet`) {
		t.Error("expected empty-state headline copy")
	}
	if strings.Contains(out, `data-test="history-list"`) {
		t.Error("did not expect history-list when no activities present")
	}
	if strings.Contains(out, `data-test="history-row"`) {
		t.Error("did not expect history rows when no activities present")
	}
	if strings.Contains(out, `data-test="history-pagination"`) {
		t.Error("did not expect pagination block when none provided")
	}
}

// TestRenderHistory_PaginationControlsIntact asserts that when the handler
// supplies a Pagination map, the prev/next nav controls render with their
// hrefs preserved.
func TestRenderHistory_PaginationControlsIntact(t *testing.T) {
	activities := []map[string]interface{}{
		{
			"Type":        "checkin",
			"Title":       "Manual Check-in",
			"Description": "Confirmed via web",
			"Timestamp":   "Apr 27, 2026 at 9:00 AM",
		},
	}
	pagination := map[string]interface{}{
		"Prev": "/history?page=1",
		"Next": "/history?page=3",
	}

	out := renderHistory(t, historyData(activities, pagination))

	if !strings.Contains(out, `data-test="history-pagination"`) {
		t.Error("expected pagination wrapper when Pagination data present")
	}
	for _, want := range []string{
		`href="/history?page=1"`,
		`href="/history?page=3"`,
		`rel="prev"`,
		`rel="next"`,
		`← Older`,
		`Newer →`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected pagination markup %q in output", want)
		}
	}
}

// TestRenderHistory_RemovesOldCopy guards against regressions: the old
// Bootstrap-ish history page markup and copy must be gone.
func TestRenderHistory_RemovesOldCopy(t *testing.T) {
	activities := []map[string]interface{}{
		{
			"Type":        "checkin",
			"Title":       "Manual Check-in",
			"Description": "Confirmed via web",
			"Timestamp":   "Apr 27, 2026 at 9:00 AM",
		},
	}
	out := renderHistory(t, historyData(activities, nil))

	mustNotContain := []string{
		`Activity History</h1>`,
		`Dead Man's Switch`,
		`class="timeline-item"`,
		`class="timeline-marker`,
		`class="filter-controls"`,
		`id="activityType"`,
		`id="dateRange"`,
		`form-control`,
	}
	for _, banned := range mustNotContain {
		if strings.Contains(out, banned) {
			t.Errorf("expected old markup %q to be gone, found in output", banned)
		}
	}
}

// TestRenderHistory_UnknownTypeFallsBackToRecorded asserts the default
// label/icon path renders for activity types we don't explicitly map.
func TestRenderHistory_UnknownTypeFallsBackToRecorded(t *testing.T) {
	activities := []map[string]interface{}{
		{
			"Type":        "other",
			"Title":       "Something happened",
			"Description": "An entry with no specific category",
			"Timestamp":   "Apr 27, 2026 at 9:00 AM",
		},
	}
	out := renderHistory(t, historyData(activities, nil))

	if !strings.Contains(out, `· RECORDED`) {
		t.Error("expected RECORDED fallback badge for unknown activity type")
	}
}
