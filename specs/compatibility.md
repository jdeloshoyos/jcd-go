# JCD — Legacy Compatibility

- **Status:** Draft
- **Version:** 0.1.0
- **Last updated:** 2026-11-10

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
| `HumanFileSize` (divide while > 1024) | Same algorithm, preserved boundary quirk. |
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
discoverable and removes the hidden timer.

### 3.2 Output protocol
Legacy: `1.2*OK*C:\some\dir`, parsed by splitting on `*`. A path containing `*`
(valid on POSIX systems) would break parsing. The port prints exactly the
resolved path on stdout and signals success with exit code `0`; errors go to
stderr.

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

## 4. Parity checklist

- [x] Six sort modes, directories first
- [x] Case-insensitive substring search, nth match
- [x] Nine-character attribute mask in legacy order
- [x] `HumanFileSize` units and rounding
- [ ] Full key set in the TUI (later milestone)
- [ ] Drive-selection screen (later milestone)
- [ ] Help screen (later milestone)
- [ ] PowerShell wrapper end-to-end (later milestone)

## 5. Out of scope for parity

- Reproducing the legacy screen-buffer drawing technique.
- The `*`-delimited output format.
- The unused `KillParent`/process-listing code.
- Byte-for-byte identical colors; the port uses a themed palette.
