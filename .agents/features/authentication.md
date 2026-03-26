# Feature: Authentication

Microsoft Entra ID (OIDC) authentication for user identity and session management.

## Capabilities

- Single sign-on via Microsoft Entra ID (Azure AD) using OIDC
- Backend validates Entra Bearer tokens against the remote JWKS endpoint
- User identity extracted from token claims: `tid` (tenant/directory ID) and `oid` (object ID)
- Auto-provisioning: users are created on first login via `EnsureUser()`
- No local passwords, email, or display name stored (privacy-focused)
- Frontend uses `@azure/msal-browser` for login redirect and silent token refresh
- Development bypass mode when Entra is disabled (creates a dev user with hardcoded IDs)
- Accounts can be disabled by an administrator
- Entra configuration via YAML config or environment variables (`TW_ENTRA_*`)
