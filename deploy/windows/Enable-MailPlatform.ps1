[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][string]$OwnerEmail,
    [string]$RuntimeRoot = 'E:\pxy-runtime\Mail'
)
$ErrorActionPreference = 'Stop'
$RuntimeRoot = [IO.Path]::GetFullPath($RuntimeRoot)
if ($RuntimeRoot -ne 'E:\pxy-runtime\Mail') { throw 'Platform workspaces must run in E:\pxy-runtime\Mail.' }
$OwnerEmail = $OwnerEmail.Trim().ToLowerInvariant()
if ($OwnerEmail -notmatch '^[0-9]+@qq\.com$') { throw 'Specify the platform owner QQ email address.' }
$path = Join-Path $RuntimeRoot '.env'
$settings = @{}
foreach ($line in [IO.File]::ReadAllLines($path)) {
    if ($line -match '^([^#=]+)=(.*)$') { $settings[$matches[1]] = $matches[2] }
}
if ($settings.MAIL_SPACE_EMAIL -and $settings.MAIL_SPACE_EMAIL -ne $OwnerEmail) {
    throw 'The existing workspace is bound to another owner; an explicit data migration is required.'
}
function New-PlatformSecret {
    $bytes = New-Object byte[] 32
    $rng = [Security.Cryptography.RandomNumberGenerator]::Create()
    try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
    return [Convert]::ToBase64String($bytes).TrimEnd('=').Replace('+','-').Replace('/','_')
}
$backupRoot = Join-Path $RuntimeRoot 'backups'
New-Item -ItemType Directory -Force -Path $backupRoot | Out-Null
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
Copy-Item -LiteralPath $path -Destination (Join-Path $backupRoot "before-platform-$stamp.env")
$env:PGPASSWORD = $settings.DBPASS
try {
    & 'C:\Program Files\PostgreSQL\16\bin\pg_dump.exe' -h 127.0.0.1 -p 25433 -U billionmail -d billionmail -Fc -f (Join-Path $backupRoot "before-platform-$stamp.dump")
    if ($LASTEXITCODE -ne 0) { throw 'Owner database backup failed.' }
} finally { Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue }
$settings.MAIL_SPACE_EMAIL = $OwnerEmail
if (!$settings.MAIL_APP_SESSION_SECRET) { $settings.MAIL_APP_SESSION_SECRET = New-PlatformSecret }
if (!$settings.MAIL_BOOTSTRAP_SECRET) { $settings.MAIL_BOOTSTRAP_SECRET = New-PlatformSecret }
[IO.File]::WriteAllLines($path,@($settings.Keys | Sort-Object | ForEach-Object { "$_=$($settings[$_])" }),[Text.UTF8Encoding]::new($false))
Write-Output "Existing Mail workspace bound to $OwnerEmail. Database and configuration backup saved."
