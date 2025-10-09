# Task Wizard MCP Server

A Model Context Protocol (MCP) server implementation for Task Wizard, built with C# and the official ModelContextProtocol SDK.

## Overview

This MCP server exposes Task Wizard's tasks and labels as resources and provides tools for CRUD operations. This is a proof-of-concept implementation with stub data to demonstrate the MCP server capabilities without requiring connection to the actual Task Wizard API server.

## Features

### Resources

The server exposes two main resources:

- **`taskwizard://tasks`** - List of all tasks
- **`taskwizard://labels`** - List of all labels

### Tools

#### Task Tools

- **`list_tasks`** - List all tasks
- **`get_task`** - Get a specific task by ID
  - Parameters: `id` (number)
- **`create_task`** - Create a new task
  - Parameters: `title` (string, required), `next_due_date` (string, optional), `end_date` (string, optional), `frequency` (string, optional), `is_rolling` (boolean, optional), `labels` (array of numbers, optional)
- **`update_task`** - Update an existing task
  - Parameters: `id` (number, required), `title` (string, required), `next_due_date` (string, optional), `end_date` (string, optional), `frequency` (string, optional), `is_rolling` (boolean, optional), `labels` (array of numbers, optional)
- **`delete_task`** - Delete a task
  - Parameters: `id` (number)

#### Label Tools

- **`list_labels`** - List all labels
- **`get_label`** - Get a specific label by ID
  - Parameters: `id` (number)
- **`create_label`** - Create a new label
  - Parameters: `name` (string, required), `color` (string, optional, default: "#000000")
- **`update_label`** - Update an existing label
  - Parameters: `id` (number, required), `name` (string, required), `color` (string, optional)
- **`delete_label`** - Delete a label
  - Parameters: `id` (number)

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

The server will start and communicate via stdio (standard input/output) using the MCP protocol.

## Project Structure

```
mcpserver/
└── TaskWizardMcpServer/
    ├── Models/
    │   ├── Label.cs        # Label model and request types
    │   └── Task.cs         # Task model and request types
    ├── Services/
    │   └── StubDataService.cs  # In-memory stub data service
    ├── Program.cs          # MCP server implementation
    └── TaskWizardMcpServer.csproj
```

## Implementation Details

This is a **proof-of-concept** implementation that uses:

- **Stub Data**: The server maintains tasks and labels in memory. Data is not persisted and resets on restart.
- **No Authentication**: No authentication is required to use the server.
- **MCP Protocol**: Uses the official ModelContextProtocol SDK (version 0.4.0-preview.2) for C#.

## Future Enhancements

- Integration with the actual Task Wizard API server
- User authentication and authorization
- Persistent data storage
- Additional resources (e.g., task history)
- Additional tools (e.g., complete task, skip task)
