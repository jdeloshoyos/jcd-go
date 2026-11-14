# jcd.ps1 - change the PowerShell location using the jcd terminal navigator.
#
# Usage:
#   jcd [directory]
#
# jcd renders its interface on stderr. On exit it prints two lines to stdout:
# a banner (product name and version) followed by the final directory. This
# wrapper captures them, prints a confirmation, and applies the path with
# Set-Location. The Go process cannot change its parent shell's directory
# itself.
#
# --help and --version print an "Information" mode-selector line first, then
# the text to show. The wrapper relays that text verbatim (blank lines
# included) and changes nothing. Keep the marker in sync with
# internal/cli.ModeInfo in the Go code.
#
# Set JCD_BIN to override the executable (useful for tests).

$binary = if ($env:JCD_BIN) { $env:JCD_BIN } else { 'jcd_bin' }

$global:LASTEXITCODE = 0

# jcd writes UTF-8 to stdout. Windows PowerShell 5.1 decodes a native command's
# stdout with [Console]::OutputEncoding, which defaults to the OEM code page
# (for example CP850). That corrupts non-ASCII paths: the UTF-8 bytes for "á"
# (0xC3 0xA1) are read as two OEM characters, turning "Imágenes" into
# "Im├ígenes". Force UTF-8 for the call, then restore the previous encoding.
$oldOutputEncoding = [Console]::OutputEncoding
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
try {
    $result = & $binary @args
} catch {
    [Console]::Error.WriteLine("jcd: failed to run '$binary': $_")
    return
} finally {
    [Console]::OutputEncoding = $oldOutputEncoding
}

if ($LASTEXITCODE -ne 0) {
    # jcd already wrote a diagnostic to stderr.
    return
}

$raw = @($result)
if ($raw.Count -eq 0) {
    return
}

# Informational output (--help, --version) is prefixed with a mode-selector
# line. Relay the rest verbatim, preserving blank lines, and leave the
# location alone.
if ($raw[0] -eq 'Information') {
    $raw | Select-Object -Skip 1
    return
}

$lines = $raw |
    Where-Object { -not [string]::IsNullOrWhiteSpace($_) }

if ($lines.Count -eq 0) {
    return
}

# The banner is the first line; the target path is the last line.
if ($lines.Count -ge 2) {
    Write-Host $lines[0]
}
$target = $lines[-1]

if (Test-Path -LiteralPath $target -PathType Container) {
    Write-Host "Changing current directory to $target"
    Set-Location -LiteralPath $target
} else {
    [Console]::Error.WriteLine("jcd: '$target' is not a valid directory")
}
