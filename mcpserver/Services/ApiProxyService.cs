using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace TaskWizard.McpServer.Services;

public class ApiProxyService(IHttpClientFactory httpClientFactory, IHttpContextAccessor httpContextAccessor)
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
    };

    private HttpClient CreateClient()
    {
        var client = httpClientFactory.CreateClient("ApiServer");

        var authHeader = httpContextAccessor.HttpContext?.Request.Headers["Authorization"].FirstOrDefault();
        if (!string.IsNullOrEmpty(authHeader))
        {
            client.DefaultRequestHeaders.TryAddWithoutValidation("Authorization", authHeader);
        }

        return client;
    }

    private async Task<string> SendAsync(HttpMethod method, string path, object? body = null)
    {
        var client = CreateClient();

        using var request = new HttpRequestMessage(method, path);
        if (body is not null)
        {
            var json = JsonSerializer.Serialize(body, JsonOptions);
            request.Content = new StringContent(json, Encoding.UTF8, "application/json");
        }

        var response = await client.SendAsync(request);
        var responseBody = await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
        {
            return JsonSerializer.Serialize(new
            {
                error = true,
                status = (int)response.StatusCode,
                body = responseBody,
            });
        }

        return responseBody;
    }

    // Tasks

    public Task<string> GetAllTasks() =>
        SendAsync(HttpMethod.Get, "api/v1/tasks/");

    public Task<string> GetTasksDueBefore(string before) =>
        SendAsync(HttpMethod.Get, $"api/v1/tasks/due?before={Uri.EscapeDataString(before)}");

    public Task<string> GetTasksByLabel(int labelId) =>
        SendAsync(HttpMethod.Get, $"api/v1/tasks/label/{labelId}");

    public Task<string> SearchTasksByTitle(string query) =>
        SendAsync(HttpMethod.Get, $"api/v1/tasks/search?q={Uri.EscapeDataString(query)}");

    public Task<string> GetTask(int id) =>
        SendAsync(HttpMethod.Get, $"api/v1/tasks/{id}");

    public Task<string> CreateTask(object request) =>
        SendAsync(HttpMethod.Post, "api/v1/tasks/", request);

    public Task<string> UpdateTask(object request) =>
        SendAsync(HttpMethod.Put, "api/v1/tasks/", request);

    public Task<string> DeleteTask(int id) =>
        SendAsync(HttpMethod.Delete, $"api/v1/tasks/{id}");

    public Task<string> CompleteTask(int id) =>
        SendAsync(HttpMethod.Post, $"api/v1/tasks/{id}/do");

    public async Task<string> UncompleteTask(int id)
    {
        var historyResponse = await SendAsync(HttpMethod.Get, $"api/v1/tasks/{id}/history");

        int latestHistoryId;
        try
        {
            using var doc = JsonDocument.Parse(historyResponse);
            if (doc.RootElement.ValueKind == JsonValueKind.Object &&
                doc.RootElement.TryGetProperty("error", out _))
            {
                return historyResponse;
            }

            var history = doc.RootElement.GetProperty("history");
            if (history.ValueKind != JsonValueKind.Array || history.GetArrayLength() == 0)
            {
                return JsonSerializer.Serialize(new
                {
                    error = true,
                    message = "No completion history to undo for this task",
                });
            }

            latestHistoryId = history.EnumerateArray().Max(e => e.GetProperty("id").GetInt32());
        }
        catch (Exception ex)
        {
            return JsonSerializer.Serialize(new
            {
                error = true,
                message = $"Failed to resolve task history: {ex.Message}",
            });
        }

        return await SendAsync(HttpMethod.Post, $"api/v1/tasks/{id}/undo?history_id={latestHistoryId}");
    }

    public Task<string> SkipTask(int id) =>
        SendAsync(HttpMethod.Post, $"api/v1/tasks/{id}/skip");

    // Labels

    public Task<string> GetAllLabels() =>
        SendAsync(HttpMethod.Get, "api/v1/labels");

    public Task<string> CreateLabel(object request) =>
        SendAsync(HttpMethod.Post, "api/v1/labels", request);

    public Task<string> UpdateLabel(object request) =>
        SendAsync(HttpMethod.Put, "api/v1/labels", request);

    public Task<string> DeleteLabel(int id) =>
        SendAsync(HttpMethod.Delete, $"api/v1/labels/{id}");
}
