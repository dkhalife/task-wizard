# Feature: WebSocket Real-Time Sync

Real-time communication channel between the API server and connected frontends.

## Capabilities

- Per-user WebSocket connections with JWT authentication via protocol headers
- Server broadcasts task updates to all active connections for a user
- Frontend uses WebSocket as a transport for API calls when available (feature-flag gated)
- Fallback to HTTP when WebSocket is unavailable or disabled
- Keep-alive via ping/pong mechanism (54s ping interval, 60s pong timeout)
- Sync status indicator in the navigation bar
- Automatic reconnection and data re-fetch on tab visibility change
