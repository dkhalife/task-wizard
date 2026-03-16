using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.IdentityModel.Tokens;
using ModelContextProtocol.AspNetCore;
using ModelContextProtocol.AspNetCore.Authentication;
using TaskWizard.McpServer.Services;

var builder = WebApplication.CreateBuilder(args);

var tenantId = Environment.GetEnvironmentVariable("TW_ENTRA_TENANT_ID") ?? "";
var audience = Environment.GetEnvironmentVariable("TW_ENTRA_AUDIENCE") ?? "";
var issuer = Environment.GetEnvironmentVariable("TW_ENTRA_ISSUER")
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
    options.Authority = $"https://login.microsoftonline.com/{tenantId}/v2.0";
    options.TokenValidationParameters = new TokenValidationParameters
    {
        ValidateIssuer = true,
        ValidateAudience = true,
        ValidateLifetime = true,
        ValidateIssuerSigningKey = true,
        ValidAudience = audience,
        ValidIssuer = issuer,
    };
})
.AddMcp(options =>
{
    options.ResourceMetadata = new()
    {
        AuthorizationServers = { $"https://login.microsoftonline.com/{tenantId}/v2.0" },
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
