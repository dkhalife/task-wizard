# Feature: Email Service

SMTP-based email delivery for account lifecycle and token management events.

## Capabilities

- Welcome/activation email sent on user sign-up with a verification link
- Password reset emails with time-limited tokens
- API token expiration reminder emails sent 72 hours before expiry
- HTML-formatted email templates
- Configurable SMTP settings (host, port, credentials) via config file or environment variables
- Validation of email configuration before attempting to send
