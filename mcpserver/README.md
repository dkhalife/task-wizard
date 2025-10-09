# Task Wizard MCP Server

A Model Context Protocol (MCP) server implementation for Task Wizard, built with C# and the official ModelContextProtocol SDK.

## Overview

This MCP server exposes Task Wizard's tasks and labels through MCP tools for CRUD operations. This is a proof-of-concept implementation with stub data to demonstrate the MCP server capabilities without requiring connection to the actual Task Wizard API server.

## Features

### Tools

The server provides 10 CRUD tools for interacting with tasks and labels using modern C# attributes:

#### Task Tools

- **`ListTasks`** - List all tasks
- **`GetTask`** - Get a specific task by ID
  - Parameters: `id` (int)
- **`CreateTask`** - Create a new task
  - Parameters: `title` (string, required), `nextDueDate` (DateTime?, optional), `endDate` (DateTime?, optional), `frequency` (string, optional), `isRolling` (bool, optional), `labels` (int[]?, optional)
- **`UpdateTask`** - Update an existing task
  - Parameters: `id` (int, required), `title` (string, required), `nextDueDate` (DateTime?, optional), `endDate` (DateTime?, optional), `frequency` (string, optional), `isRolling` (bool, optional), `labels` (int[]?, optional)
- **`DeleteTask`** - Delete a task
  - Parameters: `id` (int)

#### Label Tools

- **`ListLabels`** - List all labels
- **`GetLabel`** - Get a specific label by ID
  - Parameters: `id` (int)
- **`CreateLabel`** - Create a new label
  - Parameters: `name` (string, required), `color` (string, optional, default: "#000000")
- **`UpdateLabel`** - Update an existing label
  - Parameters: `id` (int, required), `name` (string, required), `color` (string, optional)
- **`DeleteLabel`** - Delete a label
  - Parameters: `id` (int)

## Requirements

- .NET 9.0 SDK or later

## Building

```bash
cd mcpserver/TaskWizardMcpServer
dotnet build
```

## Running

```bash
cd mcpserver/TaskWizardMcpServer
dotnet run
```

The server will start and listen on `http://localhost:3001` using HTTP transport for the MCP protocol.

## Project Structure

```
mcpserver/
└── TaskWizardMcpServer/
    ├── Models/
    │   ├── Label.cs        # Label model and request types
    │   └── Task.cs         # Task model and request types
    ├── Services/
    │   └── StubDataService.cs  # In-memory stub data service
    ├── Tools/
    │   ├── TaskTools.cs    # Task CRUD tools with attributes
    │   └── LabelTools.cs   # Label CRUD tools with attributes
    ├── Program.cs          # MCP server setup with HTTP transport
    └── TaskWizardMcpServer.csproj
```

## Implementation Details

This is a **proof-of-concept** implementation that uses:

- **Namespace**: `TaskWizard.McpServer` as the root namespace
- **HTTP Transport**: Uses HTTP instead of stdio for communication
- **Modern C# Attributes**: Tools are defined using `[McpServerTool]` and `[Description]` attributes
- **Stub Data**: The server maintains tasks and labels in memory. Data is not persisted and resets on restart.
- **No Authentication**: No authentication is required to use the server.
- **MCP Protocol**: Uses the official ModelContextProtocol.AspNetCore SDK (version 0.4.0-preview.2) for C#.

## CI/CD

A GitHub Actions workflow (`.github/workflows/dotnet-build.yml`) is configured to automatically build the .NET project on push and pull requests to the main branch.

## Future Enhancements

- Integration with the actual Task Wizard API server
- User authentication and authorization
- Persistent data storage
- Additional resources (e.g., task history)
- Additional tools (e.g., complete task, skip task)
