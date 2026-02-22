# Task Wizard — Architecture Summary

Task Wizard is a self-hosted, privacy-focused task management application. It is composed of three main services and a shared protobuf contract layer.

## System Components

### 1. API Server (`apiserver/`)
- **Language**: Go
- **Framework**: Gin (HTTP), gRPC, WebSocket
- **DI**: Uber FX
- **Database**: SQLite (default) or MySQL via GORM
- **Role**: The central backend. Handles all business logic, persistence, authentication, background scheduling, notifications, and serves the frontend as static files.

### 2. Frontend (`frontend/`)
- **Language**: TypeScript
- **Framework**: React (class components) with Redux
- **Transport**: gRPC-Web as primary, HTTP REST as fallback
- **Role**: Single-page application for task management, label organization, notification configuration, API token management, and user settings.

### 3. MCP Server (`mcpserver/`)
- **Language**: C# (.NET 9)
- **Framework**: ASP.NET Core with ModelContextProtocol
- **Role**: Exposes task and label management as MCP (Model Context Protocol) tools so AI assistants can interact with Task Wizard programmatically. Currently uses in-memory stub data.

### 4. Protobuf Contracts (`proto/`)
- **Language**: Protocol Buffers v3
- **Role**: Single source of truth for all message types, enumerations, and service definitions shared between the API server (Go) and the frontend (TypeScript). Code generation scripts produce Go and TypeScript stubs.

## Communication

```
┌───────────┐  gRPC-Web / HTTP   ┌──────────────┐
│  Frontend │ ──────────────────► │  API Server  │
│  (React)  │ ◄──── WebSocket ── │  (Go / Gin)  │
└───────────┘                    └──────┬───────┘
                                       │ GORM
                                       ▼
                                 ┌────────────┐
                                 │  SQLite /   │
                                 │  MySQL      │
                                 └────────────┘

┌───────────┐  MCP (HTTP)
│ AI Client │ ──────────────────► ┌─────────────┐
└───────────┘                    │ MCP Server  │
                                 │ (.NET)      │
                                 └─────────────┘
```

## API Server Internals

| Layer | Directory | Purpose |
|-------|-----------|---------|
| HTTP Handlers | `internal/apis/` | REST + CalDAV route handlers |
| gRPC Services | `internal/grpc/` | Generated gRPC service implementations |
| Middleware | `internal/middleware/` | JWT auth, scope enforcement |
| Models | `internal/models/` | GORM data models |
| Repositories | `internal/repos/` | Database access layer |
| Services | `internal/services/` | Business logic, scheduler, notifications, housekeeping |
| Utilities | `internal/utils/` | Auth helpers, email, CalDAV parsing, DB setup |
| WebSocket | `internal/ws/` | Real-time push to connected clients |
| Migrations | `internal/migrations/` | Schema versioning |
| Config | `config/` | YAML-based configuration with env var overrides |

## Frontend Internals

| Layer | Directory | Purpose |
|-------|-----------|---------|
| Views | `src/views/` | Page-level React components (Tasks, Labels, Settings, Auth, etc.) |
| Store | `src/store/` | Redux slices for tasks, labels, user, tokens, feature flags, WebSocket, status |
| API | `src/api/` | HTTP/gRPC transport abstraction |
| gRPC | `src/grpc/` | Generated gRPC-Web client and type definitions |
| Models | `src/models/` | TypeScript interfaces and helpers |
| Components | `src/components/` | Shared UI components (ErrorBoundary, StatusList) |

## Key Architectural Patterns

- **Repository pattern** for data access abstraction
- **Service layer** for business logic separation
- **Dependency injection** (Uber FX) for wiring
- **Scope-based authorization** on API tokens (e.g. `task:read`, `label:write`, `dav:read`)
- **Background scheduler** for notifications, token cleanup, password reset expiration
- **Smart transport** in the frontend — tries WebSocket/gRPC first, falls back to HTTP
- **Feature flags** to toggle behaviors like WebSocket transport and auto-refresh
- **Real-time sync** via WebSocket with per-user connection tracking
