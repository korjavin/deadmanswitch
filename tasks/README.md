# Dead Man's Switch - Task List

This directory contains detailed task specifications for addressing TODOs, FIXMEs, and production readiness issues identified in the codebase.

## 📋 Task Overview

Each task file is a self-contained specification that includes:
- **Problem description** - What needs to be fixed and why
- **Current state** - Code references and examples
- **Proposed solution** - Detailed implementation approach
- **Acceptance criteria** - Definition of done
- **Testing requirements** - How to verify the fix
- **Files to modify** - Complete list of affected files
- **Effort estimates** - Complexity and time estimates
- **Dependencies** - Related tasks and prerequisites

## 🔴 Critical Priority Tasks

These tasks address security vulnerabilities and must be completed before production use:

| Task | Title | Effort | Status |
|------|-------|--------|--------|
| [TASK-001](./TASK-001-implement-master-key-management.md) | Implement Master Key Management System | 4-6h | Not Started |
| [TASK-002](./TASK-002-implement-access-code-secure-storage.md) | Implement Secure Access Code Storage with TTL | 6-8h | Not Started |

**Why Critical:**
- TASK-001: Replaces hardcoded encryption keys (MAJOR security vulnerability)
- TASK-002: Ensures access codes are securely stored with expiration

**Must Complete Before:** Any production deployment

## 🟠 High Priority Tasks

These tasks improve security and core functionality:

| Task | Title | Effort | Status |
|------|-------|--------|--------|
| [TASK-004](./TASK-004-complete-secret-questions-implementation.md) | Complete Secret Questions Implementation | 8-10h | Partially Complete |

**Why High Priority:**
- Secret questions feature is partially implemented but critical re-encryption logic is missing

## 🟡 Medium Priority Tasks

These tasks enhance user experience and maintainability:

| Task | Title | Effort | Status |
|------|-------|--------|--------|
| [TASK-003](./TASK-003-enhance-ping-urgency-levels.md) | Enhance Ping Messages with Urgency Levels | 3-4h | Not Started |
| [TASK-008](./TASK-008-add-golangci-lint.md) | Add golangci-lint and Fix Warnings | 5-7h | Not Started |

**Why Medium Priority:**
- TASK-003: Improves UX by showing urgency in notifications (removes TODOs)
- TASK-008: Improves code quality and catches potential bugs

## 🟢 Low Priority Tasks

These tasks improve code organization and maintainability:

| Task | Title | Effort | Status |
|------|-------|--------|--------|
| [TASK-005](./TASK-005-consolidate-hardcoded-time-constants.md) | Consolidate Hardcoded Time Constants | 6-8h | Not Started |
| [TASK-006](./TASK-006-move-email-templates-to-folder.md) | Move Email Templates to Dedicated Folder | 4-6h | Not Started |
| [TASK-007](./TASK-007-remove-user-name-field.md) | Remove User Name Field | 3-4h | Not Started |

**Why Low Priority:**
- These are code quality improvements that don't affect functionality
- Can be done incrementally over time

## 📊 Task Statistics

- **Total Tasks:** 8
- **Critical:** 2 (25%)
- **High:** 1 (12.5%)
- **Medium:** 2 (25%)
- **Low:** 3 (37.5%)
- **Total Estimated Effort:** 39-53 hours

## 🎯 Recommended Implementation Order

### Phase 1: Security Fixes (Critical)
**Goal:** Make application production-ready from security perspective

1. **TASK-001** - Master Key Management (4-6h)
   - Removes hardcoded encryption keys
   - Adds proper key management
   - **Blocker for production use**

2. **TASK-002** - Access Code Storage (6-8h)
   - Securely stores access codes
   - Implements TTL and rate limiting
   - Depends on TASK-001

**Phase 1 Total:** 10-14 hours

### Phase 2: Core Functionality (High Priority)
**Goal:** Complete partially implemented features

3. **TASK-004** - Secret Questions (8-10h)
   - Completes timelock + Shamir implementation
   - Adds re-encryption scheduler task
   - Advanced feature for secret recovery

**Phase 2 Total:** 8-10 hours

### Phase 3: User Experience (Medium Priority)
**Goal:** Improve notifications and code quality

4. **TASK-003** - Ping Urgency Levels (3-4h)
   - Makes notifications more effective
   - Removes TODOs in scheduler

5. **TASK-008** - golangci-lint (5-7h)
   - Catches potential bugs
   - Enforces code quality
   - Sets up CI/CD quality gates

**Phase 3 Total:** 8-11 hours

### Phase 4: Code Quality (Low Priority)
**Goal:** Clean up code organization

6. **TASK-006** - Email Templates (4-6h)
   - Prerequisite for TASK-003 (urgency templates)
   - Makes emails easier to maintain

7. **TASK-005** - Time Constants (6-8h)
   - Makes durations configurable
   - Improves code maintainability

8. **TASK-007** - Remove Name Field (3-4h)
   - Reduces PII collection
   - Simplifies database schema

**Phase 4 Total:** 13-18 hours

## 📚 Additional TODOs from todo.md

The following items from `/todo.md` are not yet converted to task files but should be considered:

### Features to Implement
- [ ] Profile update functionality with Telegram disconnect (line 32)
- [ ] ActivityPub account monitoring (line 56)
- [ ] Telegram channel monitoring (line 57)
- [ ] Audit logging for Telegram connections (line 27)
- [ ] User check-ins shown in activity log (line 65)
- [ ] Passkey as second factor option (line 66)
- [ ] Login notifications via Telegram (line 34)

### Configuration & Settings
- [ ] Remove settings functionality - use env vars (line 28)
- [ ] Make ping sending via Telegram configurable (line 33)

### Testing
- [ ] Add integration tests (target 80% coverage) (line 64)
- [ ] Add security linters (line 60)

### Research Tasks
- [ ] Research passwordless passkey login (line 46)

## 🔗 Task Dependencies

```
TASK-001 (Master Key)
    ↓
    ├── TASK-002 (Access Codes) - Needs secure encryption
    └── TASK-004 (Secret Questions) - Needs proper key management

TASK-003 (Urgency Levels)
    ↓
    └── TASK-006 (Email Templates) - Should use template system

TASK-008 (golangci-lint) - Independent, can be done anytime
TASK-005 (Time Constants) - Independent, can be done anytime
TASK-007 (Remove Name) - Independent, can be done anytime
```

## 📝 Task File Format

Each task follows this structure:

```markdown
# TASK-XXX: Title

## Priority: [CRITICAL/HIGH/MEDIUM/LOW] 🔴🟠🟡🟢

## Status: [Not Started/In Progress/Completed/Blocked]

## Category: [Security/Features/Code Quality/etc.]

## Description
Brief overview of the task

## Problem
Current state and why it needs fixing

## Proposed Solution
Detailed implementation approach

## Acceptance Criteria
- [ ] Checklist of requirements

## Testing Requirements
Unit tests, integration tests, manual testing

## Files to Modify
Complete list with line numbers

## Estimated Effort
- Complexity: [Low/Medium/High]
- Time: X-Y hours
- Risk: [Low/Medium/High]

## Dependencies
Related tasks and prerequisites

## References
Links to code, docs, related issues
```

## 🚀 Getting Started

### For Project Maintainers
1. Review task priorities and adjust based on project goals
2. Create GitHub issues from task files
3. Assign tasks to milestones
4. Start with Phase 1 (critical security fixes)

### For Contributors
1. Choose a task matching your skills and available time
2. Read the full task specification
3. Check dependencies and prerequisites
4. Implement following the proposed solution
5. Verify all acceptance criteria
6. Submit PR with reference to task number

### Creating New Tasks
If you find additional TODOs or issues:

1. Create a new task file: `TASK-XXX-short-description.md`
2. Follow the task file format above
3. Add to this README in the appropriate priority section
4. Update statistics and dependencies

## 📖 Additional Documentation

- [Project Overview](../docs/PROJECT_OVERVIEW.md) - Architecture and design
- [Security Documentation](../docs/security.md) - Security model and threats
- [TODO List](../todo.md) - Original TODO items
- [Contributing Guide](../CONTRIBUTING.md) - Development guidelines

## 🔍 Finding TODOs in Code

To find TODO/FIXME comments in the codebase:

```bash
# Find all TODO comments
rg "// TODO" --type go

# Find all FIXME comments
rg "// FIXME" --type go

# Find hardcoded values that might need configuration
rg "time\.(Hour|Minute|Second)" --type go | grep "\*"

# Find potential security issues
rg "hardcoded|placeholder|demo-only" --type go -i
```

## 📊 Progress Tracking

Create a GitHub project board with columns:
- **Backlog** - All tasks not yet started
- **In Progress** - Tasks being worked on
- **Review** - PRs submitted
- **Done** - Completed and merged

Or use GitHub Issues with labels:
- `priority: critical` 🔴
- `priority: high` 🟠
- `priority: medium` 🟡
- `priority: low` 🟢
- `category: security`
- `category: feature`
- `category: quality`
- `category: docs`

## 💡 Tips for Task Completion

1. **Read the full task** - Don't skip sections, especially testing requirements
2. **Check dependencies** - Complete prerequisite tasks first
3. **Test thoroughly** - Follow all testing requirements
4. **Update documentation** - Keep docs in sync with code changes
5. **Small PRs** - One task per PR for easier review
6. **Reference task number** - In commits and PRs

## 🤝 Questions or Suggestions?

- Open an issue for task clarifications
- Propose new tasks via PR to this directory
- Discuss priorities in project discussions

---

**Last Updated:** 2025-11-16
**Total Tasks:** 8
**Total Estimated Effort:** 39-53 hours
