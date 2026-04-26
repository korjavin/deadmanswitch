package web

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"testing"

	tpl "github.com/korjavin/deadmanswitch/internal/web/templates"
)

// renderIndex parses layout.html + index.html together and returns
// the rendered HTML in the unauthenticated state — landing is only
// shown to anonymous visitors.
func renderIndex(t *testing.T, data tpl.TemplateData) string {
	t.Helper()
	root := projectRoot(t)
	layoutPath := filepath.Join(root, "web", "templates", "layout.html")
	indexPath := filepath.Join(root, "web", "templates", "index.html")

	tmpl, err := template.New("layout.html").Funcs(tpl.FuncMap()).
		ParseFiles(layoutPath, indexPath)
	if err != nil {
		t.Fatalf("parse landing templates: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		t.Fatalf("execute landing: %v", err)
	}
	return buf.String()
}

func TestLandingRendersHeartbeatHero(t *testing.T) {
	out := renderIndex(t, tpl.TemplateData{
		Title:           "A quiet kind of safety net",
		ActivePage:      "home",
		IsAuthenticated: false,
	})

	mustContain := []string{
		// Hero kicker + serif headline + lede
		`· A QUIET KIND OF SAFETY NET`,
		`Write the things`,
		`you'd want them to know`,
		`if you couldn't tell them.`,
		`Heartbeat is a small service that watches for your absence`,
		// Hero CTAs
		`href="/register"`,
		`Begin writing`,
		`I already have an account`,
		// Meta row
		`SELF-HOSTED · DATA SOVEREIGN`,
		`TIME-BOXED ENCRYPTION`,
		`OPEN SOURCE`,
		// Stacked envelopes
		`class="hero-letters"`,
		`class="envelope envelope-a"`,
		`class="envelope envelope-b"`,
		`class="envelope envelope-c"`,
		`For Mom`,
		`For Sam`,
		`For my kids, when they're older`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("landing hero missing snippet %q", s)
		}
	}
}

func TestLandingRendersThreeStepRhythm(t *testing.T) {
	out := renderIndex(t, tpl.TemplateData{
		IsAuthenticated: false,
	})

	mustContain := []string{
		`· THE WHOLE THING`,
		`Three small habits, one big peace of mind.`,
		`01 · STEP`,
		`02 · STEP`,
		`03 · STEP`,
		`Write your letters`,
		`Live your life`,
		`If we go silent`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("3-step rhythm missing snippet %q", s)
		}
	}
}

func TestLandingRendersSixChannels(t *testing.T) {
	out := renderIndex(t, tpl.TemplateData{
		IsAuthenticated: false,
	})

	mustContain := []string{
		`· QUIETLY EVERYWHERE`,
		`We watch in six soft ways.`,
		`✓ Just signing in`,
		`✓ GitHub commits`,
		`✓ Telegram reply`,
		`✓ Email link tap`,
		`✓ Passkey check`,
		`✓ Personal URL`,
		// Sample nudge email card
		`EXAMPLE NUDGE · DAY 31`,
		`From: heartbeat`,
		`I'm here`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("six-channels section missing snippet %q", s)
		}
	}
}

func TestLandingRendersUseCases(t *testing.T) {
	out := renderIndex(t, tpl.TemplateData{
		IsAuthenticated: false,
	})

	mustContain := []string{
		`· WHAT PEOPLE WRITE`,
		`Letters, instructions, and the keys to everything.`,
		// Three glyphs/kinds
		`✉`,
		`☐`,
		`⚿`,
		`· A LETTER`,
		`· INSTRUCTIONS`,
		`· SECRETS &amp; KEYS`,
		`For the people you love`,
		`For when they have to figure things out`,
		`Passwords, vaults, recovery codes`,
		// The third (secrets) card is highlighted
		`class="usecase usecase-highlight"`,
		// Vault detail block
		`· ON SECRETS, SPECIFICALLY`,
		`Your vault, on your server, unsealed only when time runs out.`,
		`time-boxed encryption`,
		`✓ Self-hosted (Docker / k8s)`,
		`✓ Encrypted at rest (AES-256)`,
		`✓ Time-boxed key release`,
		// Mono-pre sample
		`1PASSWORD RECOVERY`,
		`SEALED · 2,140 BYTES`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("use-cases / vault section missing snippet %q", s)
		}
	}
}

func TestLandingRendersTestimonialAndCTA(t *testing.T) {
	out := renderIndex(t, tpl.TemplateData{
		IsAuthenticated: false,
	})

	mustContain := []string{
		`· FROM A USER`,
		`I used to keep a sealed envelope in my desk drawer.`,
		`MARGUERITE H. · USING HEARTBEAT FOR 2 YEARS`,
		`Begin with one letter.`,
		`Pick someone you love. Write the thing. We'll keep it safe.`,
		`Start your first letter`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("testimonial/CTA missing snippet %q", s)
		}
	}
}

func TestLandingShowsMarketingChromeAndFooter(t *testing.T) {
	out := renderIndex(t, tpl.TemplateData{
		IsAuthenticated: false,
	})

	// Layout supplies marketing-bar + footer; landing should sit inside
	// the unauthenticated shell.
	mustContain := []string{
		`class="marketing-bar"`,
		`class="marketing-footer"`,
		`HEARTBEAT · MADE WITH CARE · ©2026`,
		`href="/login"`,
		`href="/register"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("landing marketing chrome missing snippet %q", s)
		}
	}
	// Sidebar must not appear on the landing page (anon).
	if strings.Contains(out, `class="sidebar"`) {
		t.Errorf("landing unexpectedly contains authenticated sidebar")
	}
}

// TestLandingRemovesOldCopy guards against regressions of the
// pre-redesign landing — the agent should have replaced every
// piece of the old hero/feature deck.
func TestLandingRemovesOldCopy(t *testing.T) {
	out := renderIndex(t, tpl.TemplateData{
		IsAuthenticated: false,
	})

	mustNotContain := []string{
		`Secure Your Digital Legacy`,
		`Our dead man's switch`,
		`Get Started`,
		`feature-card`,
		`feature-icon`,
		`Store Your Secrets`,
		`Designate Recipients`,
		`Regular Check-ins`,
		`Automatic Delivery`,
		`Security First`,
		`End-to-End Encryption`,
		`Ready to Secure Your Digital Future`,
		`👥`,
		`⏰`,
		`📨`,
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("landing still contains old copy/markup %q", s)
		}
	}
}
