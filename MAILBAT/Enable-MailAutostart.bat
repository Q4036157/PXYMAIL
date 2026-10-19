@echo off
title %~nx0
if /I "%PXY_AGENT_CLI%"=="1" goto run
if not defined PXY_FIXED_TAB_TITLE (
    wt.exe -w -1 new-tab --title "%~nx0" --suppressApplicationTitle cmd.exe /d /c "set PXY_FIXED_TAB_TITLE=1&& call ""%~f0"" %*"
    exit /b
)
:run
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0..\deploy\windows\Install-MailAutostart.ps1" %*
set "operation_code=%ERRORLEVEL%"
if /I not "%PXY_AGENT_CLI%"=="1" pause
exit /b %operation_code%
