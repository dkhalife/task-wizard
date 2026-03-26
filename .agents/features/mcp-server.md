# Feature: MCP Server (AI Integration)

A Model Context Protocol server that exposes Task Wizard operations as tools for AI assistants.

## Capabilities

- Task tools: list, get, create, create with custom recurrence, update, delete, complete, uncomplete, skip, list due before date, list by label
- Label tools: list, create, update, delete
- Runs as a standalone .NET 9 web service on port 3001
- Uses HTTP transport for MCP communication
- Proxies all requests to the real API server via `ApiProxyService` (configurable via `TW_API_URL`, default `http://localhost:2021`)
- Authenticates using Microsoft Entra ID (requires `TW_ENTRA_TENANT_ID`, `TW_ENTRA_CLIENT_ID`, `TW_ENTRA_AUDIENCE`)
- Enables AI assistants (e.g. Claude, Copilot) to manage tasks programmatically
