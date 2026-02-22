# Feature: MCP Server (AI Integration)

A Model Context Protocol server that exposes Task Wizard operations as tools for AI assistants.

## Capabilities

- Task tools: list, get, create, update, delete tasks
- Label tools: list, get, create, update, delete labels
- Runs as a standalone .NET 9 web service on port 3001
- Uses HTTP transport for MCP communication
- Currently backed by an in-memory stub data service with sample data
- Enables AI assistants (e.g. Claude, Copilot) to manage tasks programmatically
