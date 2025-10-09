using System.ComponentModel;
using System.Text.Json;
using ModelContextProtocol.Server;
using TaskWizard.McpServer.Models;
using TaskWizard.McpServer.Services;

namespace TaskWizard.McpServer.Tools;

[McpServerToolType]
public class LabelTools
{
    private readonly StubDataService _dataService;

    public LabelTools(StubDataService dataService)
    {
        _dataService = dataService;
    }

    [McpServerTool, Description("List all labels")]
    public string ListLabels()
    {
        return JsonSerializer.Serialize(_dataService.GetAllLabels(), new JsonSerializerOptions { WriteIndented = true });
    }

    [McpServerTool, Description("Get a specific label by ID")]
    public string GetLabel([Description("Label ID")] int id)
    {
        var label = _dataService.GetLabel(id);
        return label != null
            ? JsonSerializer.Serialize(label, new JsonSerializerOptions { WriteIndented = true })
            : JsonSerializer.Serialize(new { error = "Label not found" });
    }

    [McpServerTool, Description("Create a new label")]
    public string CreateLabel(
        [Description("Label name")] string name,
        [Description("Label color (hex format)")] string color = "#000000")
    {
        var createRequest = new CreateLabelRequest
        {
            Name = name,
            Color = color
        };
        var label = _dataService.CreateLabel(createRequest);
        return JsonSerializer.Serialize(label, new JsonSerializerOptions { WriteIndented = true });
    }

    [McpServerTool, Description("Update an existing label")]
    public string UpdateLabel(
        [Description("Label ID")] int id,
        [Description("Label name")] string name,
        [Description("Label color (hex format)")] string color = "#000000")
    {
        var updateRequest = new UpdateLabelRequest
        {
            Id = id,
            Name = name,
            Color = color
        };
        var label = _dataService.UpdateLabel(updateRequest);
        return label != null
            ? JsonSerializer.Serialize(label, new JsonSerializerOptions { WriteIndented = true })
            : JsonSerializer.Serialize(new { error = "Label not found" });
    }

    [McpServerTool, Description("Delete a label")]
    public string DeleteLabel([Description("Label ID")] int id)
    {
        var success = _dataService.DeleteLabel(id);
        return JsonSerializer.Serialize(new { success });
    }
}
