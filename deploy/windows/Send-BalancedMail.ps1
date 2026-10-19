param(
 [Parameter(Mandatory=$true)][string]$CsvPath,
 [Parameter(Mandatory=$true)][string]$FinalRecipient,
 [int]$TemplateId=86,
 [Parameter(Mandatory=$true)][string]$BatchKey,
 [switch]$OnlyCustomers
)
$ErrorActionPreference='Stop'
$root='E:\pxy-runtime\Mail'
if($BatchKey -notmatch '^[a-zA-Z0-9-]+$'){throw 'Invalid batch key'}
$checkpoint=Join-Path $root ('migration-data\'+$BatchKey+'.json')
if(Test-Path $checkpoint){throw 'Batch exists; inspect before resuming to prevent duplicates'}
$cfg=@{};Get-Content -Encoding UTF8 (Join-Path $root '.env')|ForEach-Object{$p=$_ -split '=',2;if($p.Count -eq 2){$cfg[$p[0]]=$p[1]}}
$l=Invoke-RestMethod 'http://127.0.0.1:3040/api/login' -Method Post -ContentType application/json -Body (@{username=$cfg.ADMIN_USERNAME;password=$cfg.ADMIN_PASSWORD}|ConvertTo-Json)
if($l.code -ne 0){throw 'Login failed'}
$headers=@{Authorization=('Bearer '+$l.data.token)}
function Api($Path,$Body){$r=Invoke-RestMethod ('http://127.0.0.1:3040/api'+$Path) -Headers $headers -Method Post -ContentType application/json -Body ([Text.Encoding]::UTF8.GetBytes(($Body|ConvertTo-Json -Depth 8)));if($r.code -ne 0){throw ($Path+': '+$r.msg)};return $r}
function Db($Sql){$env:PGPASSWORD=$cfg.DBPASS;try{$r=& 'C:\Program Files\PostgreSQL\16\bin\psql.exe' -h $cfg.DBHOST -p $cfg.DBPORT -U $cfg.DBUSER -d $cfg.DBNAME -At -v ON_ERROR_STOP=1 -c $Sql;if($LASTEXITCODE -ne 0){throw 'DB failed'};return $r}finally{Remove-Item Env:PGPASSWORD}}
$entries=@(Import-Csv -Encoding UTF8 -LiteralPath $CsvPath)
if(!$OnlyCustomers){$entries += [pscustomobject]@{email=$FinalRecipient;attributes=(@{csv_position='2147483647';source='user_added_test'}|ConvertTo-Json -Compress)}}
if(@($entries.email|Select-Object -Unique).Count -ne $entries.Count){throw 'Duplicate recipients'}
$unsub=Get-Content -Encoding UTF8 'E:\pxy-runtime\Mail\configuration\unsubscribe.json' -Raw|ConvertFrom-Json
foreach($entry in $entries){
 if($entry.email -notmatch '^[a-zA-Z0-9._+%-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'){throw 'Invalid recipient'}
 if($entry.email -eq $FinalRecipient){
  if(@(Db ("SELECT id FROM recipient_info WHERE lower(recipient)='"+$entry.email+"' AND is_sent=2;")).Count -gt 0){throw 'Final test recipient already in flight'}
 }elseif(@(Db ("SELECT id FROM recipient_info WHERE lower(recipient)='"+$entry.email+"' AND is_sent IN(1,2,3);")).Count -gt 0){throw ('Prior sending history: '+$entry.email)}
 $check=Invoke-RestMethod ($unsub.internal_url.TrimEnd('/')+'/internal/check') -NoProxy -Method Post -TimeoutSec 8 -Headers @{Authorization=('Bearer '+$unsub.api_key)} -ContentType application/json -Body (@{email=$entry.email}|ConvertTo-Json)
 if($check.blocked -ne $false){throw 'Suppression check failed'}
}
$profiles=Get-Content -Encoding UTF8 'E:\pxy-runtime\Mail\configuration\profiles.json' -Raw|ConvertFrom-Json
$available=@($profiles.PSObject.Properties.Value|Where-Object {$_.enabled -and $_.from -match '@(126|163)\.com$'})
$left=@($available|Where-Object from -like '*@163.com');$right=@($available|Where-Object from -like '*@126.com');$ordered=@()
for($i=0;$i -lt [Math]::Max($left.Count,$right.Count);$i++){if($i -lt $left.Count){$ordered+=$left[$i]};if($i -lt $right.Count){$ordered+=$right[$i]}}
$verified=@();$unavailable=@()
foreach($p in $ordered){
 if([int](Db ("SELECT count(*) FROM bm_relay_config WHERE active=1 AND lower(auth_user)='"+$p.from+"';")) -eq 0){$unavailable+=@{sender=$p.from;reason='No active native SMTP configuration'};continue}
 try{$null=Api '/relay/test_connection' @{relay_host=$p.host;relay_port=[string]$p.port;auth_user=$p.username;auth_password=$p.password};$verified+=$p.from;Write-Output ('Verified '+$p.from)}catch{$unavailable+=@{sender=$p.from;reason=$_.Exception.Message};Write-Output ('Skipping failed SMTP verification: '+$p.from)}
}
if($verified.Count -eq 0 -or $entries.Count -gt ($verified.Count*2)){throw 'Not enough verified senders for maximum two emails each'}
$policyPath=Join-Path $root 'balanced-send-policy.json'
$cursor=0
if(Test-Path $policyPath){$previousPolicy=Get-Content -Encoding UTF8 $policyPath -Raw|ConvertFrom-Json;$cursor=[int]$previousPolicy.next_sender_index}
$policy=[ordered]@{mode='balanced_round_robin';providers=@('163.com','126.com');interval_seconds=20;maximum_per_sender_per_batch=2;next_sender_index=$cursor;template_id=$TemplateId;updated_at=[DateTimeOffset]::Now.ToString('o')}
$policy|ConvertTo-Json|Set-Content $policyPath -Encoding utf8
$state=[ordered]@{status='prepared';template_id=$TemplateId;csv_path=$CsvPath;recipients=$entries.email;verified_senders=$verified;unavailable_senders=$unavailable;tasks=@();created_at=[DateTimeOffset]::Now.ToString('o')}
function SaveState{$state|ConvertTo-Json -Depth 8|Set-Content $checkpoint -Encoding utf8}
SaveState
Db "INSERT INTO bm_options(name,value) VALUES('native_send_interval','20') ON CONFLICT(name) DO UPDATE SET value=EXCLUDED.value;"|Out-Null
for($i=0;$i -lt $entries.Count;$i++){
 $entry=$entries[$i];$email=$entry.email;$sender=$verified[($cursor+$i)%$verified.Count]
 $group=Api '/contact/group/create' @{name=($BatchKey+'-'+($i+1));description='Approved balanced batch; original CSV order; one task per recipient';create_type=2;file_type='csv';file_data=(($entry|ConvertTo-Csv -NoTypeInformation)-join "`n");double_optin=0}
 $task=Api '/batch_mail/task/create' @{addresser=$sender;full_name='真模型';subject='AI 工具、量化研究与兼职交流，欢迎按需联系';template_id=$TemplateId;group_id=[int]$group.data.group_id;is_record=1;unsubscribe=1;threads=1;track_open=0;track_click=0;start_time=([DateTimeOffset]::UtcNow.ToUnixTimeSeconds()+86400);warmup=0;remark=($BatchKey+' balanced '+($i+1)+'/'+$entries.Count);tag_logic='AND'}
 $id=[int]$task.data.id;$state.tasks+=@{id=$id;sender=$sender;recipient=$email;status='scheduled'};SaveState
 $actual=@(Db "SELECT recipient FROM recipient_info WHERE task_id=$id;")
 if($actual.Count -ne 1 -or $actual[0] -ne $email){throw 'Recipient mismatch; task not started'}
 $now=[DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
 $started=@(Db "UPDATE email_tasks SET recipient_count=1,start_time=$now WHERE id=$id AND task_process=0 AND pause=0 RETURNING id;")
 if($started -notcontains [string]$id){throw 'Task start failed'}
 $state.status='sending';SaveState;Write-Output ("Started task ${id}: $sender -> $email")
 $deadline=[DateTimeOffset]::Now.AddMinutes(3)
 while($true){Start-Sleep -Seconds 5;$sent=[int](Db "SELECT is_sent FROM recipient_info WHERE task_id=$id;");if($sent -eq 1){break};if($sent -eq 3 -or [DateTimeOffset]::Now -gt $deadline){$null=Api '/batch_mail/task/pause' @{task_id=$id};$state.status='paused';SaveState;throw 'Failed or timed out; no automatic retry'}}
 $state.tasks[$i].status='smtp_accepted';SaveState
 $policy.next_sender_index=($cursor+$i+1)%$verified.Count;$policy|ConvertTo-Json|Set-Content $policyPath -Encoding utf8
 Write-Output ("Accepted "+($i+1)+'/'+$entries.Count)
}
$state.status='completed';$state.smtp_accepted=$entries.Count;$state.failed=0;$state.completed_at=[DateTimeOffset]::Now.ToString('o');SaveState
$manifest=Join-Path (Join-Path $root 'recipient-lists') (([IO.Path]::GetFileNameWithoutExtension($CsvPath))+'.json')
if(Test-Path $manifest){$reservation=Get-Content -Encoding UTF8 $manifest -Raw|ConvertFrom-Json;$reservation.status='sent';$reservation|Add-Member -NotePropertyName task_ids -NotePropertyValue $state.tasks.id -Force;$reservation|ConvertTo-Json -Depth 8|Set-Content $manifest -Encoding utf8}
Write-Output ('Completed '+$entries.Count+' emails; final recipient '+$FinalRecipient)
