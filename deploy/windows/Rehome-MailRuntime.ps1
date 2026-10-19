[CmdletBinding()]
param([string]$SourceRoot='D:\pxy-runtime\Mail',[string]$DestinationRoot='E:\pxy-runtime\Mail',[string]$OwnerSid)
$ErrorActionPreference='Stop'
$SourceRoot=[IO.Path]::GetFullPath($SourceRoot).TrimEnd('\')
$DestinationRoot=[IO.Path]::GetFullPath($DestinationRoot).TrimEnd('\')
if($SourceRoot -ne 'D:\pxy-runtime\Mail' -or $DestinationRoot -ne 'E:\pxy-runtime\Mail'){throw 'This migration is restricted to the requested Mail runtime paths.'}
if(!$OwnerSid -or $OwnerSid -notmatch '^S-1-5-21-'){throw 'Supply the workstation owner SID for the private runtime ACL.'}
$stamp=Get-Date -Format 'yyyyMMdd-HHmmss'
$parent=Split-Path $DestinationRoot
$archive=Join-Path $parent ('Mail-before-rehome-'+$stamp)
$retired=Join-Path (Split-Path $SourceRoot) ('Mail-inactive-after-E-move-'+$stamp)
foreach($target in @($archive,$retired)){
 $expectedParent=if($target -eq $archive){$parent}else{Split-Path $SourceRoot}
 if(!( [IO.Path]::GetFullPath($target).StartsWith($expectedParent+'\',[StringComparison]::OrdinalIgnoreCase))){throw 'Archive path is outside the requested runtime parent.'}
 if(Test-Path -LiteralPath $target){throw 'Archive path already exists.'}
}
$manager=Join-Path $PSScriptRoot 'Manage-MailNative.ps1'
$autostart=Join-Path $PSScriptRoot 'Install-MailAutostart.ps1'
$psql='C:\Program Files\PostgreSQL\16\bin\psql.exe'
$shell=Join-Path $env:SystemRoot 'System32\WindowsPowerShell\v1.0\powershell.exe'
New-Item -ItemType Directory -Path $parent -Force | Out-Null
$reportPath=Join-Path $parent ('mail-runtime-migration-'+$stamp+'.json')
$report=@{status='running';stage='preflight';source=$SourceRoot;destination=$DestinationRoot;started=(Get-Date -Format o);previous_e_backup=$archive;inactive_d_backup=$retired}
function Save-Report { $report|ConvertTo-Json -Depth 8|Set-Content -LiteralPath $reportPath -Encoding UTF8 }
function Read-Settings([string]$Root){
 $values=@{};foreach($line in [IO.File]::ReadAllLines((Join-Path $Root '.env'))){if($line -match '^([^#=]+)=(.*)$'){$values[$matches[1]]=$matches[2]}};return $values
}
function Invoke-Sql([string]$SQL){
 $result=& $psql -h 127.0.0.1 -p 25433 -U billionmail -d billionmail -v ON_ERROR_STOP=1 -tAc $SQL
 if($LASTEXITCODE -ne 0){throw 'Mail database query failed.'};return ([string]($result -join "`n")).Trim()
}
function Snapshot {
 $counts=Invoke-Sql "SELECT json_build_object('smtp',(SELECT count(*) FROM bm_relay_config),'templates',(SELECT count(*) FROM email_templates),'contacts',(SELECT count(*) FROM bm_contacts),'tasks',(SELECT count(*) FROM email_tasks),'accepted',(SELECT count(*) FROM recipient_info WHERE is_sent=1),'inbox',(SELECT count(*) FROM bm_native_inbox_messages),'cached_bodies',(SELECT count(*) FROM bm_native_inbox_messages WHERE body_cached),'inbox_accounts',(SELECT count(*) FROM bm_native_inbox_accounts))::text"
 return $counts|ConvertFrom-Json
}
function Invoke-Manager([string]$Action,[string]$Root){
 $args='-NoProfile -ExecutionPolicy Bypass -File "'+$manager+'" -Action '+$Action+' -RuntimeRoot "'+$Root+'" -NoBrowser'
 $operation=Start-Process -FilePath $shell -ArgumentList $args -WindowStyle Hidden -PassThru
 if(!$operation.WaitForExit(120000)){throw "Mail $Action process did not finish for $Root."}
 if($operation.ExitCode -ne 0){throw "Mail $Action failed for $Root; inspect last-operation.json."}
}
$sourceStopped=$false
$destinationStarted=$false
try {
 Save-Report
 $settings=Read-Settings $SourceRoot;$env:PGPASSWORD=$settings.DBPASS
 $dataDir=Invoke-Sql 'SHOW data_directory'
 if($dataDir.Replace('/','\') -ne (Join-Path $SourceRoot 'postgresql')){throw 'The active database is not the requested D drive source.'}
 $running=Invoke-Sql 'SELECT count(*) FROM email_tasks WHERE pause=0 AND task_process<>2'
 if([int]$running -ne 0){throw 'An outbound task is active; migrate after that task is paused.'}
 $task=Get-ScheduledTask -TaskName 'PXY-MailNative-Start' -ErrorAction SilentlyContinue
 if($task){Disable-ScheduledTask -TaskName $task.TaskName|Out-Null}
 $report.stage='stopping-source';Save-Report
 $sourceApp=Get-Process -Id ([int][IO.File]::ReadAllText((Join-Path $SourceRoot 'application.pid')).Trim()) -ErrorAction SilentlyContinue
 if($sourceApp){
  if($sourceApp.Path -ne (Join-Path $SourceRoot 'core\billionmail.exe')){throw 'Source process ownership mismatch.'}
  Stop-Process -Id $sourceApp.Id -Force -Confirm:$false;$sourceApp.WaitForExit(30000)|Out-Null
  $remaining=Get-Process -Id $sourceApp.Id -ErrorAction SilentlyContinue
  if($remaining -and !$remaining.HasExited){throw 'The D drive application did not stop.'}
 }
 $sourceStopped=$true
 $report.before=Snapshot
 Invoke-Manager 'Stop' $SourceRoot;$sourceStopped=$true
 $report.stage='copying-data';Save-Report
 if(Test-Path -LiteralPath $DestinationRoot){Move-Item -LiteralPath $DestinationRoot -Destination $archive}
 New-Item -ItemType Directory -Path $DestinationRoot|Out-Null
 & icacls.exe $DestinationRoot /inheritance:r /grant:r "*$($OwnerSid):(OI)(CI)F" '*S-1-5-18:(OI)(CI)F' '*S-1-5-32-544:(OI)(CI)F' > $null
 if($LASTEXITCODE -ne 0){throw 'Cannot secure the E drive runtime.'}
 & robocopy.exe $SourceRoot $DestinationRoot /E /COPY:DAT /DCOPY:DAT /R:1 /W:1 /XJ /XD (Join-Path $SourceRoot 'browser-check') /XF 'application.pid' 'redis.pid' 'operation.lock' /NFL /NDL /NJH /NJS /NP *> (Join-Path $parent ('mail-migration-'+$stamp+'-copy.log'))
 if($LASTEXITCODE -ge 8){throw 'Runtime data copy failed.'}
 foreach($leaf in @('.env','login.txt')){
  if((Get-FileHash -LiteralPath (Join-Path $SourceRoot $leaf)).Hash -ne (Get-FileHash -LiteralPath (Join-Path $DestinationRoot $leaf)).Hash){throw "Configuration verification failed: $leaf"}
 }
 foreach($file in Get-ChildItem -LiteralPath $DestinationRoot -File | Where-Object {$_.Extension -in @('.ps1','.cjs')}){
  $text=[IO.File]::ReadAllText($file.FullName).Replace($SourceRoot,$DestinationRoot).Replace($SourceRoot.Replace('\','/'),$DestinationRoot.Replace('\','/'))
  [IO.File]::WriteAllText($file.FullName,$text,(New-Object Text.UTF8Encoding($false)))
 }
 $report.stage='starting-on-e';Save-Report
 Invoke-Manager 'Start' $DestinationRoot;$destinationStarted=$true
 $newDataDir=Invoke-Sql 'SHOW data_directory'
 if($newDataDir.Replace('/','\') -ne (Join-Path $DestinationRoot 'postgresql')){throw 'Database did not start from E drive.'}
 Invoke-Sql "UPDATE bm_options SET value='E:/pxy-runtime/Mail/configuration/unsubscribe.json' WHERE name='native_legacy_unsubscribe_config'"|Out-Null
 $report.after=Snapshot
 foreach($property in $report.before.PSObject.Properties){
  if($property.Value -ne $report.after.($property.Name)){throw "Record count changed during migration: $($property.Name)"}
 }
 $report.stage='registering-autostart';Save-Report
 & $autostart -RuntimeRoot $DestinationRoot
 Enable-ScheduledTask -TaskName 'PXY-MailNative-Start'|Out-Null
 $task=Get-ScheduledTask -TaskName 'PXY-MailNative-Start'
 $report.autostart=@{state=[string]$task.State;arguments=[string]$task.Actions.Arguments}
 if($task.Actions.Arguments -notlike '*E:\pxy-runtime\Mail*'){throw 'Autostart still targets a different runtime.'}
 $pidPath=Join-Path $DestinationRoot 'application.pid'
 $app=Get-Process -Id ([int][IO.File]::ReadAllText($pidPath).Trim())
 if($app.Path -ne (Join-Path $DestinationRoot 'core\billionmail.exe')){throw 'The running executable is not on E drive.'}
 $report.executable=$app.Path;$report.data_directory=$newDataDir
 $report.stage='retiring-d-runtime';Save-Report
 [IO.File]::WriteAllText((Join-Path $SourceRoot '已迁移到E盘.txt'),"当前运行目录：E:\pxy-runtime\Mail。此目录仅为迁移前备份，不再启动。`r`n源码和 BAT：D:\x1\x2\Mail。",(New-Object Text.UTF8Encoding($false)))
 Move-Item -LiteralPath $SourceRoot -Destination $retired
 $report.status='ok';$report.stage='complete';$report.completed=(Get-Date -Format o);Save-Report
 Copy-Item -LiteralPath $reportPath -Destination (Join-Path $DestinationRoot 'runtime-migration-result.json')
} catch {
 $report.status='failed';$report.error=$_.Exception.Message
 if($sourceStopped -and (Test-Path -LiteralPath $SourceRoot)){
  try {
   if($destinationStarted){Invoke-Manager 'Stop' $DestinationRoot}
   Invoke-Manager 'Start' $SourceRoot
   & $autostart -RuntimeRoot $SourceRoot
   Enable-ScheduledTask -TaskName 'PXY-MailNative-Start'|Out-Null
   $report.rollback='D source restored'
  } catch {$report.rollback_error=$_.Exception.Message}
 }
 Save-Report;exit 1
} finally {Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue}
