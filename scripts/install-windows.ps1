param(
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA 'NEURA'),
    [switch]$NoBuild,
    [switch]$NoShortcut
)

$ErrorActionPreference = 'Stop'
$RepoRoot = Split-Path -Parent $PSScriptRoot
$DistExe = Join-Path $RepoRoot 'dist\neura.exe'

if (-not $NoBuild) {
    & (Join-Path $PSScriptRoot 'build-windows.ps1')
}
if (-not (Test-Path -LiteralPath $DistExe -PathType Leaf)) {
    throw "NEURA executable not found at $DistExe. Build it first or omit -NoBuild."
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$TargetExe = Join-Path $InstallDir 'neura.exe'
$Launcher = Join-Path $InstallDir 'Start NEURA.cmd'
Copy-Item -LiteralPath $DistExe -Destination $TargetExe -Force

$launcherBody = @"
@echo off
setlocal
cd /d "%USERPROFILE%"
start "NEURA" /min "$TargetExe" -workspace "%USERPROFILE%"
timeout /t 2 /nobreak >nul
start "" "http://127.0.0.1:8765/health"
endlocal
"@
Set-Content -LiteralPath $Launcher -Value $launcherBody -Encoding ASCII

if (-not $NoShortcut) {
    $Programs = [Environment]::GetFolderPath('Programs')
    $ShortcutPath = Join-Path $Programs 'NEURA.lnk'
    $Shell = New-Object -ComObject WScript.Shell
    $Shortcut = $Shell.CreateShortcut($ShortcutPath)
    $Shortcut.TargetPath = $Launcher
    $Shortcut.WorkingDirectory = $InstallDir
    $Shortcut.Description = 'Start NEURA local agent'
    $Shortcut.Save()
}

$Installed = Get-Item -LiteralPath $TargetExe
if ($Installed.Length -le 0) { throw 'Installed NEURA executable is empty.' }
Write-Host "NEURA installed for the current user in: $InstallDir"
Write-Host "Start it with: $Launcher"
