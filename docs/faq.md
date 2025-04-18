# Frequently Asked Questions

## Privacy & Security

### How secure is my data in Dead Man's Switch?
Your data is encrypted using industry-standard AES-256 encryption. The encryption happens on your server, and only encrypted data is stored in the database. This means that even if someone gained access to your database, they couldn't read your secrets without the encryption keys.

### Who can access my stored secrets?
Only you and the recipients you explicitly designate can access your secrets. Since this is a self-hosted application, no third parties (including the developer) have access to your data. Your secrets remain entirely under your control.

### Is my data sent to any external servers?
No. Dead Man's Switch is completely self-contained. The only external communications are the notifications sent to your recipients when the switch is triggered, and those only contain the information needed to access the secrets you've designated for them.

### What happens if someone hacks into my server?
If someone gains access to your server, they would only find encrypted data. Without your master password, they cannot decrypt your secrets. However, it's still important to secure your server with proper authentication, firewalls, and regular security updates.

## Functionality

### How does the check-in system work?
You'll receive regular check-in requests via email or Telegram (based on your configuration). You simply need to respond to these requests to confirm you're still active. If you miss a check-in, you'll receive additional reminders before the switch is triggered.

### What happens if I lose access to my check-in methods?
You can configure multiple check-in methods (both email and Telegram). If you lose access to one, you can still check in using the other. Additionally, you can log in to the web interface and manually check in at any time.

### Can I customize how often I need to check in?
Yes, you can configure both the frequency of check-ins and the deadline (how long after a missed check-in the switch is triggered). This allows you to balance security with convenience based on your personal needs.

## Technical Details

### What technologies does Dead Man's Switch use?
Dead Man's Switch is built with Go for the backend, using SQLite for the database. The frontend uses standard HTML, CSS, and JavaScript without heavy frameworks. This makes it lightweight and easy to deploy on almost any server.

### Can I run this on a Raspberry Pi or other low-power device?
Yes! The application is designed to be lightweight and can run on modest hardware, including a Raspberry Pi or similar single-board computer.

### How can I back up my data?
Since all data is stored in a SQLite database file, you can simply back up the data directory. We recommend setting up regular backups to a secure location. The database file is located at the path specified in your configuration (default: `/app/data/db.sqlite`).

### Is there an API for integration with other systems?
Currently, there is no public API, but this is planned for future releases. If you have specific integration needs, please open an issue on GitHub.
