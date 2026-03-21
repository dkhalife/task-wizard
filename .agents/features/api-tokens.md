# Feature: API Tokens

Fine-grained access tokens that allow external integrations to interact with Task Wizard.

## Capabilities

- Create named API tokens with an expiration date
- Scoped permissions: Tasks.Read, Tasks.Write, Labels.Read, Labels.Write, User.Read, User.Write, Tokens.Write
- Write scopes automatically include their corresponding read scope
- Tokens are validated the same way as JWT sessions but carry scope restrictions
- List and delete existing tokens from the settings UI
- Background housekeeping: expiration reminders sent 72h before expiry, expired tokens auto-cleaned
