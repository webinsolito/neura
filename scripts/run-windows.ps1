$ErrorActionPreference = 'Stop'

$RepoRoot = Split-Path -Parent $PSScriptRoot
$PackageDir = Join-Path $RepoRoot 'dist\NEURA-Windows'
$Exe = Join-Path $PackageDir 'neura.exe'
$Launcher = Join-Path $PackageDir 'Start-NEURA.ps1'

if (-not (Test-Path -LiteralPath $Exe -PathType Leaf) -or -not (Test-Path -LiteralPath $Launcher -PathType Leaf)) {
    & (Join-Path $PSScriptRoot 'build-windows.ps1')
}

if (-not (Test-Path -LiteralPath $Exe -PathType Leaf)) {
    throw 'NEURA executable was not produced.'
}
if (-not (Test-Path -LiteralPath $Launcher -PathType Leaf)) {
    throw 'NEURA launcher was not produced.'
}

& $Launcher @args
exit $LASTEXITCODE
