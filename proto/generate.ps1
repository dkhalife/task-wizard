# Generate protobuf files for all languages
# This script executes all generate-*.ps1 scripts in the proto directory

Get-ChildItem -Path $PSScriptRoot -Filter "generate-*.ps1" | ForEach-Object {
    Write-Host "Executing: $($_.Name)" -ForegroundColor Cyan
    & $_.FullName
}

Write-Host "All generation scripts completed!" -ForegroundColor Green
