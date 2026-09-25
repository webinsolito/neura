@echo off
setlocal
title NEURA
where powershell.exe >nul 2>&1
if errorlevel 1 (
  echo.
  echo [NEURA] Windows PowerShell non e disponibile.
  echo Apri Windows Update oppure usa un Windows 10/11 aggiornato.
  pause
  exit /b 1
)
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%~dp0Start-NEURA.ps1"
set "NEURA_EXIT=%ERRORLEVEL%"
if not "%NEURA_EXIT%"=="0" (
  echo.
  echo NEURA non e partita. Leggi il messaggio sopra: indica cosa correggere e dove trovare il log.
  pause
)
exit /b %NEURA_EXIT%
