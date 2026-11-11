# JCD — Behaviour Specification

- **Status:** Draft
- **Version:** 1.2.0
- **Last updated:** 2026-11-11

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
| `quick-access`| Quick Access popup for tracked directories. |

Esc in `searching` returns to `browsing`; Esc in `browsing` quits. F1 toggles
`help`. Backspace at a drive root switches to `drive-select`. F4 opens
`quick-access` from `browsing`; Esc there closes it and returns to `browsing`.

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
| `F3`         | Enter search mode with an empty query. |
| Any visible character | Enter search mode with that character as the query. |
| `Esc`        | Quit and report the current directory. |

Keys already bound to an action keep it: only visible, printable, non-space
characters that are not otherwise mapped start a search (so `?` still toggles
help, and space does not search).

### 2.2 Searching

Typing a query moves the selection to the first match immediately; the matched
substring is highlighted in every matching row. Matching ignores case and
accents. Search can be opened with `F3` (empty query) or by typing any visible
character, which seeds the query.

| Key     | Action |
| ------- | ------ |
| `Enter` | Open the selected entry (enter a directory or launch a file), then leave search mode. |
| `Tab`   | Jump to the next match, wrapping around. |
| `Shift+Tab` | Jump to the previous match. |
| `Esc`   | Cancel search mode; the selection is left where it is. |
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

### 2.5 Quick Access

| Key          | Action |
| ------------ | ------ |
| `F4`         | Open Quick Access from browsing. |
| `↑` / `↓`    | Move the selection between directory entries (section headers are skipped). |
| `PgUp`/`PgDn`| Move one page of entries. |
| `Home`/`End` | First / last entry. |
| `Enter`      | Jump to the selected directory and quit, so the wrapper changes the shell's directory. |
| `F`          | Toggle the selected entry's favorite status. |
| `Esc`        | Close the popup without navigating. |

## 3. Search semantics

- Search is case- and diacritic-insensitive and matches **anywhere** in the
  entry name (substring), not only at the start. Accents are stripped to their
  base character, so `ima` matches `imágenes`, `imá`, `imà`, and `ïma`.
- An empty query matches nothing.
- Typing moves the selection to the first match as soon as one exists; when
  nothing matches, the selection is left unchanged.
- The matched substring is highlighted in every matching row (selected or not)
  while search mode is active.
- `Enter` opens the selected entry — entering a directory or launching a file —
  and leaves search mode.
- `Tab` advances to the next match; after the last match it wraps to the first.
- If there are no matches, the footer shows a zero match count and the
  selection is unchanged.
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

### 5.2 Quick Access popup

The popup is a centered, bordered window overlaid on the directory listing, so
the listing stays visible around it. It contains a single scrollable list
grouped into sections, in this order: Favorites (`★ Favorites`), Recent
(`🕒 Recent`), and Frequent (`🔥 Frequent`). A directory may appear in more than
one section, but never twice within the same section. Empty sections are
omitted.

- Favorite entries are prefixed with `★ `; other entries are indented with two
  spaces so paths stay aligned.
- Each section has a title row and a dashed rule beneath it; the cursor skips
  both.
- The popup footer shows `[↑↓]=Move [Enter]=Go [F]=Favorite [Esc]=Close`.
- When no tracked directory is available, the popup shows
  `No tracked directories yet.`

Paths wider than the popup are **middle-truncated at segment boundaries**: the
beginning (volume/root and as many leading segments as fit) and the end
(ideally the parent and leaf) are preserved, and removed middle segments are
replaced by `...`. For example:

```
C:\users\jaime\documents\projects\doccompose\backend
C:\users\jaime\...\doccompose\backend
C:\users\...\doccompose\backend
C:\...\doccompose\backend
```

Segment names are not cut while a shorter segment combination still fits. Only
as a last resort is the leaf itself truncated.

## 6. Size formatting

Sizes use standard IEC binary units. Values below 1 KiB are shown as whole
bytes (`512 B`). Larger values are divided by 1024 and shown with one decimal
place and an IEC unit: `KiB`, `MiB`, `GiB`, `TiB`, `PiB`, `EiB`, `ZiB`, `YiB`.

A value that would round to `1024.0` is promoted to the next unit, so
1 MiB − 1 byte renders as `1.0 MiB` rather than `1024.0 KiB`. This is a
deliberate change from the legacy `KB`/`MB` formatting.

## 7. Exit and output protocol

### 7.1 Success

The app restores the terminal, then writes **two lines** to stdout: a banner
followed by the target path.

```
JCD Terminal Directory Navigator v<version>
<absolute target path>
```

and exits with status `0`. Nothing else is written to stdout.

### 7.2 Failure

Errors are written to stderr and the process exits non-zero:

| Exit code | Meaning |
| --------- | ------- |
| `0`       | Success; banner and path written to stdout. |
| `1`       | Runtime error (e.g. cannot read directory). |
| `2`       | Usage error (e.g. invalid argument). |
| `3`       | Terminal too small. |

### 7.3 Shell wrapper

A thin wrapper consumes the stdout lines and changes the parent shell's working
directory. Because the wrapper captures stdout, the TUI renders to **stderr**;
stdout carries only the banner and the path. The wrapper prints the banner and a
confirmation line, then applies the path (the last line). For PowerShell:

```powershell
$binary = if ($env:JCD_BIN) { $env:JCD_BIN } else { 'jcd' }
$global:LASTEXITCODE = 0

# jcd writes UTF-8; force UTF-8 decoding for the native command and restore it.
$oldOutputEncoding = [Console]::OutputEncoding
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
try {
    $result = & $binary @args
} finally {
    [Console]::OutputEncoding = $oldOutputEncoding
}

if ($LASTEXITCODE -ne 0) { return }

$lines = @($result) | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
if ($lines.Count -eq 0) { return }

if ($lines.Count -ge 2) { Write-Host $lines[0] }   # banner
$target = $lines[-1]
if (Test-Path -LiteralPath $target -PathType Container) {
    Write-Host "Changing current directory to $target"
    Set-Location -LiteralPath $target
}
```

The path is transferred as **UTF-8**. Windows PowerShell 5.1 decodes a native
command's stdout with `[Console]::OutputEncoding`, which defaults to the OEM code
page (for example CP850). Reading UTF-8 bytes that way corrupts non-ASCII names:
the bytes for `á` (`0xC3 0xA1`) become `├í`, so `Imágenes` arrives as
`Im├ígenes`. A wrapper must therefore save `[Console]::OutputEncoding`, set it
to UTF-8 for the duration of the call, and restore it afterwards. This only
affects the captured stdout; the TUI renders directly to stderr and is
unaffected.

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
7. A leading `--` ends option parsing, so a directory whose name begins with
   `-` can be passed as `jcd -- -dirname`.
8. Unknown options are usage errors (exit code `2`).

## 9. Terminal requirements

- Minimum size: 80 columns × 20 rows. Smaller terminals cause exit code `3`
  with a stderr message.
- The terminal is restored on every exit path (normal quit, error, signal).

## 10. Configuration

An optional JSON file configures the interface. It is resolved in this order:

1. `--config <path>`, `-c <path>`, or `--config=<path>`
2. the `JCD_CONFIG` environment variable
3. `<UserConfigDir>/jcd/config.json`

A missing file is not an error. Recognised fields:

| Field   | Values | Default |
| ------- | ------ | ------- |
| `theme` | `default`, `mono` | `default` |
| `sort`  | `name_asc`, `name_desc`, `date_asc`, `date_desc`, `size_asc`, `size_desc` | `name_asc` |
| `mouse` | boolean | `false` |

An invalid file, theme, or sort value is a usage error (exit code `2`).

When `mouse` is enabled, the mouse wheel moves the selection (and scrolls the
help screen) and left-clicking an entry selects it. Mouse support is off by
default so the terminal's native text selection keeps working.

## 11. Themes

Two themes are available:

- `default` — color-coded directories, files, links, and drives.
- `mono` — no colors; relies on bold, underline, and reverse video.

The help screen is a full-screen bordered panel in both themes.

## 12. Quick Access tracking

JCD keeps a single persisted collection of the directories the user has
selected. There are no separate "recent" and "frequent" collections: all views
are derived from this one list.

- A visit is recorded whenever the app exits normally with a reported
  directory (for example Esc, Ctrl+C, or Enter in Quick Access), just before the
  banner and path are written to stdout.
- If the directory is already tracked, its visit count is incremented and its
  last-visited time is updated; otherwise a new entry is created with a visit
  count of `1`.
- At most 100 directories are kept. When the collection exceeds the limit, the
  lowest-ranked entries are removed, favoring favorites, then frequently used
  (higher visit count), then recently used directories.
- Directories that no longer exist are skipped when the popup is built and
  removed when the collection is saved.
- The history is stored separately from the configuration in
  `<UserConfigDir>/jcd/tracked.json` (`%AppData%\jcd\tracked.json` on Windows)
  as `{"directories": [...]}`. Each entry records `path`, `visit_count`,
  `last_visited_utc`, and `favorite`.
- The Favorites section lists every favorite, ordered by recency. Recent lists
  the 10 most recently visited directories. Frequent lists the 10 most visited
  directories, newest first on ties.
