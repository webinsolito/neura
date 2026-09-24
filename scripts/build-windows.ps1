$ErrorActionPreference = 'Stop'

$RepoRoot = Split-Path -Parent $PSScriptRoot
$SourceDir = Join-Path $RepoRoot 'v1_bridge'
$DistDir = Join-Path $RepoRoot 'dist'

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw 'Go is required to build NEURA. Install Go, then run this script again.'
}

New-Item -ItemType Directory -Force -Path $DistDir | Out-Null
Push-Location $SourceDir
try {
    go test ./...
    go vet ./...
    $env:CGO_ENABLED = '0'
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    go build -trimpath -ldflags '-s -w' -o (Join-Path $DistDir 'neura.exe') ./cmd/neura
} finally {
    Pop-Location
}

Write-Host "NEURA Windows build ready: $DistDir\neura.exe"
