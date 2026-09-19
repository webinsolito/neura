$ErrorActionPreference = "Stop"
$repo = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$boardPath = Join-Path $repo "automation\mission_board.json"
if (-not (Test-Path $boardPath)) { throw "mission_board_missing" }
$board = Get-Content $boardPath -Raw | ConvertFrom-Json
$now = [DateTime]::UtcNow
$validUntil = [DateTime]::Parse($board.valid_until_utc).ToUniversalTime()
if ($now -gt $validUntil) { throw "mission_board_expired" }

$mission = $board.missions | Where-Object { $_.status -eq "PENDING" -and $_.id -notlike "BASE-*" } | Select-Object -First 1
if (-not $mission) {
  Write-Host "No worker mission pending."
  exit 0
}

Write-Host "NEURA H24 mission: $($mission.id) / focus=$($board.activity_controlled)"
if ($env:NEURA_H24_WORKER_ENABLED -ne "1") {
  Write-Host "Worker disabled: supervisor-only mode. No code will be changed."
  exit 0
}
if ([string]::IsNullOrWhiteSpace($env:NEURA_H24_WORKER_CMD)) {
  throw "worker_command_not_configured"
}

$branchSafe = ($mission.id.ToLower() -replace '[^a-z0-9-]', '-')
$branch = "candidate/hourly-$([DateTime]::UtcNow.ToString('yyyyMMdd-HHmm'))-$branchSafe"
git -C $repo checkout candidate/autonomy
git -C $repo pull --ff-only origin candidate/autonomy
git -C $repo checkout -b $branch

$env:NEURA_H24_MISSION_ID = $mission.id
$env:NEURA_H24_FOCUS = $board.activity_controlled
$env:NEURA_H24_MISSION = $mission.objective
$env:NEURA_H24_BASE_SHA = $board.head_sha

& powershell -NoProfile -ExecutionPolicy Bypass -Command $env:NEURA_H24_WORKER_CMD
if ($LASTEXITCODE -ne 0) { throw "worker_failed_exit_$LASTEXITCODE" }

Push-Location (Join-Path $repo "capo_local\src")
try {
  gofmt -w (Get-ChildItem -Recurse -Filter *.go | ForEach-Object FullName)
  go vet ./...
  go test -count=1 -skip 'TestStaticRecoveryBundleMatchesCurrentRelease|TestWindowsBootstrapC927Contract|TestCleanWindowsReadinessHarnessExists' ./...
} finally { Pop-Location }

$changes = git -C $repo status --porcelain
if ([string]::IsNullOrWhiteSpace(($changes -join ""))) {
  Write-Host "Mission produced no source change; no commit created."
  exit 0
}

git -C $repo add -A
git -C $repo -c user.name="NEURA H24 Worker" -c user.email="neura-h24@users.noreply.github.com" commit -m "candidate: $($mission.id) $($board.activity_controlled)"
git -C $repo push -u origin $branch
Write-Host "Candidate pushed: $branch"
