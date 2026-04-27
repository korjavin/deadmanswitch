package web

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"testing"

	tpl "github.com/korjavin/deadmanswitch/internal/web/templates"
)

// renderLayout parses layout.html alongside a stub content child and
// returns the rendered HTML. We supply empty defines for any blocks
// child templates would normally fill so the layout can run alone.
func renderLayout(t *testing.T, data tpl.TemplateData) string {
	t.Helper()
	root := projectRoot(t)
	layoutPath := filepath.Join(root, "web", "templates", "layout.html")

	stub := `{{ define "content" }}<div data-test="content-block"></div>{{ end }}
{{ define "styles" }}{{ end }}
{{ define "scripts" }}{{ end }}`

	tmpl, err := template.New("layout.html").Funcs(tpl.FuncMap()).ParseFiles(layoutPath)
	if err != nil {
		t.Fatalf("parse layout: %v", err)
	}
	if _, err := tmpl.Parse(stub); err != nil {
		t.Fatalf("parse stub: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		t.Fatalf("execute: %v", err)
	}
	return buf.String()
}

func TestLayoutTitleSaysHeartbeat(t *testing.T) {
	out := renderLayout(t, tpl.TemplateData{Title: "Dashboard"})
	if !strings.Contains(out, "<title>Dashboard — Heartbeat</title>") {
		t.Errorf("title missing or wrong; got snippet:\n%s", titleSnippet(out))
	}
	if strings.Contains(out, "Dead Man's Switch") {
		t.Errorf("rendered output still contains old brand string 'Dead Man's Switch'")
	}
}

func TestLayoutAuthenticatedShowsSidebar(t *testing.T) {
	out := renderLayout(t, tpl.TemplateData{
		Title:           "Dashboard",
		IsAuthenticated: true,
		ActivePage:      "dashboard",
		User:            map[string]interface{}{"Email": "alex@example.com"},
	})
	mustContain := []string{
		`class="app"`,
		`class="topbar"`,
		`class="sidebar"`,
		`class="nav-group-label">· TODAY`,
		`class="nav-group-label">· PEOPLE`,
		`class="nav-group-label">· LETTERS`,
		`class="nav-group-label">· ACCOUNT`,
		`alex@example.com`,
		`heartbeat`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("authenticated layout missing expected snippet %q", s)
		}
	}
	mustNotContain := []string{
		`class="marketing-bar"`,
		`class="marketing-footer"`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("authenticated layout unexpectedly contains %q", s)
		}
	}
}

func TestLayoutUnauthenticatedShowsMarketing(t *testing.T) {
	out := renderLayout(t, tpl.TemplateData{
		Title:           "Welcome",
		IsAuthenticated: false,
	})
	mustContain := []string{
		`class="marketing"`,
		`class="marketing-bar"`,
		`class="marketing-footer"`,
		`HEARTBEAT · MADE WITH CARE · ©2026`,
		`href="/login"`,
		`href="/register"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("unauthenticated layout missing expected snippet %q", s)
		}
	}
	mustNotContain := []string{
		`class="sidebar"`,
		`grid-area: sidebar`,
		`<nav class="sidebar"`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("unauthenticated layout unexpectedly contains %q", s)
		}
	}
}

// TestLayoutActiveNavFlips renders the layout for each authenticated
// ActivePage value the sidebar cares about and asserts only the
// matching nav-item gets the .active class.
func TestLayoutActiveNavFlips(t *testing.T) {
	cases := []struct {
		page    string
		hrefHit string
	}{
		{"dashboard", "/dashboard"},
		{"recipients", "/recipients"},
		{"secrets", "/secrets"},
		{"settings", "/settings"},
		{"profile", "/profile"},
		{"history", "/history"},
	}
	for _, tc := range cases {
		t.Run(tc.page, func(t *testing.T) {
			out := renderLayout(t, tpl.TemplateData{
				Title:           "X",
				IsAuthenticated: true,
				ActivePage:      tc.page,
				User:            map[string]interface{}{"Email": "x@x"},
			})

			activeMarker := `href="` + tc.hrefHit + `" class="nav-item active"`
			if !strings.Contains(out, activeMarker) {
				t.Errorf("expected %q to be active for ActivePage=%q\n--- output ---\n%s",
					tc.hrefHit, tc.page, sidebarSnippet(out))
			}

			// Every other listed page's nav-item must not carry .active.
			for _, other := range cases {
				if other.page == tc.page {
					continue
				}
				badMarker := `href="` + other.hrefHit + `" class="nav-item active"`
				if strings.Contains(out, badMarker) {
					t.Errorf("page %q unexpectedly active when ActivePage=%q",
						other.page, tc.page)
				}
			}
		})
	}
}

func TestLayoutLoadsBrandLogo(t *testing.T) {
	out := renderLayout(t, tpl.TemplateData{
		Title:           "X",
		IsAuthenticated: true,
		User:            map[string]interface{}{"Email": "x@x"},
	})
	// Old emoji brand should be gone.
	if strings.Contains(out, "🔐") {
		t.Errorf("layout still contains old emoji brand")
	}
	// New brand should include the heart-pulse SVG path.
	if !strings.Contains(out, `class="brand-mark"`) {
		t.Errorf("layout missing brand SVG mark")
	}
	if !strings.Contains(out, `M5 12h3l1.5-3 3 6L15 12h4`) {
		t.Errorf("layout missing heart-pulse trace path")
	}
}

func TestLayoutHeartbeatVariantReflectsTemplateData(t *testing.T) {
	cases := []struct {
		name    string
		variant string
		label   string
		wantCls string
		wantTxt string
	}{
		{"default", "", "", `class="heartbeat ok"`, "HEARTBEAT · OK"},
		{"warn", "warn", "OVERDUE", `class="heartbeat warn"`, "HEARTBEAT · OVERDUE"},
		{"crit", "crit", "URGENT", `class="heartbeat crit"`, "HEARTBEAT · URGENT"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := renderLayout(t, tpl.TemplateData{
				Title:            "X",
				IsAuthenticated:  true,
				User:             map[string]interface{}{"Email": "x@x"},
				HeartbeatVariant: tc.variant,
				HeartbeatLabel:   tc.label,
			})
			if !strings.Contains(out, tc.wantCls) {
				t.Errorf("layout missing %q for variant %q", tc.wantCls, tc.variant)
			}
			if !strings.Contains(out, tc.wantTxt) {
				t.Errorf("layout missing %q for label %q", tc.wantTxt, tc.label)
			}
		})
	}
}

func TestLayoutFlashRendersAsMonoLabel(t *testing.T) {
	out := renderLayout(t, tpl.TemplateData{
		Title:           "X",
		IsAuthenticated: true,
		User:            map[string]interface{}{"Email": "x@x"},
		Flash:           map[string]string{"info": "Welcome back."},
	})
	if !strings.Contains(out, `class="flash-stack"`) {
		t.Errorf("flash block missing")
	}
	if !strings.Contains(out, `class="alert alert-info"`) {
		t.Errorf("flash alert class missing")
	}
	if !strings.Contains(out, `class="label">· info`) {
		t.Errorf("flash label not rendered as mono dotted label")
	}
}

func titleSnippet(html string) string {
	i := strings.Index(html, "<title>")
	if i < 0 {
		return "(no <title> in output)"
	}
	end := strings.Index(html[i:], "</title>")
	if end < 0 {
		end = 80
	}
	return html[i : i+end+len("</title>")]
}

func sidebarSnippet(html string) string {
	i := strings.Index(html, `class="sidebar"`)
	if i < 0 {
		return "(no sidebar)"
	}
	end := strings.Index(html[i:], "</nav>")
	if end < 0 {
		end = 1500
	}
	return html[i : i+end]
}
