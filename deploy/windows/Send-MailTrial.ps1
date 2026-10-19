param([switch]$Send)
$ErrorActionPreference = 'Stop'
$runtime = 'E:\pxy-runtime\Mail'
$checkpoint = Join-Path $runtime 'migration-data\mail-trial-20261004.json'
$config = @{}
Get-Content -Encoding UTF8 (Join-Path $runtime '.env') | ForEach-Object {
    $parts = $_ -split '=', 2
    if ($parts.Count -eq 2) { $config[$parts[0]] = $parts[1] }
}
$login = Invoke-RestMethod 'http://127.0.0.1:3040/api/login' -Method Post -ContentType 'application/json' -Body (@{username=$config.ADMIN_USERNAME; password=$config.ADMIN_PASSWORD}|ConvertTo-Json)
if ($login.code -ne 0) { throw 'Mail login failed.' }
$headers = @{Authorization=('Bearer '+$login.data.token)}
if (Test-Path -LiteralPath $checkpoint) {
    $previous = Get-Content -Encoding UTF8 -LiteralPath $checkpoint -Raw | ConvertFrom-Json
    if ($previous.task_id) { Write-Output "Existing trial task: $($previous.task_id). No new task created."; return }
}
$state = Get-Content -Encoding UTF8 'E:\pxy-runtime\Mail\configuration\state.json' -Raw | ConvertFrom-Json
$profiles = Get-Content -Encoding UTF8 'E:\pxy-runtime\Mail\configuration\profiles.json' -Raw | ConvertFrom-Json
$blocked = @{}
foreach ($entry in $profiles.PSObject.Properties) { $blocked[$entry.Value.from.ToLower()]=1 }
foreach ($job in $state.jobs.PSObject.Properties) {
    foreach ($recipient in $job.Value.recipients) {
        if ($recipient.status -ne 'pending' -or $job.Value.status -ne 'paused') { $blocked[$recipient.email.ToLower()]=1 }
    }
}
foreach ($entry in $state.suppressed.PSObject.Properties) { $blocked[$entry.Name.ToLower()]=1 }
$source = Get-Content -Encoding UTF8 (Join-Path $runtime 'migration-data\yxhk-contacts.json') -Raw | ConvertFrom-Json
foreach ($dnc in $source.do_not_contact) {
    foreach ($contact in $source.contacts | Where-Object id -eq $dnc.lead_id) { $blocked[$contact.email.ToLower()]=1 }
}
$rows = Import-Csv -Encoding UTF8 -LiteralPath 'E:\pxy-runtime\Mail\configuration\original-customers.csv'
$lastJob = $state.jobs.PSObject.Properties | Where-Object {$_.Value.recipients.Count -ge 10} | Sort-Object {$_.Value.created} | Select-Object -Last 1
if ($lastJob.Value.status -ne 'completed') { throw 'Previous YXHK batch is not complete; refusing overlapping sends.' }
$lastEmail = $lastJob.Value.recipients[-1].email
$start = -1
for ($i=0;$i -lt $rows.Count;$i++) { if ($rows[$i].email -eq $lastEmail) { $start=$i+1; break } }
if ($start -lt 0) { throw 'Could not locate the previous batch cursor in the original CSV.' }
$unsubscribe = Get-Content -Encoding UTF8 'E:\pxy-runtime\Mail\configuration\unsubscribe.json' -Raw | ConvertFrom-Json
$selected = @()
$seen = @{}
for ($i=$start;$i -lt $rows.Count -and $selected.Count -lt 10;$i++) {
    $row = $rows[$i]
    $email = $row.email.Trim().ToLower()
    if ($row.role -ne '成员' -or $blocked.ContainsKey($email) -or $seen.ContainsKey($email)) { continue }
    try { $parsed = [System.Net.Mail.MailAddress]::new($email) } catch { continue }
    if ($parsed.Address -ne $email) { continue }
    $domain = ($email -split '@')[-1]
    $mx = Resolve-DnsName -Name $domain -Type MX -DnsOnly -ErrorAction Stop | Where-Object QueryType -eq 'MX'
    if (!$mx) { throw "No confirmed MX route for $domain. Trial preparation stopped." }
    $prepared = Invoke-RestMethod ($unsubscribe.internal_url.TrimEnd('/')+'/internal/prepare') -NoProxy -Method Post -TimeoutSec 8 -Headers @{Authorization=('Bearer '+$unsubscribe.api_key)} -ContentType 'application/json' -Body (@{email=$email}|ConvertTo-Json)
    if ($prepared.blocked -eq $true) { continue }
    if ($prepared.blocked -ne $false -or $prepared.url -notlike 'https://*') { throw 'Unsubscribe service returned an invalid result.' }
    $selected += [pscustomobject]@{email=$email;csv_position=$i+1}
    $seen[$email]=1
}
if ($selected.Count -ne 10) { throw 'Could not safely prepare exactly 10 recipients.' }
$record = @{task_id=$null; group_id=$null; template='YXHK #14 01 通用版 - 单人测试 20260930'; sender='xiaotongzmx001@163.com'; interval=20; recipients=$selected}
$record | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $checkpoint -Encoding utf8
Write-Output "Prepared 10 recipients after original CSV row $start."
if (!$Send) { return }
$env:PGPASSWORD=$config.DBPASS
try {
    & 'C:\Program Files\PostgreSQL\16\bin\psql.exe' -h $config.DBHOST -p $config.DBPORT -U $config.DBUSER -d $config.DBNAME -v ON_ERROR_STOP=1 -c "INSERT INTO bm_options(name,value) VALUES('native_legacy_unsubscribe_config','E:/pxy-runtime/Mail/configuration/unsubscribe.json'),('native_send_interval','20') ON CONFLICT(name) DO UPDATE SET value=EXCLUDED.value;"
    if ($LASTEXITCODE -ne 0) { throw 'Failed to configure native sending.' }
} finally { Remove-Item Env:\PGPASSWORD }
$templates=Invoke-RestMethod 'http://127.0.0.1:3040/api/email_template/list?page=1&page_size=50' -Headers $headers
$template = $templates.data.list | Where-Object temp_name -eq $record.template
if (!$template -or @($template).Count -ne 1) { throw 'The original general template could not be matched uniquely.' }
$fileData = ($selected | Select-Object email | ConvertTo-Csv -NoTypeInformation) -join "`n"
$group=Invoke-RestMethod 'http://127.0.0.1:3040/api/contact/group/create' -Method Post -Headers $headers -ContentType 'application/json' -Body (@{name='Mail 试发10人 20261004';description="Original CSV continuation after row $start; history and unsubscribe checked.";create_type=2;file_type='csv';file_data=$fileData;double_optin=0}|ConvertTo-Json)
if ($group.code -ne 0) { throw "Contact group failed: $($group.msg)" }
$record.group_id=$group.data.group_id
$record | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $checkpoint -Encoding utf8
$request=@{addresser=$record.sender;full_name='真模型';subject='AI 工具、量化研究与兼职交流，欢迎按需联系';template_id=$template.id;group_id=$record.group_id;is_record=1;unsubscribe=1;threads=1;track_open=0;track_click=0;start_time=[DateTimeOffset]::UtcNow.ToUnixTimeSeconds();warmup=0;remark='Mail 首次试发10人 20261004；原CSV接续，20秒间隔';tag_logic='AND'}
$task=Invoke-RestMethod 'http://127.0.0.1:3040/api/batch_mail/task/create' -Method Post -Headers $headers -ContentType 'application/json' -Body ($request|ConvertTo-Json)
if ($task.code -ne 0) { throw "Task creation failed: $($task.msg)" }
$record.task_id=$task.data.id
$record | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $checkpoint -Encoding utf8
Write-Output "Mail task created: $($record.task_id); recipients: 10; interval: 20 seconds."
