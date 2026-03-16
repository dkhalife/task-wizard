using System.ComponentModel;
using ModelContextProtocol.Server;
using TaskWizard.McpServer.Models;
using TaskWizard.McpServer.Services;

namespace TaskWizard.McpServer.Tools;

[McpServerToolType]
public class LabelTools(ApiProxyService api)
{
    [McpServerTool, Description("List all labels")]
    public Task<string> ListLabels() =>
        api.GetAllLabels();

    [McpServerTool, Description("Create a new label")]
    public Task<string> CreateLabel(
        [Description("Label name")] string name,
        [Description("Label color (hex format, e.g. #FF5733)")] string color = "#000000") =>
        api.CreateLabel(new CreateLabelRequest { Name = name, Color = color });

    [McpServerTool, Description("Update an existing label")]
    public Task<string> UpdateLabel(
        [Description("Label ID")] int id,
        [Description("Label name")] string name,
        [Description("Label color (hex format, e.g. #FF5733)")] string color = "#000000") =>
        api.UpdateLabel(new UpdateLabelRequest { Id = id, Name = name, Color = color });

    [McpServerTool, Description("Delete a label")]
    public Task<string> DeleteLabel([Description("Label ID")] int id) =>
        api.DeleteLabel(id);
}
