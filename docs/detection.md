# Activity Detection and Dead Man's Switch Mechanism

This document explains how the Dead Man's Switch application detects user activity, sends reminders, and enforces deadlines.

## Activity Detection

The system monitors user activity from multiple sources to determine if a user is still active. This multi-source approach provides redundancy and convenience, ensuring the switch is only triggered when a user is genuinely inactive.

### Activity Sources

1. **Web Application Activity**
   - Logging into the application
   - Navigating through the application
   - Manually checking in via the web interface
   - Responding to email verification links

2. **Telegram Bot Activity**
   - Sending messages to the bot
   - Responding to ping messages via the "I'm OK" button
   - Using any bot commands

3. **GitHub Activity** (if configured)
   - Public commits
   - Issue comments
   - Pull request activity
   - Any other public GitHub events

4. **ActivityPub Activity** (if configured)
   - Posts and interactions on ActivityPub-compatible platforms

5. **Telegram Channel Activity** (if configured)
   - Posts and interactions in configured Telegram channels

### Activity Tracking

All detected activity updates the user's `LastActivity` timestamp in the database. This timestamp is used to determine when the next ping should be sent and whether the user has exceeded their deadline.

## Reminder Flow

When a user hasn't shown activity for their configured ping frequency (e.g., 7 days), the system initiates the reminder flow:

1. **Initial Reminder**
   - If the user has Telegram configured, a message is sent to their Telegram account first
   - An email reminder is sent if Telegram is not configured or as a backup

2. **Escalation**
   - If no response is received within a configurable period (default: 24 hours), a second reminder is sent
   - This reminder emphasizes the approaching deadline

3. **Final Warning**
   - If the deadline is approaching (e.g., 48 hours before), a final warning is sent
   - This warning clearly states when the switch will be triggered

## Deadline Enforcement

If a user fails to respond to reminders and exceeds their configured deadline (e.g., 14 days since last activity), the system enforces the deadline:

1. **Pre-Trigger Verification**
   - The system performs a final check of all activity sources
   - If any activity is detected, the deadline is reset

2. **Switch Triggering**
   - If no activity is detected, the switch is triggered
   - All configured secrets are prepared for delivery

3. **Secret Delivery**
   - Secrets are delivered to the designated recipients via email
   - Each recipient receives only the secrets assigned to them
   - Recipients receive a secure link to access the secrets

## Security Considerations

1. **False Positives**
   - The multi-source activity detection minimizes the risk of false triggers
   - Users can configure longer deadlines for added safety

2. **False Negatives**
   - The system prioritizes avoiding false triggers over ensuring timely delivery
   - Multiple notification methods increase the chance of user response

3. **Audit Trail**
   - All activity detection, reminders, and deadline enforcements are logged
   - Users can review their activity history in the web interface

## Configuration Options

Users can customize their Dead Man's Switch behavior:

1. **Ping Frequency**: How often the system checks for activity (in days)
2. **Ping Deadline**: How long after the last activity before the switch triggers (in days)
3. **Notification Methods**: Choose between Telegram, email, or both
4. **Activity Sources**: Enable or disable specific activity detection sources

## Implementation Details

The scheduler runs several periodic tasks:

1. **Ping Task** (every 5 minutes)
   - Checks for users who need to be pinged based on their last activity and ping frequency
   - Sends appropriate notifications

2. **Dead Switch Task** (every 15 minutes)
   - Checks for users who have exceeded their deadline
   - Triggers the switch for inactive users

3. **External Activity Task** (hourly)
   - Checks for activity on external platforms (GitHub, etc.)
   - Updates user activity timestamps accordingly

4. **Cleanup Task** (daily)
   - Removes expired sessions and other temporary data
