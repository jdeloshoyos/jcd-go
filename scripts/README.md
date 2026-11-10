# Shell wrappers

Thin shell integrations consume the path printed by the `jcd` executable and
change the parent shell's working directory.

## PowerShell — `jcd.ps1`

The wrapper runs the navigator, captures the final path from stdout, and applies
it with `Set-Location`. The interface renders to stderr so stdout stays clean.

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

Set `JCD_BIN` to point at a specific executable (defaults to `jcd`). The
integration tests use this override.

## Planned

- `jcd.cmd` — cmd.exe wrapper (`for /f` + `cd /d`).
- `jcd.sh` / `jcd.fish` — POSIX shell and fish functions.

See [`../specs/behavior.md`](../specs/behavior.md) §7 for the output protocol.
