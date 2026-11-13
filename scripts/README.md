# Shell wrappers

Thin shell integrations consume the path printed by the `jcd` executable and
change the parent shell's working directory.

This folder also contains the release installer and packaging tooling:

- `install.ps1` — per-user installer shipped inside the release ZIP. It also
  unblocks the installed files (clears the "Mark of the Web") so the wrapper
  runs without a per-invocation `-ExecutionPolicy Bypass`. See
  [Installing from a Release](../README.md#installing-from-a-release).
- `install-readme.txt` — plain-text installation guide staged into the release
  ZIP as `readme.txt`.
- `build_release.ps1` — maintainer script that builds and packages the release
  ZIP. See
  [Creating a Release Package](../README.md#creating-a-release-package).

## PowerShell — `jcd.ps1`

The wrapper runs the navigator, captures the banner and final path from stdout,
prints the banner and a "Changing current directory to …" confirmation, then
applies the path with `Set-Location`. The interface renders to stderr.

The path crosses the process boundary as UTF-8. Windows PowerShell 5.1 decodes
native command output with `[Console]::OutputEncoding` (the OEM code page by
default), which corrupts non-ASCII names — for example `Imágenes` becomes
`Im├ígenes`. The wrapper saves that setting, switches to UTF-8 for the call, and
restores it afterwards, so accented and non-Latin directory names work.

Add one of these to your PowerShell profile:

```powershell
# Recommended: a function that passes arguments through.
function jcd { & "C:\tools\jcd\scripts\jcd.ps1" @args }

# Or an alias.
Set-Alias jcd "C:\tools\jcd\scripts\jcd.ps1"
```

Then:

```powershell
jcd
jcd C:\some\directory
```

Set `JCD_BIN` to point at a specific executable (defaults to `jcd_bin`). The
integration tests use this override.

## Planned

- `jcd.cmd` — cmd.exe wrapper (`for /f` + `cd /d`).
- `jcd.sh` / `jcd.fish` — POSIX shell and fish functions.

See [`../specs/behavior.md`](../specs/behavior.md) §7 for the output protocol.
