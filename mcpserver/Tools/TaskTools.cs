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
        [Description("Frequency type: once, daily, weekly, monthly, yearly")] string frequencyType = "once",
        [Description("Rolling: reschedule from completion date instead of original due date")] bool isRolling = false,
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

    [McpServerTool, Description("Create a new task with a custom recurrence schedule")]
    public Task<string> CreateCustomTask(
        [Description("Task title")] string title,
        [Description("Recurrence mode: interval, days_of_the_week, day_of_the_months")] string on,
        [Description("Next due date (ISO 8601 format, e.g. 2025-01-15T00:00:00Z)")] string? nextDueDate = null,
        [Description("End date (ISO 8601 format)")] string? endDate = null,
        [Description("Repeat interval (required when on=interval)")] int? every = null,
        [Description("Interval unit: hours, days, weeks, months, years (required when on=interval)")] string? unit = null,
        [Description("Days of the week (0=Sun..6=Sat, required when on=days_of_the_week)")] int[]? days = null,
        [Description("Months (0=Jan..11=Dec, required when on=day_of_the_months)")] int[]? months = null,
        [Description("Rolling: reschedule from completion date instead of original due date")] bool isRolling = false,
        [Description("Label IDs")] int[]? labels = null) =>
        api.CreateTask(new CreateTaskRequest
        {
            Title = title,
            NextDueDate = nextDueDate,
            EndDate = endDate,
            IsRolling = isRolling,
            Frequency = new Frequency
            {
                Type = "custom",
                On = on,
                Every = every,
                Unit = unit,
                Days = days?.ToList(),
                Months = months?.ToList(),
            },
            Labels = labels?.ToList() ?? []
        });

    [McpServerTool, Description("Update an existing task")]
    public Task<string> UpdateTask(
        [Description("Task ID")] int id,
        [Description("Task title")] string title,
        [Description("Next due date (ISO 8601 format, e.g. 2025-01-15T00:00:00Z)")] string? nextDueDate = null,
        [Description("End date (ISO 8601 format)")] string? endDate = null,
        [Description("Frequency type: once, daily, weekly, monthly, yearly")] string frequencyType = "once",
        [Description("Rolling: reschedule from completion date instead of original due date")] bool isRolling = false,
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
