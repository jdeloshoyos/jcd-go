# Shell wrappers

This directory will contain the thin shell integrations that consume the path
printed by the `jcd` executable and change the parent shell's working
directory.

Planned for the TUI milestone (M2):

- `jcd.ps1` — PowerShell wrapper (`Set-Location`).

Later (cross-platform milestone):

- `jcd.cmd` — cmd.exe wrapper (`for /f` + `cd /d`).
- `jcd.sh` / `jcd.fish` — POSIX shell and fish functions.

The Go process never changes the parent shell's directory itself; it prints the
resolved path on stdout, and these wrappers apply it. See
[`../specs/behavior.md`](../specs/behavior.md) §7 for the output protocol.
