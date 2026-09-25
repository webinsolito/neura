$ErrorActionPreference = 'Stop'

$RepoRoot = Split-Path -Parent $PSScriptRoot
$SourceDir = Join-Path $RepoRoot 'v1_bridge'
$DistDir = Join-Path $RepoRoot 'dist'
$PackageDir = Join-Path $DistDir 'NEURA-Windows'

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw 'Go is required only to build NEURA from source. The published NEURA-Windows package does not require Go.'
}

if (Test-Path -LiteralPath $PackageDir) {
    Remove-Item -LiteralPath $PackageDir -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $PackageDir | Out-Null

Push-Location $SourceDir
try {
    go test ./...
    go vet ./...
    $env:CGO_ENABLED = '0'
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    go build -trimpath -ldflags '-s -w' -o (Join-Path $PackageDir 'neura.exe') .
} finally {
    Pop-Location
}

Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'Start-NEURA.ps1') -Destination (Join-Path $PackageDir 'Start-NEURA.ps1')
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'Launch-NEURA.cmd') -Destination (Join-Path $PackageDir 'Launch-NEURA.cmd')
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'NEURA-Windows-LEGGIMI.txt') -Destination (Join-Path $PackageDir 'LEGGIMI.txt')

$Exe = Join-Path $PackageDir 'neura.exe'
if (-not (Test-Path -LiteralPath $Exe -PathType Leaf)) {
    throw 'NEURA executable was not produced.'
}
$hash = (Get-FileHash -LiteralPath $Exe -Algorithm SHA256).Hash
$commit = 'unknown'
try {
    $commit = (& git -C $RepoRoot rev-parse HEAD 2>$null).Trim()
} catch {}
$manifest = @(
    'NEURA Windows Package'
    'architecture=windows-amd64'
    "commit=$commit"
    "neura.exe.sha256=$hash"
    'launcher=Launch-NEURA.cmd'
    'preflight=Start-NEURA.ps1'
) -join [Environment]::NewLine
Set-Content -LiteralPath (Join-Path $PackageDir 'PACKAGE-MANIFEST.txt') -Value $manifest -Encoding UTF8

Write-Host "NEURA Windows package ready: $PackageDir"
Write-Host "SHA256 neura.exe: $hash"
