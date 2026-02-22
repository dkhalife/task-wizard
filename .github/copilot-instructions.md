# Task Wizard — AI Coding Instructions

For detailed architecture, data flow, and feature documentation, see the `.agents/` directory:
- `.agents/architecture.md` — full system architecture, database schema, internal layers
- `.agents/features/` — per-feature documentation

## Project Overview

Task Wizard is a self-hosted, privacy-focused task management app: React frontend communicates via gRPC-Web/HTTP → Go API server handles business logic and persistence → SQLite/MySQL via GORM. A .NET MCP server exposes tools for AI assistants.

**Components:** API Server (`apiserver/`), Frontend (`frontend/`), MCP Server (`mcpserver/`), Proto (`proto/`)

## Development Commands

```powershell
# Backend
cd apiserver && go build ./...                      # Build
cd apiserver && golangci-lint run                    # Lint
cd apiserver && go test ./...                        # Test

# Frontend
cd frontend && yarn install --frozen-lockfile        # Install deps
cd frontend && yarn lint                             # Lint
cd frontend && yarn tsc                              # Type check
cd frontend && yarn build                            # Build (Vite)
cd frontend && yarn test                             # Unit tests (Jest)
cd frontend && yarn test:e2e                         # E2E tests (Playwright)

# MCP Server
cd mcpserver && dotnet build                         # Build

# Proto generation (after modifying .proto files)
cd proto && .\generate.ps1
```

## Code Style

### Comments
- **Never add comments that restate what the code does** — code must be self-documenting
- Comment only to explain **why** a decision was made or **how** a non-obvious algorithm works
- No commented-out / dead code — remove it, version control has history

### No Dead Code
- Do not leave unused functions, variables, imports, or commented-out blocks
- If code is no longer needed, delete it

### Go
- Error handling: return errors with context via `fmt.Errorf("message: %s", err.Error())`, log with `log.Errorf()` before returning HTTP errors
- Naming: PascalCase for exported, camelCase for unexported, short receivers (`r`, `s`, `h`)
- Imports: stdlib → internal (`dkhalife.com/tasks/core/internal/...`) → external; use short aliases for long imports
- DI: Uber FX with `fx.Provide(NewXxx)`, `fx.Invoke()` for route registration, `fx.Supply()` for config
- Models: GORM struct tags (`json`, `gorm`, `binding`), snake_case JSON field names
- Testing: testify suites extending `test.DatabaseTestSuite`, `s.Require().NoError()` assertions
- Context: propagate `context.Context` through all layers
- Logging: `logging.FromContext(ctx)` for context-aware logging

### TypeScript / React
- Class components with Redux `connect()` for page-level views
- Props and state defined as TypeScript `interface`
- Use `import type` for type-only imports
- Path aliases: `@/*` maps to `src/*`
- Naming: PascalCase for components, camelCase for functions/variables, UPPER_CASE for constants
- State management: Redux Toolkit slices with `createAsyncThunk` for async operations
- Typed hooks: `useAppDispatch`, `useAppSelector`

### C# / .NET
- .NET 9 with ASP.NET Core DI
- Private fields: `_camelCase` with underscore prefix
- MCP tools: `[McpServerToolType]` class-level + `[McpServerTool]` method-level attributes
- Null safety: explicit nullable returns (`Task?`), default string values (`= string.Empty`)

## Workflow Rules

- **Always build and test before completing work**: run `go build ./... && go test ./...` for backend, `yarn lint && yarn tsc && yarn build` for frontend, `dotnet build` for MCP server
- After proto changes, run `cd proto && .\generate.ps1` and verify all builds still pass

## CI/CD

GitHub Actions workflows in `.github/workflows/`:
- `api-build.yml`: Build + lint + test Go (Go 1.25), upload coverage to Codecov
- `frontend-build.yml`: Lint + type-check + build + Playwright E2E (Node 20.x, 22.x)
- `mcp-build.yml`: Build .NET 9
- `full-release.yml` / `cut-release.yml`: Release automation
