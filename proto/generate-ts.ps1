$PROTO_DIR = $PSScriptRoot
$OUT_DIR = Join-Path -Path $PSScriptRoot -ChildPath ".." | Join-Path -ChildPath "frontend" | Join-Path -ChildPath "src" | Join-Path -ChildPath "grpc"
$FRONTEND_DIR = Join-Path -Path $PSScriptRoot -ChildPath ".." | Join-Path -ChildPath "frontend"
$NODE_MODULES_BIN = Join-Path -Path $FRONTEND_DIR -ChildPath "node_modules" | Join-Path -ChildPath ".bin"

# Detect the correct protoc-gen-ts_proto executable based on OS
if ($IsWindows -or $PSVersionTable.Platform -eq 'Win32NT') {
    $PROTOC_GEN_TS_PROTO = Join-Path -Path $NODE_MODULES_BIN -ChildPath "protoc-gen-ts_proto.cmd"
} else {
    $PROTOC_GEN_TS_PROTO = Join-Path -Path $NODE_MODULES_BIN -ChildPath "protoc-gen-ts_proto"
}

$PROTO_FILES = @(
  (Join-Path -Path $PROTO_DIR -ChildPath "common.proto"),
  (Join-Path -Path $PROTO_DIR -ChildPath "label.proto"),
  (Join-Path -Path $PROTO_DIR -ChildPath "task.proto"),
  (Join-Path -Path $PROTO_DIR -ChildPath "user.proto"),
  (Join-Path -Path $PROTO_DIR -ChildPath "task_service.proto")
)

Write-Host "Creating output directory..."
New-Item -ItemType Directory -Force -Path $OUT_DIR | Out-Null

# Generate clean TypeScript code with ts-proto
Write-Host "Generating TypeScript code with ts-proto..."
& protoc `
  --proto_path="$PROTO_DIR" `
  --plugin="protoc-gen-ts_proto=$PROTOC_GEN_TS_PROTO" `
  --ts_proto_out="$OUT_DIR" `
  --ts_proto_opt=esModuleInterop=true `
  --ts_proto_opt=outputServices=generic-definitions,useExactTypes=false,env=browser,exportCommonSymbols=false `
  $PROTO_FILES

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Successfully generated TypeScript code"
} else {
    Write-Host "✗ Failed to generate TypeScript code"
    exit 1
}
