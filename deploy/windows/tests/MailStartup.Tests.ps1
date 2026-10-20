$deployRoot=Split-Path $PSScriptRoot -Parent

Describe 'Mail startup survives a renamed source folder' {
    BeforeEach {
        $runtime=Join-Path $TestDrive ([guid]::NewGuid().ToString('N'))
        foreach($folder in @('core','postgresql')) {
            New-Item -ItemType Directory -Force -Path (Join-Path $runtime $folder)|Out-Null
        }
        foreach($file in @('.env','core\billionmail.exe','postgresql\PG_VERSION')) {
            Set-Content -LiteralPath (Join-Path $runtime $file) -Value 'fixture'
        }
        Mock Register-ScheduledTask { }
    }

    It 'registers the stable runtime copy instead of the source path' {
        & (Join-Path $deployRoot 'Install-MailAutostart.ps1') -RuntimeRoot $runtime
        $installed=Join-Path $runtime 'operations\Manage-MailNative.ps1'
        (Get-FileHash $installed).Hash | Should Be (Get-FileHash (Join-Path $deployRoot 'Manage-MailNative.ps1')).Hash
        Assert-MockCalled Register-ScheduledTask -Times 1 -Exactly -Scope It -ParameterFilter {
            $TaskName -eq 'PXY-MailNative-Start' -and
            $Action.Arguments.Contains($installed) -and
            $Action.Arguments.Contains('-Action Start') -and
            !$Action.Arguments.Contains($deployRoot)
        }
    }

    It 'rejects a wrong runtime without registering or creating an empty workspace' {
        $missing=Join-Path $TestDrive 'wrong-runtime'
        $caught=$null
        try { & (Join-Path $deployRoot 'Install-MailAutostart.ps1') -RuntimeRoot $missing }
        catch { $caught=$_.Exception.Message }
        $caught | Should Match 'Cannot register autostart'
        Test-Path $missing | Should Be $false
        Assert-MockCalled Register-ScheduledTask -Times 0 -Exactly -Scope It
    }

    It 'fails Start before creating directories when existing data is missing' {
        $missing=Join-Path $TestDrive 'missing-start-runtime'
        & powershell.exe -NoProfile -ExecutionPolicy Bypass -File (Join-Path $deployRoot 'Manage-MailNative.ps1') -Action Start -RuntimeRoot $missing -NoBrowser 2>$null
        $LASTEXITCODE | Should Be 1
        Test-Path $missing | Should Be $false
    }
}
