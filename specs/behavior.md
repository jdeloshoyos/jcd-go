# JCD — Behaviour Specification

- **Status:** Draft
- **Version:** 0.1.0
- **Last updated:** 2026-11-10

This document defines exactly how the app behaves. It is the reference for the
test suite.

## 1. Application states

The TUI is a small state machine:

| State         | Description |
| ------------- | ----------- |
| `browsing`    | Normal directory listing and navigation. |
| `searching`   | Search-input mode is active. |
| `help`        | Help screen is displayed. |
| `drive-select`| Windows drive-selection screen. |

Esc in `searching` returns to `browsing`; Esc in `browsing` quits. F1 toggles
`help`. Backspace at a drive root switches to `drive-select`.

## 2. Key bindings

### 2.1 Browsing

| Key          | Action |
| ------------ | ------ |
| `↑` / `↓`    | Move cursor one entry; scroll when at the view edge. |
| `PgUp`/`PgDn`| Move cursor one page. |
| `Home`/`End` | Jump to first/last entry. |
| `Enter`      | Enter directory, or open file. |
| `Backspace`  | Go to parent directory (or drive-select at a drive root). |
| `F1` / `?`   | Toggle help. |
| `F2`         | Cycle sort mode. |
| `/`          | Enter search mode. |
| `Esc`        | Quit and report the current directory. |

### 2.2 Searching

| Key     | Action |
| ------- | ------ |
| `Enter` | Jump to the first match; remain in search mode. |
| `Tab`   | Jump to the next match, wrapping around. |
| `Esc`   | Cancel search mode; selection is unchanged. |
| `←`/`→` | Move the text cursor within the query. |
| `Backspace` | Delete the previous character. |

### 2.3 Help

| Key          | Action |
| ------------ | ------ |
| `↑` / `↓`    | Scroll one line. |
| `PgUp`/`PgDn`| Scroll one page. |
| `Esc` / `F1` / `?` | Close help. |

### 2.4 Drive selection (Windows)

| Key   | Action |
| ----- | ------ |
| `↑`/`↓`| Move selection. |
| `Enter`| Navigate to the selected drive root. |
| `F1` / `?` | Toggle help; closing it returns here. |
| `Esc`  | Quit and report the last browsed directory. |

## 3. Search semantics

- Search is case-insensitive and matches **anywhere** in the entry name
  (substring), not only at the start.
- An empty query matches nothing.
- `Enter` selects the first match.
- `Tab` advances to the next match; after the last match it wraps to the first.
- If there are no matches, the app emits a short feedback signal (beep or a
  visual flash) and keeps the current selection.
- Search mode does not modify the listing; it only moves the selection.

## 4. Sorting rules

Six modes, cycled by F2:

| Mode | Order |
| ---- | ----- |
| 1    | Name ascending |
| 2    | Last-write time ascending |
| 3    | Size ascending |
| 4    | Name descending |
| 5    | Last-write time descending |
| 6    | Size descending |

Rules that apply to every mode:

1. Directories and drives always precede files.
2. Name comparisons are case-insensitive.
3. For date/size modes, equal keys are broken by ascending name.
4. Name-descending reverses only the name comparison; directories still come
   first.

## 5. Display layout

```
<path>                                                      <- row 1
<N> directories, <M> files; <size> used.      [F2]=... <pos>/<total>  <- row 2
<entry rows>                                                <- rows 3..H-1
[F1]=Show help [ESC]=Quit                 Search: <pattern>  <- row H
```

- When the path is wider than the terminal, it is abbreviated as
  `<leading>[...]<trailing>`.
- Entry row format:
  `<9-char attr mask> <YYYY/MM/DD HH:MM:SS> <name>[ (<size>)]`.
- Drives show `<letter>:` followed by a drive-type label.

### 5.1 Attribute mask

Nine positions, in order. A dash means the attribute is absent.

| Pos | Letter | Meaning |
| --- | ------ | ------- |
| 1   | `d`    | Directory |
| 2   | `l`    | Reparse point / symlink |
| 3   | `r`    | Read-only |
| 4   | `h`    | Hidden |
| 5   | `s`    | System |
| 6   | `a`    | Archive |
| 7   | `c`    | Compressed |
| 8   | `i`    | Not content indexed |
| 9   | `e`    | Encrypted |

On non-Windows platforms the mask is derived from portable metadata: `d` for
directories, `l` for symlinks, `h` for dotfiles, and `r` when the owner lacks
write permission.

## 6. Size formatting

Sizes are rendered in binary steps using the legacy algorithm: divide by 1024
while the value is **strictly greater** than 1024. Values below that threshold
are shown as `<n> bytes`; larger values are shown with one decimal and a unit
(`KB`, `MB`, `GB`, `TB`, `PB`, `EB`, `ZB`, `YB`). This boundary behaviour is
intentionally preserved from the legacy app.

## 7. Exit and output protocol

### 7.1 Success

The app restores the terminal, then writes **exactly one line** to stdout:

```
<absolute target path>
```

and exits with status `0`. Nothing else is written to stdout.

### 7.2 Failure

Errors are written to stderr and the process exits non-zero:

| Exit code | Meaning |
| --------- | ------- |
| `0`       | Success; path written to stdout. |
| `1`       | Runtime error (e.g. cannot read directory). |
| `2`       | Usage error (e.g. invalid argument). |
| `3`       | Terminal too small. |

### 7.3 Shell wrapper

A thin wrapper consumes the stdout line and changes the parent shell's working
directory. Because the wrapper captures stdout, the TUI renders to **stderr**;
stdout carries only the path. For PowerShell:

```powershell
$binary = if ($env:JCD_BIN) { $env:JCD_BIN } else { 'jcd' }
$global:LASTEXITCODE = 0
$result = & $binary @args
if ($LASTEXITCODE -ne 0) { return }

$target = $result | Where-Object { $_ } | Select-Object -First 1
if ($target -and (Test-Path -LiteralPath $target -PathType Container)) {
    Set-Location -LiteralPath $target
}
```

The Go process never attempts to modify the parent shell's directory directly
(a child process cannot portably change its parent's working directory). The
wrapper must therefore run in the caller's session (an alias or function in the
PowerShell profile), not in a separate process.

The repository ships this wrapper as `scripts/jcd.ps1`. `JCD_BIN` overrides the
executable it invokes, which also makes the wrapper easy to test.

## 8. Path resolution at startup

1. With no argument, use the current working directory.
2. Resolve relative arguments against the current working directory.
3. On Windows, an argument beginning with `\` is relative to the current drive
   root.
4. Trim trailing separators.
5. Verify the result is an existing directory; otherwise exit `2`.
6. If the app started on `drive-select` and the user quits without choosing a
   drive, report the last browsed directory.

## 9. Terminal requirements

- Minimum size: 80 columns × 20 rows. Smaller terminals cause exit code `3`
  with a stderr message.
- The terminal is restored on every exit path (normal quit, error, signal).
