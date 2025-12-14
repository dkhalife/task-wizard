# Generate Go protobuf and gRPC files
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc
#
# Install dependencies:
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

$ErrorActionPreference = "Stop"

$ProtoDir = $PSScriptRoot
$OutputDir = Join-Path $ProtoDir ".." "apiserver" "internal" "grpc"

# Create output directory if it doesn't exist
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

Write-Host "Generating Go protobuf files..." -ForegroundColor Yellow

# Get all proto files
$ProtoFiles = Get-ChildItem -Path $ProtoDir -Filter "*.proto" | Select-Object -ExpandProperty Name

# Generate Go files
protoc `
    --proto_path=$ProtoDir `
    --go_out=$OutputDir `
    --go_opt=paths=source_relative `
    --go-grpc_out=$OutputDir `
    --go-grpc_opt=paths=source_relative `
    $ProtoFiles

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error generating Go files" -ForegroundColor Red
    exit 1
}

Write-Host "Go protobuf files generated successfully at: $OutputDir" -ForegroundColor Green
