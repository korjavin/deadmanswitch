# TASK-001: Implement Master Key Management System

## Priority: CRITICAL 🔴

## Status: Not Started

## Category: Security / Cryptography

## Description

The application currently uses a hardcoded dummy master key (`"this-is-a-dummy-master-key-for-demo-only"`) for encrypting and decrypting secrets. This is a critical security vulnerability that makes the application unsuitable for production use.

## Problem

**Current State:**
- Hardcoded master key is used in 3 locations:
  - `internal/web/handlers/secrets.go:484`
  - `internal/web/handlers/secrets.go:584`
  - `internal/web/handlers/secrets.go:706`
  - `internal/web/handlers/secret_questions.go:247`
  - `internal/web/handlers/secret_questions.go:474`

**Why This is Critical:**
1. Anyone with access to the source code knows the encryption key
2. All secrets are encrypted with the same static key
3. If the key is compromised, all secrets for all users are compromised
4. There's no way to rotate or change the key
5. This completely undermines the security model of the application

**Current Code Example:**
```go
// From internal/web/handlers/secrets.go:484
// In a real implementation, we would get the master key from the user's session
// For now, we'll use a dummy master key for demonstration
masterKey := []byte("this-is-a-dummy-master-key-for-demo-only")
```

## Proposed Solution

Implement a proper master key management system with these components:

### 1. Add Master Key to Configuration
Add a new environment variable `MASTER_KEY` to the configuration system.

**Files to modify:**
- `internal/config/config.go` - Add `MasterKey []byte` field
- `internal/config/config.go` - Load from `MASTER_KEY` env variable
- `internal/config/config.go` - Add validation (minimum 32 bytes)
- `.env.example` - Document the variable
- `README.md` - Document key generation requirements

**Key Requirements:**
- Must be at least 32 bytes (256 bits)
- Should be randomly generated using cryptographically secure RNG
- Must be base64 encoded in environment variable
- Should validate on startup

### 2. Provide Key Generation Utility
Create a CLI command or script to generate secure master keys.

**Implementation:**
```go
// Example: cmd/keygen/main.go
package main

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "io"
)

func main() {
    key := make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, key); err != nil {
        panic(err)
    }
    encoded := base64.StdEncoding.EncodeToString(key)
    fmt.Printf("MASTER_KEY=%s\n", encoded)
}
```

### 3. Update Handlers to Use Config Master Key
Replace all hardcoded master keys with the config value.

**Files to modify:**
- `internal/web/handlers/secrets.go` (3 locations)
- `internal/web/handlers/secret_questions.go` (2 locations)

**Implementation:**
```go
// Instead of:
masterKey := []byte("this-is-a-dummy-master-key-for-demo-only")

// Use:
masterKey := h.config.MasterKey
```

### 4. Add Key Rotation Support (Future Enhancement)
While not required for this task, the architecture should support future key rotation:
- Store key version with encrypted data
- Support multiple active keys during transition
- Implement background re-encryption task

## Acceptance Criteria

- [ ] `MASTER_KEY` environment variable added to config
- [ ] Config validates master key (minimum 32 bytes)
- [ ] Key generation utility created
- [ ] All 5 hardcoded master key usages replaced
- [ ] Documentation updated (README, .env.example)
- [ ] Application fails to start if `MASTER_KEY` not set
- [ ] Unit tests added for config validation
- [ ] Security documentation updated to reflect new implementation

## Testing Requirements

1. **Unit Tests:**
   - Test config loading with valid key
   - Test config validation with invalid key (too short)
   - Test config validation with missing key
   - Test encryption/decryption with config key

2. **Integration Tests:**
   - Test secret creation with config key
   - Test secret retrieval with config key
   - Test secret editing with config key
   - Test secret questions with config key

3. **Manual Testing:**
   - Generate a new master key
   - Start application with new key
   - Create secrets and verify encryption
   - Restart application and verify secrets can be decrypted

## Files to Modify

1. **Configuration:**
   - `/internal/config/config.go` - Add MasterKey field and loading logic

2. **Handlers:**
   - `/internal/web/handlers/secrets.go` - Replace hardcoded keys (lines 484, 584, 706)
   - `/internal/web/handlers/secret_questions.go` - Replace hardcoded keys (lines 247, 474)

3. **Utilities:**
   - `/cmd/keygen/main.go` - New file for key generation

4. **Documentation:**
   - `/README.md` - Add master key setup instructions
   - `/.env.example` - Add MASTER_KEY example
   - `/docs/security.md` - Update security implementation status
   - `/docker-compose.yml` - Add MASTER_KEY to environment

5. **Tests:**
   - `/internal/config/config_test.go` - Add master key tests
   - `/internal/web/handlers/secrets_test.go` - Update to use config key

## Migration Considerations

**⚠️ Breaking Change Warning:**
This change will make all existing encrypted secrets unreadable unless you set the master key to the old hardcoded value temporarily.

**Migration Path:**
1. For existing deployments, set `MASTER_KEY` to base64 of old hardcoded key
2. Use the old key temporarily to decrypt existing secrets
3. Implement key rotation to migrate to a new secure key

**Old Key (for migration only):**
```bash
echo -n "this-is-a-dummy-master-key-for-demo-only" | base64
# dGhpcy1pcy1hLWR1bW15LW1hc3Rlci1rZXktZm9yLWRlbW8tb25seQ==
```

## References

- Current implementation: `/internal/crypto/crypto.go`
- Security documentation: `/docs/security.md` (lines 198-200)
- Related TODO item: `/todo.md` (line 12)
- Encryption strategy docs: `/docs/security.md` (lines 5-36)

## Estimated Effort

- **Complexity:** Medium
- **Time:** 4-6 hours
- **Risk:** High (critical security component)

## Dependencies

None - this is a foundational security fix

## Follow-up Tasks

After completing this task, consider:
- TASK-002: Implement key rotation mechanism
- TASK-003: Add key derivation from user passwords
- TASK-004: Implement per-user encryption keys
