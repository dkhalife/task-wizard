# Stage 1: Build the application
FROM mcr.microsoft.com/dotnet/sdk:9.0 AS builder

WORKDIR /usr/src/app

# Copy project file and restore dependencies
COPY TaskWizardMcpServer.csproj ./
RUN dotnet restore

# Copy source code and build
COPY . ./
RUN dotnet publish -c Release -o /app/publish --no-restore

# Stage 2: Create a smaller runtime image
FROM mcr.microsoft.com/dotnet/aspnet:9.0

WORKDIR /app

RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/publish .

ENV ASPNETCORE_URLS="http://+:3001"

EXPOSE 3001

ENTRYPOINT ["dotnet", "TaskWizardMcpServer.dll"]
