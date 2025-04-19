# Dead Man's Switch Todo List

This document tracks remaining tasks and implementation status for the Dead Man's Switch application.

## Core Components
- [x] Basic project structure
- [x] Configuration management system
- [x] Database models and storage layer
- [x] Web server implementation
- [x] User authentication system
- [x] Secret storage
- [ ] Secret proper encryption/re-encryption
- [ ] Scheduler for check-in reminders and deadlines
- [ ] Email notification system
- [ ] Telegram bot integration

## Deployment
- [x] Dockerfile created
- [x] Docker Compose configuration
- [ ] Environment variable documentation
- [ ] Production deployment guide

## Features & Improvements
- [x] We don't use phone method for contacting receipients, remove it from everywhere. we use only email for recipients
- [ ] Add tracing of adding tg accounts and connecting/disconnecting to tg bot into audit log and show it in history page
- [ ] Remove settings functionality - remove 2FA and complex notification settings, just use defaults from environment variables
- [x] Show connected telegram bots in profile page
- [x] Telegram bot on login, should determine it's login from bot.Me and write this info to database, then it's name should be shown in profile page in istruction to connect
- [x] On succeful tg bot connect, we store tg user id and username in database, and show this connection on profile page
- [ ] Implement profile update functionality, and allow user to disconect tg account from profile page, before disconecting message to tg user should be sent to warn user
- [ ] After connecting telegram bot, schedule sending pings with tg bot also, make it configurable in profile page
- [ ] When user logins send message to tg bot first if it's connected, then to email with explanation that someone just logged in and a link that can immediately invalidate all the sessions if user is not recognizing this activity.
- [ ] If we add recipient to already created secret we need to re-encrypt secret with key for this recipient
- [ ] Implement re-encoding of all secrets when user changes password
- [x] Implement actual secret and recipients creation with proper encryption
- [x] Implement recipient creation, editing, and deletion functionality
- [x] Implement recipient-secret management functionality
- [x] Replace hardcoded mock activity data in history.go with data retrieved from database
- [x] Incorporate ping-checks history in activity data
- [x] Fix template error on /secrets page
- [x] Implement real email sending for test contact with recipients
- [x] Add recipient confirmation functionality
- [ ] Fix "in real implementation" TODOs in code
- [x] Implement 2FA with Google Authenticator or similar for login
- [x] Implement passkeys to login, user can create up to 5 passkeys to his account (using go-webauthn/webauthn library)
- [ ] Research how can we let user login with passkey without make him to enter his email, and implement it if possilbe
- [ ] implement key derivation for encrypting secrets from master password
- [ ] Add ability to register github handle to user account, and then scheduler can check if user was active recently on github, and if yes reschedule pings on tg/email, as we know that user is okay. There we should think of some interface of ActivityProvider, as we are going to have many of them. Describe this well in userguide and in readme
- [ ] Similar to github, add ability to subscribe to activitypub account and implement the same logic
- [ ] similar to github and activitypub, add ability to monitor telegram channel
- [ ] understand how can we passover encrypted secrets to user since we have no way to know the master password. I think we have to create a key for every recipient, encode copy of the secret with this key, and provide key to user if switch is triggered. need to describe it well in diagram and security doc, brainstorm the idea and threats
- [ ] Find a way to secure bind telegram handle to registered user, that somebody else couldn't do it, and by those prevent switch trigger. Expalin this threat in security doc, and explain the solution well
- [ ] Add golangci-lint and fix all the warnings
- [ ] Add some security linters and fix the warnings
- [ ] Add some frontend tests and fix the warnings
- [x] Add unit tests for all components, targeting code coverage to 80%, and show coverage in github interface
- [ ] Add integration tests for all components, targeting code coverage to 80%, and show coverage in github interface

## Next Steps
1. Complete the scheduler implementation:
   - Implement time-based triggers for check-ins
   - Set up deadline monitoring
   - Create notification queue

2. Finish the notification systems:
   - Email notifications for check-ins and deadlines
   - Telegram bot commands and notification delivery
   - Secret release mechanism when deadlines are missed

3. Testing and Security:
   - Write unit and integration tests
   - Perform security audit of encryption methods
   - Test the entire workflow from secret creation to delivery

4. Documentation:
   - Complete user guide
   - Document API endpoints
   - Create admin documentation

## Testing Checklist
- [ ] User registration
- [ ] User authentication
- [ ] Secret creation with encryption
- [ ] Contact management
- [ ] Check-in functionality
- [ ] Deadline enforcement
- [ ] Email notifications
- [ ] Telegram notifications
- [ ] Secret delivery
