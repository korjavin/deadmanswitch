# Security Documentation

> ⚠️ **CRITICAL SECURITY WARNING**: The current implementation uses placeholder encryption keys and simplified cryptography methods for development purposes only. The encryption described in this document represents the target security model, but is **NOT FULLY IMPLEMENTED** in the current version. Using this application for sensitive information in its current state is **STRONGLY DISCOURAGED**.

## Encryption Strategy

### Secret Encryption

The Dead Man's Switch application uses a layered encryption approach to protect your sensitive information:

1. **AES-256-GCM** for symmetric encryption of the secret content
   - Authenticated encryption provides both confidentiality and integrity
   - Unique nonce (IV) generated for each encryption operation
   - Authentication tags verify the integrity of the ciphertext

2. **Key Derivation**
   - User master password is processed through Argon2id with high compute parameters
   - Salt is randomly generated and stored with the encrypted data
   - Stretches the password to create a strong encryption key

3. **Per-Secret Encryption**
   - Each secret is encrypted with a unique data encryption key (DEK)
   - DEKs are themselves encrypted with the user's master key (key encryption key, KEK)
   - This approach allows sharing specific secrets without exposing others

### Storage Security

1. **No Plaintext Storage**
   - Secrets are never stored in plaintext
   - Master password is never stored
   - Only encrypted data and necessary metadata are persisted

2. **Database Security**
   - All sensitive database fields are encrypted
   - Access to the database is restricted and authenticated
   - Regular backups are encrypted

## Threat Model

### In-Scope Threats

1. **Unauthorized Server Access**
   - Even with full server access, an attacker cannot decrypt secrets without the master password
   - Encrypted data remains protected even if database is compromised

2. **Network Eavesdropping**
   - All communications use TLS encryption
   - API calls containing sensitive data use additional application-level encryption

3. **Brute Force Attacks**
   - Argon2id with high compute parameters makes password brute-forcing impractical
   - Rate limiting on authentication endpoints
   - Account lockout mechanisms after multiple failed attempts

4. **Server Compromise**
   - Secrets remain encrypted even if the server is fully compromised
   - No keys are stored that would allow decryption without user authentication

### Out-of-Scope Threats

1. **Client-side Compromise**
   - Malware on the user's device could intercept the master password
   - Users should ensure their devices are secure

2. **Social Engineering**
   - Users must protect their master password against phishing and other social attacks

3. **Physical Security**
   - Physical access to an unlocked user device is outside our protection scope

## Security Recommendations

1. **Strong Master Password**
   - Use a long, random, unique password
   - Consider using a password manager

2. **Regular Backups**
   - Backup your encrypted database regularly
   - Test restoration procedures

3. **Access Control**
   - Limit server access to authorized personnel only
   - Use strong authentication for server administration

4. **Updates**
   - Keep the application and its dependencies up to date
   - Monitor security advisories

## External Activity Monitoring

### GitHub Activity Monitoring

The application can monitor your public GitHub activity to automatically postpone check-ins:

1. **Privacy Considerations**
   - Only public GitHub API data is accessed
   - No GitHub authentication or personal access tokens are required
   - Only your GitHub username is stored in the database
   - No GitHub data is stored other than the timestamp of your latest activity

2. **Security Implications**
   - Reduces the frequency of manual check-ins, improving usability
   - Provides an additional signal of user activity beyond explicit check-ins
   - All activity checks are logged for transparency and auditability
   - Users can disconnect GitHub integration at any time

3. **Potential Risks**
   - If an attacker gains access to your GitHub account, they could potentially prevent the dead man's switch from triggering
   - This risk is mitigated by the fact that the switch will still trigger if no manual check-ins are performed within the configured deadline
   - Users should maintain strong security on their GitHub accounts if using this feature

For more details, see the [GitHub Activity Monitoring Guide](./github-activity.md).

## Authentication Security

### Password Authentication

The application supports traditional password-based authentication with the following security measures:

1. **Password Hashing** - Passwords are hashed using Argon2id with secure parameters
2. **Rate Limiting** - Failed login attempts are rate-limited to prevent brute force attacks
3. **Session Management** - Secure HTTP-only cookies with appropriate flags

### Two-Factor Authentication (2FA)

Users can enable 2FA using TOTP (Time-based One-Time Password) for additional security:

1. **TOTP Implementation** - Compatible with standard authenticator apps (Google Authenticator, Authy, etc.)
2. **Secure Backup Codes** - Recovery codes provided for account recovery

### WebAuthn Passkeys

The application supports WebAuthn passkeys as a phishing-resistant authentication method:

1. **Implementation** - Uses the go-webauthn/webauthn library for server-side WebAuthn operations
2. **Storage** - Passkey data is stored securely in the database with the following information:
   - Credential ID: Unique identifier for the credential (stored as binary)
   - Public Key: The public key component used for verification (stored as binary)
   - User ID: Association with the user account
   - Metadata: Name, creation time, last used time
   - Sign Count: Counter to prevent cloning attacks
   - AAGUID: Authenticator Attestation GUID identifying the authenticator model
3. **Security Benefits**:
   - Phishing resistance - credentials are bound to the origin
   - No shared secrets - private keys never leave the authenticator
   - Biometric or PIN protection on the device side
   - No password transmission over the network

## Telegram Account Binding Security

The application allows users to bind their Telegram account for notifications and activity monitoring. This feature has specific security considerations:

### Threat Model

1. **Unauthorized Binding**:
   - An attacker could attempt to bind their Telegram account to a user's profile
   - This would allow them to receive notifications and potentially prevent the dead man's switch from triggering
   - Particularly dangerous if combined with other compromised authentication methods

2. **Notification Spoofing**:
   - Fake notifications could be sent to mislead the user
   - Could be used in phishing attempts or to hide legitimate security alerts

### Security Measures

1. **Binding Verification**:
   - Telegram binding requires authentication through an existing method (password + 2FA or passkey)
   - A unique verification code is sent to the user's registered email
   - The code must be entered in the Telegram chat to complete binding

2. **Binding Confirmation**:
   - Successful binding triggers an immediate notification to both email and Telegram
   - Users can review and revoke Telegram connections in their profile
   - All binding events are logged in the audit trail

3. **Rate Limiting**:
   - Binding attempts are rate-limited to prevent brute force attacks
   - Suspicious activity triggers additional verification steps

4. **Pending Verification State**:
   - When a user initiates Telegram binding, their Telegram ID is stored in a pending state
   - The pending state requires explicit verification through a unique code
   - Until verified, the Telegram ID cannot be used to receive notifications or affect the dead man's switch
   - Pending bindings expire after 24 hours if not verified
   - Only one pending binding is allowed per user at a time

### Threat Mitigation
- **Preventing Unauthorized Binding**:
  - The pending verification state ensures no Telegram account can be bound without explicit user confirmation
  - Even if an attacker gains temporary access to the user's account, they cannot complete binding without access to the verification code
  - The pending state prevents attackers from pre-binding their own Telegram account


## Current Implementation Status

The current version of the application has the following limitations:

1. **Hardcoded Master Key** - A static, hardcoded master key is used for development purposes
2. **Simplified Key Management** - The full key derivation and management system is not yet implemented
3. **Limited Authentication** - Some password security features are partially implemented
4. **Incomplete Audit Logging** - Not all security events are properly logged and monitored

These issues will be addressed before the first stable release. The application should only be used in isolated, trusted environments for testing and development purposes until these issues are resolved.

## Security Disclosure

If you discover a security vulnerability, please send an email to [security contact]. We take all security concerns seriously and will respond promptly.