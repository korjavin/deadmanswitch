# Dead Man's Switch Implementation Plan

## Project Overview
The Dead Man's Switch application allows users to store sensitive information that will be sent to designated contacts if the user fails to check in regularly. This document outlines the implementation plan and tracks progress.

## Implementation Status

### Core Components
- [x] Basic project structure
- [x] Configuration management system
- [x] Database models and storage layer
- [x] Web server implementation
- [x] User authentication system
- [x] Secret storage and encryption
- [ ] Scheduler for check-in reminders and deadlines
- [ ] Email notification system
- [ ] Telegram bot integration

### Deployment
- [x] Dockerfile created
- [x] Docker Compose configuration
- [ ] Environment variable documentation
- [ ] Production deployment guide

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

## Environment Variables

The application requires the following environment variables to be set:

| Variable | Required | Description | Default |
|----------|----------|-------------|---------|
| BASE_DOMAIN | Yes | Base domain for application URLs | - |
| TG_BOT_TOKEN | Yes | Telegram bot token for notifications | - |
| ADMIN_EMAIL | Yes | Administrator email address | - |
| DBPath | No | Path to SQLite database file | ./data/deadmanswitch.db |
| SMTP_HOST | No | SMTP server host | - |
| SMTP_PORT | No | SMTP server port | 587 |
| SMTP_USERNAME | No | SMTP authentication username | - |
| SMTP_PASSWORD | No | SMTP authentication password | - |
| SMTP_FROM | No | Email sender address | - |
| PING_FREQUENCY | No | How often users need to check in | 24h |
| PING_DEADLINE | No | Time after which secrets are released | 72h |
| DEBUG | No | Enable debug mode | false |
| LOG_LEVEL | No | Log verbosity level | info |

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