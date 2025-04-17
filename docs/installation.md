# Installation Guide

This guide will help you set up and run the Dead Man's Switch application on your own server.

## Requirements

- A server running Linux with Docker installed
- Domain name pointing to your server (for HTTPS and email verification)
- Email account for sending notifications (or use a service like SendGrid)
- Telegram account (to create a bot for notifications)

## Quick Start with Docker

The simplest way to deploy the Dead Man's Switch is using our pre-built Docker image:

```bash
docker run -d \
  --name deadmanswitch \
  -p 443:443 -p 80:80 \
  -v deadmanswitch-data:/app/data \
  -e BASE_DOMAIN="your-domain.com" \
  -e TG_BOT_TOKEN="your_telegram_bot_token" \
  -e SMTP_HOST="smtp.example.com" \
  -e SMTP_PORT="587" \
  -e SMTP_USERNAME="your_email@example.com" \
  -e SMTP_PASSWORD="your_email_password" \
  -e ADMIN_EMAIL="admin@example.com" \
  -e PING_FREQUENCY="3" \
  -e PING_DEADLINE="14" \
  ghcr.io/korjavin/deadmanswitch:latest
```

## Required Environment Variables

| Variable | Description |
|----------|-------------|
| BASE_DOMAIN | The domain name where your application is hosted |
| TG_BOT_TOKEN | Token for your Telegram bot (get from BotFather) |
| ADMIN_EMAIL | Email address for admin notifications |

## Optional Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| SMTP_HOST | SMTP server hostname | (required for email) |
| SMTP_PORT | SMTP server port | 587 |
| SMTP_USERNAME | SMTP username | (required for email) |
| SMTP_PASSWORD | SMTP password | (required for email) |
| SMTP_FROM | From address for emails | admin@yourdomain.com |
| PING_FREQUENCY | How often to ping users (days) | 3 |
| PING_DEADLINE | Time until switch activates (days) | 14 |
| DB_PATH | Database file location | /app/data/db.sqlite |
| LOG_LEVEL | Logging verbosity (debug, info, warn, error) | info |
| ENABLE_METRICS | Enable Prometheus metrics | false |
| DEBUG | Enable debug mode | false |

## Setting up a Telegram Bot

1. Open Telegram and search for "BotFather"
2. Start a chat and send `/newbot` command
3. Follow the instructions to create your bot
4. Copy the API token provided
5. Use this token as the `TG_BOT_TOKEN` environment variable

## Data Persistence

The application stores all data in `/app/data`. Mount this directory as a volume to ensure data persistence:

```bash
docker volume create deadmanswitch-data
```

## Manual Build

If you prefer to build the Docker image yourself:

```bash
git clone https://github.com/korjavin/deadmanswitch.git
cd deadmanswitch
docker build -t deadmanswitch .
```

## SSL/TLS Certificates

The application includes Caddy as a reverse proxy, which automatically obtains and renews Let's Encrypt SSL certificates. Ensure that:

1. Your domain is correctly pointing to your server's IP address
2. Ports 80 and 443 are accessible from the internet
3. The BASE_DOMAIN environment variable is set correctly

## Security Recommendations

1. **Firewall**: Configure your server's firewall to only allow necessary ports (80, 443)
2. **Regular Backups**: Backup the data volume regularly
3. **Updates**: Keep the application updated with the latest security patches
4. **Monitoring**: Set up monitoring for the container to ensure it's running

## Troubleshooting

### Check Logs

```bash
docker logs deadmanswitch
```

### Container Not Starting

Check if ports 80 and 443 are already in use by other services.

### Email Not Working

Verify your SMTP settings and ensure your email provider allows sending through SMTP.

### Certificate Issues

Make sure your domain is properly configured and pointing to your server.

## Support

If you encounter any issues, please open a ticket on our [GitHub repository](https://github.com/korjavin/deadmanswitch/issues).