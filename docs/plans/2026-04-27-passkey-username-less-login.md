# Passkey Username-less Login (Discoverable Credentials + Conditional UI)

## Overview
Today, signing in with a passkey requires the user to type their email first
(`web/templates/login.html` → `/login/passkey/begin` looks up the user, then
issues a challenge with `allowCredentials` populated from that user's stored
passkeys). This is more friction than necessary for the common case where the
authenticator already has a discoverable credential for the site.

This change adds a true username-less flow with two entry points:

1. A **"Sign in with passkey"** button that triggers
   `navigator.credentials.get({ publicKey })` with **no `allowCredentials`** —
   the browser shows a native picker of every passkey scoped to our RPID and
   the server identifies the user from the credential's `userHandle`.
2. **Conditional UI / autofill** on the email field — passkeys appear in the
   browser's autofill dropdown automatically, so signing in is one tap.

The existing email-first flow stays as a fallback path for users whose old
passkeys were registered without the resident-key bit (we cannot tell from the
server side whether a stored credential is discoverable).

## Context (from discovery)

Files/components involved:

- `internal/auth/webauthn.go` — `BeginLogin` (line 254) and `FinishLogin` (line
  354) both take `*models.User` and pre-populate `AllowedCredentials` from the
  user's stored passkeys. Need new variants that do neither.
- `internal/auth/webauthn.go:75` — `BeginRegistration` uses default options;
  add `webauthn.WithAuthenticatorSelection({ResidentKey: "preferred"})` so new
  enrollments produce discoverable credentials when the authenticator supports
  it.
- `internal/web/handlers/passkey.go:286` — `HandleBeginLogin` and
  `HandleFinishLogin` require an `email` field in the request body. Need new
  handlers that don't.
- `internal/web/server.go:155-156` — current routes `/login/passkey/begin` and
  `/login/passkey/finish`. Add `/login/passkey/discover/begin` and
  `/login/passkey/discover/finish`.
- `web/templates/login.html:105-114` — passkey button currently hard-fails if
  the email field is empty (`web/templates/login.html:160`). Remove that
  requirement; add `autocomplete="username webauthn"` to the email input;
  trigger conditional mediation on page load.
- `internal/storage/passkey.go` — `GetPasskeyByCredentialID` already exists
  (used at `webauthn.go:439`). No new repository methods needed for lookup; we
  only need a quick wrapper to fetch the user from the resolved passkey.
- `internal/models/passkey.go` — `WebAuthnID()` returns `[]byte(u.ID)`; this
  is what comes back as `userHandle` from the authenticator and the value we
  use to find the user.
- `tests/frontend/login.spec.js` — existing Playwright test covers
  password-based login. Real passkey UX cannot be exercised in Playwright
  without WebAuthn virtual authenticator; we'll add a test that verifies the
  username-less button is present, calls the begin endpoint, and that the
  endpoint returns a challenge with no `allowCredentials`.

Related patterns found:

- The codebase uses an in-memory `sessions` map (`webauthn.go:33`) keyed by a
  cookie value `webauthn_session_id`. The discoverable variant must follow the
  same convention so cookies / TTLs / cleanup behave identically.
- JSON request handling in handlers uses a manual decode → restore body
  pattern (e.g. `passkey.go:137-160`); new handlers should match.

Dependencies identified:

- `github.com/go-webauthn/webauthn` v0.x — already vendored. The library
  exposes `BeginDiscoverableLogin(opts...)` and
  `FinishDiscoverableLogin(handler DiscoverableUserHandler, ...)` which we
  will use rather than reinventing the manual challenge construction the code
  currently does in `BeginLogin`.

## Development Approach
- **Testing approach**: Regular (code first, then tests) — chosen because
  WebAuthn behavior is mostly exercised end-to-end through the browser; unit
  tests cover the server-side state transitions.
- Complete each task fully before moving to the next.
- Make small, focused changes.
- **CRITICAL: every task MUST include new/updated tests** for code changes in
  that task — write unit tests for new functions, error paths, and new test
  cases for new code paths.
- **CRITICAL: all tests must pass before starting next task** — no exceptions.
- **CRITICAL: update this plan file when scope changes during implementation.**
- Run tests after each change.
- Maintain backward compatibility: the old email-first endpoints
  (`/login/passkey/begin`, `/login/passkey/finish`) and password-based login
  must keep working.

## Testing Strategy
- **Unit tests**: required for every task. Add to
  `internal/auth/webauthn_test.go` and a new
  `internal/web/handlers/passkey_test.go` (if absent) or extend existing
  `internal/web/auth_test.go`.
- **E2E tests**: extend `tests/frontend/login.spec.js`. We cannot fully drive
  a real authenticator from Playwright in CI, so the e2e test will:
  - assert the "Sign in with passkey" button exists and is enabled with an
    empty email field
  - assert the email input has `autocomplete="username webauthn"`
  - assert that `POST /login/passkey/discover/begin` returns 200 with a JSON
    body that contains a `challenge` and no `allowCredentials` (or an empty
    array)

## Progress Tracking
- Mark completed items with `[x]` immediately when done.
- Add newly discovered tasks with ➕ prefix.
- Document issues/blockers with ⚠️ prefix.
- Update plan if implementation deviates from original scope.

## What Goes Where
- **Implementation Steps** (`[ ]` checkboxes): code, tests, docs achievable in
  this repo.
- **Post-Completion** (no checkboxes): manual smoke test on a real device with
  a real authenticator.

## Implementation Steps

### Task 1: Make new passkey registrations discoverable (resident-key preferred)
- [x] in `internal/auth/webauthn.go:75`, change the `BeginRegistration` call
      to pass `webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
        ResidentKey: protocol.ResidentKeyRequirementPreferred,
        UserVerification: protocol.VerificationPreferred,
      })`
- [x] verify the existing `excludeCredentials` slice continues to be honored
      (now threaded through `webauthn.WithExclusions(...)`)
- [x] write a unit test in `internal/auth/webauthn_test.go` asserting that
      `BeginRegistration` returns options with
      `AuthenticatorSelection.ResidentKey == "preferred"`
- [x] write a unit test for `BeginRegistration` error path (e.g. nil user)
- [x] run `go test ./...` — must pass before next task

### Task 2: Add discoverable BeginLogin to WebAuthnService
- [x] add method `BeginDiscoverableLogin(ctx context.Context, w http.ResponseWriter) (*protocol.CredentialAssertion, error)`
      to `internal/auth/webauthn.go`
- [x] inside, call `s.webAuthn.BeginDiscoverableLogin()` (library helper) to
      get options + session data; do **not** populate `AllowedCredentials`
- [x] reuse the existing in-memory session storage pattern: generate session
      ID, store `*webauthn.SessionData`, set the `webauthn_session_id` cookie
      with the same TTL/flags as `BeginLogin` uses today
- [x] write a unit test asserting the returned options have an empty
      `allowCredentials`, a non-empty `challenge`, and the correct RPID
- [x] write a unit test asserting the session cookie is set on the response
      writer
- [x] run `go test ./...` — must pass before next task

### Task 3: Add discoverable FinishLogin to WebAuthnService
- [ ] add method `FinishDiscoverableLogin(ctx context.Context, r *http.Request) (*models.User, *models.Passkey, error)`
      to `internal/auth/webauthn.go`
- [ ] read the `webauthn_session_id` cookie and pop the session from the map
      (same pattern as `FinishLogin`)
- [ ] parse the credential JSON body the same way `FinishLogin` does (the
      client wraps it in `{ "credential": ... }` — keep the wrapper for
      consistency, no `email` field this time)
- [ ] call `s.webAuthn.FinishDiscoverableLogin(handler, *sessionData, parsedRequest)`
      where `handler` is a `webauthn.DiscoverableUserHandler` closure that:
        - decodes the supplied `userHandle` (`[]byte` → user ID string)
        - fetches the user via `s.repo.GetUserByID(ctx, ...)`
        - loads the user's passkeys via `getUserCredentials(ctx, user)` and
          attaches them so the library can verify the signature
        - returns the `webauthn.User`
- [ ] after success, look up the matching `*models.Passkey` via
      `GetPasskeyByCredentialID`, update `LastUsedAt` and `SignCount` (mirror
      `FinishLogin:459-465`)
- [ ] return both the resolved user and passkey
- [ ] write a unit test for the happy path with a fixture passkey + session
      (use the same test helpers as `FinishLogin` tests)
- [ ] write unit tests for: missing cookie, missing session, unknown
      `userHandle`, signature failure
- [ ] run `go test ./...` — must pass before next task

### Task 4: Add HTTP handlers and routes for discoverable login
- [ ] in `internal/web/handlers/passkey.go`, add
      `HandleBeginDiscoverableLogin(w, r)` that calls
      `webAuthnService.BeginDiscoverableLogin(r.Context(), w)`, marshals the
      options to JSON, returns them
- [ ] add `HandleFinishDiscoverableLogin(w, r)` that calls
      `webAuthnService.FinishDiscoverableLogin(r.Context(), r)`, then mirrors
      the session-creation block in `HandleFinishLogin:454-507` (create
      `models.Session`, set `session_token` cookie, write audit log,
      `{"success": true, "redirect": "/dashboard"}`)
- [ ] in `internal/web/server.go` near lines 155-156, register:
        - `r.HandleFunc("/login/passkey/discover/begin", s.handlers.passkey.HandleBeginDiscoverableLogin)`
        - `r.HandleFunc("/login/passkey/discover/finish", s.handlers.passkey.HandleFinishDiscoverableLogin)`
- [ ] write a handler unit test for `HandleBeginDiscoverableLogin` (uses
      `httptest.NewRecorder`, asserts 200 + JSON body shape — empty
      `allowCredentials`)
- [ ] write a handler unit test for `HandleFinishDiscoverableLogin` happy
      path using a stub `WebAuthnService` (or interface refactor if needed —
      flag as ⚠️ if it requires more than a trivial extraction)
- [ ] write a handler unit test for the failure path (invalid credential →
      4xx, no session cookie set)
- [ ] run `go test ./...` — must pass before next task

### Task 5: Update login template for username-less + conditional UI
- [ ] in `web/templates/login.html`, remove the standalone "EMAIL FOR PASSKEY"
      input block (lines 110-113) — the passkey button no longer needs it
- [ ] keep the passkey button (line 105-108) but rewire its handler to call
      `/login/passkey/discover/begin` → `navigator.credentials.get({ publicKey })`
      → `/login/passkey/discover/finish`; on failure, show the existing
      `auth-status err` div with the error text
- [ ] on the password form's email input (line 121), add
      `autocomplete="username webauthn"`
- [ ] in the `scripts` block, add an `IIFE` on `DOMContentLoaded` that:
        - feature-checks `PublicKeyCredential.isConditionalMediationAvailable?.()`
        - if available, fetches `/login/passkey/discover/begin`, decodes the
          challenge, calls `navigator.credentials.get({ publicKey, mediation: 'conditional' })`
        - on resolution, posts the credential to
          `/login/passkey/discover/finish` and redirects on success
        - silently ignores `NotAllowedError` / `AbortError` (user cancel)
- [ ] reuse the existing `base64URLToArrayBuffer` / `arrayBufferToBase64URL`
      helpers (lines 254-275) — extract them into a small shared scope so both
      the click handler and the conditional-UI IIFE can call them
- [ ] manually verify the template still renders by running
      `go run ./cmd/deadmanswitch` and loading `/login` (no JS errors in
      console; button disabled state cleared)
- [ ] write a Playwright test in `tests/frontend/login.spec.js`:
        - the "Sign in with passkey" button is visible without filling the
          email field
        - the password-form email input has
          `autocomplete="username webauthn"`
        - `POST /login/passkey/discover/begin` returns 200 with a JSON
          response that has `publicKey.challenge` and no
          `publicKey.allowCredentials` (or an empty array)
- [ ] run `go test ./...` and the Playwright suite — must pass before next
      task

### Task 6: Verify acceptance criteria
- [ ] verify all requirements from Overview are implemented (button
      username-less, autofill on email field, email-first fallback intact)
- [ ] verify edge cases: stale session cookie, unknown userHandle, cancelled
      browser prompt
- [ ] run full unit test suite (`go test ./...`)
- [ ] run e2e tests (`npx playwright test`)
- [ ] run linter (`golangci-lint run` if configured, otherwise `go vet ./...`)
      — all issues fixed
- [ ] verify test coverage for the two new methods is ≥ 80% via
      `go test -cover ./internal/auth/... ./internal/web/handlers/...`

### Task 7: [Final] Update documentation
- [ ] update `README.md` (and any auth-related docs) to mention username-less
      passkey login as the primary flow; note that the email-first flow stays
      as a fallback
- [ ] add a one-line note in `CLAUDE.md` (if it has a passkeys section)
      pointing future Claude sessions at the discoverable endpoints

## Technical Details

### Data structures and changes
- **No DB schema changes.** The existing `passkeys` table already stores
  `user_id` and `credential_id`; that's everything we need to resolve a
  credential back to a user.
- The `webauthn.SessionData` for discoverable flows has empty
  `AllowedCredentialIDs` and a synthesized `UserID` — the library handles
  this; we just store/restore the struct.

### Endpoints
- `POST /login/passkey/discover/begin` — body: `{}` (empty); response:
  serialized `protocol.CredentialAssertion` with `allowCredentials: []`.
- `POST /login/passkey/discover/finish` — body:
  `{ "credential": <PublicKeyCredential> }`; response:
  `{ "success": true, "redirect": "/dashboard" }` and a `session_token`
  cookie on success.

### Processing flow
1. Browser calls `…/discover/begin` → server stores SessionData + cookie,
   returns options.
2. Browser calls `navigator.credentials.get({ publicKey, mediation: ... })`.
3. User picks a passkey; authenticator returns `userHandle` (= our
   `User.WebAuthnID()` = `user.ID` bytes).
4. Browser POSTs the assertion to `…/discover/finish`.
5. Server pops SessionData, runs `FinishDiscoverableLogin` with a callback
   that resolves the user via `userHandle`, then verifies the signature
   against that user's stored credential.
6. Server creates the normal session cookie (same as today's password and
   email-first passkey flows) and returns the redirect URL.

### Backward compatibility
- Existing `/login/passkey/begin` and `/login/passkey/finish` endpoints stay
  in place and untouched. Old passkeys without resident-key flag continue to
  work via that path.
- New passkeys created after Task 1 will be discoverable on most modern
  authenticators (`ResidentKey: "preferred"`), but we make no DB-level
  guarantees — the username-less flow may simply not show a credential and
  the user falls back to the email-first form.

## Post-Completion
*Items requiring manual intervention or external systems — no checkboxes,
informational only.*

**Manual verification:**

- Smoke-test on at least three platforms with real authenticators:
  - macOS Safari + Touch ID (platform authenticator, definitely
    discoverable)
  - Chrome desktop + phone passkey via QR / hybrid transport
  - 1Password browser extension (resident-key support since 2023)
- Verify conditional UI shows passkeys in the email autofill dropdown on
  Safari and Chrome.
- Verify a user with only a non-discoverable hardware key still sees the
  email-first flow work without errors.

**External system updates:**

- None. No deployment/config changes; the new endpoints share existing
  infra (in-memory session map, repository, cookie scheme).
