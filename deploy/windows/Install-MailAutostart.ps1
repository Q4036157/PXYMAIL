[CmdletBinding()]
param([string]$RuntimeRoot='E:\pxy-runtime\Mail')
$ErrorActionPreference='Stop'
$manager=Join-Path $PSScriptRoot 'Manage-MailNative.ps1'
$arguments='-NoProfile -ExecutionPolicy Bypass -File "'+$manager+'" -Action Start -RuntimeRoot "'+$RuntimeRoot+'" -NoBrowser'
$action=New-ScheduledTaskAction -Execute "$env:SystemRoot\System32\WindowsPowerShell\v1.0\powershell.exe" -Argument $arguments
$triggers=@((New-ScheduledTaskTrigger -AtStartup),(New-ScheduledTaskTrigger -AtLogOn))
$principal=New-ScheduledTaskPrincipal -UserId 'SYSTEM' -LogonType ServiceAccount -RunLevel Highest
$settings=New-ScheduledTaskSettingsSet -StartWhenAvailable -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -MultipleInstances IgnoreNew -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
Register-ScheduledTask -TaskName 'PXY-MailNative-Start' -Action $action -Trigger $triggers -Principal $principal -Settings $settings -Force | Out-Null
Write-Output "Mail autostart registered: $RuntimeRoot"
