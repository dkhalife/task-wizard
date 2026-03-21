using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.IdentityModel.Tokens;
using ModelContextProtocol.AspNetCore;
using ModelContextProtocol.AspNetCore.Authentication;
using TaskWizard.McpServer.Services;

var builder = WebApplication.CreateBuilder(args);

var tenantId = Environment.GetEnvironmentVariable("TW_ENTRA_TENANT_ID") ?? "";
var audience = Environment.GetEnvironmentVariable("TW_ENTRA_AUDIENCE") ?? "";
var clientId = Environment.GetEnvironmentVariable("TW_ENTRA_CLIENT_ID") ?? "";
var mcpResource = Environment.GetEnvironmentVariable("TW_MCP_RESOURCE") ?? "";

if (string.IsNullOrWhiteSpace(tenantId))
    throw new InvalidOperationException("TW_ENTRA_TENANT_ID must be set to a valid Entra tenant ID.");

if (string.IsNullOrWhiteSpace(audience))
    throw new InvalidOperationException("TW_ENTRA_AUDIENCE must be set to a valid Entra audience.");

if (string.IsNullOrWhiteSpace(clientId))
    throw new InvalidOperationException("TW_ENTRA_CLIENT_ID must be set to the Entra app registration's client ID.");

if (string.IsNullOrWhiteSpace(mcpResource))
    throw new InvalidOperationException("TW_MCP_RESOURCE must be set to the canonical URL of this MCP server (e.g. https://mcp.example.com).");

var authority = Environment.GetEnvironmentVariable("TW_ENTRA_ISSUER")
    ?? $"https://login.microsoftonline.com/{tenantId}/v2.0";
var apiUrl = Environment.GetEnvironmentVariable("TW_API_URL") ?? "http://localhost:2021";

builder.Services.AddHttpContextAccessor();

builder.Services.AddHttpClient("ApiServer", client =>
{
    client.BaseAddress = new Uri(apiUrl.TrimEnd('/') + "/");
});

builder.Services.AddScoped<ApiProxyService>();

builder.Services.AddAuthentication(options =>
{
    options.DefaultAuthenticateScheme = JwtBearerDefaults.AuthenticationScheme;
    options.DefaultChallengeScheme = McpAuthenticationDefaults.AuthenticationScheme;
})
.AddJwtBearer(options =>
{
    options.Authority = authority;
    options.TokenValidationParameters = new TokenValidationParameters
    {
        ValidateIssuer = true,
        ValidateAudience = true,
        ValidateLifetime = true,
        ValidateIssuerSigningKey = true,
        ValidAudiences = new[] { audience, clientId },
        ValidIssuer = authority,
    };
})
.AddMcp(options =>
{
    options.ResourceMetadata = new()
    {
        Resource = mcpResource,
        AuthorizationServers = { authority },
        ScopesSupported = {
            $"{audience}/User.Read",
            $"{audience}/Labels.Read",
            $"{audience}/Labels.Write",
            $"{audience}/Tasks.Read",
            $"{audience}/Tasks.Write",
        },
        BearerMethodsSupported = { "header" },
    };
});

builder.Services.AddAuthorization();

builder.Services
    .AddMcpServer()
    .WithHttpTransport()
    .WithToolsFromAssembly();

var app = builder.Build();

app.UseAuthentication();
app.UseAuthorization();
app.MapMcp().RequireAuthorization();

app.Run();
