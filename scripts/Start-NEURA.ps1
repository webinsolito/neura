param(
    [string]$Listen = '127.0.0.1:8765',
    [string]$Workspace = '',
    [string]$DataDir = '',
    [switch]$NoBrowser,
    [switch]$PreflightOnly,
    [switch]$SmokeTest
)

$ErrorActionPreference = 'Stop'

$PackageDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Exe = Join-Path $PackageDir 'neura.exe'
$LocalAppData = [Environment]::GetFolderPath('LocalApplicationData')
if ([string]::IsNullOrWhiteSpace($LocalAppData)) { $LocalAppData = $env:TEMP }
$NeuraHome = Join-Path $LocalAppData 'NEURA'
$LogDir = Join-Path $NeuraHome 'logs'
$LaunchLog = Join-Path $LogDir 'launcher.log'

function Write-LaunchLog([string]$Message) {
    New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
    $line = '{0:o} {1}' -f (Get-Date).ToUniversalTime(), $Message
    Add-Content -LiteralPath $LaunchLog -Value $line -Encoding UTF8
}

function Assert-WritableDirectory([string]$Path, [string]$Label) {
    New-Item -ItemType Directory -Force -Path $Path | Out-Null
    $probe = Join-Path $Path ('.neura-write-test-' + [Guid]::NewGuid().ToString('N') + '.tmp')
    try {
        [IO.File]::WriteAllText($probe, 'ok')
    } catch {
        throw "$Label non scrivibile: $Path"
    } finally {
        Remove-Item -LiteralPath $probe -Force -ErrorAction SilentlyContinue
    }
}

function Get-ListenParts([string]$Value) {
    if ($Value -match '^127\.0\.0\.1:(\d{1,5})$') {
        return @{ Host = '127.0.0.1'; Port = [int]$Matches[1]; Address = [Net.IPAddress]::Loopback }
    }
    if ($Value -match '^\[::1\]:(\d{1,5})$') {
        return @{ Host = '::1'; Port = [int]$Matches[1]; Address = [Net.IPAddress]::IPv6Loopback }
    }
    throw 'Indirizzo non valido: usa solo 127.0.0.1:PORTA oppure [::1]:PORTA.'
}

function Test-NeuraHealth([string]$Uri) {
    try {
        $health = Invoke-RestMethod -Uri $Uri -TimeoutSec 1
        return ($health.ok -eq $true)
    } catch {
        return $false
    }
}

$proc = $null
try {
    if (-not (Test-Path -LiteralPath $Exe -PathType Leaf)) {
        throw "neura.exe non trovato nel pacchetto: $Exe. Riscarica o riestrai NEURA-Windows."
    }
    if ((Get-Item -LiteralPath $Exe).Length -lt 1024) {
        throw 'neura.exe sembra incompleto o danneggiato. Riscarica il pacchetto.'
    }

    $listenParts = Get-ListenParts $Listen
    if ($listenParts.Port -lt 1 -or $listenParts.Port -gt 65535) {
        throw 'Porta non valida: scegli un valore tra 1 e 65535.'
    }

    if ([string]::IsNullOrWhiteSpace($DataDir)) {
        $DataDir = Join-Path $NeuraHome 'data'
    }
    if ([string]::IsNullOrWhiteSpace($Workspace)) {
        $documents = [Environment]::GetFolderPath('MyDocuments')
        if ([string]::IsNullOrWhiteSpace($documents)) { $documents = $env:USERPROFILE }
        $Workspace = Join-Path $documents 'NEURA Workspace'
    }

    Assert-WritableDirectory $DataDir 'Cartella dati'
    Assert-WritableDirectory $Workspace 'Workspace'
    Assert-WritableDirectory $LogDir 'Cartella log'

    $healthUri = "http://$Listen/health"
    $uiUri = "http://$Listen/"

    if (Test-NeuraHealth $healthUri) {
        if ($SmokeTest) { throw "Una NEURA è già attiva su $Listen; smoke test non isolato." }
        Write-Host "NEURA è già attiva su $uiUri"
        Write-LaunchLog "Existing healthy NEURA detected at $Listen"
        if (-not $NoBrowser) { Start-Process $uiUri }
        return
    }

    $listener = $null
    try {
        $listener = [Net.Sockets.TcpListener]::new($listenParts.Address, $listenParts.Port)
        $listener.Start()
    } catch {
        throw "La porta $($listenParts.Port) è occupata da un'altra applicazione. Chiudila oppure avvia NEURA con un'altra porta."
    } finally {
        if ($listener) { $listener.Stop() }
    }

    Write-LaunchLog "Preflight PASS listen=$Listen data=$DataDir workspace=$Workspace"
    if ($PreflightOnly) {
        Write-Host 'NEURA_PREFLIGHT_OK'
        return
    }

    $stdout = Join-Path $LogDir 'neura.stdout.log'
    $stderr = Join-Path $LogDir 'neura.stderr.log'
    $argLine = '-listen "{0}" -data "{1}" -workspace "{2}"' -f $Listen, $DataDir, $Workspace
    $proc = Start-Process -FilePath $Exe -ArgumentList $argLine -PassThru -WindowStyle Hidden -RedirectStandardOutput $stdout -RedirectStandardError $stderr

    $ready = $false
    for ($i = 0; $i -lt 40; $i++) {
        Start-Sleep -Milliseconds 250
        if ($proc.HasExited) { break }
        if (Test-NeuraHealth $healthUri) { $ready = $true; break }
    }

    if (-not $ready) {
        $detail = ''
        if (Test-Path -LiteralPath $stderr) {
            $detail = (Get-Content -LiteralPath $stderr -Tail 8 -ErrorAction SilentlyContinue) -join ' '
        }
        if ($proc -and -not $proc.HasExited) { Stop-Process -Id $proc.Id -Force }
        throw "NEURA non ha superato il controllo di avvio. Controlla $stderr. $detail"
    }

    Write-LaunchLog "Startup PASS pid=$($proc.Id) listen=$Listen"
    if ($SmokeTest) {
        Write-Host 'NEURA_SMOKE_OK'
        if ($proc -and -not $proc.HasExited) { Stop-Process -Id $proc.Id -Force }
        return
    }

    Write-Host "NEURA è pronta: $uiUri"
    Write-Host "Workspace: $Workspace"
    Write-Host "Log: $LogDir"
    if (-not $NoBrowser) { Start-Process $uiUri }
} catch {
    $message = $_.Exception.Message
    try { Write-LaunchLog "ERROR $message" } catch {}
    if ($proc -and -not $proc.HasExited) { Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue }
    [Console]::Error.WriteLine('')
    [Console]::Error.WriteLine('[NEURA] Avvio non riuscito.')
    [Console]::Error.WriteLine($message)
    [Console]::Error.WriteLine("Log launcher: $LaunchLog")
    exit 1
}
