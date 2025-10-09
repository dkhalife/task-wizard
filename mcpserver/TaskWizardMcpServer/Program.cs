using System.Text.Json;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;
using TaskWizardMcpServer.Models;
using TaskWizardMcpServer.Services;

var dataService = new StubDataService();

McpServerOptions options = new()
{
    ServerInfo = new Implementation { Name = "task-wizard-mcp-server", Version = "1.0.0" },
    Handlers = new McpServerHandlers()
    {
        // Resource handlers
        ListResourcesHandler = (request, cancellationToken) =>
        {
            var resources = new List<Resource>
            {
                new()
                {
                    Uri = "taskwizard://tasks",
                    Name = "All Tasks",
                    Description = "List of all tasks",
                    MimeType = "application/json"
                },
                new()
                {
                    Uri = "taskwizard://labels",
                    Name = "All Labels",
                    Description = "List of all labels",
                    MimeType = "application/json"
                }
            };
            
            return ValueTask.FromResult(new ListResourcesResult { Resources = [.. resources] });
        },
        
        ReadResourceHandler = (request, cancellationToken) =>
        {
            var result = (request.Params?.Uri ?? "") switch
            {
                "taskwizard://tasks" => new ReadResourceResult
                {
                    Contents =
                    [
                        new TextResourceContents
                        {
                            Uri = "taskwizard://tasks",
                            MimeType = "application/json",
                            Text = JsonSerializer.Serialize(dataService.GetAllTasks(), new JsonSerializerOptions { WriteIndented = true })
                        }
                    ]
                },
                "taskwizard://labels" => new ReadResourceResult
                {
                    Contents =
                    [
                        new TextResourceContents
                        {
                            Uri = "taskwizard://labels",
                            MimeType = "application/json",
                            Text = JsonSerializer.Serialize(dataService.GetAllLabels(), new JsonSerializerOptions { WriteIndented = true })
                        }
                    ]
                },
                _ => throw new McpException($"Unknown resource URI: {request.Params?.Uri}")
            };
            
            return ValueTask.FromResult(result);
        },
        
        // Tool handlers
        ListToolsHandler = (request, cancellationToken) =>
        {
            var tools = new List<Tool>
            {
                // Task CRUD tools
                new()
                {
                    Name = "list_tasks",
                    Description = "List all tasks",
                    InputSchema = JsonSerializer.Deserialize<JsonElement>("""
                        {
                            "type": "object",
                            "properties": {}
                        }
                        """)
                },
                new()
                {
                    Name = "get_task",
                    Description = "Get a specific task by ID",
                    InputSchema = JsonSerializer.Deserialize<JsonElement>("""
                        {
                            "type": "object",
                            "properties": {
                                "id": {
                                    "type": "number",
                                    "description": "Task ID"
                                }
                            },
                            "required": ["id"]
                        }
                        """)
                },
                new()
                {
                    Name = "create_task",
                    Description = "Create a new task",
                    InputSchema = JsonSerializer.Deserialize<JsonElement>("""
                        {
                            "type": "object",
                            "properties": {
                                "title": {
                                    "type": "string",
                                    "description": "Task title"
                                },
                                "next_due_date": {
                                    "type": "string",
                                    "description": "Next due date (ISO 8601 format)"
                                },
                                "end_date": {
                                    "type": "string",
                                    "description": "End date (ISO 8601 format)"
                                },
                                "frequency": {
                                    "type": "string",
                                    "description": "Task frequency",
                                    "default": "once"
                                },
                                "is_rolling": {
                                    "type": "boolean",
                                    "description": "Is rolling task",
                                    "default": false
                                },
                                "labels": {
                                    "type": "array",
                                    "items": { "type": "number" },
                                    "description": "Label IDs"
                                }
                            },
                            "required": ["title"]
                        }
                        """)
                },
                new()
                {
                    Name = "update_task",
                    Description = "Update an existing task",
                    InputSchema = JsonSerializer.Deserialize<JsonElement>("""
                        {
                            "type": "object",
                            "properties": {
                                "id": {
                                    "type": "number",
                                    "description": "Task ID"
                                },
                                "title": {
                                    "type": "string",
                                    "description": "Task title"
                                },
                                "next_due_date": {
                                    "type": "string",
                                    "description": "Next due date (ISO 8601 format)"
                                },
                                "end_date": {
                                    "type": "string",
                                    "description": "End date (ISO 8601 format)"
                                },
                                "frequency": {
                                    "type": "string",
                                    "description": "Task frequency"
                                },
                                "is_rolling": {
                                    "type": "boolean",
                                    "description": "Is rolling task"
                                },
                                "labels": {
                                    "type": "array",
                                    "items": { "type": "number" },
                                    "description": "Label IDs"
                                }
                            },
                            "required": ["id", "title"]
                        }
                        """)
                },
                new()
                {
                    Name = "delete_task",
                    Description = "Delete a task",
                    InputSchema = JsonSerializer.Deserialize<JsonElement>("""
                        {
                            "type": "object",
                            "properties": {
                                "id": {
                                    "type": "number",
                                    "description": "Task ID"
                                }
                            },
                            "required": ["id"]
                        }
                        """)
                },
                // Label CRUD tools
                new()
                {
                    Name = "list_labels",
                    Description = "List all labels",
                    InputSchema = JsonSerializer.Deserialize<JsonElement>("""
                        {
                            "type": "object",
                            "properties": {}
                        }
                        """)
                },
                new()
                {
                    Name = "get_label",
                    Description = "Get a specific label by ID",
                    InputSchema = JsonSerializer.Deserialize<JsonElement>("""
                        {
                            "type": "object",
                            "properties": {
                                "id": {
                                    "type": "number",
                                    "description": "Label ID"
                                }
                            },
                            "required": ["id"]
                        }
                        """)
                },
                new()
                {
                    Name = "create_label",
                    Description = "Create a new label",
                    InputSchema = JsonSerializer.Deserialize<JsonElement>("""
                        {
                            "type": "object",
                            "properties": {
                                "name": {
                                    "type": "string",
                                    "description": "Label name"
                                },
                                "color": {
                                    "type": "string",
                                    "description": "Label color (hex format)",
                                    "default": "#000000"
                                }
                            },
                            "required": ["name"]
                        }
                        """)
                },
                new()
                {
                    Name = "update_label",
                    Description = "Update an existing label",
                    InputSchema = JsonSerializer.Deserialize<JsonElement>("""
                        {
                            "type": "object",
                            "properties": {
                                "id": {
                                    "type": "number",
                                    "description": "Label ID"
                                },
                                "name": {
                                    "type": "string",
                                    "description": "Label name"
                                },
                                "color": {
                                    "type": "string",
                                    "description": "Label color (hex format)"
                                }
                            },
                            "required": ["id", "name"]
                        }
                        """)
                },
                new()
                {
                    Name = "delete_label",
                    Description = "Delete a label",
                    InputSchema = JsonSerializer.Deserialize<JsonElement>("""
                        {
                            "type": "object",
                            "properties": {
                                "id": {
                                    "type": "number",
                                    "description": "Label ID"
                                }
                            },
                            "required": ["id"]
                        }
                        """)
                }
            };
            
            return ValueTask.FromResult(new ListToolsResult { Tools = [.. tools] });
        },
        
        CallToolHandler = (request, cancellationToken) =>
        {
            if (request.Params is null)
            {
                throw new McpException("Missing request parameters");
            }
            
            var text = request.Params.Name switch
            {
                // Task tools
                "list_tasks" => JsonSerializer.Serialize(dataService.GetAllTasks(), new JsonSerializerOptions { WriteIndented = true }),
                "get_task" => HandleGetTask(request),
                "create_task" => HandleCreateTask(request),
                "update_task" => HandleUpdateTask(request),
                "delete_task" => HandleDeleteTask(request),
                // Label tools
                "list_labels" => JsonSerializer.Serialize(dataService.GetAllLabels(), new JsonSerializerOptions { WriteIndented = true }),
                "get_label" => HandleGetLabel(request),
                "create_label" => HandleCreateLabel(request),
                "update_label" => HandleUpdateLabel(request),
                "delete_label" => HandleDeleteLabel(request),
                _ => throw new McpException($"Unknown tool: {request.Params.Name}")
            };
            
            return ValueTask.FromResult(new CallToolResult
            {
                Content = [new TextContentBlock { Text = text, Type = "text" }]
            });
        }
    }
};

await using McpServer server = McpServer.Create(new StdioServerTransport("task-wizard-mcp-server"), options);
await server.RunAsync();

// Helper methods for tool handlers
string HandleGetTask(RequestContext<CallToolRequestParams> request)
{
    var id = GetIntArgument(request, "id");
    var task = dataService.GetTask(id);
    return task != null
        ? JsonSerializer.Serialize(task, new JsonSerializerOptions { WriteIndented = true })
        : JsonSerializer.Serialize(new { error = "Task not found" });
}

string HandleCreateTask(RequestContext<CallToolRequestParams> request)
{
    var createRequest = new CreateTaskRequest
    {
        Title = GetStringArgument(request, "title") ?? "",
        NextDueDate = GetDateTimeArgument(request, "next_due_date"),
        EndDate = GetDateTimeArgument(request, "end_date"),
        Frequency = GetStringArgument(request, "frequency") ?? "once",
        IsRolling = GetBoolArgument(request, "is_rolling"),
        Labels = GetIntArrayArgument(request, "labels")
    };
    var task = dataService.CreateTask(createRequest);
    return JsonSerializer.Serialize(task, new JsonSerializerOptions { WriteIndented = true });
}

string HandleUpdateTask(RequestContext<CallToolRequestParams> request)
{
    var updateRequest = new UpdateTaskRequest
    {
        Id = GetIntArgument(request, "id"),
        Title = GetStringArgument(request, "title") ?? "",
        NextDueDate = GetDateTimeArgument(request, "next_due_date"),
        EndDate = GetDateTimeArgument(request, "end_date"),
        Frequency = GetStringArgument(request, "frequency") ?? "once",
        IsRolling = GetBoolArgument(request, "is_rolling"),
        Labels = GetIntArrayArgument(request, "labels")
    };
    var task = dataService.UpdateTask(updateRequest);
    return task != null
        ? JsonSerializer.Serialize(task, new JsonSerializerOptions { WriteIndented = true })
        : JsonSerializer.Serialize(new { error = "Task not found" });
}

string HandleDeleteTask(RequestContext<CallToolRequestParams> request)
{
    var id = GetIntArgument(request, "id");
    var success = dataService.DeleteTask(id);
    return JsonSerializer.Serialize(new { success });
}

string HandleGetLabel(RequestContext<CallToolRequestParams> request)
{
    var id = GetIntArgument(request, "id");
    var label = dataService.GetLabel(id);
    return label != null
        ? JsonSerializer.Serialize(label, new JsonSerializerOptions { WriteIndented = true })
        : JsonSerializer.Serialize(new { error = "Label not found" });
}

string HandleCreateLabel(RequestContext<CallToolRequestParams> request)
{
    var createRequest = new CreateLabelRequest
    {
        Name = GetStringArgument(request, "name") ?? "",
        Color = GetStringArgument(request, "color") ?? "#000000"
    };
    var label = dataService.CreateLabel(createRequest);
    return JsonSerializer.Serialize(label, new JsonSerializerOptions { WriteIndented = true });
}

string HandleUpdateLabel(RequestContext<CallToolRequestParams> request)
{
    var updateRequest = new UpdateLabelRequest
    {
        Id = GetIntArgument(request, "id"),
        Name = GetStringArgument(request, "name") ?? "",
        Color = GetStringArgument(request, "color") ?? "#000000"
    };
    var label = dataService.UpdateLabel(updateRequest);
    return label != null
        ? JsonSerializer.Serialize(label, new JsonSerializerOptions { WriteIndented = true })
        : JsonSerializer.Serialize(new { error = "Label not found" });
}

string HandleDeleteLabel(RequestContext<CallToolRequestParams> request)
{
    var id = GetIntArgument(request, "id");
    var success = dataService.DeleteLabel(id);
    return JsonSerializer.Serialize(new { success });
}

// Helper methods to extract arguments
int GetIntArgument(RequestContext<CallToolRequestParams> request, string key)
{
    if (request.Params?.Arguments?.TryGetValue(key, out var value) == true && value is JsonElement element)
    {
        return element.GetInt32();
    }
    return 0;
}

string? GetStringArgument(RequestContext<CallToolRequestParams> request, string key)
{
    if (request.Params?.Arguments?.TryGetValue(key, out var value) == true && value is JsonElement element)
    {
        return element.GetString();
    }
    return null;
}

bool GetBoolArgument(RequestContext<CallToolRequestParams> request, string key)
{
    if (request.Params?.Arguments?.TryGetValue(key, out var value) == true && value is JsonElement element)
    {
        return element.GetBoolean();
    }
    return false;
}

DateTime? GetDateTimeArgument(RequestContext<CallToolRequestParams> request, string key)
{
    var str = GetStringArgument(request, key);
    if (string.IsNullOrEmpty(str)) return null;
    return DateTime.TryParse(str, out var dt) ? dt : null;
}

List<int> GetIntArrayArgument(RequestContext<CallToolRequestParams> request, string key)
{
    if (request.Params?.Arguments?.TryGetValue(key, out var value) == true && value is JsonElement element && element.ValueKind == JsonValueKind.Array)
    {
        return element.EnumerateArray().Select(e => e.GetInt32()).ToList();
    }
    return new List<int>();
}
