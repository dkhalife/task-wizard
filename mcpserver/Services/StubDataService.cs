using TaskWizard.McpServer.Models;

namespace TaskWizard.McpServer.Services;

public class StubDataService
{
    private readonly List<Label> _labels = new()
    {
        new Label { Id = 1, Name = "Work", Color = "#FF5733" },
        new Label { Id = 2, Name = "Personal", Color = "#33FF57" },
        new Label { Id = 3, Name = "Urgent", Color = "#3357FF" }
    };

    private readonly List<Models.Task> _tasks = new()
    {
        new Models.Task
        {
            Id = 1,
            Title = "Complete project documentation",
            NextDueDate = DateTime.UtcNow.AddDays(3),
            Frequency = "once",
            IsRolling = false,
            Labels = new List<Label>()
        },
        new Models.Task
        {
            Id = 2,
            Title = "Weekly team meeting",
            NextDueDate = DateTime.UtcNow.AddDays(1),
            Frequency = "weekly",
            IsRolling = false,
            Labels = new List<Label>()
        }
    };

    private int _nextTaskId = 3;
    private int _nextLabelId = 4;

    // Task CRUD operations
    public List<Models.Task> GetAllTasks() => _tasks.ToList();

    public Models.Task? GetTask(int id) => _tasks.FirstOrDefault(t => t.Id == id);

    public Models.Task CreateTask(CreateTaskRequest request)
    {
        var task = new Models.Task
        {
            Id = _nextTaskId++,
            Title = request.Title,
            NextDueDate = request.NextDueDate,
            EndDate = request.EndDate,
            Frequency = request.Frequency,
            IsRolling = request.IsRolling,
            Labels = request.Labels.Select(id => _labels.FirstOrDefault(l => l.Id == id))
                           .Where(l => l != null)
                           .Cast<Label>()
                           .ToList()
        };
        _tasks.Add(task);
        return task;
    }

    public Models.Task? UpdateTask(UpdateTaskRequest request)
    {
        var task = _tasks.FirstOrDefault(t => t.Id == request.Id);
        if (task == null) return null;

        task.Title = request.Title;
        task.NextDueDate = request.NextDueDate;
        task.EndDate = request.EndDate;
        task.Frequency = request.Frequency;
        task.IsRolling = request.IsRolling;
        task.Labels = request.Labels.Select(id => _labels.FirstOrDefault(l => l.Id == id))
                           .Where(l => l != null)
                           .Cast<Label>()
                           .ToList();
        return task;
    }

    public bool DeleteTask(int id)
    {
        var task = _tasks.FirstOrDefault(t => t.Id == id);
        if (task == null) return false;
        _tasks.Remove(task);
        return true;
    }

    // Label CRUD operations
    public List<Label> GetAllLabels() => _labels.ToList();

    public Label? GetLabel(int id) => _labels.FirstOrDefault(l => l.Id == id);

    public Label CreateLabel(CreateLabelRequest request)
    {
        var label = new Label
        {
            Id = _nextLabelId++,
            Name = request.Name,
            Color = request.Color
        };
        _labels.Add(label);
        return label;
    }

    public Label? UpdateLabel(UpdateLabelRequest request)
    {
        var label = _labels.FirstOrDefault(l => l.Id == request.Id);
        if (label == null) return null;

        label.Name = request.Name;
        label.Color = request.Color;
        return label;
    }

    public bool DeleteLabel(int id)
    {
        var label = _labels.FirstOrDefault(l => l.Id == id);
        if (label == null) return false;
        _labels.Remove(label);
        return true;
    }
}
