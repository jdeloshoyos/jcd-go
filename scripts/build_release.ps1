#Requires -Version 5.1
<#
.SYNOPSIS
    Builds a JCD release package for GitHub Releases.

.DESCRIPTION
    Builds the stripped production binary, stages the distribution files, and
    creates dist\jcd-<version>-windows-amd64.zip. The archive contains
    install.ps1, jcd_bin.exe, and jcd.ps1 at its root, ready for users to
    extract and run install.ps1.

    The version is read from internal/version/version.go, which is the same
    constant the binary reports. That keeps the archive name and the shipped
    binary in sync, so bump that constant before creating a release.
#>
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$distDir = Join-Path $repoRoot 'dist'
$binaryPath = Join-Path $repoRoot 'jcd_bin.exe'
$versionFile = Join-Path $repoRoot 'internal\version\version.go'

# --- Read the version from the code ------------------------------------------

$versionSource = Get-Content -LiteralPath $versionFile -Raw
$versionMatch = [regex]::Match($versionSource, 'const\s+Version\s*=\s*"([^"]+)"')
if (-not $versionMatch.Success) {
    throw "Could not read the version from $versionFile."
}
$Version = $versionMatch.Groups[1].Value

if ($Version -notmatch '^\d+\.\d+\.\d+([-+][0-9A-Za-z.-]+)?$') {
    throw "Version '$Version' in $versionFile is not a valid semantic version."
}

$zipName = "jcd-$Version-windows-amd64.zip"
$zipPath = Join-Path $distDir $zipName

# --- Build the production binary ---------------------------------------------

Write-Host "Building jcd_bin.exe (version $Version)..."
Push-Location $repoRoot
try {
    & go build -ldflags="-s -w" -trimpath -o $binaryPath '.\cmd\jcd'
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed with exit code $LASTEXITCODE."
    }
} finally {
    Pop-Location
}

if (-not (Test-Path -LiteralPath $binaryPath -PathType Leaf)) {
    throw "The build did not produce $binaryPath."
}

# Verify the binary reports the version the archive is named after.
$reported = (& $binaryPath --version) -join ' '
if ($LASTEXITCODE -ne 0) {
    throw "The built binary failed to report its version (exit code $LASTEXITCODE)."
}
if ($reported -notmatch [regex]::Escape($Version)) {
    throw "Built binary reports '$($reported.Trim())', expected version '$Version'."
}

# --- Stage the distribution files --------------------------------------------

if (Test-Path -LiteralPath $distDir) {
    Remove-Item -LiteralPath $distDir -Recurse -Force
}
New-Item -ItemType Directory -Path $distDir -Force | Out-Null

$installer = Join-Path $PSScriptRoot 'install.ps1'
$wrapper = Join-Path $PSScriptRoot 'jcd.ps1'

foreach ($source in @($installer, $wrapper, $binaryPath)) {
    if (-not (Test-Path -LiteralPath $source -PathType Leaf)) {
        throw "Missing required file: $source"
    }
}

Copy-Item -LiteralPath $installer -Destination (Join-Path $distDir 'install.ps1') -Force
Copy-Item -LiteralPath $wrapper -Destination (Join-Path $distDir 'jcd.ps1') -Force
Copy-Item -LiteralPath $binaryPath -Destination (Join-Path $distDir 'jcd_bin.exe') -Force

# --- Create the archive -------------------------------------------------------

# List the staged files explicitly so the destination ZIP (also under dist\)
# is never included in the archive.
$stagedFiles = @(
    (Join-Path $distDir 'install.ps1')
    (Join-Path $distDir 'jcd.ps1')
    (Join-Path $distDir 'jcd_bin.exe')
)
Compress-Archive -Path $stagedFiles -DestinationPath $zipPath -Force

Write-Host ''
Write-Host 'Release package created:'
Write-Host ''
Write-Host "  Version: $Version"
Write-Host "  dist\$zipName"
