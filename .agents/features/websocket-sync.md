# Feature: WebSocket Real-Time Sync

Real-time communication channel between the API server and connected frontends.

## Capabilities

- Per-user WebSocket connections with JWT authentication via protocol headers
- Server broadcasts task updates to all active connections for a user
- Frontend uses WebSocket for real-time push updates when connected
- WebSocket connection is established when authenticated (not feature-flag gated)
- Keep-alive via ping/pong mechanism (54s ping interval, 60s pong timeout)
- Sync status indicator in the navigation bar
- Automatic reconnection and data re-fetch on tab visibility change
