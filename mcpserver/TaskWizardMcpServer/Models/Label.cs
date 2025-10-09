namespace TaskWizard.McpServer.Models;

public class Label
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public string Color { get; set; } = string.Empty;
}

public class CreateLabelRequest
{
    public string Name { get; set; } = string.Empty;
    public string Color { get; set; } = string.Empty;
}

public class UpdateLabelRequest
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public string Color { get; set; } = string.Empty;
}
