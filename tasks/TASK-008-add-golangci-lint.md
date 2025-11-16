# TASK-008: Add golangci-lint and Fix Warnings

## Priority: MEDIUM 🟡

## Status: Not Started

## Category: Code Quality / CI/CD

## Description

Add golangci-lint to the project to enforce code quality standards, catch potential bugs, and maintain consistent code style. Integrate it into GitHub Actions and optionally create a badge for the README.

## Problem

**Current State:**
From `/todo.md:59`:
> Add golangci-lint and fix all the warnings, add it to github actions and if possible create badge for it

**Why This is Important:**
1. **Code Quality** - Catches common mistakes and code smells
2. **Security** - Identifies potential security issues
3. **Consistency** - Enforces consistent code style
4. **Maintainability** - Makes code easier to read and maintain
5. **Best Practices** - Ensures Go best practices are followed

## Proposed Solution

### 1. Create golangci-lint Configuration

**New file:** `.golangci.yml`

```yaml
# golangci-lint configuration
run:
  timeout: 5m
  tests: true
  modules-download-mode: readonly

linters:
  enable:
    # Enabled by default
    - errcheck      # Check for unchecked errors
    - gosimple      # Simplify code
    - govet         # Vet examines Go source code
    - ineffassign   # Detect ineffectual assignments
    - staticcheck   # Static analysis
    - unused        # Find unused code

    # Additional recommended linters
    - bodyclose     # Check HTTP response body is closed
    - dogsled       # Find assignments with too many blank identifiers
    - dupl          # Find code duplication
    - exhaustive    # Check exhaustiveness of enum switch statements
    - exportloopref # Find pointers to loop variables
    - gochecknoinits # Check there are no init functions
    - goconst       # Find repeated strings that could be constants
    - gocritic      # Comprehensive Go linter
    - gocyclo       # Compute cyclomatic complexity
    - godot         # Check comments end in period
    - gofmt         # Check code is gofmt-ed
    - goimports     # Check imports are formatted
    - gosec         # Security checker
    - misspell      # Find misspelled words
    - nakedret      # Find naked returns in large functions
    - nolintlint    # Check nolint directives
    - prealloc      # Find slice declarations that could be preallocated
    - revive        # Fast, configurable, extensible, flexible, and beautiful linter
    - rowserrcheck  # Check sql.Rows.Err is checked
    - sqlclosecheck # Check sql.Rows/sql.Stmt are closed
    - stylecheck    # Replacement for golint
    - typecheck     # Type-checks Go code
    - unconvert     # Remove unnecessary type conversions
    - unparam       # Find unused function parameters
    - whitespace    # Find unnecessary whitespace

  disable:
    - gochecknoglobals # We use some globals for config
    - gocognit         # Similar to gocyclo, less strict
    - godox            # Don't fail on TODO comments
    - goerr113         # Too strict on error wrapping
    - gomnd            # Magic number detector (too noisy)
    - lll              # Line length (handled by gofmt)
    - nlreturn         # Too strict on blank lines
    - wsl              # Too opinionated on whitespace

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true

  govet:
    check-shadowing: true
    enable-all: true

  gocyclo:
    min-complexity: 15

  dupl:
    threshold: 100

  goconst:
    min-len: 3
    min-occurrences: 3

  misspell:
    locale: US

  gosec:
    excludes:
      - G404 # Use of weak random number generator - we use crypto/rand where needed

  revive:
    rules:
      - name: exported
        disabled: true # Don't require doc comments on all exports
      - name: package-comments
        disabled: true

  stylecheck:
    checks: ["all", "-ST1000", "-ST1003"]

issues:
  exclude-rules:
    # Exclude some linters from running on tests
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec

    # Exclude specific gosec rules for test files
    - path: _test\.go
      text: "G404:"  # Weak random in tests is OK

    # Ignore long lines in templates
    - path: web/templates/
      linters:
        - lll

  max-issues-per-linter: 0
  max-same-issues: 0

output:
  format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
```

### 2. Add to GitHub Actions

**File to modify:** `.github/workflows/ci.yml`

Add a new job for linting:

```yaml
jobs:
  # ... existing jobs ...

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24.1'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m
          # Optional: show only new issues
          only-new-issues: false

  # Make lint a required check
  ci-success:
    name: CI Success
    needs: [test, lint, build]  # Add lint to required checks
    runs-on: ubuntu-latest
    steps:
      - name: Success
        run: echo "All CI checks passed!"
```

### 3. Add Badge to README

**File to modify:** `README.md`

Add badge near the top:

```markdown
# Dead Man's Switch

[![Go Report Card](https://goreportcard.com/badge/github.com/YOUR_USERNAME/deadmanswitch)](https://goreportcard.com/report/github.com/YOUR_USERNAME/deadmanswitch)
[![golangci-lint](https://github.com/YOUR_USERNAME/deadmanswitch/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_USERNAME/deadmanswitch/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/YOUR_USERNAME/deadmanswitch/branch/main/graph/badge.svg)](https://codecov.io/gh/YOUR_USERNAME/deadmanswitch)
```

### 4. Add Makefile Target

**File to modify:** `Makefile` (or create if doesn't exist)

```makefile
.PHONY: lint
lint:
	golangci-lint run --timeout=5m

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix --timeout=5m

.PHONY: lint-install
lint-install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 5. Fix Common Issues

Based on typical Go codebases, here are likely issues to fix:

#### A. Unchecked Errors
```go
// Before:
user, _ := repo.GetUserByID(ctx, id)

// After:
user, err := repo.GetUserByID(ctx, id)
if err != nil {
    return fmt.Errorf("failed to get user: %w", err)
}
```

#### B. Ineffectual Assignments
```go
// Before:
result := someValue
result = otherValue  // First assignment is ineffectual

// After:
result := otherValue
```

#### C. HTTP Body Not Closed
```go
// Before:
resp, err := http.Get(url)

// After:
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
```

#### D. Context Not Passed
```go
// Before:
func (r *Repository) GetUser(id string) (*User, error) {
    rows, err := r.db.Query("SELECT * FROM users WHERE id = ?", id)

// After:
func (r *Repository) GetUser(ctx context.Context, id string) (*User, error) {
    rows, err := r.db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", id)
```

#### E. Error Strings
```go
// Before:
return errors.New("Failed to connect")  // Should not be capitalized

// After:
return errors.New("failed to connect")
```

### 6. Create Pre-commit Hook (Optional)

**New file:** `.githooks/pre-commit`

```bash
#!/bin/bash

# Run golangci-lint before commit
echo "Running golangci-lint..."
golangci-lint run --timeout=2m

if [ $? -ne 0 ]; then
    echo "❌ golangci-lint failed. Commit aborted."
    echo "Run 'make lint-fix' to auto-fix some issues."
    exit 1
fi

echo "✅ golangci-lint passed"
```

Install hook:
```bash
git config core.hooksPath .githooks
chmod +x .githooks/pre-commit
```

## Implementation Plan

### Phase 1: Setup (1 hour)
1. Create `.golangci.yml` configuration
2. Test locally: `golangci-lint run`
3. Review all warnings
4. Categorize: auto-fixable vs manual fixes

### Phase 2: Fix Issues (3-5 hours)
1. Run `golangci-lint run --fix` for auto-fixes
2. Manually fix remaining issues
3. Add `//nolint` directives for acceptable exceptions (with comments)
4. Test that all changes don't break functionality

### Phase 3: CI Integration (1 hour)
1. Add GitHub Actions workflow
2. Test workflow in PR
3. Make lint a required check

### Phase 4: Documentation (30 min)
1. Add badge to README
2. Document linting in CONTRIBUTING.md
3. Add Makefile targets

## Acceptance Criteria

- [ ] `.golangci.yml` configuration created
- [ ] Local linting works: `golangci-lint run` passes
- [ ] GitHub Actions workflow added
- [ ] Workflow runs on PRs and pushes
- [ ] Badge added to README
- [ ] Makefile targets added
- [ ] All linting issues fixed or justified
- [ ] Pre-commit hook created (optional)
- [ ] Documentation updated
- [ ] No build breakages from fixes

## Testing Requirements

1. **Verify Linting:**
   - Run `golangci-lint run` locally
   - Check all enabled linters pass
   - Verify configuration works

2. **Verify CI:**
   - Push changes and verify GitHub Action runs
   - Verify it fails on lint errors
   - Verify it passes on clean code

3. **Verify Fixes:**
   - Run full test suite after fixes
   - Verify no functionality broken
   - Check build still succeeds

## Common Issues and Fixes

### Issue: Too Many Warnings
**Solution:** Start with fewer linters, gradually add more

### Issue: False Positives
**Solution:** Use `//nolint:rulename // reason` with justification

### Issue: CI Timeout
**Solution:** Increase timeout or enable caching

### Issue: Different Results Locally vs CI
**Solution:** Pin golangci-lint version in both places

## Configuration Tuning

Start conservative, then tighten:

**Week 1:** Enable only critical linters
```yaml
enable:
  - errcheck
  - govet
  - staticcheck
```

**Week 2:** Add security and quality linters
```yaml
enable:
  - gosec
  - ineffassign
  - unused
```

**Week 3:** Add style linters
```yaml
enable:
  - gofmt
  - goimports
  - stylecheck
```

## Files to Create/Modify

**New Files:**
1. `.golangci.yml` - Linter configuration
2. `.githooks/pre-commit` - Pre-commit hook (optional)

**Files to Modify:**
1. `.github/workflows/ci.yml` - Add lint job
2. `README.md` - Add badge
3. `Makefile` - Add lint targets
4. `CONTRIBUTING.md` - Document linting requirements
5. Various Go files - Fix linting issues

## Benefits

1. **Catch Bugs Early** - Many issues found before runtime
2. **Code Quality** - Enforces consistent, high-quality code
3. **Security** - gosec catches security issues
4. **Review Efficiency** - Automated checks reduce manual review burden
5. **Onboarding** - New contributors get immediate feedback

## References

- TODO item: `/todo.md:59`
- golangci-lint docs: https://golangci-lint.run/
- GitHub Action: https://github.com/golangci/golangci-lint-action

## Estimated Effort

- **Complexity:** Medium
- **Time:** 5-7 hours (including fixing all warnings)
- **Risk:** Low-Medium (fixes might introduce bugs if not tested)

## Dependencies

None - can be done independently

## Follow-up Tasks

- Add security-specific linters (from TASK-009)
- Set up automatic dependency updates (Dependabot)
- Add Go Report Card monitoring
- Consider SonarQube integration
