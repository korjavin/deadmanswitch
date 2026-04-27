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

// renderSecretsTemplate parses layout.html + the named template and
// returns the rendered HTML in the authenticated app shell.
func renderSecretsTemplate(t *testing.T, name string, data tpl.TemplateData) string {
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

func authedSecretsUser() map[string]interface{} {
	return map[string]interface{}{
		"Email": "alex@example.com",
		"Name":  "alex@example.com",
	}
}

// ---------------------------------------------------------------------------
// secrets.html — letter grid
// ---------------------------------------------------------------------------

func TestRenderSecrets_LetterGrid(t *testing.T) {
	updated := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	created := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)

	secrets := []map[string]interface{}{
		{
			"ID":        "let-1",
			"Title":     "Letter to Sam",
			"Type":      "encrypted",
			"Content":   "Dear Sam, if you're reading this it means I'm gone for a while...",
			"CreatedAt": created,
			"UpdatedAt": updated,
			"Recipients": []map[string]interface{}{
				{"ID": "r1", "Name": "Sam Carter", "Email": "sam@example.com"},
				{"ID": "r2", "Name": "Jamie Lee", "Email": "jamie@example.com"},
			},
		},
		{
			"ID":        "let-2",
			"Title":     "Banking notes",
			"Type":      "encrypted",
			"Content":   "The savings account is at...",
			"CreatedAt": created,
			"UpdatedAt": updated,
			"Recipients": []map[string]interface{}{
				{"ID": "r1", "Name": "Sam Carter", "Email": "sam@example.com"},
			},
		},
		{
			"ID":         "let-3",
			"Title":      "1Password recovery",
			"Type":       "encrypted",
			"Content":    "Master password recovery hint...",
			"CreatedAt":  created,
			"UpdatedAt":  updated,
			"Recipients": []map[string]interface{}{},
		},
	}

	out := renderSecretsTemplate(t, "secrets.html", tpl.TemplateData{
		Title:           "Letters",
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User:            authedSecretsUser(),
		Data: map[string]interface{}{
			"Secrets": secrets,
		},
	})

	mustContain := []string{
		`data-test="letters"`,
		`· YOUR LETTERS`,
		`What you've written`,
		`encrypted at rest`,
		`data-test="add-letter"`,
		`href="/secrets/new"`,
		`New letter`,
		// Grid + cards
		`data-test="letter-grid"`,
		`data-test="letter-card"`,
		`data-test="letter-kind"`,
		`✉ LETTER`,
		// Titles + previews
		`Letter to Sam`,
		`Dear Sam, if you&#39;re reading this`,
		`Banking notes`,
		`1Password recovery`,
		// Recipient summaries (mixed)
		`to Sam Carter +1`,
		`to Sam Carter`,
		`· no recipient`,
		// Per-letter URLs preserved
		`href="/secrets/let-1"`,
		`href="/secrets/let-1/assign"`,
		`action="/secrets/let-1"`,
		`name="_method" value="DELETE"`,
		// Trailing add card
		`data-test="add-another-letter"`,
		`Write another letter`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("secrets.html (grid) missing snippet %q", s)
		}
	}

	// Empty-state markers must NOT appear when populated.
	if strings.Contains(out, `data-test="letters-empty"`) {
		t.Error("secrets.html should not show empty-state when letters present")
	}

	// Old Bootstrap markup must be gone.
	mustNotContain := []string{
		`secret-card`,
		`btn btn-primary`,
		`Add New Secret`,
		`No Secrets Yet`,
		`Dead Man's Switch`,
		`Add First Secret`,
		`alert alert-info`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("secrets.html still contains old markup %q", s)
		}
	}
}

func TestRenderSecrets_EmptyState(t *testing.T) {
	out := renderSecretsTemplate(t, "secrets.html", tpl.TemplateData{
		Title:           "Letters",
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User:            authedSecretsUser(),
		Data: map[string]interface{}{
			"Secrets": []map[string]interface{}{},
		},
	})

	mustContain := []string{
		`data-test="letters-empty"`,
		`Nothing written yet.`,
		`Write your first letter`,
		`href="/secrets/new"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("secrets.html (empty) missing snippet %q", s)
		}
	}

	if strings.Contains(out, `data-test="letter-grid"`) {
		t.Error("secrets.html should not render letter-grid when empty")
	}
	if strings.Contains(out, `data-test="letter-card"`) {
		t.Error("secrets.html should not render letter-card when empty")
	}
}

// ---------------------------------------------------------------------------
// new-secret.html — kind selector + editor
// ---------------------------------------------------------------------------

func TestRenderNewSecret_KindSelector(t *testing.T) {
	out := renderSecretsTemplate(t, "new-secret.html", tpl.TemplateData{
		Title:           "Begin a letter",
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User:            authedSecretsUser(),
		Data: map[string]interface{}{
			"Recipients": []map[string]interface{}{
				{"ID": "r1", "Name": "Sam Carter", "Email": "sam@example.com"},
				{"ID": "r2", "Name": "Jamie Lee", "Email": "jamie@example.com"},
			},
		},
	})

	mustContain := []string{
		`data-test="new-letter"`,
		`· WRITE A NEW LETTER`,
		`Begin a letter`,
		// Form contract preserved
		`action="/secrets/new"`,
		`method="POST"`,
		`name="title"`,
		`name="content"`,
		`name="recipients"`,
		// Kind selector — three radio cards, letter selected by default
		`data-test="kind-group"`,
		`data-test="kind-letter"`,
		`data-test="kind-instruction"`,
		`data-test="kind-credential"`,
		`name="kind"`,
		`value="letter"`,
		`value="instruction"`,
		`value="credential"`,
		`✉`,
		`☐`,
		`⚿`,
		// Letter is the default selected kind (radio checked + class selected)
		`hb-kind selected" data-test="kind-letter"`,
		// Body uses serif font by default (kind-letter — no kind-credential class on initial)
		`hb-body-text`,
		// Mono labels
		`· KIND`,
		`· TITLE`,
		`· BODY`,
		`· SEND TO`,
		// Encrypted-at-rest banner
		`data-test="encryption-note"`,
		`ENCRYPTED AT REST`,
		`TIME-BOXED`,
		`ONLY UNSEALED ON DELIVERY`,
		// Recipient list
		`data-test="recipient-list"`,
		`Sam Carter`,
		`Jamie Lee`,
		`value="r1"`,
		`value="r2"`,
		// Submit
		`Save letter`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("new-secret.html missing snippet %q", s)
		}
	}

	// Old Bootstrap markup must be gone.
	mustNotContain := []string{
		`form-control`,
		`form-check-input`,
		`Add New Secret</h1>`,
		`Save Secret`,
		`Dead Man's Switch`,
		`Manage Recipients</h3>`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("new-secret.html still contains old markup %q", s)
		}
	}
}

func TestRenderNewSecret_NoRecipientsEmpty(t *testing.T) {
	out := renderSecretsTemplate(t, "new-secret.html", tpl.TemplateData{
		Title:           "Begin a letter",
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User:            authedSecretsUser(),
		Data: map[string]interface{}{
			"Recipients": []map[string]interface{}{},
		},
	})

	mustContain := []string{
		`data-test="recipient-empty"`,
		`No people in your circle yet.`,
		`href="/recipients/new"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("new-secret.html (no recipients) missing snippet %q", s)
		}
	}
	if strings.Contains(out, `data-test="recipient-list"`) {
		t.Error("recipient-list should not render when no recipients")
	}
}

// ---------------------------------------------------------------------------
// view-secret.html — metadata strip + pre-filled
// ---------------------------------------------------------------------------

func TestRenderViewSecret_MetadataStrip(t *testing.T) {
	created := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 4, 20, 14, 30, 0, 0, time.UTC)

	out := renderSecretsTemplate(t, "view-secret.html", tpl.TemplateData{
		Title:           "Edit Letter",
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User:            authedSecretsUser(),
		Data: map[string]interface{}{
			"Secret": map[string]interface{}{
				"ID":             "let-42",
				"Name":           "Letter to Sam",
				"Type":           "note",
				"Content":        "Dear Sam, if you're reading this...",
				"CreatedAt":      created,
				"LastModified":   updated,
				"EncryptionType": "aes-256-gcm",
			},
			"Recipients": []map[string]interface{}{
				{"ID": "r1", "Name": "Sam Carter", "Email": "sam@example.com", "IsAssigned": true},
				{"ID": "r2", "Name": "Jamie Lee", "Email": "jamie@example.com", "IsAssigned": false},
			},
		},
	})

	mustContain := []string{
		`data-test="view-letter"`,
		`· EDIT LETTER`,
		// Title rendered as h1
		`<h1>Letter to Sam</h1>`,
		// Metadata strip
		`data-test="meta-strip"`,
		`· CREATED`,
		`· LAST EDITED`,
		`· ENCRYPTION`,
		`· STATUS`,
		`Mar 15, 2026`,
		`Apr 20, 2026`,
		`aes-256-gcm`,
		`🔒 Sealed`,
		// Form contract preserved
		`action="/secrets/let-42"`,
		`method="POST"`,
		`name="title"`,
		`name="content"`,
		`name="recipients"`,
		`name="kind"`,
		// Pre-filled values
		`value="Letter to Sam"`,
		`Dear Sam, if you&#39;re reading this...`,
		// Kind selector default = letter
		`hb-kind selected" data-test="kind-letter"`,
		// Recipients pre-checked correctly
		`id="recipient-r1"`,
		`id="recipient-r2"`,
		// Encrypted note
		`ENCRYPTED AT REST`,
		// Submit + delete
		`Save changes`,
		`name="_method" value="DELETE"`,
		`Delete this letter`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("view-secret.html missing snippet %q", s)
		}
	}

	// Sam (assigned) should have checked attribute on its <input>; Jamie's must not.
	idxR1 := strings.Index(out, `id="recipient-r1"`)
	idxR2 := strings.Index(out, `id="recipient-r2"`)
	if idxR1 < 0 || idxR2 < 0 {
		t.Fatal("expected both recipient inputs in output")
	}
	// The `checked` attribute renders right after id="recipient-{ID}" within the same <input> tag.
	r1Tail := out[idxR1:min0(idxR1+200, len(out))]
	r1Tag := r1Tail[:strings.Index(r1Tail, ">")+1]
	if !strings.Contains(r1Tag, "checked") {
		t.Errorf("expected recipient r1 input to be checked, got tag: %q", r1Tag)
	}
	r2Tail := out[idxR2:min0(idxR2+200, len(out))]
	r2Tag := r2Tail[:strings.Index(r2Tail, ">")+1]
	if strings.Contains(r2Tag, "checked") {
		t.Errorf("expected recipient r2 input to NOT be checked, got tag: %q", r2Tag)
	}

	// Old markup gone
	mustNotContain := []string{
		`form-control`,
		`form-check-input`,
		`Edit Secret</h1>`,
		`Manage Recipients</h3>`,
		`alert alert-info`,
		`timeline-content`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("view-secret.html still contains old markup %q", s)
		}
	}
}

// ---------------------------------------------------------------------------
// manage-secret-recipients.html — envelope rows
// ---------------------------------------------------------------------------

func TestRenderManageSecretRecipients_EnvelopeRows(t *testing.T) {
	out := renderSecretsTemplate(t, "manage-secret-recipients.html", tpl.TemplateData{
		Title:           "Who gets this letter",
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User:            authedSecretsUser(),
		Data: map[string]interface{}{
			"Secret": map[string]interface{}{
				"ID":    "let-7",
				"Title": "Letter to Sam",
			},
			"Recipients": []map[string]interface{}{
				{"ID": "r1", "Name": "Sam Carter", "Email": "sam@example.com", "IsAssigned": true},
				{"ID": "r2", "Name": "Jamie Lee", "Email": "jamie@example.com", "IsAssigned": false},
				{"ID": "r3", "Name": "Robin Park", "Email": "robin@example.com", "IsAssigned": true},
			},
		},
	})

	mustContain := []string{
		`data-test="manage-secret-recipients"`,
		`· ENVELOPES · WHO RECEIVES THIS LETTER`,
		`Who gets &ldquo;Letter to Sam&rdquo;`,
		// Letter summary mini-card
		`data-test="letter-summary"`,
		`· LETTER`,
		// Section head
		`· ENVELOPES`,
		`TICK TO INCLUDE`,
		// Form contract preserved
		`action="/secrets/let-7/assign"`,
		`method="POST"`,
		`name="recipients"`,
		// Envelope rows
		`data-test="envelope-list"`,
		`data-test="envelope-row"`,
		`Sam Carter`,
		`sam@example.com`,
		`Jamie Lee`,
		`jamie@example.com`,
		`Robin Park`,
		`robin@example.com`,
		// Checkbox states
		`value="r1"`,
		`value="r2"`,
		`value="r3"`,
		`id="recipient-r1"`,
		`id="recipient-r2"`,
		`id="recipient-r3"`,
		// Sub-line markers
		`· CURRENTLY HOLDS THIS LETTER`,
		`· DOES NOT HOLD THIS LETTER`,
		// Submit
		`Save assignments`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("manage-secret-recipients.html missing snippet %q", s)
		}
	}

	// Old markup gone
	mustNotContain := []string{
		`form-control`,
		`form-check-input`,
		`Dead Man's Switch`,
		`alert alert-info`,
		`Secret Details</h3>`,
		`Assign Recipients</h3>`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("manage-secret-recipients.html still contains old markup %q", s)
		}
	}

	// Empty-state must not appear when populated.
	if strings.Contains(out, `data-test="recipients-empty"`) {
		t.Error("manage-secret-recipients.html should not show empty when recipients present")
	}
}

func TestRenderManageSecretRecipients_EmptyRecipients(t *testing.T) {
	out := renderSecretsTemplate(t, "manage-secret-recipients.html", tpl.TemplateData{
		Title:           "Who gets this letter",
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User:            authedSecretsUser(),
		Data: map[string]interface{}{
			"Secret": map[string]interface{}{
				"ID":    "let-7",
				"Title": "Letter to Sam",
			},
			"Recipients": []map[string]interface{}{},
		},
	})

	mustContain := []string{
		`data-test="recipients-empty"`,
		`Nobody in your circle yet`,
		`Add your first person`,
		`href="/recipients/new"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("manage-secret-recipients.html (empty) missing snippet %q", s)
		}
	}
	if strings.Contains(out, `data-test="envelope-list"`) {
		t.Error("envelope-list should not render when recipients empty")
	}
}

func min0(a, b int) int {
	if a < b {
		return a
	}
	return b
}
