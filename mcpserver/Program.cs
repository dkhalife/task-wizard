using ModelContextProtocol.Server;
using TaskWizard.McpServer.Services;

var builder = WebApplication.CreateBuilder(args);

// Register services
builder.Services.AddSingleton<StubDataService>();

// Configure MCP server with HTTP transport
builder.Services
    .AddMcpServer()
    .WithHttpTransport()
    .WithToolsFromAssembly();

var app = builder.Build();

app.MapMcp();

app.Run("http://localhost:3001");
