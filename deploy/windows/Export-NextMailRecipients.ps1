param([ValidateRange(1,1000)][int]$Count=10)
$ErrorActionPreference='Stop'
$root='E:\pxy-runtime\Mail'
$registry=Join-Path $root 'recipient-lists'
New-Item -ItemType Directory -Path $registry -Force | Out-Null
$cfg=@{}
Get-Content -Encoding UTF8 (Join-Path $root '.env')|ForEach-Object{$p=$_ -split '=',2;if($p.Count -eq 2){$cfg[$p[0]]=$p[1]}}
$blocked=@{}
$env:PGPASSWORD=$cfg.DBPASS
try {
    $history=& 'C:\Program Files\PostgreSQL\16\bin\psql.exe' -h $cfg.DBHOST -p $cfg.DBPORT -U $cfg.DBUSER -d $cfg.DBNAME -At -v ON_ERROR_STOP=1 -c 'SELECT DISTINCT lower(recipient) FROM recipient_info WHERE is_sent IN (1,2,3);'
    if($LASTEXITCODE -ne 0){throw 'Could not read Mail sending history.'}
    foreach($email in $history){$blocked[$email.Trim()]=1}
}finally{Remove-Item Env:\PGPASSWORD}
foreach($file in Get-ChildItem -LiteralPath $registry -Filter '*.json'){
    $reserved=Get-Content -Encoding UTF8 -LiteralPath $file.FullName -Raw|ConvertFrom-Json
    foreach($r in $reserved.recipients){$blocked[$r.email.ToLower()]=1}
}
$state=Get-Content -Encoding UTF8 'E:\pxy-runtime\Mail\configuration\state.json' -Raw|ConvertFrom-Json
foreach($j in $state.jobs.PSObject.Properties){foreach($r in $j.Value.recipients){if($r.status -ne 'pending' -or $j.Value.status -ne 'paused'){$blocked[$r.email.ToLower()]=1}}}
foreach($s in $state.suppressed.PSObject.Properties){$blocked[$s.Name.ToLower()]=1}
$profiles=Get-Content -Encoding UTF8 'E:\pxy-runtime\Mail\configuration\profiles.json' -Raw|ConvertFrom-Json
foreach($p in $profiles.PSObject.Properties){$blocked[$p.Value.from.ToLower()]=1}
$source=Get-Content -Encoding UTF8 (Join-Path $root 'migration-data\yxhk-contacts.json') -Raw|ConvertFrom-Json
foreach($d in $source.do_not_contact){foreach($c in $source.contacts|Where-Object id -eq $d.lead_id){$blocked[$c.email.ToLower()]=1}}
$rows=Import-Csv -Encoding UTF8 -LiteralPath 'E:\pxy-runtime\Mail\configuration\original-customers.csv'
$last=$state.jobs.PSObject.Properties|Where-Object{$_.Value.recipients.Count -ge 10}|Sort-Object{$_.Value.created}|Select-Object -Last 1
if($last.Value.status -ne 'completed'){throw 'Previous YXHK campaign is not complete.'}
$start=-1
for($i=0;$i -lt $rows.Count;$i++){if($rows[$i].email -eq $last.Value.recipients[-1].email){$start=$i+1;break}}
if($start -lt 0){throw 'Original CSV continuation cursor could not be located.'}
$unsub=Get-Content -Encoding UTF8 'E:\pxy-runtime\Mail\configuration\unsubscribe.json' -Raw|ConvertFrom-Json
$selected=@();$seen=@{};$mxCache=@{}
for($i=$start;$i -lt $rows.Count -and $selected.Count -lt $Count;$i++){
    $email=$rows[$i].email.Trim().ToLower()
    if($rows[$i].role -ne '成员' -or $blocked.ContainsKey($email) -or $seen.ContainsKey($email)){continue}
    try{$address=[System.Net.Mail.MailAddress]::new($email)}catch{continue}
    if($address.Address -ne $email){continue}
    $domain=($email -split '@')[-1]
    if(!$mxCache.ContainsKey($domain)){
        $mx=Resolve-DnsName -Name $domain -Type MX -DnsOnly -ErrorAction Stop|Where-Object QueryType -eq 'MX'
        if(!$mx){throw "Cannot confirm MX route for $domain."}
        $mxCache[$domain]=1
    }
    $check=Invoke-RestMethod ($unsub.internal_url.TrimEnd('/')+'/internal/check') -NoProxy -Method Post -TimeoutSec 8 -Headers @{Authorization=('Bearer '+$unsub.api_key)} -ContentType 'application/json' -Body (@{email=$email}|ConvertTo-Json)
    if($check.blocked -eq $true){continue}
    if($check.blocked -ne $false){throw 'Could not confirm unsubscribe state.'}
    $selected+=[pscustomobject]@{email=$email;csv_position=$i+1}
    $seen[$email]=1
}
if($selected.Count -ne $Count){throw 'Not enough eligible recipients; no list created.'}
$stamp=Get-Date -Format 'yyyyMMdd-HHmmss'
$name="Mail-Next-$Count-Rows-$($selected[0].csv_position)-$($selected[-1].csv_position)-$stamp"
$destination=Join-Path 'C:\Users\Work\Downloads' ($name+'.csv')
$selected|ForEach-Object{[pscustomobject]@{email=$_.email;attributes=(@{csv_position=[string]$_.csv_position;list_id=$name}|ConvertTo-Json -Compress)}}|Export-Csv -LiteralPath $destination -NoTypeInformation -Encoding utf8
@{list_id=$name;file=$destination;status='prepared_not_sent';created_at=[DateTimeOffset]::Now.ToString('o');recipients=$selected}|ConvertTo-Json -Depth 5|Set-Content -LiteralPath (Join-Path $registry ($name+'.json')) -Encoding utf8
@{file=$destination;count=$Count;first_csv_row=$selected[0].csv_position;last_csv_row=$selected[-1].csv_position;emails_sent=0}|ConvertTo-Json -Compress
