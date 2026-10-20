[CmdletBinding()]
param([string]$RuntimeRoot='E:\pxy-runtime\Mail')
$ErrorActionPreference='Stop'
$RuntimeRoot=[IO.Path]::GetFullPath($RuntimeRoot)
foreach($required in @('.env','core\billionmail.exe','postgresql\PG_VERSION')) {
    if(!(Test-Path -LiteralPath (Join-Path $RuntimeRoot $required) -PathType Leaf)) {
        throw "Cannot register autostart for an incomplete Mail runtime: $required"
    }
}
# 自启动只依赖稳定运行目录，源码文件夹改名后仍可启动现有数据。
$operationsRoot=Join-Path $RuntimeRoot 'operations'
New-Item -ItemType Directory -Force -Path $operationsRoot|Out-Null
$manager=Join-Path $operationsRoot 'Manage-MailNative.ps1'
$candidate=Join-Path $operationsRoot ('manager-'+[guid]::NewGuid().ToString('N')+'.tmp')
try {
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'Manage-MailNative.ps1') -Destination $candidate
    Move-Item -LiteralPath $candidate -Destination $manager -Force
} finally {
    if(Test-Path -LiteralPath $candidate){Remove-Item -LiteralPath $candidate}
}
$arguments='-NoProfile -ExecutionPolicy Bypass -File "'+$manager+'" -Action Start -RuntimeRoot "'+$RuntimeRoot+'" -NoBrowser'
$action=New-ScheduledTaskAction -Execute "$env:SystemRoot\System32\WindowsPowerShell\v1.0\powershell.exe" -Argument $arguments
$triggers=@((New-ScheduledTaskTrigger -AtStartup),(New-ScheduledTaskTrigger -AtLogOn))
$principal=New-ScheduledTaskPrincipal -UserId 'SYSTEM' -LogonType ServiceAccount -RunLevel Highest
$settings=New-ScheduledTaskSettingsSet -StartWhenAvailable -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -MultipleInstances IgnoreNew -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
Register-ScheduledTask -TaskName 'PXY-MailNative-Start' -Action $action -Trigger $triggers -Principal $principal -Settings $settings -Force | Out-Null
Write-Output "Mail autostart registered: $RuntimeRoot"
