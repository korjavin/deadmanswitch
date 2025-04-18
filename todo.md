# Dead Man's Switch Todo List

This document tracks remaining tasks and implementation status for the Dead Man's Switch application.

## Core Components
- [x] Basic project structure
- [x] Configuration management system
- [x] Database models and storage layer
- [x] Web server implementation
- [x] User authentication system
- [x] Secret storage and encryption
- [ ] Scheduler for check-in reminders and deadlines
- [ ] Email notification system
- [ ] Telegram bot integration

## Deployment
- [x] Dockerfile created
- [x] Docker Compose configuration
- [ ] Environment variable documentation
- [ ] Production deployment guide

## Features & Improvements
- [ ] Reduce settings - remove 2FA and complex notification settings, just use defaults from environment variables
- [ ] Implement re-encoding of all secrets when user changes password
- [ ] Implement actual secret and recipients creation
- [ ] Replace hardcoded mock activity data in history.go with data retrieved from database
- [ ] Incorporate ping-checks history in activity data
- [ ] Fix template error on /secrets page
- [ ] Fix "in real implementation" TODOs in code
- [ ] Implement 2FA with Google Authenticator or similar for login

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
