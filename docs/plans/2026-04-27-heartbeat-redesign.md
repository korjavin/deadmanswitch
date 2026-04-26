# Heartbeat Redesign — Apply Claude Design Bundle

## Overview

Port the **Heartbeat** design (delivered as a React/JSX prototype bundle from Claude Design) onto the existing Go-templated web UI. The design replaces the current Bootstrap-ish look with a "warm-paper, monospace, hand-drawn" aesthetic and re-frames the product around three metaphors:

- **Heartbeat** — the soft name for the dead-man's-switch mechanic ("you're good", "next nudge in 5 days") instead of alarm/medical language
- **Envelopes** — each recipient is rendered as a sealed envelope card; secrets become "letters" placed inside
- **Multi-channel check-in** — the dashboard surfaces six soft ways the system already counts you as alive (web login, GitHub, Telegram, email link, passkey, personal URL)

Decision recap (from planning conversation):
- **Naming**: rebrand only the *user-facing UI* to Heartbeat / Letters / People. Backend Go packages, repo, env vars, and docs keep "deadmanswitch" / "secret" / "recipient" terminology — no churn outside `web/`.
- **Scope**: redesign **all 19 templates** in `web/templates/` for a cohesive rebrand.
- **Approach**: foundation (tokens + fonts + base CSS + layout shell) first, then one template (or tight cluster) per task.

## Context (from discovery)

**Existing app structure:**
- Go templates in `web/templates/` (19 files): `layout.html`, `index.html`, `login.html`, `register.html`, `login-2fa.html`, `2fa-setup.html`, `passkeys.html`, `verify-success.html`, `confirmation.html`, `dashboard.html`, `recipients.html`, `new-recipient.html`, `manage-recipient-secrets.html`, `secrets.html`, `new-secret.html`, `view-secret.html`, `manage-secret-recipients.html`, `settings.html`, `profile.html`, `history.html`
- Static assets: `web/static/css/{normalize.css, main.css}`, `web/static/js/main.js`
- Server: `internal/web/server.go` parses templates; handlers in `internal/web/handlers/`
- Existing handler tests: `dashboard_test.go`, `recipients_test.go`, `secrets_test.go`, `history_test.go`, `api_test.go` — these execute templates and inspect output, which gives us a natural place to add render assertions for each redesigned template.
- Playwright is currently disabled in CI (commit `2d7600d`). E2E coverage stays out of scope; we rely on Go-level template render tests.

**Design bundle (cached locally at `.design-fetch/deadmanswitch/`):**
- `project/Heartbeat.html` — entry shell, loads tokens.css, app.css, JSX prototype files
- `project/tokens.css` — paper-warmth color tokens, JetBrains Mono font face, type scale
- `project/app.css` — buttons, inputs, cards, sidebar, topbar, heartbeat pulse, modal, toast
- `project/landing.jsx` — full marketing page (hero with stacked envelope illustration, 3-step rhythm, multi-channel section, 3 use cases incl. secrets vault detail, testimonial, CTA)
- `project/dashboard.jsx` — soft-status copy, status timeline strip, 6 channel cards, recent activity + circle
- `project/recipients.jsx` — envelope grid + add-modal + recipient drawer
- `project/letters.jsx`, `auth.jsx`, `other-pages.jsx` — auth, settings, profile, history
- `project/primitives.jsx` — Logo, Icon (inline SVG set), Card, Badge, PageHead, SectionHead, Modal, Field
- `project/fonts/JetBrainsMono-400.woff2` (+italic) — self-host these
- `project/assets/dotgrid-bg.svg`, `blueprint-plate-01.svg` — copy to `web/static/assets/`

**Patterns identified:**
- Color system uses **oklch** in tokens.css (paper-warmth ivory) — modern browsers support this natively; no Sass needed.
- Buttons use **offset shadow** (`box-shadow: 2px 2px 0 0 var(--line-1)`) and translate on `:active` for the analog stationery feel.
- Layout shell: `display: grid; grid-template: "topbar topbar" "sidebar main" / 256px 1fr` for authenticated views; sticky `marketing-bar` for unauth.
- Iconography: small inline SVG path set inside primitives.jsx (`heart`, `mail`, `send`, `log-in`, `git-branch`, `fingerprint`, `key`, `users-round`, `user-plus`, `settings`, `feather`, `bell`, `check`, etc.). Port to `web/static/js/icons.js` or inline as `{{template "icon" "name"}}` partial.

**Constraints carried over from earlier project decisions:**
- Self-hosted, data-sovereign — **self-host the woff2 fonts**, do not load Google Fonts CDN. Drop the `<link href="fonts.googleapis.com…">` from the design.
- No SMS/phone, no pricing nav (already removed in design's chat round 2).
- Hero copy emphasizes self-host + time-boxed encryption + open source.

## Development Approach

- **Testing approach**: Regular (HTML changes first, then Go render tests). Pure UI work — TDD doesn't fit because the "expected output" is the design itself, which we're transcribing.
- Complete each task fully before moving to the next.
- Make small, focused changes — one template (or tight cluster) per task so review is tractable.
- **CRITICAL: every task MUST include new/updated tests** for templates touched in that task:
  - For each redesigned template: write a Go test (in or alongside the matching `internal/web/handlers/*_test.go`) that renders the template with representative data and asserts the new structural markers appear (e.g. `class="page-head"`, `· HEARTBEAT`, `class="envelope"`).
  - Update existing handler tests if their assertions reference now-removed copy (e.g. "Dead Man's Switch" in titles).
  - Tests cover a happy path plus one empty/edge state per template (e.g. recipients with zero entries → empty state markup).
- **CRITICAL: all tests must pass before starting next task** — `go test ./...` must be green.
- **CRITICAL: update this plan file when scope changes during implementation.**
- Run `go test ./...` and `./scripts/run-golangci-lint.sh` after each task.
- Maintain backward compatibility: no template-data shape changes; only HTML structure and CSS change. Handlers stay untouched.

## Testing Strategy

- **Unit/render tests**: required for every task that touches a template. Pattern:
  ```go
  func TestRenderDashboard_Heartbeat(t *testing.T) {
      out := renderTemplate(t, "dashboard.html", sampleDashboardData())
      assert.Contains(t, out, `class="page-head"`)
      assert.Contains(t, out, "You're good") // soft status copy
      assert.NotContains(t, out, "System Active") // old alarm copy
  }
  ```
- **E2E tests**: project Playwright suite is currently disabled in CI; we will not re-enable it as part of this work. Out of scope.
- **Manual verification** lives in Post-Completion (no checkbox) — the agent cannot truly evaluate visual fidelity, so the human signs off.

## Progress Tracking

- Mark completed items with `[x]` immediately when done.
- Add newly discovered tasks with ➕ prefix.
- Document issues/blockers with ⚠️ prefix.
- Update plan if implementation deviates from original scope.
- Keep plan in sync with actual work done.

## What Goes Where

- **Implementation Steps** (`[ ]` checkboxes): in-repo edits — CSS, templates, Go render tests, asset copies.
- **Post-Completion** (no checkboxes): browser-based visual review, deployment, README screenshot updates.

## Implementation Steps

### Task 1: Foundation — design tokens, fonts, base CSS, assets
- [x] copy `JetBrainsMono-400.woff2` and `JetBrainsMono-400italic.woff2` from `.design-fetch/deadmanswitch/project/fonts/` to `web/static/fonts/`
- [x] copy `dotgrid-bg.svg` and `blueprint-plate-01.svg` from `.design-fetch/.../project/assets/` to `web/static/assets/`
- [x] create `web/static/css/tokens.css` based on the design's `tokens.css`: paper palette, type scale, spacing, motion vars, JetBrains Mono `@font-face` rules pointing to `/static/fonts/` (note: design ships hex paper palette, not oklch; we kept hex 1:1 with the bundle)
- [x] create `web/static/css/heartbeat.css` based on the design's `app.css`: button/input/card/badge/toggle/checkbox primitives, sidebar+topbar+main grid, page-head, section-head, heartbeat pulse, empty state, mono-pre, modal, toast
- [x] add `Inter` and `Source Serif 4` self-hosted woff2 (or CSS `@font-face` with system-fallback stacks) — chose system-fallback stacks (`system-ui`/`Georgia`) to avoid font-download cost; tokens declare `--serif`/`--sans` so a future task can swap in self-hosted woff2 without churn
- [x] update `web/templates/layout.html` `<head>` to load `tokens.css` + `heartbeat.css` instead of `normalize.css` + `main.css`; keep `main.css` available but unused for now (delete in final task)
- [x] write Go test that asserts the static handler serves `tokens.css` and `heartbeat.css` with `text/css` content-type and the woff2 fonts with `font/woff2` (added `internal/web/static_test.go`; woff2 check verifies the `wOF2` magic header rather than relying on host mime DB)
- [x] write a smoke test: render `index.html` (or any layout consumer) and assert the new stylesheet links appear in the `<head>` and the old ones do not (implemented as `TestLayoutLoadsHeartbeatStylesheets` reading layout.html directly — full template-render plumbing arrives in Task 2 along with the layout rewrite)
- [x] run `go test ./...` and `./scripts/run-golangci-lint.sh` — must pass before next task (all tests green; the 51 golangci-lint warnings are pre-existing gosec/gofmt issues in untouched files — none introduced by Task 1)

### Task 2: App shell — `layout.html` (topbar + sidebar + footer)
- [x] rewrite `web/templates/layout.html`:
  - Title format: `{{ .Title }} — Heartbeat` (only the user-visible string changes)
  - Authenticated layout: `.app` grid with `.topbar` (Logo + Heartbeat status pulse + user dropdown) and `.sidebar` (nav groups: TODAY · Dashboard, PEOPLE · Recipients, LETTERS · Secrets, ACCOUNT · Settings/Profile/History)
  - Unauthenticated layout: sticky `.marketing-bar` with Logo, links (How it works / The promise / Self-host / FAQ), Sign in + Begin buttons
  - Replace 🔐 emoji brand with the SVG heartbeat-and-pulse Logo from `primitives.jsx`
  - Replace footer with the design's terse mono footer (`HEARTBEAT · MADE WITH CARE · ©2026` + Privacy/Security/GitHub/Contact)
  - Keep all template `{{ block }}` / `{{ define }}` blocks named identically so child templates don't break
- [x] inline the heart-pulse SVG and small icon set used by sidebar (heart, mail, users-round, settings, clock, log-out, key) — chose `{{ define "icon-<name>" }}` partials inside `layout.html` so child templates can reuse them via `{{ template "icon-heart" . }}` without an extra file or JS dependency
- [x] preserve `Flash` / alert rendering but restyle with mono `.label` + dashed border
- [x] active-nav highlighting via existing `{{ if eq .ActivePage "..." }}` checks → swap to `class="nav-item active"`
- [x] write Go test: render layout in both authenticated and unauthenticated states, assert sidebar absent in unauth and present in auth, assert `Heartbeat` appears in `<title>`, assert old `Dead Man's Switch` brand is gone from rendered output (added `internal/web/layout_test.go` with `TestLayoutTitleSaysHeartbeat`, `TestLayoutAuthenticatedShowsSidebar`, `TestLayoutUnauthenticatedShowsMarketing`)
- [x] write Go test: assert active-nav class flips correctly when `ActivePage` changes (added `TestLayoutActiveNavFlips` covering all 6 sidebar pages)
- [x] run `go test ./...` — must pass before next task (all tests green; lint warnings unchanged from Task 1 baseline)

### Task 3: Landing — `index.html`
- [x] rewrite `web/templates/index.html` to mirror `landing.jsx`:
  - Hero (left): kicker `· A QUIET KIND OF SAFETY NET` + serif h1 ("Write the things you'd want them to know if you couldn't tell them.") + lede + Begin/Already-have-account buttons + meta row (SELF-HOSTED · DATA SOVEREIGN | TIME-BOXED ENCRYPTION | OPEN SOURCE)
  - Hero (right): `HeroLetters` — three rotated envelope cards, hand-drawn dotgrid backdrop. Implement as static HTML+CSS (rotation via `transform`, three fixed envelope contents)
  - Section "Three small habits, one big peace of mind." — three-column step grid (01/02/03), monospace numbering, dashed underline
  - Section "We watch in six soft ways." — two-column with sample nudge email card (mono `From:` + body + `I'm here →` button)
  - Section "What people write" — three-column use-case cards (✉ Letter / ☐ Instructions / ⚿ Secrets & Keys, the secrets one highlighted)
  - Section "On secrets, specifically" — two-column self-host vault card with mono pre showing 1Password recovery sample
  - Testimonial-as-letter (centered serif italic quote)
  - CTA ("Begin with one letter.") + serif h2
- [x] inline page-specific styles for envelope rotations and dotgrid radial-gradient backdrop in the `{{ define "styles" }}` block
- [x] update or remove now-irrelevant external links to GitHub README sections in nav (handled in Task 2 layout) — also updated `IndexHandler` title from "Dead Man's Switch" → "A quiet kind of safety net" so the rendered `<title>` reads "A quiet kind of safety net — Heartbeat"
- [x] write Go test for `/` route: assert hero copy ("Write the things"), three step labels (01/02/03), all six channel checkmarks, three use cases (LETTER, INSTRUCTIONS, SECRETS & KEYS), testimonial, footer (added `internal/web/landing_test.go` with `TestLandingRendersHeartbeatHero`, `…ThreeStepRhythm`, `…SixChannels`, `…UseCases`, `…TestimonialAndCTA`, `…ShowsMarketingChromeAndFooter`)
- [x] write test for old-copy removal: assert "Secure Your Digital Legacy" hero is gone, no `🔒` emoji in feature cards (added `TestLandingRemovesOldCopy` covering old hero/feature/CTA strings + 👥/⏰/📨 emoji; the `🔒` in the new vault mono-pre is intentional design copy and not a feature-card icon, so it is allowed)
- [x] run `go test ./...` — must pass before next task (all green; lint warnings unchanged from Task 2 baseline)

### Task 4: Auth flow — `login.html`, `register.html`, `login-2fa.html`, `2fa-setup.html`, `passkeys.html`, `verify-success.html`, `confirmation.html`
- [x] rewrite `login.html` and `register.html` to match `auth.jsx`:
  - Centered card on paper background, `box-shadow: 4px 4px 0 0 var(--line-1)`
  - Mono labels (`· EMAIL`, `· PASSWORD`), input with `border: 1px solid var(--line-1)` and offset focus
  - Primary "Sign in" / "Begin" button full-width
  - Footer link to alternate (Sign in ↔ Begin)
  - Add small marketing-bar at top with just the Logo so users can navigate back (the unauthenticated layout's existing `marketing-bar` already supplies this — the auth pages render inside it without changes)
- [x] rewrite `login-2fa.html` and `2fa-setup.html`: same card pattern, mono input for code (centered, large letter-spacing), soft instructional copy
- [x] rewrite `passkeys.html`: list of registered passkeys with `· REGISTERED` mono badges + "Add a passkey" primary button; envelope-style empty state (dashed border, serif "No passkeys yet" headline)
- [x] rewrite `verify-success.html` and `confirmation.html`: large serif heading, mono kicker, single CTA — keep them brief, consistent with the calm tone
- [x] preserve all form field `name=""` attributes, action URLs, CSRF tokens, hidden inputs — backend handlers keep working unchanged (login keeps `name="email"`/`"password"`/`"remember"` action `/login`; register keeps `name="confirmPassword"` action `/register`; login-2fa keeps hidden `name="email"`/`"password"`/`"remember"`/`"totp_code"` action `/login`; 2fa-setup keeps `name="code"` action `/2fa/verify`; passkeys keeps DELETE form action `/profile/passkeys/{id}` with `name="_method" value="DELETE"`)
- [x] write Go tests for each redesigned auth template: render with sample data, assert new structural markers and that form `action`/`method`/`name` attributes are intact (added `internal/web/auth_test.go` with `TestRenderLogin_Heartbeat`, `TestRenderRegister_Heartbeat`, `TestRenderLogin2FA_Heartbeat`, `TestRenderLogin2FA_RendersErrorAlert`, `TestRender2FASetup_Heartbeat`, `TestRenderPasskeys_WithEntries`, `TestRenderVerifySuccess_Heartbeat`, `TestRenderConfirmation_Heartbeat`)
- [x] write test for empty-passkeys-state on `passkeys.html` (`TestRenderPasskeys_EmptyState` asserts `data-test="passkey-empty"` appears and `passkey-list` does not)
- [x] run `go test ./...` — must pass before next task (all green; lint baseline unchanged from Task 3 — 51 pre-existing gosec/gofmt warnings in untouched handler files, none introduced by Task 4)

### Task 5: Dashboard — `dashboard.html`
- [x] rewrite `dashboard.html` to mirror `dashboard.jsx`:
  - `PageHead` with kicker `TODAY · <weekday, month day>`, soft-status title (`You're good, {{ .Data.User.FirstName }}.` / `Just checking in.` / `Are you there?`), lede explaining next nudge / silent stretch
  - Status timeline strip card: three columns LAST SEEN / NEXT NUDGE / DELIVERY (IF SILENT) with values + sub-date
  - Big check-in card with primary "I'm here" button (preserve existing `/api/check-in` POST flow and `checkInButton` JS hook from current `dashboard.html`)
  - Section "How we know you're alive" — six channel cards in a 3-column grid (Signing in, GitHub activity, Telegram reply, Email link, Passkey tap, Personal URL)
  - Two-column footer: Recent activity (left, last 4 items) + Your circle (right, up to 4 recipients with confirmed/awaiting badges)
  - Modal "How Heartbeat works" triggered by ghost button in PageHead actions
- [x] map existing `.Data` fields to the new layout — handler enriched additively (no removals) to expose `StatusVariant`, `Timeline.{LastSeen,NextNudge,Delivery}`, `Circle`, `User.{FirstName,GitHubConnected,TelegramConnected,HasPasskey}`; existing `Status`, `Activities`, `Stats`, `LatestPing`, `PingFrequency`, `PingDeadline`, etc. all preserved so backward-compat is intact:
  - `Status` → `StatusVariant` (`ok`/`warn`/`crit`) drives both soft-copy block and `.heartbeat ok|warn|crit` pulse class on the status card
  - `LastActivity`, `NextCheckIn`, `Deadline` → fed into `humanizeAgo`/`humanizeUntil` helpers and surfaced in the three timeline-strip cells (value + sub-date)
  - Stats counters left in `.Data.Stats` for any future panels — section header meta for "Your circle" instead links straight to `/recipients` to match the design
  - `Activities` → first 4 rendered in the bottom-left list card (existing `[]map[string]string` shape kept)
- [x] preserve check-in JS in `{{ define "scripts" }}` but update DOM selectors and success/error toast copy ("Got it. We see you.") — new toast element `#hbToast` replaces the old in-page alert injection; check-in still POSTs `/api/check-in` and reloads on success
- [x] write Go test: render dashboard with mocked status `active`/`caution`/`danger`, assert correct soft-copy variant + correct heartbeat pulse class (`TestRenderDashboard_OkVariant` / `…WarnVariant` / `…CritVariant` in `internal/web/dashboard_test.go`)
- [x] write Go test: render dashboard with empty recipients, assert empty-state link in "Your circle" (`TestRenderDashboard_EmptyCircle` plus `TestRenderDashboard_PopulatedCircle` and `TestRenderDashboard_SingularLetter` for the populated path and pluralization)
- [x] write Go test: assert old alarm copy ("System Active", "Critical Action Required", `class="status-indicator"`) absent (`TestRenderDashboard_RemovesOldAlarmCopy` covers those plus `Welcome back!`, `Check In Now`, `Total Secrets` and the old `check-in-box` class)
- [x] run `go test ./...` — must pass before next task (all green; lint baseline unchanged from Task 4 — 51 pre-existing gosec/gofmt warnings, none introduced by Task 5)

### Task 6: People — `recipients.html`, `new-recipient.html`, `manage-recipient-secrets.html`
- [ ] rewrite `recipients.html` to mirror the envelope grid in `recipients.jsx`:
  - `PageHead` (kicker `THE PEOPLE · YOUR CIRCLE`, title "People who matter", lede about envelopes, primary "Add someone" action)
  - `auto-fill, minmax(320px, 1fr)` envelope card grid
  - Each envelope: dashed-bottom flap (`TO ·` + confirmed/awaiting badge), name + relation + email body, sunken footer (`CONTAINS · {n} letters` or "nothing yet")
  - Trailing dashed `+ Add another person` placeholder card
  - Empty state: dashed bordered block with `users-round` icon, "Nobody yet" headline, descriptive paragraph, primary button
- [ ] rewrite `new-recipient.html`: stepped form (step 1 name+email+relation, step 2 confirmation copy showing the email Heartbeat will send) — matches `AddRecipientModal` structure but as a full page since this is a separate route
- [ ] rewrite `manage-recipient-secrets.html`: list letters/secrets currently assigned to a recipient with checkboxes; mono `· CONTAINS` section head; "save" primary button
- [ ] preserve existing form fields and action URLs; restyle only
- [ ] write Go tests: render `recipients.html` with 0 recipients (empty state) and 3 recipients (envelope grid), assert markers
- [ ] write Go test: render `new-recipient.html`, assert step labels and form fields
- [ ] write Go test: render `manage-recipient-secrets.html` with assigned + unassigned letters, assert checkbox states
- [ ] run `go test ./...` — must pass before next task

### Task 7: Letters — `secrets.html`, `new-secret.html`, `view-secret.html`, `manage-secret-recipients.html`
- [ ] rewrite `secrets.html` ("Letters") to mirror `letters.jsx`:
  - `PageHead` kicker `· YOUR LETTERS`, title "What you've written", primary "New letter" action
  - List view with kind badges (`✉ LETTER` / `☐ INSTRUCTIONS` / `⚿ KEYS`), title, last-updated relative time, recipient envelopes preview
  - Empty state: serif "Nothing written yet." + primary button
- [ ] rewrite `new-secret.html` (the editor):
  - Kind selector at top (three radio cards: Letter / Instructions / Secrets & Keys) — visually highlight selected
  - Title input (serif italic placeholder for letters; mono for secrets/instructions)
  - Body textarea (serif body for letters; mono for secrets — switch class based on selected kind)
  - Recipient assignment: multi-select envelope row at bottom with confirmed badges
  - Right rail or footer note: `🔒 ENCRYPTED AT REST · TIME-BOXED · ONLY UNSEALED ON DELIVERY`
- [ ] rewrite `view-secret.html`: full-letter rendered in serif (or mono if kind=secret), with metadata strip (created, last edited, recipients, encryption status) above
- [ ] rewrite `manage-secret-recipients.html`: mirror `manage-recipient-secrets.html` shape but inverted (which envelopes contain this letter)
- [ ] preserve form action URLs, hidden ID fields, CSRF; restyle only
- [ ] write Go tests: render each, assert kind selector states, assert encrypted-at-rest banner appears, assert serif-vs-mono class switches by kind
- [ ] run `go test ./...` — must pass before next task

### Task 8: Settings + Profile — `settings.html`, `profile.html`
- [ ] rewrite `settings.html`:
  - `PageHead` kicker `· ACCOUNT · CADENCE`, title "How often we check in"
  - Form with mono labels for `cadence_days` (slider + numeric input) and `grace_days` (slider + numeric input)
  - Channel toggles: Email, Telegram, Web — show "always on" mono badge for things that don't need user action
  - Visual cadence preview: mono-pre showing `today ──30d──● ──60d──● delivered` style timeline based on current values (matches the design's modal pre block)
- [ ] rewrite `profile.html`:
  - Sections: Identity (name, email), Connected channels (GitHub, Telegram), Authentication (password change, passkeys link, 2FA link)
  - Mono `· CONNECTED` badges or `· NOT YET` dashed badges per channel
  - Remove the SMS/phone field if still present (per chat round 2)
- [ ] preserve form actions, field names, validation messages
- [ ] write Go test: render `settings.html` with cadence=30/grace=60, assert mono-pre timeline contains `30d` and `60d`
- [ ] write Go test: render `profile.html` with GitHub disconnected → `· NOT YET` badge; with GitHub connected → `· CONNECTED` + username
- [ ] write Go test: assert no SMS/phone field appears in `profile.html`
- [ ] run `go test ./...` — must pass before next task

### Task 9: History — `history.html`
- [ ] rewrite `history.html`:
  - `PageHead` kicker `· AUDIT TRAIL`, title "Everything we've recorded", lede explaining append-only nature
  - Single-column timeline with mono dates on left, soft activity description on right, dashed dividers between days
  - Event type badges (LOGIN / GITHUB / CHECKIN / NUDGE / LETTER / RECIPIENT / SETTINGS) with the design's icon set
  - Pagination preserved (existing pattern)
- [ ] write Go test: render with 5 sample events of mixed types, assert each badge label appears, dates render in mono with tabular-nums
- [ ] write Go test: assert pagination controls intact
- [ ] run `go test ./...` — must pass before next task

### Task 10: Verify acceptance criteria
- [ ] verify all 19 templates have been redesigned and reference `heartbeat.css` (no template still pulls `main.css`)
- [ ] delete `web/static/css/main.css` and `web/static/css/normalize.css` once confirmed no template references them; if any HTMX/JS still references them, leave with a deprecation comment and surface a `⚠️` blocker
- [ ] verify all redesigned templates render with sample data (every existing handler test passes; new ones added per task pass)
- [ ] verify edge cases handled: empty recipients, empty letters, unauthenticated layout, all three dashboard status variants
- [ ] run full test suite: `go test ./...`
- [ ] run linter: `./scripts/run-golangci-lint.sh` — fix any new issues
- [ ] verify test coverage hasn't regressed: `go test -coverprofile=coverage.out ./internal/web/... && go tool cover -func=coverage.out` — target ≥80%
- [ ] grep for residual "Dead Man's Switch" in any template file: `grep -rn "Dead Man" web/templates/` should return empty (only acceptable in `<!-- comment -->` form referencing the project name, not in user-facing copy)
- [ ] grep for any leftover phone/SMS references in templates per chat round 2
- [ ] clean up `.design-fetch/` working directory — *do not commit it*; add to `.gitignore` if not already covered

*Note: ralphex automatically moves completed plans to `docs/plans/completed/`*

## Technical Details

**Design tokens (oklch paper):**
```css
--bg-0: oklch(0.985 0.008 80);
--bg-1: oklch(0.965 0.012 80);
--bg-elev: oklch(0.995 0.006 80);
--fg-0: oklch(0.22 0.02 60);
--line-1: oklch(0.55 0.04 60);
--line-2: oklch(0.85 0.02 70);
--orange: oklch(0.62 0.16 45);
```

**Type stack:**
- Mono: `'JetBrains Mono', ui-monospace, SFMono-Regular, monospace` (self-hosted woff2)
- Serif (display only): `'Source Serif 4', Georgia, serif` (system fallback if not self-hosted)
- Sans (body in non-mono regions): `'Inter', system-ui, sans-serif` (system fallback OK)

**Signature pattern — offset-shadow card:**
```css
.card { background: var(--bg-elev); border: 1px solid var(--line-1); box-shadow: 2px 2px 0 0 var(--line-1); padding: 24px; }
.btn:hover { transform: translate(-1px, -1px); box-shadow: 2px 2px 0 0 var(--line-1); }
.btn:active { transform: translate(0, 0); box-shadow: none; }
```

**Heartbeat pulse (used in topbar status indicator):**
```css
.heartbeat .pulse { width: 8px; height: 8px; background: var(--fg-0); border-radius: 50%; animation: pulse 2s ease-in-out infinite; }
.heartbeat.warn .pulse { animation-duration: 0.8s; }
.heartbeat.crit .pulse { animation-duration: 0.4s; }
@keyframes pulse { 0%, 100% { opacity: 1; transform: scale(1); } 50% { opacity: 0.3; transform: scale(0.7); } }
```

**Soft status copy mapping (dashboard):**
| existing `Status` | new title                   | new lede                                                                    | pulse class |
|-------------------|-----------------------------|-----------------------------------------------------------------------------|-------------|
| `active`          | `You're good, {first}.`     | `Next gentle nudge in N days. Nothing else to do today.`                    | `ok`        |
| `caution`         | `Just checking in.`         | `It's been a quiet stretch. Tap below to let us know you're around.`        | `warn`      |
| `danger`          | `Are you there?`            | `If we don't hear from you soon, your letters will start to go out.`        | `crit`      |

**Naming map (UI strings only):**
| backend term      | UI string                |
|-------------------|--------------------------|
| Dead Man's Switch | Heartbeat                |
| recipient         | person / envelope        |
| recipients        | your circle / people     |
| secret            | letter / instructions / keys |
| ping              | nudge                    |
| trigger / dead    | letters delivered (silent)   |
| check-in          | I'm here / heartbeat     |

## Post-Completion

*Items requiring manual intervention or external systems — no checkboxes, informational only.*

**Manual visual review** (cannot be automated):
- Open every redesigned page in a browser at desktop and mobile widths and confirm pixel-fidelity to the design's intent (the agent's render tests assert structure but cannot evaluate visual quality)
- Verify font loading: woff2 served from same origin, no FOUT/FOIT regressions, fallback fonts look acceptable if Source Serif 4 / Inter not self-hosted
- Verify heartbeat pulse animation runs smoothly and respects `prefers-reduced-motion`
- Test the check-in flow end-to-end: dashboard `I'm here` → toast → reload → status reset
- Test envelope hover/translate animation on `recipients.html`
- Verify accessibility basics: focus rings visible (offset shadow doesn't replace them on inputs), color contrast on `oklch(0.42)` secondary text passes WCAG AA, dotgrid background is decorative-only

**Documentation updates** (out of scope for this plan, capture if discovered):
- Update repo README screenshots once redesign is shipped
- Update `docs/installation.md` if any new static-asset paths affect Docker volumes
- Consider linking the design bundle's chat transcript or a screenshot in `docs/` for future contributors who want to understand the design language

**External / deployment:**
- Re-enable Playwright e2e tests in CI as a follow-up (commit `2d7600d` disabled them); update Playwright selectors that targeted old DOM (e.g. `.feature-card`, `.status-indicator.active` no longer exist)
- Verify the GHCR Docker image still builds and the Portainer auto-deploy pipeline picks up the static asset additions
