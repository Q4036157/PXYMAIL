[CmdletBinding()]
param(
    [ValidateSet('Deploy','Start','Stop','Status')][string]$Action = 'Deploy',
    [string]$RuntimeRoot = 'E:\pxy-runtime\Mail',
    [switch]$NoBrowser
)
$ErrorActionPreference = 'Stop'
$sourceRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..'))
$RuntimeRoot = [IO.Path]::GetFullPath($RuntimeRoot)
$coreRoot = Join-Path $RuntimeRoot 'core'
$configPath = Join-Path $RuntimeRoot '.env'
$postgresData = Join-Path $RuntimeRoot 'postgresql'
$redisRoot = Join-Path $RuntimeRoot 'redis'
$logsRoot = Join-Path $RuntimeRoot 'logs'
$exePath = Join-Path $coreRoot 'billionmail.exe'
$url = 'http://127.0.0.1:3040'
$pgCommand = Get-Command initdb.exe -ErrorAction SilentlyContinue
$pgBin = if ($pgCommand) { Split-Path $pgCommand.Source } else { 'C:\Program Files\PostgreSQL\16\bin' }
$redisCommand = Get-Command redis-server.exe -ErrorAction SilentlyContinue
$redisExe = if ($redisCommand) { $redisCommand.Source } else { 'C:\Program Files\Redis\redis-server.exe' }
if (!(Test-Path (Join-Path $pgBin 'pg_ctl.exe')) -or !(Test-Path $redisExe)) {
    throw 'PostgreSQL or Redis native runtime was not found.'
}

function Invoke-Native([string]$File, [string[]]$Arguments) {
    & $File @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$File failed with exit code $LASTEXITCODE" }
}
function New-Secret {
    $bytes = New-Object byte[] 24
    $rng = [Security.Cryptography.RandomNumberGenerator]::Create()
    try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
    return [Convert]::ToBase64String($bytes).TrimEnd('=').Replace('+','A').Replace('/','B')
}
function Read-Settings {
    $settings = @{}
    foreach ($line in [IO.File]::ReadAllLines($configPath)) {
        if ($line -match '^([^#=]+)=(.*)$') { $settings[$matches[1]] = $matches[2] }
    }
    return $settings
}
function Test-Listener([int]$Port) {
    $client = New-Object Net.Sockets.TcpClient
    try { $client.Connect('127.0.0.1', $Port); return $true } catch { return $false } finally { $client.Dispose() }
}
function Test-App {
    try {
        $health = Invoke-RestMethod "$url/health/native" -TimeoutSec 3
        return $health.status -eq 'ok' -and $health.application -eq 'BillionMail'
    } catch { return $false }
}
function Test-Gateway {
    try {
        $health = Invoke-RestMethod 'http://127.0.0.1:3041/health/native' -TimeoutSec 3
        return $health.status -eq 'ok' -and $health.auth -eq 'platform_isolated'
    } catch { return $false }
}
function Stop-OwnedProcess([string]$PidPath, [string]$ExpectedExe) {
    if (!(Test-Path -LiteralPath $PidPath)) { return }
    $ownedId = [int]([IO.File]::ReadAllText($PidPath).Trim())
    $process = Get-Process -Id $ownedId -ErrorAction SilentlyContinue
    if ($process) {
        if (![string]::Equals($process.Path, $ExpectedExe, [StringComparison]::OrdinalIgnoreCase)) {
            throw "Process ownership mismatch for PID $ownedId; refusing to stop it."
        }
        Stop-Process -Id $ownedId -Force -Confirm:$false
        $process.WaitForExit(30000) | Out-Null
        $remaining = Get-Process -Id $ownedId -ErrorAction SilentlyContinue
        if ($remaining -and !$remaining.HasExited) {
            throw "Process $ownedId did not stop; the running executable was not replaced."
        }
    }
    Remove-Item -LiteralPath $PidPath
}
foreach ($dir in @($RuntimeRoot,$coreRoot,$logsRoot,$redisRoot)) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
}
$lock = $null
try {
    $lock = [IO.File]::Open((Join-Path $RuntimeRoot 'operation.lock'),'OpenOrCreate','ReadWrite','None')
    if ($Action -eq 'Status') {
        Write-Host "Application healthy: $(Test-App)"
        Write-Host "PostgreSQL listener: $(Test-Listener 25433)"
        Write-Host "Redis listener: $(Test-Listener 26380)"
        Write-Host "URL: $url"
        Write-Host "Platform gateway healthy: $(Test-Gateway)"
        exit 0
    }
    if ($Action -eq 'Stop') {
        Stop-OwnedProcess (Join-Path $RuntimeRoot 'gateway.pid') $exePath
        Stop-OwnedProcess (Join-Path $RuntimeRoot 'application.pid') $exePath
        Stop-OwnedProcess (Join-Path $RuntimeRoot 'redis.pid') $redisExe
        if (Test-Path (Join-Path $postgresData 'postmaster.pid')) {
            Invoke-Native (Join-Path $pgBin 'pg_ctl.exe') @('-D',$postgresData,'stop','-m','fast','-w')
        }
        @{status='ok'; action=$Action; time=(Get-Date -Format o)} |
            ConvertTo-Json | Set-Content -LiteralPath (Join-Path $RuntimeRoot 'last-operation.json') -Encoding UTF8
        Write-Host 'Mail native services stopped. Other projects were not stopped.'
        exit 0
    }
    if (!(Test-Path -LiteralPath $configPath)) {
        $adminPassword = New-Secret
        $dbPassword = New-Secret
        $redisPassword = New-Secret
        $environmentLines = @(
            'ADMIN_USERNAME=admin', "ADMIN_PASSWORD=$adminPassword", 'SafePath=',
            'DBHOST=127.0.0.1', 'DBPORT=25433', 'DBUSER=billionmail', 'DBNAME=billionmail', "DBPASS=$dbPassword",
            'REDIS_ADDRESS=127.0.0.1:26380', 'REDISDB=1', "REDISPASS=$redisPassword",
            'HTTP_PORT=3040', 'BILLIONMAIL_HOSTNAME=localhost', 'TZ=Asia/Shanghai', 'FAIL2BAN_INIT=n'
        )
        [IO.File]::WriteAllLines($configPath,$environmentLines,(New-Object Text.UTF8Encoding $false))
        "URL: $url`r`nUsername: admin`r`nPassword: $adminPassword" |
            Set-Content -LiteralPath (Join-Path $RuntimeRoot 'login.txt') -Encoding UTF8
        $identity = [Security.Principal.WindowsIdentity]::GetCurrent().Name
        Invoke-Native 'icacls.exe' @($RuntimeRoot,'/inheritance:r','/grant:r',"${identity}:(OI)(CI)F",'SYSTEM:(OI)(CI)F')
    }
    $settings = Read-Settings
    if ($Action -eq 'Deploy') {
        Write-Host 'Building the browser interface.'
        Push-Location (Join-Path $sourceRoot 'core\frontend')
        try {
            Invoke-Native 'pnpm.cmd' @('build')
        } finally { Pop-Location }
        & robocopy.exe (Join-Path $sourceRoot 'core\frontend\dist') (Join-Path $sourceRoot 'core\public\dist') /E /NFL /NDL /NJH /NJS /NP
        if ($LASTEXITCODE -ge 8) { throw 'Frontend artifact copy failed.' }
        Write-Host 'Building the Go backend for Windows. No Docker is used.'
        Push-Location (Join-Path $sourceRoot 'core')
        try {
            Invoke-Native 'go.exe' @('build','-o',(Join-Path $coreRoot 'billionmail.next.exe'),'main.go')
        } finally { Pop-Location }
        $appPidPath = Join-Path $RuntimeRoot 'application.pid'
        if ((Test-Listener 3040) -and !(Test-Path -LiteralPath $appPidPath)) {
            throw 'Port 3040 is occupied and this deployment has no process ownership record.'
        }
        if (Test-Path -LiteralPath $appPidPath) {
            Write-Host 'Build succeeded. Restarting the Mail backend to apply the new build.'
            Stop-OwnedProcess (Join-Path $RuntimeRoot 'gateway.pid') $exePath
            Stop-OwnedProcess $appPidPath $exePath
        }
        Move-Item -LiteralPath (Join-Path $coreRoot 'billionmail.next.exe') -Destination $exePath -Force
        foreach ($folder in @('public','languages','template','resource')) {
            & robocopy.exe (Join-Path $sourceRoot "core\$folder") (Join-Path $coreRoot $folder) /E /NFL /NDL /NJH /NJS /NP
            if ($LASTEXITCODE -ge 8) { throw "Copy failed: $folder" }
        }
        $configDir = Join-Path $coreRoot 'manifest\config'
        New-Item -ItemType Directory -Force $configDir | Out-Null
        $configText = [IO.File]::ReadAllText((Join-Path $sourceRoot 'core\manifest\config\config.yaml'))
        $configText.Replace('cookieSecure: true','cookieSecure: false') |
            Set-Content -LiteralPath (Join-Path $configDir 'config.yaml') -Encoding UTF8
    }
    if (!(Test-Path -LiteralPath $exePath)) { throw 'Backend is not built. Run Deploy-Mail.bat first.' }
    $env:BILLIONMAIL_NATIVE = '1'
    $env:BILLIONMAIL_ENV_FILE = $configPath
    $env:PGPASSWORD = $settings.DBPASS
    if (!(Test-Path -LiteralPath (Join-Path $postgresData 'PG_VERSION'))) {
        $passwordFile = Join-Path $RuntimeRoot 'initdb-password.tmp'
        try {
            [IO.File]::WriteAllText($passwordFile,$settings.DBPASS)
            Invoke-Native (Join-Path $pgBin 'initdb.exe') @('-D',$postgresData,'-U','billionmail',"--pwfile=$passwordFile",'--auth=scram-sha-256','--encoding=UTF8','--locale=C')
        } finally { Remove-Item -LiteralPath $passwordFile -ErrorAction SilentlyContinue }
    }
    if (!(Test-Listener 25433)) {
        Invoke-Native (Join-Path $pgBin 'pg_ctl.exe') @('-D',$postgresData,'-l',(Join-Path $logsRoot 'postgresql.log'),'-o','-h 127.0.0.1 -p 25433 -c max_connections=40 -c shared_buffers=64MB','start','-w')
    }
    $database = & (Join-Path $pgBin 'psql.exe') -h 127.0.0.1 -p 25433 -U billionmail -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='billionmail'"
    if ($LASTEXITCODE -ne 0) { throw 'PostgreSQL connection failed or port 25433 is owned by another server.' }
    $activeData = & (Join-Path $pgBin 'psql.exe') -h 127.0.0.1 -p 25433 -U billionmail -d postgres -tAc 'SHOW data_directory'
    if ($LASTEXITCODE -ne 0 -or ![string]::Equals(([string]$activeData).Trim().Replace('/','\'),$postgresData,[StringComparison]::OrdinalIgnoreCase)) {
        throw 'Port 25433 belongs to a different runtime directory; refusing to start a second Mail deployment.'
    }
    if ($database -notcontains '1') {
        Invoke-Native (Join-Path $pgBin 'createdb.exe') @('-h','127.0.0.1','-p','25433','-U','billionmail','billionmail')
    }
    if (!(Test-Listener 26380)) {
        $redisConfig = Join-Path $redisRoot 'redis.conf'
        @('bind 127.0.0.1','port 26380',"requirepass $($settings.REDISPASS)",
            'appendonly yes', 'maxmemory 64mb', 'maxmemory-policy noeviction',
            ('dir "' + $redisRoot.Replace('\','/') + '"')) |
            Set-Content -LiteralPath $redisConfig -Encoding ASCII
        $redisProcess = Start-Process -FilePath $redisExe -ArgumentList ('"' + $redisConfig + '"') -WorkingDirectory $redisRoot -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $logsRoot 'redis.out.log') -RedirectStandardError (Join-Path $logsRoot 'redis.err.log')
        $redisProcess.Id | Set-Content -LiteralPath (Join-Path $RuntimeRoot 'redis.pid') -Encoding ASCII
    }
    if (!(Test-App)) {
        if (Test-Listener 3040) { throw 'Port 3040 is occupied by an unhealthy or different application.' }
        $app = Start-Process -FilePath $exePath -WorkingDirectory $coreRoot -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $logsRoot 'application.out.log') -RedirectStandardError (Join-Path $logsRoot 'application.err.log')
        $app.Id | Set-Content -LiteralPath (Join-Path $RuntimeRoot 'application.pid') -Encoding ASCII
        $ready = $false
        for ($attempt = 0; $attempt -lt 45; $attempt++) {
            if (Test-App) { $ready = $true; break }
            $app.Refresh()
            if ($app.HasExited) { throw "Backend exited. See $logsRoot\application.err.log" }
            Start-Sleep -Seconds 1
        }
        if (!$ready) { throw "Startup timed out. See $logsRoot" }
    }
    if ($settings.MAIL_APP_SESSION_SECRET -and !(Test-Gateway)) {
        if (Test-Listener 3041) { throw 'Port 3041 is occupied by another or unhealthy gateway.' }
        $gateway = Start-Process -FilePath $exePath -ArgumentList '--platform-gateway' -WorkingDirectory $coreRoot -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $logsRoot 'gateway.out.log') -RedirectStandardError (Join-Path $logsRoot 'gateway.err.log')
        $gateway.Id | Set-Content -LiteralPath (Join-Path $RuntimeRoot 'gateway.pid') -Encoding ASCII
        $gatewayReady = $false
        for ($attempt = 0; $attempt -lt 30; $attempt++) {
            if (Test-Gateway) { $gatewayReady = $true; break }
            $gateway.Refresh()
            if ($gateway.HasExited) { throw 'Platform gateway exited. Check gateway.err.log.' }
            Start-Sleep -Seconds 1
        }
        if (!$gatewayReady) { throw 'Platform gateway startup timed out.' }
    }
    @{status='ok'; action=$Action; time=(Get-Date -Format o); url=$url; executable=$exePath} |
        ConvertTo-Json | Set-Content -LiteralPath (Join-Path $RuntimeRoot 'last-operation.json') -Encoding UTF8
    Write-Host "Mail is running: $url"
    Write-Host "Login details: $RuntimeRoot\login.txt"
    Write-Host "Logs: $logsRoot"
    if (!$NoBrowser) { Start-Process $url }
} catch {
    @{status='failed'; action=$Action; time=(Get-Date -Format o); error=$_.Exception.Message; location=$_.ScriptStackTrace} |
        ConvertTo-Json | Set-Content -LiteralPath (Join-Path $RuntimeRoot 'last-operation.json') -Encoding UTF8
    Write-Host ("Mail operation failed: " + $_.Exception.Message) -ForegroundColor Red
    exit 1
} finally {
    if ($lock) { $lock.Dispose() }
    Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue
}
