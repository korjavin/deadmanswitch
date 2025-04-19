# Authentication Guide

This guide explains the authentication methods available in the Dead Man's Switch application.

## Authentication Methods

The Dead Man's Switch application supports multiple authentication methods to provide both security and convenience:

1. **Password Authentication** - Traditional username/password login
2. **Two-Factor Authentication (2FA)** - Additional security layer using TOTP
3. **WebAuthn Passkeys** - Modern, phishing-resistant authentication using hardware security keys or biometrics

## Password Authentication

Password authentication is the default method and requires:

- Email address
- Strong password (minimum 8 characters, recommended to use a mix of letters, numbers, and symbols)

Passwords are securely hashed using Argon2id before storage, ensuring that even if the database is compromised, your actual password remains protected.

## Two-Factor Authentication (2FA)

After enabling 2FA in your profile settings:

1. Scan the provided QR code with an authenticator app (Google Authenticator, Authy, etc.)
2. Enter the 6-digit code from your authenticator app to verify setup
3. Save the provided backup codes in a secure location

When logging in with 2FA enabled, you'll need to provide:
- Your email and password
- The current 6-digit code from your authenticator app

## WebAuthn Passkeys

WebAuthn passkeys provide a more secure and convenient authentication method using:
- Hardware security keys (like YubiKey)
- Platform authenticators (like Windows Hello, Touch ID, or Android biometrics)

### How Passkeys Work

1. **Registration**: When you register a passkey, your device generates a public-private key pair. The private key never leaves your device, while the public key is stored on our server.

2. **Authentication**: When you log in, our server sends a challenge to your device. Your device signs this challenge with the private key, and our server verifies the signature using the stored public key.

### Security Benefits of Passkeys

- **Phishing Resistance**: Passkeys are bound to the specific domain, making them immune to phishing attacks
- **No Shared Secrets**: Your private key never leaves your device
- **No Password Transmission**: Nothing secret is transmitted over the network during authentication
- **Biometric Protection**: Many passkeys are protected by biometrics or PINs on your device

### Technical Implementation

The Dead Man's Switch application uses the [go-webauthn/webauthn](https://github.com/go-webauthn/webauthn) library for WebAuthn operations. Passkey data stored in the database includes:

- Credential ID: Unique identifier for the credential
- Public Key: Used to verify authentication attempts
- User ID: Links the passkey to your account
- Metadata: Name, creation time, last used time
- Sign Count: Security counter to prevent cloning attacks

### Managing Your Passkeys

You can manage your passkeys in the Profile section:

1. **Add a New Passkey**: Give it a recognizable name and follow the prompts on your device
2. **View Existing Passkeys**: See when each passkey was created and last used
3. **Delete Passkeys**: Remove passkeys you no longer use or trust

## Best Practices

For maximum security, we recommend:

1. Use a strong, unique password as your base authentication
2. Enable 2FA for an additional layer of security
3. Register at least two passkeys (preferably on different devices) for convenient and secure access
4. Keep your authenticator app and security keys in secure locations
5. Store your 2FA backup codes separately from your password

## Troubleshooting

### Password Issues
- If you forget your password, use the "Forgot Password" link on the login page
- Password reset links are sent to your registered email address

### 2FA Issues
- If you lose access to your authenticator app, use one of your backup codes to log in
- If you lose both your authenticator app and backup codes, contact support

### Passkey Issues
- If your security key is lost or damaged, use your password (and 2FA if enabled) to log in
- Register multiple passkeys on different devices as backups
- Some older browsers may not support WebAuthn - use a modern browser for the best experience
