# GitHub Activity Monitoring

The Dead Man's Switch application includes a feature to monitor your GitHub activity and automatically postpone check-ins when you're active. This document explains how this feature works, how to set it up, and privacy considerations.

## How It Works

1. The system periodically checks the public GitHub API for activity from your GitHub username
2. If activity is detected since your last check-in, your next scheduled ping is automatically extended
3. This reduces the need for manual check-ins while maintaining the security of the dead man's switch

## Setting Up GitHub Activity Monitoring

1. Log in to your Dead Man's Switch account
2. Navigate to your Profile page
3. In the GitHub Integration section, enter your GitHub username
4. Click "Connect GitHub"
5. Your GitHub username will be saved and activity monitoring will begin immediately

## Disconnecting GitHub Integration

1. Log in to your Dead Man's Switch account
2. Navigate to your Profile page
3. In the GitHub Integration section, click "Disconnect GitHub"
4. Your GitHub username will be removed and activity monitoring will stop

## Technical Details

- The application checks for GitHub activity once per hour
- Only public activity is monitored (commits, pull requests, issues, comments, etc.)
- No GitHub authentication is required - only your public username is needed
- When activity is detected, your next scheduled ping is extended based on your ping frequency settings
- All activity checks are recorded in your audit log for transparency

## Privacy Considerations

- The application only accesses publicly available GitHub data
- No personal access tokens or GitHub authentication is required
- Your GitHub username is stored in the application's database
- Activity checks are performed using the public GitHub API
- No GitHub data is stored other than the timestamp of your latest activity

## Limitations

- Only public GitHub activity can be detected
- Activity in private repositories will not be detected
- The GitHub API has rate limits that may affect the frequency of checks
- The system relies on the GitHub API being available

## Troubleshooting

If your GitHub activity is not being detected:

1. Verify that your GitHub username is correct in your profile settings
2. Ensure that you have public activity on GitHub
3. Check the audit log to see if activity checks are being performed
4. If problems persist, contact the system administrator

## Future Enhancements

Future versions may include:

- Support for additional activity sources (GitLab, Bitbucket, etc.)
- More granular control over which types of GitHub activity to monitor
- Options to configure how activity affects check-in scheduling
