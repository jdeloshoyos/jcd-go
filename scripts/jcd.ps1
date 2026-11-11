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
# Set JCD_BIN to override the executable (useful for tests).

$binary = if ($env:JCD_BIN) { $env:JCD_BIN } else { 'jcd_bin' }

$global:LASTEXITCODE = 0
try {
    $result = & $binary @args
} catch {
    [Console]::Error.WriteLine("jcd: failed to run '$binary': $_")
    return
}

if ($LASTEXITCODE -ne 0) {
    # jcd already wrote a diagnostic to stderr.
    return
}

$lines = @($result) |
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
