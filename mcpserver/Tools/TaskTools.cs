using System.ComponentModel;
using ModelContextProtocol.Server;
using TaskWizard.McpServer.Models;
using TaskWizard.McpServer.Services;

namespace TaskWizard.McpServer.Tools;

[McpServerToolType]
public class TaskTools(ApiProxyService api)
{
    [McpServerTool, Description("List all tasks")]
    public Task<string> ListTasks() =>
        api.GetAllTasks();

    [McpServerTool, Description("Get a specific task by ID")]
    public Task<string> GetTask([Description("Task ID")] int id) =>
        api.GetTask(id);

    [McpServerTool, Description("Create a new task")]
    public Task<string> CreateTask(
        [Description("Task title")] string title,
        [Description("Next due date (ISO 8601 format, e.g. 2025-01-15T00:00:00Z)")] string? nextDueDate = null,
        [Description("End date (ISO 8601 format)")] string? endDate = null,
        [Description("Frequency type: once, daily, weekly, monthly, yearly, custom")] string frequencyType = "once",
        [Description("Is rolling task")] bool isRolling = false,
        [Description("Label IDs")] int[]? labels = null) =>
        api.CreateTask(new CreateTaskRequest
        {
            Title = title,
            NextDueDate = nextDueDate,
            EndDate = endDate,
            IsRolling = isRolling,
            Frequency = new Frequency { Type = frequencyType },
            Labels = labels?.ToList() ?? []
        });

    [McpServerTool, Description("Update an existing task")]
    public Task<string> UpdateTask(
        [Description("Task ID")] int id,
        [Description("Task title")] string title,
        [Description("Next due date (ISO 8601 format, e.g. 2025-01-15T00:00:00Z)")] string? nextDueDate = null,
        [Description("End date (ISO 8601 format)")] string? endDate = null,
        [Description("Frequency type: once, daily, weekly, monthly, yearly, custom")] string frequencyType = "once",
        [Description("Is rolling task")] bool isRolling = false,
        [Description("Label IDs")] int[]? labels = null) =>
        api.UpdateTask(new UpdateTaskRequest
        {
            Id = id,
            Title = title,
            NextDueDate = nextDueDate,
            EndDate = endDate,
            IsRolling = isRolling,
            Frequency = new Frequency { Type = frequencyType },
            Labels = labels?.ToList() ?? []
        });

    [McpServerTool, Description("Delete a task")]
    public Task<string> DeleteTask([Description("Task ID")] int id) =>
        api.DeleteTask(id);

    [McpServerTool, Description("Mark a task as complete / done")]
    public Task<string> CompleteTask([Description("Task ID")] int id) =>
        api.CompleteTask(id);

    [McpServerTool, Description("Undo a task completion (mark as not done)")]
    public Task<string> UncompleteTask([Description("Task ID")] int id) =>
        api.UncompleteTask(id);

    [McpServerTool, Description("Skip a task (advance to next due date without completing)")]
    public Task<string> SkipTask([Description("Task ID")] int id) =>
        api.SkipTask(id);
}
