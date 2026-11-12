#Requires -Version 5.1
<#
.SYNOPSIS
    Installs JCD for the current user.

.DESCRIPTION
    Run this script from an extracted JCD release package. It expects
    jcd_bin.exe and jcd.ps1 to sit next to this script.

    The installer:
      * copies jcd_bin.exe and jcd.ps1 to %LOCALAPPDATA%\Programs\JCD;
      * adds that directory to the current user's PATH (never the system PATH);
      * adds a jcd function to the current user's PowerShell profile.

    It is safe to run repeatedly. Running it again upgrades the installed files
    and never touches user configuration in %AppData%\jcd.
#>
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

function Write-Ok {
    param([Parameter(Mandatory = $true)][string]$Message)
    Write-Host "[OK] $Message"
}

function Write-ErrorLine {
    param([Parameter(Mandatory = $true)][string]$Message)
    Write-Host "Error: $Message" -ForegroundColor Red
}

# Returns a comparable form of a PATH entry (case-insensitive, no trailing '\').
function Get-PathKey {
    param([string]$Entry)
    if ([string]::IsNullOrWhiteSpace($Entry)) { return '' }
    return $Entry.Trim().TrimEnd('\')
}

Write-Host 'Installing JCD...'
Write-Host ''

# --- Locate the release payload ----------------------------------------------

$sourceBinary = Join-Path $PSScriptRoot 'jcd_bin.exe'
$sourceWrapper = Join-Path $PSScriptRoot 'jcd.ps1'

if (-not (Test-Path -LiteralPath $sourceBinary -PathType Leaf)) {
    Write-ErrorLine 'jcd_bin.exe was not found next to install.ps1.'
    Write-Host "Expected: $sourceBinary"
    Write-Host 'Extract the full release ZIP and run install.ps1 from that folder.'
    exit 1
}
Write-Ok 'Found jcd_bin.exe'

if (-not (Test-Path -LiteralPath $sourceWrapper -PathType Leaf)) {
    Write-ErrorLine 'jcd.ps1 was not found next to install.ps1.'
    Write-Host "Expected: $sourceWrapper"
    Write-Host 'Extract the full release ZIP and run install.ps1 from that folder.'
    exit 1
}
Write-Ok 'Found jcd.ps1'

# --- Resolve the installation directory --------------------------------------

$localAppData = $env:LOCALAPPDATA
if ([string]::IsNullOrWhiteSpace($localAppData)) {
    $localAppData = [Environment]::GetFolderPath('LocalApplicationData')
}
if ([string]::IsNullOrWhiteSpace($localAppData)) {
    Write-ErrorLine 'Could not determine the LOCALAPPDATA folder.'
    exit 1
}

$installDir = Join-Path $localAppData 'Programs\JCD'

if (Test-Path -LiteralPath $installDir -PathType Container) {
    Write-Ok 'Installation directory already exists'
} else {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    Write-Ok 'Created installation directory'
}

# --- Install the files (overwrite on upgrade) --------------------------------

Copy-Item -LiteralPath $sourceBinary -Destination (Join-Path $installDir 'jcd_bin.exe') -Force
Write-Ok 'Installed jcd_bin.exe'

Copy-Item -LiteralPath $sourceWrapper -Destination (Join-Path $installDir 'jcd.ps1') -Force
Write-Ok 'Installed jcd.ps1'

# --- Add the installation directory to the User PATH -------------------------

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$installKey = Get-PathKey $installDir

$alreadyOnPath = $false
if (-not [string]::IsNullOrEmpty($userPath)) {
    foreach ($entry in ($userPath -split ';')) {
        if ((Get-PathKey $entry) -ieq $installKey) {
            $alreadyOnPath = $true
            break
        }
    }
}

if ($alreadyOnPath) {
    Write-Ok 'User PATH already configured'
} else {
    if ([string]::IsNullOrEmpty($userPath)) {
        $newPath = $installDir
    } else {
        $newPath = $userPath.TrimEnd(';') + ';' + $installDir
    }
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Ok 'Added installation directory to User PATH'
}

# --- Add the integration function to the PowerShell profile ------------------

$profilePath = $PROFILE
if ([string]::IsNullOrWhiteSpace($profilePath)) {
    Write-ErrorLine 'Could not determine the PowerShell profile path ($PROFILE).'
    exit 1
}

$profileDir = Split-Path -Parent $profilePath
if (-not (Test-Path -LiteralPath $profileDir -PathType Container)) {
    New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
}

$profileBlock = @'
# JCD integration
function jcd {
& "$env:LOCALAPPDATA\Programs\JCD\jcd.ps1" @args
}
'@

$profileText = ''
if (Test-Path -LiteralPath $profilePath -PathType Leaf) {
    $profileText = Get-Content -LiteralPath $profilePath -Raw
    if ($null -eq $profileText) { $profileText = '' }
}

if ($profileText.Contains('# JCD integration')) {
    Write-Ok 'PowerShell profile already configured'
} else {
    $newline = [Environment]::NewLine
    $prefix = ''
    if ($profileText.Length -gt 0) {
        if (-not $profileText.EndsWith("`n")) {
            $prefix += $newline
        }
        $prefix += $newline
    }
    $textToAppend = $prefix + $profileBlock.TrimEnd() + $newline
    [System.IO.File]::AppendAllText(
        $profilePath,
        $textToAppend,
        [System.Text.UTF8Encoding]::new($false)
    )
    Write-Ok 'Added JCD integration to PowerShell profile'
}

# --- Done ---------------------------------------------------------------------

Write-Host ''
Write-Host 'Installation complete.'
Write-Host ''
Write-Host 'Restart PowerShell and run:'
Write-Host ''
Write-Host 'jcd'
exit 0
