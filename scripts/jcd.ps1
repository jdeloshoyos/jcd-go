# jcd.ps1 - change the PowerShell location using the jcd terminal navigator.
#
# Usage:
#   jcd [directory]
#
# jcd renders its interface on stderr and prints the final directory to
# stdout, so this wrapper captures only the path and applies it with
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

$target = @($result) |
    Where-Object { -not [string]::IsNullOrWhiteSpace($_) } |
    Select-Object -First 1

if ([string]::IsNullOrWhiteSpace($target)) {
    return
}

if (Test-Path -LiteralPath $target -PathType Container) {
    Set-Location -LiteralPath $target
} else {
    [Console]::Error.WriteLine("jcd: '$target' is not a valid directory")
}
