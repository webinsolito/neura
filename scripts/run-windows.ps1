$ErrorActionPreference = 'Stop'

$RepoRoot = Split-Path -Parent $PSScriptRoot
$Exe = Join-Path $RepoRoot 'dist\neura.exe'

if (-not (Test-Path $Exe)) {
    & (Join-Path $PSScriptRoot 'build-windows.ps1')
}

if (-not (Test-Path $Exe)) {
    throw 'NEURA executable was not produced.'
}

& $Exe @args
exit $LASTEXITCODE
