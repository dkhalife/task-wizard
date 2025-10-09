using System.ComponentModel;
using System.Text.Json;
using ModelContextProtocol.Server;
using TaskWizard.McpServer.Models;
using TaskWizard.McpServer.Services;

namespace TaskWizard.McpServer.Tools;

[McpServerToolType]
public class TaskTools
{
    private readonly StubDataService _dataService;

    public TaskTools(StubDataService dataService)
    {
        _dataService = dataService;
    }

    [McpServerTool, Description("List all tasks")]
    public string ListTasks()
    {
        return JsonSerializer.Serialize(_dataService.GetAllTasks(), new JsonSerializerOptions { WriteIndented = true });
    }

    [McpServerTool, Description("Get a specific task by ID")]
    public string GetTask([Description("Task ID")] int id)
    {
        var task = _dataService.GetTask(id);
        return task != null
            ? JsonSerializer.Serialize(task, new JsonSerializerOptions { WriteIndented = true })
            : JsonSerializer.Serialize(new { error = "Task not found" });
    }

    [McpServerTool, Description("Create a new task")]
    public string CreateTask(
        [Description("Task title")] string title,
        [Description("Next due date (ISO 8601 format)")] DateTime? nextDueDate = null,
        [Description("End date (ISO 8601 format)")] DateTime? endDate = null,
        [Description("Task frequency")] string frequency = "once",
        [Description("Is rolling task")] bool isRolling = false,
        [Description("Label IDs")] int[]? labels = null)
    {
        var createRequest = new CreateTaskRequest
        {
            Title = title,
            NextDueDate = nextDueDate,
            EndDate = endDate,
            Frequency = frequency,
            IsRolling = isRolling,
            Labels = labels?.ToList() ?? new List<int>()
        };
        var task = _dataService.CreateTask(createRequest);
        return JsonSerializer.Serialize(task, new JsonSerializerOptions { WriteIndented = true });
    }

    [McpServerTool, Description("Update an existing task")]
    public string UpdateTask(
        [Description("Task ID")] int id,
        [Description("Task title")] string title,
        [Description("Next due date (ISO 8601 format)")] DateTime? nextDueDate = null,
        [Description("End date (ISO 8601 format)")] DateTime? endDate = null,
        [Description("Task frequency")] string frequency = "once",
        [Description("Is rolling task")] bool isRolling = false,
        [Description("Label IDs")] int[]? labels = null)
    {
        var updateRequest = new UpdateTaskRequest
        {
            Id = id,
            Title = title,
            NextDueDate = nextDueDate,
            EndDate = endDate,
            Frequency = frequency,
            IsRolling = isRolling,
            Labels = labels?.ToList() ?? new List<int>()
        };
        var task = _dataService.UpdateTask(updateRequest);
        return task != null
            ? JsonSerializer.Serialize(task, new JsonSerializerOptions { WriteIndented = true })
            : JsonSerializer.Serialize(new { error = "Task not found" });
    }

    [McpServerTool, Description("Delete a task")]
    public string DeleteTask([Description("Task ID")] int id)
    {
        var success = _dataService.DeleteTask(id);
        return JsonSerializer.Serialize(new { success });
    }
}
