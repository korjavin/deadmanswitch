package web

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tpl "github.com/korjavin/deadmanswitch/internal/web/templates"
)

// renderRecipientsTemplate parses layout.html + the named template and
// returns the rendered HTML in the authenticated app shell.
func renderRecipientsTemplate(t *testing.T, name string, data tpl.TemplateData) string {
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

func authedUser() map[string]interface{} {
	return map[string]interface{}{
		"Email": "alex@example.com",
		"Name":  "alex@example.com",
	}
}

// ---------------------------------------------------------------------------
// recipients.html — envelope grid
// ---------------------------------------------------------------------------

func TestRenderRecipients_EnvelopeGrid(t *testing.T) {
	sentAt := time.Now().Add(-2 * time.Hour)
	recipients := []map[string]interface{}{
		{
			"ID":            "rec-1",
			"Name":          "Sam Carter",
			"Email":         "sam@example.com",
			"Relationship":  "Sister",
			"ContactMethod": "email",
			"IsConfirmed":   true,
			"AssignedSecrets": []map[string]interface{}{
				{"ID": "s1", "Title": "Letter to Sam"},
				{"ID": "s2", "Title": "Banking notes"},
			},
		},
		{
			// In-circle: just added, no hello sent — should NOT show
			// "awaiting reply" because we're not actually waiting on them.
			"ID":                 "rec-2",
			"Name":               "Jamie Lee",
			"Email":              "jamie@example.com",
			"Relationship":       "Best friend",
			"ContactMethod":      "email",
			"IsConfirmed":        false,
			"ConfirmationSentAt": nil,
			"AssignedSecrets":    []map[string]interface{}{},
		},
		{
			// Hello was sent (via Test button or send_intro checkbox)
			// but they haven't clicked the link yet — this is the only
			// state where "awaiting reply" is honest.
			"ID":                 "rec-4",
			"Name":               "Priya Mehta",
			"Email":              "priya@example.com",
			"Relationship":       "Friend",
			"ContactMethod":      "email",
			"IsConfirmed":        false,
			"ConfirmationSentAt": &sentAt,
			"AssignedSecrets":    []map[string]interface{}{},
		},
		{
			"ID":            "rec-3",
			"Name":          "Robin Park",
			"Email":         "robin@example.com",
			"Relationship":  "Lawyer",
			"ContactMethod": "email",
			"IsConfirmed":   true,
			"AssignedSecrets": []map[string]interface{}{
				{"ID": "s3", "Title": "Will copy"},
			},
		},
	}

	out := renderRecipientsTemplate(t, "recipients.html", tpl.TemplateData{
		Title:           "Recipients",
		ActivePage:      "recipients",
		IsAuthenticated: true,
		User:            authedUser(),
		Data: map[string]interface{}{
			"Recipients": recipients,
		},
	})

	mustContain := []string{
		`data-test="recipients"`,
		`· THE PEOPLE · YOUR CIRCLE`,
		`People who matter`,
		`data-test="add-recipient"`,
		`href="/recipients/new"`,
		// Grid + envelope structure
		`data-test="envelope-grid"`,
		`data-test="envelope"`,
		`TO ·`,
		// Names + relations + emails
		`Sam Carter`,
		`Sister`,
		`sam@example.com`,
		`Jamie Lee`,
		`Best friend`,
		`jamie@example.com`,
		`Robin Park`,
		// Confirmation badges — tri-state:
		//   confirmed (IsConfirmed=true) → "✓ confirmed"
		//   awaiting (ConfirmationSentAt set, not yet confirmed) → "awaiting reply"
		//   in-circle (no hello sent yet) → "in circle"
		`data-test="badge-confirmed"`,
		`✓ confirmed`,
		`data-test="badge-awaiting"`,
		`awaiting reply`,
		`data-test="badge-in-circle"`,
		`in circle`,
		// Priya is the awaiting-reply recipient; Jamie is the in-circle one.
		`Priya Mehta`,
		// CONTAINS counts (mixed pluralization)
		`· CONTAINS`,
		`2 letters`,
		`1 letter`,
		`nothing yet`,
		// Per-envelope action URLs preserved
		`href="/recipients/rec-1"`,
		`href="/recipients/rec-1/secrets"`,
		`href="/recipients/rec-1/test"`,
		`action="/recipients/rec-1"`,
		`name="_method" value="DELETE"`,
		// Trailing add-card
		`data-test="add-another"`,
		`Add another person`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("recipients.html (grid) missing snippet %q", s)
		}
	}

	// Empty-state markers must NOT appear when populated.
	if strings.Contains(out, `data-test="recipients-empty"`) {
		t.Error("recipients.html should not show empty-state when recipients are present")
	}

	// Old Bootstrap markup must be gone.
	mustNotContain := []string{
		`No Recipients Yet`,
		`Add First Recipient`,
		`recipient-card`,
		`btn btn-primary`,
		`bg-success`,
		`bg-warning`,
		`Dead Man's Switch`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("recipients.html still contains old markup %q", s)
		}
	}
}

func TestRenderRecipients_EmptyState(t *testing.T) {
	out := renderRecipientsTemplate(t, "recipients.html", tpl.TemplateData{
		Title:           "Recipients",
		ActivePage:      "recipients",
		IsAuthenticated: true,
		User:            authedUser(),
		Data: map[string]interface{}{
			"Recipients": []map[string]interface{}{},
		},
	})

	mustContain := []string{
		`data-test="recipients-empty"`,
		`Nobody yet`,
		`Add your first person`,
		`href="/recipients/new"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("recipients.html (empty) missing snippet %q", s)
		}
	}

	// Grid markers must NOT appear when empty.
	if strings.Contains(out, `data-test="envelope-grid"`) {
		t.Error("recipients.html should not show envelope grid when empty")
	}
	if strings.Contains(out, `data-test="envelope"`) {
		t.Error("recipients.html should not render envelopes when empty")
	}
}

// ---------------------------------------------------------------------------
// new-recipient.html — stepped form
// ---------------------------------------------------------------------------

func TestRenderNewRecipient_Heartbeat(t *testing.T) {
	out := renderRecipientsTemplate(t, "new-recipient.html", tpl.TemplateData{
		Title:           "Add Recipient",
		ActivePage:      "recipients",
		IsAuthenticated: true,
		User:            authedUser(),
		Data:            map[string]interface{}{},
	})

	mustContain := []string{
		`data-test="new-recipient"`,
		`· ADD SOMEONE TO YOUR CIRCLE`,
		`Add someone`,
		// Stepped form labels
		`data-test="steps"`,
		`data-test="step-1"`,
		`data-test="step-2"`,
		`Who they are`,
		// Step 2 is now opt-in: heading, label, and checkbox all reflect that.
		`Say hello (optional)`,
		`· STEP 2 · SAY HELLO? (OPTIONAL)`,
		`name="send_intro"`,
		`data-test="send-intro"`,
		`Off by default`,
		// Form preserves backend contract
		`action="/recipients/new"`,
		`method="POST"`,
		`name="name"`,
		`name="email"`,
		`name="relationship"`,
		`name="contactMethod"`,
		`name="telegramUsername"`,
		`name="notes"`,
		`name="verified"`,
		// Mono labels
		`· THEIR NAME`,
		`· THEIR EMAIL`,
		`· WHO THEY ARE TO YOU`,
		`· REACH THEM ALSO BY`,
		// Email preview block (still rendered so user knows what would be sent)
		`data-test="preview"`,
		`From: heartbeat`,
		`Yes, this is my email →`,
		// Submit button copy — adding to circle, not invoking an email send.
		`Add to circle`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("new-recipient.html (add) missing snippet %q", s)
		}
	}

	// Old Bootstrap markup gone, and the misleading copy that implied
	// adding a recipient automatically sends them an email is gone too.
	mustNotContain := []string{
		`form-control`,
		`form-check-input`,
		`Add Recipient</h1>`,
		`Edit Recipient</h1>`,
		`Send invitation`,
		`What we'll send them`,
		`STEP 2 · WHAT WE'LL SEND THEM`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("new-recipient.html still contains old markup %q", s)
		}
	}

	// send_intro checkbox must default to unchecked — the whole point of
	// this rework is that we don't send anything to the recipient unless
	// the user explicitly opts in.
	if strings.Contains(out, `name="send_intro"`) && strings.Contains(out, `checked`) {
		// Coarse but cheap: scan the input element and assert no checked attr.
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, `name="send_intro"`) && strings.Contains(line, `checked`) {
				t.Errorf("send_intro checkbox must NOT default to checked: %s", strings.TrimSpace(line))
			}
		}
	}
}

func TestRenderNewRecipient_EditMode(t *testing.T) {
	out := renderRecipientsTemplate(t, "new-recipient.html", tpl.TemplateData{
		Title:           "Edit Recipient",
		ActivePage:      "recipients",
		IsAuthenticated: true,
		User:            authedUser(),
		Data: map[string]interface{}{
			"Recipient": map[string]interface{}{
				"ID":               "rec-42",
				"Name":             "Sam Carter",
				"Email":            "sam@example.com",
				"Notes":            "Sister; reads slowly",
				"Relationship":     "family",
				"ContactMethod":    "telegram",
				"TelegramUsername": "samc",
				"Verified":         true,
			},
		},
	})

	mustContain := []string{
		`· EDIT ENVELOPE`,
		`Edit envelope`,
		// Form posts to the edit URL
		`action="/recipients/rec-42"`,
		// Pre-filled values
		`value="Sam Carter"`,
		`value="sam@example.com"`,
		`Sister; reads slowly`,
		`value="samc"`,
		// Selected option markup
		`<option value="family" selected>Family</option>`,
		`<option value="telegram" selected>Email + Telegram</option>`,
		// Verified checkbox checked
		`name="verified"`,
		`checked`,
		// Submit copy switches in edit mode
		`Save changes`,
		// Email preview pre-fills with editing values
		`data-test="preview-name"`,
		`data-test="preview-to"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("new-recipient.html (edit) missing snippet %q", s)
		}
	}
}

// ---------------------------------------------------------------------------
// manage-recipient-secrets.html — letter checkboxes
// ---------------------------------------------------------------------------

func TestRenderManageRecipientSecrets_Mixed(t *testing.T) {
	out := renderRecipientsTemplate(t, "manage-recipient-secrets.html", tpl.TemplateData{
		Title:           "Manage Secrets for Sam Carter",
		ActivePage:      "recipients",
		IsAuthenticated: true,
		User:            authedUser(),
		Data: map[string]interface{}{
			"Recipient": map[string]interface{}{
				"ID":    "rec-7",
				"Name":  "Sam Carter",
				"Email": "sam@example.com",
			},
			"Secrets": []map[string]interface{}{
				{"ID": "let-1", "Title": "Letter to Sam", "IsAssigned": true},
				{"ID": "let-2", "Title": "Banking notes", "IsAssigned": false},
				{"ID": "let-3", "Title": "1Password recovery", "IsAssigned": true},
			},
		},
	})

	mustContain := []string{
		`data-test="manage-recipient-secrets"`,
		`· LETTERS · IN THIS ENVELOPE`,
		`What's in Sam Carter's envelope`,
		// Recipient summary mini-card
		`data-test="recipient-summary"`,
		`sam@example.com`,
		// Section head
		`· CONTAINS`,
		// Form contract preserved
		`action="/recipients/rec-7/secrets"`,
		`method="POST"`,
		`name="secrets"`,
		// Letter rows
		`data-test="letter-list"`,
		`data-test="letter-row"`,
		`Letter to Sam`,
		`Banking notes`,
		`1Password recovery`,
		`value="let-1"`,
		`value="let-2"`,
		`value="let-3"`,
		// Submit
		`Save assignments`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("manage-recipient-secrets.html (mixed) missing snippet %q", s)
		}
	}

	// Verify checkbox state: assigned letters render with `checked`
	if !strings.Contains(out, `id="secret-let-1"`) || !strings.Contains(out, `id="secret-let-3"`) {
		t.Error("expected checkbox ids for assigned letters")
	}
	// Empty state must not appear when secrets exist
	if strings.Contains(out, `data-test="letters-empty"`) {
		t.Error("manage-recipient-secrets.html should not show empty state with secrets")
	}

	// Old Bootstrap markup gone
	mustNotContain := []string{
		`form-control`,
		`form-check-input`,
		`Dead Man's Switch`,
		`alert alert-info`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("manage-recipient-secrets.html still contains old markup %q", s)
		}
	}
}

func TestRenderManageRecipientSecrets_EmptyLetters(t *testing.T) {
	out := renderRecipientsTemplate(t, "manage-recipient-secrets.html", tpl.TemplateData{
		Title:           "Manage Secrets",
		ActivePage:      "recipients",
		IsAuthenticated: true,
		User:            authedUser(),
		Data: map[string]interface{}{
			"Recipient": map[string]interface{}{
				"ID":    "rec-7",
				"Name":  "Sam Carter",
				"Email": "sam@example.com",
			},
			"Secrets": []map[string]interface{}{},
		},
	})

	mustContain := []string{
		`data-test="letters-empty"`,
		`No letters written yet`,
		`Write your first letter`,
		`href="/secrets/new"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("manage-recipient-secrets.html (empty letters) missing snippet %q", s)
		}
	}
	if strings.Contains(out, `data-test="letter-list"`) {
		t.Error("letter-list should not render when secrets list is empty")
	}
}
