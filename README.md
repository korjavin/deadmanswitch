# Dead Man's Switch

> ⚠️ **SECURITY WARNING**: This project is currently under active development with cryptography implementation still in progress. **DO NOT** use this application in production or for truly sensitive information until a stable release is available. The current implementation should only be used in trusted, isolated environments for testing purposes.

A secure, self-hosted application that acts as a digital "dead man's switch" - ensuring your sensitive information is only shared with specified recipients if you're unable to respond to regular check-ins.

## What is a Dead Man's Switch?

A dead man's switch is a device or system that is triggered if the human operator becomes incapacitated or fails to respond. In the digital context, it's a system that will execute predetermined actions if you don't regularly confirm you're still active.

## How it Works

1. **Create encrypted secrets** - Store sensitive information securely with strong encryption
2. **Set up recipients** - Designate who receives which secrets if the switch is triggered
3. **Configure check-in methods** - Set up Telegram and/or email verification
4. **Respond to regular pings** - Simply respond to a Telegram message or click an email link
5. **If you don't respond** - After a customizable deadline passes, your secrets are securely delivered to your designated recipients

## Why Self-Hosted?

I strongly believe that users should own their data, especially when it comes to highly sensitive information like what's stored in a dead man's switch. This is why I created this as a free, self-hosted solution rather than a paid service.

Self-hosting offers several critical advantages:

1. **Complete privacy** - Your sensitive data never leaves your control
2. **No third-party dependencies** - No reliance on external services that might disappear
3. **Full transparency** - You can verify exactly how your data is handled
4. **Customization** - Tailor the application to your specific needs
5. **Security** - Eliminate potential vulnerabilities from third-party hosting

For truly confidential information, entrusting it to any third-party service creates unnecessary risk. By self-hosting, you maintain complete control over your sensitive data throughout its lifecycle.

This project will always remain free and open source. If you find it valuable, please consider [sponsoring me on GitHub](https://github.com/sponsors/korjavin) to support ongoing development and maintenance.

## Key Features

- **Strong encryption** - All secrets are encrypted using industry-standard algorithms
- **Flexible recipient management** - Assign different secrets to different recipients
- **Dual verification methods** - Choose between Telegram and email for check-ins
- **Customizable schedules** - Configure ping frequency and response deadlines
- **Simple web interface** - Easily manage your secrets and recipients
- **Self-contained Docker image** - Simple deployment with automatic HTTPS
- **Complete audit logs** - Track all system activities

## Getting Started

### Quick Start with Docker Compose

1. Clone this repository
2. Copy `.env.example` to `.env` and modify the variables as needed
3. Run `docker-compose up -d`

For more detailed instructions, see the [Installation Guide](./docs/installation.md).

## Security

> ⚠️ **IMPORTANT**: The cryptography implementation is still under development. The current version uses placeholder encryption keys and simplified methods that are NOT suitable for production use.

Security is the highest priority for this application. See our [Security Documentation](./docs/security.md) for details on the encryption methods and threat model.

## License

This project is licensed under the MIT License - see the LICENSE file for details.