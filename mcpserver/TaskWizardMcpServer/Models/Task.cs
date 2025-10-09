namespace TaskWizardMcpServer.Models;

public class Task
{
    public int Id { get; set; }
    public string Title { get; set; } = string.Empty;
    public DateTime? NextDueDate { get; set; }
    public DateTime? EndDate { get; set; }
    public string Frequency { get; set; } = "once";
    public bool IsRolling { get; set; }
    public List<Label> Labels { get; set; } = new();
}

public class CreateTaskRequest
{
    public string Title { get; set; } = string.Empty;
    public DateTime? NextDueDate { get; set; }
    public DateTime? EndDate { get; set; }
    public string Frequency { get; set; } = "once";
    public bool IsRolling { get; set; }
    public List<int> Labels { get; set; } = new();
}

public class UpdateTaskRequest
{
    public int Id { get; set; }
    public string Title { get; set; } = string.Empty;
    public DateTime? NextDueDate { get; set; }
    public DateTime? EndDate { get; set; }
    public string Frequency { get; set; } = "once";
    public bool IsRolling { get; set; }
    public List<int> Labels { get; set; } = new();
}
