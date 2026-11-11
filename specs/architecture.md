# JCD — Architecture

- **Status:** Draft
- **Version:** 1.2.0
- **Last updated:** 2026-11-11

## 1. Technology stack

| Concern        | Choice |
| -------------- | ------ |
| Language       | Go 1.27.3 |
| TUI runtime    | Bubble Tea v2 (`charm.land/bubbletea/v2`, v2.1.x) |
| Styling/layout | Lip Gloss v2 (`charm.land/lipgloss/v2`) |
| Components     | Bubbles v2 (`charm.land/bubbles/v2`): `textinput`, `viewport` |
| CLI/config     | Cobra and Huh are candidates for a later milestone, not initial |

Bubble Tea v2 is chosen because it provides the Elm-style event loop, an
alternate-screen cell renderer, resize handling, and mouse support, while
leaving the listing layout fully under our control. Lower-level libraries
(`tcell`, `termbox`, `gocui`) would require re-implementing scrolling, input
parsing, and redraw.

## 2. Package layout

```
cmd/jcd/              command entry point: arguments, run loop, output protocol
internal/cli/         start-directory resolution, argument parsing, exit codes
internal/config/      optional JSON configuration (theme, sort, mouse)
internal/fsentry/     portable model + pure logic (no TUI, no OS-specific code)
internal/history/     tracked-directory store for Quick Access
internal/tui/         Bubble Tea model, update, view, styles, help
internal/platform/    the only package with build-tagged OS-specific code
internal/version/     semantic version, single source of truth
internal/integration/ end-to-end tests for the shell wrapper contract
scripts/              shell wrappers (PowerShell first)
specs/                living specifications
testdata/             fixture trees for tests
```

`internal/shellout` is intentionally absent: path emission and exit codes are
small enough to live in `cmd/jcd` and `internal/cli`.

## 3. Dependency rules

```
cmd/jcd  ──────────────►  internal/tui  ──────►  internal/fsentry
   │                          │                     ▲
   │                          └──────►  internal/history
   └──────────►  internal/platform  ──────────────┘
```

- `internal/fsentry` imports **only the standard library**. It must stay free
  of terminal and OS-specific dependencies.
- `internal/history` also imports only the standard library. It owns the
  tracked-directory store used by Quick Access and has no terminal
  dependencies.
- `internal/platform` imports `internal/fsentry` to construct entries and
  attribute models. It is the only package with `//go:build` OS tags.
- `internal/tui` imports `internal/fsentry` for the model, `internal/history`
  for Quick Access, and `internal/platform` for OS actions (open file, drive
  list). It does not import `cmd`.
- `cmd/jcd` wires everything together and owns process exit. It loads the
  history store, passes it to the TUI, records the final selection, and saves
  the store.

Because OS-specific attribute extraction is injected into `fsentry.ScanDir` as
a function (`fsentry.AttrFunc`), the platform package implements the interface
without creating an import cycle.

## 4. Core model (`internal/fsentry`)

- `Entry` — name, path, kind, size, modification time, attributes, drive type.
- `Attributes` — portable attribute model with `Mask()` rendering.
- `Kind` — file, directory, symlink, drive, other.
- `DriveType` — mirrors the Windows `GetDriveType` values.
- `SortMode` — the six legacy modes, with `String`, `Next`, and `Normalize`.
- `Listing` — scan result with entries, directory count, and total file bytes.
- `ScanDir(dir, AttrFunc, SortMode)` — reads a directory, builds entries,
  computes summary counts, and sorts.
- `Sort`, `Matches`, `CountMatches`, `NthMatch`, `HumanFileSize` — pure
  helpers.

`TotalBytes` sums **regular files only**, so it is meaningful and stable across
platforms (on Windows, directory sizes are zero, so this matches the legacy
value).

## 5. TUI model (`internal/tui`)

Follows the Elm architecture:

- `Model` holds the current `Listing`, cursor/offset, sort mode, search query,
  active theme, mouse flag, the tracked-directory store, and the active state
  (`browsing`, `searching`, `help`, `drive-select`, `quick-access`).
- `Update` handles key, mouse, and resize messages, switching on state.
- `View` renders the header, entry viewport, and footer with Lip Gloss and
  sets the terminal mouse mode when enabled.
- The listing is a **hand-rolled viewport** for pixel-faithful columns. Help
  uses `bubbles/viewport`; search uses `bubbles/textinput`.
- Quick Access is a centered bordered popup overlaid on the directory listing
  (the listing stays visible around it), flattened into a single list of
  section and entry rows. Navigation skips section rows; long paths are
  middle-truncated at segment boundaries (`truncateMiddlePath`).
- `Theme` bundles all styles. `default` is color-coded; `mono` uses bold,
  underline, and reverse video only.

Terminal actions (alternate screen) are declared on `tea.View` in v2.

### 5.1 Rendering and streams

The interface renders to **stderr** (`tea.WithOutput(os.Stderr)`). The shell
wrapper captures stdout to obtain the final directory, so stdout must carry
exactly one line: the resolved path. Input still comes from the terminal;
Bubble Tea v2 opens the TTY automatically when stdin is not a terminal.

The wrapper is invoked in the caller's session and applies the path with
`Set-Location`. The Go process never changes its parent's directory directly,
because that is not portable (or possible) for a child process.

### 5.2 Tracked directories (`internal/history`)

`history.Store` owns a single in-memory collection and its JSON file. It
exposes `Load`, `Save`, `Record`, `ToggleFavorite`, `PruneMissing`, and
`Sections`; the TUI derives the Favorites/Recent/Frequent views from
`Sections`. Recording deduplicates case-insensitively on Windows, updates the
visit count and timestamp, and trims to the 100 best-ranked entries. The store
is independent of the user-authored configuration file.

## 6. Platform abstraction (`internal/platform`)

| File                 | Responsibility |
| -------------------- | -------------- |
| `attrs_windows.go`   | Map Win32 file attributes to `fsentry.Attributes`. |
| `attrs_unix.go`      | Map Unix mode bits/dotfiles to `fsentry.Attributes`. |
| `drives_windows.go`  | Enumerate `C:`–`Z:` via `GetLogicalDrives`/`GetDriveTypeW`. |
| `drives_unix.go`     | Return `/` as the single root. |
| `open_windows.go`    | Open a file with `cmd /c start`. |
| `open_unix.go`       | Open a file with `open` (macOS) or `xdg-open` (Linux). |

## 7. Testing strategy

- **Unit tests** for `fsentry`: table-driven coverage of scan (empty, hidden,
  symlinks, Unicode, special characters, counts, totals), all six sort modes,
  search matching/nth-match, IEC size formatting, and attribute masks.
- **Config tests** for defaults, missing/partial/invalid files, and sort parsing.
- **History tests** for recording/upsert, Windows case-insensitive dedup,
  favorite toggling, eviction ranking, missing-directory pruning, section
  derivation, and JSON round-tripping.
- **Platform tests** for attribute mapping: raw Win32 bitmasks plus real
  read-only/hidden/symlink files on Windows; dotfiles, permissions, and
  symlinks on Unix.
- **TUI tests** driven through `Update`: every key, state transitions, error
  paths, Quick Access navigation/favorite toggling/enter-to-jump, and a
  Program-level smoke test of the real Bubble Tea loop.
- **Golden tests** for `View()` covering browsing, search, help, drives, Quick
  Access, and small terminals.
- **Integration tests** for the PowerShell wrapper: success, failure, and
  special/Unicode paths.
- **Benchmarks** for sorting, searching, and scanning large listings.
- **CI matrix** on `windows-latest`, `ubuntu-latest`, and `macos-latest`, with
  a formatting gate and a race-detector run on Linux.

## 8. Build and release

- `go build ./...` and `go test ./...` must pass on all CI platforms.
- `build_prd.ps1` produces a smaller Windows binary (`jcd_bin.exe`) with
  `-trimpath` and `-ldflags="-s -w"`, stripping DWARF symbols and debug info.
- Semantic versioning; the current release is `1.2.0`.

## 9. Deliberate design decisions

- **Hand-rolled listing** instead of `bubbles/list` for exact column control.
- **Explicit search mode** instead of the legacy silent timed search.
- **Clean stdout protocol** (a banner line plus the path line) instead of the
  `*`-delimited format, which breaks on paths containing `*`.
- **UI on stderr, result on stdout** so a capturing wrapper receives only the
  banner and path.
- **Windows-first**; the platform layer keeps Unix support a small addition.
- **Quick Access uses a single tracked collection** and derives Favorites,
  Recent, and Frequent views from it, avoiding separate "recent" and "frequent"
  stores. The history lives in its own JSON file so user configuration is never
  rewritten.
