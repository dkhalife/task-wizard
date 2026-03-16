using System.Text.Json.Serialization;

namespace TaskWizard.McpServer.Models;

public class Task
{
    [JsonPropertyName("id")]
    public int Id { get; set; }

    [JsonPropertyName("title")]
    public string Title { get; set; } = string.Empty;

    [JsonPropertyName("frequency")]
    public Frequency Frequency { get; set; } = new();

    [JsonPropertyName("next_due_date")]
    public string? NextDueDate { get; set; }

    [JsonPropertyName("end_date")]
    public string? EndDate { get; set; }

    [JsonPropertyName("is_rolling")]
    public bool IsRolling { get; set; }

    [JsonPropertyName("notification")]
    public NotificationTriggerOptions Notification { get; set; } = new();

    [JsonPropertyName("labels")]
    public List<Label> Labels { get; set; } = new();
}

public class Frequency
{
    [JsonPropertyName("type")]
    public string Type { get; set; } = "once";

    [JsonPropertyName("on")]
    public string? On { get; set; }

    [JsonPropertyName("every")]
    public int? Every { get; set; }

    [JsonPropertyName("unit")]
    public string? Unit { get; set; }

    [JsonPropertyName("days")]
    public List<int>? Days { get; set; }

    [JsonPropertyName("months")]
    public List<int>? Months { get; set; }
}

public class NotificationTriggerOptions
{
    [JsonPropertyName("enabled")]
    public bool Enabled { get; set; }

    [JsonPropertyName("due_date")]
    public bool DueDate { get; set; }

    [JsonPropertyName("pre_due")]
    public bool PreDue { get; set; }

    [JsonPropertyName("overdue")]
    public bool Overdue { get; set; }
}

public class CreateTaskRequest
{
    [JsonPropertyName("title")]
    public string Title { get; set; } = string.Empty;

    [JsonPropertyName("next_due_date")]
    public string? NextDueDate { get; set; }

    [JsonPropertyName("end_date")]
    public string? EndDate { get; set; }

    [JsonPropertyName("is_rolling")]
    public bool IsRolling { get; set; }

    [JsonPropertyName("frequency")]
    public Frequency Frequency { get; set; } = new();

    [JsonPropertyName("notification")]
    public NotificationTriggerOptions Notification { get; set; } = new();

    [JsonPropertyName("labels")]
    public List<int> Labels { get; set; } = new();
}

public class UpdateTaskRequest
{
    [JsonPropertyName("id")]
    public int Id { get; set; }

    [JsonPropertyName("title")]
    public string Title { get; set; } = string.Empty;

    [JsonPropertyName("next_due_date")]
    public string? NextDueDate { get; set; }

    [JsonPropertyName("end_date")]
    public string? EndDate { get; set; }

    [JsonPropertyName("is_rolling")]
    public bool IsRolling { get; set; }

    [JsonPropertyName("frequency")]
    public Frequency Frequency { get; set; } = new();

    [JsonPropertyName("notification")]
    public NotificationTriggerOptions Notification { get; set; } = new();

    [JsonPropertyName("labels")]
    public List<int> Labels { get; set; } = new();
}
