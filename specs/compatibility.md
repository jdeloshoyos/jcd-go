# JCD — Legacy Compatibility

- **Status:** Draft
- **Version:** 1.3.0
- **Last updated:** 2026-11-11

This document maps the PowerBASIC legacy application (`legacy-app/`) to the Go
port and records every deliberate behavioural difference.

## 1. Legacy components

| Legacy file              | Role |
| ------------------------ | ---- |
| `jcd_ps.ps1`             | PowerShell wrapper: runs `jcd_bin.exe p <args>`, parses `version*STATUS*result`, and `Set-Location`s. |
| `jcd.BAS`                | Main program: argument handling, event loop, key bindings, output. |
| `DirectoryInfo_class.INC`| Directory scanning, sorting, search, attributes, drive enumeration. |
| `Console_class.INC`      | Console drawing: header, footer, entries, help, cursor. |
| `pb_proc.inc`            | Unused process helpers from an earlier `KillParent` approach. |

## 2. Behaviour mapping

| Legacy behaviour | Go port |
| ---------------- | ------- |
| Start dir from `command$(2)` with Windows path fixups | `cmd/jcd` path resolution (see `behavior.md` §8). |
| Require console ≥ 80×20 | Same requirement, exit code `3`. |
| `ScanDir` with NORMAL+HIDDEN+SYSTEM+SUBDIR | `fsentry.ScanDir` via `os.ReadDir` (all entries). |
| Six sort modes, dirs first | `fsentry.SortMode` / `Sort`. |
| Type-to-search with 1 s reset | **Explicit `/` search mode** (deliberate change). |
| Tab cycles search matches | Same, wrapping. |
| Attribute mask `d l r h s a c i e` | `fsentry.Attributes.Mask()` (identical order). |
| `HumanFileSize` (divide while > 1024) | Standard IEC units (deliberate change; see §3.6). |
| Drive screen `C:`–`Z:` | `platform.Drives()` (Windows); `/` elsewhere. |
| Enter opens file via `ShellExecute` | `platform.OpenFile()` (`cmd /c start`, `open`, `xdg-open`). |
| `version*OK*path` output | **Single path line on stdout + exit code** (deliberate change). |
| `^`-escaping for cmd.exe batch | Not needed with the clean protocol. |
| `KillParent`/`pb_proc.inc` | Removed; not used since legacy v1.07. |

## 3. Deliberate deltas

### 3.1 Search UX
The legacy app started a search on any printable key, with a silent one-second
reset window, and `Tab` to cycle. The port uses an explicit `/` search mode
with a visible input, `Enter` to jump, and `Tab` to cycle. This is more
discoverable and removes the hidden timer. The port also folds accents before
comparing, so searching `ima` matches `imágenes`.

### 3.2 Output protocol
Legacy: `1.2*OK*C:\some\dir`, parsed by splitting on `*`. A path containing `*`
(valid on POSIX systems) would break parsing. The port prints exactly the
resolved path on stdout and signals success with exit code `0`; errors go to
stderr. The PowerShell wrapper applies the path with `-LiteralPath`, and the
protocol is covered by tests for spaces, Unicode, and shell metacharacters
(`&`, `(`, `)`, `'`, `;`, `,`, `=`), as well as `*` names on POSIX.

### 3.3 Listing rendering
The legacy app painted directly into the console buffer with `locate`/`color`.
The port uses Bubble Tea + Lip Gloss with a hand-rolled viewport, preserving
the column layout while gaining resize handling and mouse support.

### 3.4 Cross-platform
The legacy app is Windows-only. The port isolates OS specifics in
`internal/platform` so Linux and macOS builds work with a reduced feature set
(no drive-selection screen; `/` is the single root).

### 3.5 Total size
Legacy summed the size field of every entry. On Windows, directories report a
size of zero, so this was effectively "regular files". The port sums regular
files explicitly, which keeps the same Windows result and gives a meaningful
value on Unix.

### 3.6 Size units
Legacy rendered `KB`, `MB`, and so on, with a boundary quirk (1 MiB rendered as
`1024.0 KB`). The port uses standard IEC units (`B`, `KiB`, `MiB`, `GiB`, …)
and promotes a unit when the value would round to `1024.0`, so sizes read
correctly at unit boundaries.

### 3.7 Themes, mouse, and configuration
The port adds a monochrome theme, optional mouse support (wheel and
click-to-select), and an optional JSON configuration file. The legacy app had
none of these. Mouse support is off by default so the terminal's native text
selection keeps working.

### 3.8 Quick Access
The port adds a Quick Access popup (F4) that tracks directories the user
selects and groups them into Favorites, Recent, and Frequent sections. The
legacy app had no equivalent. The history is persisted in its own
`tracked.json` file, and long paths are middle-truncated at segment boundaries
so the most distinguishing parts stay visible.

## 4. Parity checklist

- [x] Six sort modes, directories first
- [x] Case- and diacritic-insensitive substring search, nth match
- [x] Nine-character attribute mask in legacy order
- [x] IEC size formatting (deliberate delta)
- [x] Optional configuration (theme, sort, mouse)
- [x] Default and monochrome themes
- [x] Optional mouse wheel and click selection
- [x] Browsing and navigation keys in the TUI
- [x] Sort cycling with F2
- [x] Explicit search mode with Tab cycling
- [x] File opening with the default handler
- [x] PowerShell wrapper end-to-end
- [x] Scrollable help screen (F1 or `?`)
- [x] Drive-selection screen (Backspace at a drive root)
- [x] Quick Access popup with tracked directories (F4)
- [x] Full key set including help and drives
- [x] Attribute mapping parity tests (Win32 bits and real files)
- [x] Edge-case path tests (spaces, Unicode, shell metacharacters, `*` on POSIX)

## 5. Out of scope for parity

- Reproducing the legacy screen-buffer drawing technique.
- The `*`-delimited output format.
- The unused `KillParent`/process-listing code.
- Byte-for-byte identical colors; the port uses a themed palette.
