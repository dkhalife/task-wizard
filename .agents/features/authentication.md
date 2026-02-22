# Feature: Authentication

JWT-based authentication system for user identity and session management.

## Capabilities

- User sign-up with email, password (min 8 chars), and display name
- Account activation via email verification link
- Login with email and password (bcrypt hashed)
- JWT sessions with configurable timeout and refresh window
- Password reset flow via email with time-limited tokens
- Password change for logged-in users
- Registration can be disabled server-side via configuration
- Accounts can be disabled by an administrator
