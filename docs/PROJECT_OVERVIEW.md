# Dead Man's Switch - Project Overview

## Project Idea

**Dead Man's Switch** is a self-hosted security application that acts as a digital failsafe mechanism. The core concept:

- Users store encrypted sensitive information (secrets) that should only be shared with designated recipients if the user becomes incapacitated or unavailable
- The system regularly checks if the user is active through periodic "pings" (check-ins)
- If the user fails to respond within a configured deadline, the secrets are automatically delivered to pre-designated recipients
- The philosophy emphasizes **self-hosting** for complete privacy and data ownership

## Why Self-Hosted?

Unlike third-party "dead man's switch" services, this application gives users complete control over their data. By hosting it yourself, you ensure:
- No third party has access to your encrypted secrets
- You control when and how the system is maintained
- Your data never leaves infrastructure you control
- No subscription fees or service dependencies

## Core Use Cases

1. **Digital Legacy**: Share passwords, account access, and important information with family after death
2. **Emergency Access**: Ensure critical business information reaches successors if you're incapacitated
3. **Journalist/Activist Protection**: Automatically release sensitive documents if you're detained
4. **Key Person Risk**: Ensure critical information is shared if a key team member is unavailable

## Architecture Overview

### Technology Stack

**Backend:**
- **Language**: Go 1.24.1
- **Database**: SQLite (pure Go implementation, no CGO)
- **Web Framework**: Standard library + Gorilla Mux
- **Crypto**: AES-256-GCM, Argon2id, Shamir Secret Sharing
- **External Integration**: Telegram Bot API, GitHub API

**Frontend:**
- HTML templates (Go's `html/template`)
- JavaScript for dynamic interactions
- Playwright for E2E testing

**Infrastructure:**
- Docker & Docker Compose
- SQLite (file-based, no separate DB server)

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Web Interface                         │
│              (HTML Templates + JavaScript)                   │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                      Web Server Layer                        │
│         (Gorilla Mux Router + Handlers + Middleware)        │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌──────────────┬──────────────┬──────────────┬───────────────┐
│   Business   │  Auth Layer  │   Crypto     │   Scheduler   │
│    Logic     │   (2FA/Pass) │   (AES-GCM)  │   (Tasks)     │
└──────────────┴──────────────┴──────────────┴───────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                  Storage Layer (Repository)                  │
│                      SQLite Database                         │
└─────────────────────────────────────────────────────────────┘
         ↓                          ↓                  ↓
┌─────────────┐          ┌─────────────────┐  ┌──────────────┐
│   Email     │          │  Telegram Bot   │  │   Activity   │
│   Client    │          │     (API)       │  │  Providers   │
└─────────────┘          └─────────────────┘  └──────────────┘
```

### Core Components

1. **Config** (`/internal/config/`)
   - Environment-based configuration management
   - Validation and defaults

2. **Models** (`/internal/models/`)
   - Data models: User, Secret, Recipient, SecretAssignment, PingHistory, AuditLog, Session, Passkey, SecretQuestion

3. **Storage** (`/internal/storage/`)
   - Repository pattern for data access
   - SQLite implementation with migrations
   - Transaction support

4. **Crypto** (`/internal/crypto/`)
   - AES-256-GCM encryption
   - Argon2id key derivation
   - Layered encryption (DEK/KEK pattern)
   - Shamir secret sharing
   - Time-lock encryption

5. **Auth** (`/internal/auth/`)
   - TOTP (2FA) implementation
   - WebAuthn (passkey) support
   - Password hashing

6. **Web** (`/internal/web/`)
   - HTTP server and routing
   - Handlers for all routes
   - Authentication middleware
   - Template rendering

7. **Scheduler** (`/internal/scheduler/`)
   - Background task execution
   - Ping tasks (check-in reminders)
   - Dead switch tasks (deadline enforcement)
   - External activity monitoring
   - Reminder escalation system

8. **Email** (`/internal/email/`)
   - SMTP client
   - Template-based emails

9. **Telegram** (`/internal/telegram/`)
   - Bot API integration
   - Command handlers
   - User activity tracking

10. **Activity** (`/internal/activity/`)
    - Pluggable activity provider system
    - GitHub activity monitoring
    - Future: ActivityPub, Telegram channels

## Key Features

### 1. Multi-Factor Authentication
- Email/password login
- TOTP-based 2FA
- WebAuthn passkeys (hardware keys, biometrics)

### 2. Secret Management
- Create encrypted secrets (notes, credentials, files)
- AES-256-GCM encryption with unique DEK per secret
- Assign secrets to specific recipients

### 3. Recipient Management
- Add recipients with email contact
- Custom messages per recipient
- Secret questions (Shamir sharing + time-lock)

### 4. Dead Man's Switch Mechanism
- Configurable ping frequency (1-7 days)
- Configurable deadline (7-30 days)
- Multiple notification methods (email, Telegram)
- Escalating reminder system (normal → urgent → final warning)

### 5. Multi-Source Activity Detection
- Web application login/usage
- Manual check-ins
- Telegram bot interactions
- GitHub activity monitoring
- Extensible provider system

### 6. Automatic Secret Delivery
- Triggers when deadline expires
- Email delivery to recipients
- Access codes for security
- Per-recipient custom messages

## User Flows

### Initial Setup
1. User registers (email + password)
2. Enable 2FA (optional)
3. Add passkeys (optional)
4. Create encrypted secrets
5. Add recipients
6. Assign secrets to recipients
7. Configure ping settings
8. Enable pinging

### Regular Check-in (Email)
1. Scheduler detects user needs ping
2. System generates verification code
3. Email sent to user
4. User clicks verification link
5. LastActivity updated
6. Next ping rescheduled

### Regular Check-in (Telegram)
1. Scheduler detects user needs ping
2. Telegram message with "I'm OK" button
3. User clicks button
4. LastActivity updated
5. Ping marked as responded

### External Activity Detection
1. Hourly scheduler task runs
2. GitHub API checked for recent events
3. If activity found → LastActivity updated
4. Next ping automatically rescheduled

### Dead Switch Trigger
1. Deadline expires without response
2. Final activity check (all sources)
3. If no activity → trigger switch
4. Retrieve all recipient assignments
5. For each recipient:
   - Generate access code
   - Send delivery email
   - Log delivery event
6. Disable pinging
7. Create audit log entry

## Security Model

### Encryption Layers

1. **Per-Secret Data Encryption Keys (DEK)**
   - Each secret encrypted with unique random key
   - Allows selective sharing without exposing all secrets

2. **Key Encryption Key (KEK)**
   - DEKs encrypted with user's master key
   - Master key derived from user password via Argon2id

3. **Shamir Secret Sharing**
   - Secrets can be split into shares
   - Requires threshold answers to reconstruct

4. **Time-Lock Encryption**
   - Secret questions locked until deadline
   - Automatically re-encrypted if user still active

### Authentication Security

- Passwords hashed with Argon2id
- TOTP 2FA support with backup codes
- WebAuthn passkeys for phishing-resistant auth
- Secure session management
- Comprehensive audit logging

### Telegram Account Binding Security

To prevent attackers from binding their own Telegram account to trigger the dead man's switch:
1. Pending verification state for bindings
2. Unique verification codes sent via email
3. 24-hour expiration on pending bindings
4. Only one pending binding per user

## Current Development Status

### ✅ Implemented
- Core authentication system (password, 2FA, passkeys)
- Secret storage with encryption
- Recipient management
- Scheduler with ping/deadline tasks
- Email notifications
- Telegram bot integration
- GitHub activity monitoring
- Frontend E2E tests (Playwright)
- Unit test coverage (~80%)

### ⚠️ In Development (NOT Production-Ready)

**Critical Security Issues:**
1. **Hardcoded master encryption key** - Must be replaced with proper key management
2. **Simplified cryptography** - Current implementation is for development only
3. **Missing access code storage** - TODOs in scheduler for secure storage
4. **Incomplete audit logging** - Not all events properly logged

**Missing Features:**
1. Secret questions (Shamir + timelock) - partially implemented
2. ActivityPub monitoring
3. Telegram channel monitoring
4. Profile update functionality
5. Integration tests
6. Security linters

### 📋 Planned Features
- Remove unused user name field
- Consolidate hardcoded time constants
- Move email templates to dedicated folder
- Enhanced audit logging for Telegram connections
- Passkey-based second factor option

## Development Patterns

- **Repository Pattern**: Clean data access abstraction
- **Provider Pattern**: Pluggable activity monitoring
- **Task Scheduler Pattern**: Background job processing
- **Middleware Chain**: Request processing (auth, logging)
- **Layered Encryption**: Defense in depth for secrets

## Testing Strategy

- **Unit Tests**: Go's native testing framework, targeting 80% coverage
- **E2E Tests**: Playwright for frontend flows
- **Mock Repositories**: Reusable mocks in `storage_test` package
- **Dynamic Coverage**: Coverage tracking across runs

## Documentation

- `/docs/security.md` - Security model and threat analysis
- `/docs/detection.md` - Activity detection mechanisms
- `/docs/faq.md` - Frequently asked questions
- `/docs/dynamic-coverage.md` - Coverage tracking documentation
- `/todo.md` - Implementation task list

## Important Warnings

> ⚠️ **CRITICAL SECURITY WARNING**: The current implementation uses placeholder encryption keys and simplified cryptography methods for development purposes only. Using this application for sensitive information in its current state is **STRONGLY DISCOURAGED**.

See `docs/security.md` for detailed security status and limitations.

## Deployment

The application is designed to run in Docker:

```bash
docker-compose up -d
```

Configuration via environment variables:
- `BASE_DOMAIN` - Application domain
- `TG_BOT_TOKEN` - Telegram bot token
- `SMTP_*` - Email configuration
- `PING_FREQUENCY` - Check-in frequency (1-7 days)
- `PING_DEADLINE` - Inactivity deadline (7-30 days)

## Contributing

See `CONTRIBUTING.md` for development guidelines, testing requirements, and code standards.

## License

[License information not found in codebase]
