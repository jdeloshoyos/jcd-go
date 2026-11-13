# JCD — Terminal Directory Navigator

A modern Go port of the PowerBASIC **JVA Directory Navigator** (v1.2, 2010).

JCD shows a directory listing in the terminal, lets you navigate the
filesystem with the keyboard, and — on exit — reports the final directory back
to a thin shell wrapper so the wrapper can change the shell's working
directory.

> **Status:** released (`1.3.1`). The interactive TUI is implemented:
> browsing, navigation, sort cycling, search, file opening, a scrollable help
> screen, drive selection, a Quick Access popup for tracked directories,
> themes, optional mouse support, configuration, and the PowerShell `cd`
> integration.

## Requirements

- Go 1.27 or newer
- PowerShell 5.1+ or PowerShell 7+ (for the wrapper)

## Build

```sh
go build -o jcd ./cmd/jcd
```

On Windows this produces `jcd.exe`.

For a smaller production binary, strip symbols and trim paths:

```powershell
go build -ldflags="-s -w" -trimpath -o .\jcd_bin.exe .\cmd\jcd
```

`-ldflags="-s -w"` removes DWARF symbol tables and debug information, and
`-trimpath` keeps source paths out of the binary. `jcd_bin` is the executable
name the PowerShell wrapper looks for by default.

## Installing from a Release

Official builds are distributed as ZIP archives on
[GitHub Releases](https://github.com/jdeloshoyos/jcd-go/releases). Each archive
contains everything needed to install JCD for the current user.

1. Download `jcd-<version>-windows-amd64.zip` from the latest release.
2. Extract the ZIP.
3. Open PowerShell in the extracted folder.
4. Run the installer:

   ```powershell
   .\install.ps1
   ```

5. Restart PowerShell.
6. Run `jcd`.

The installer does not require administrator privileges and configures the
following for the current user only:

- **Installation directory** — copies `jcd_bin.exe` and `jcd.ps1` to
  `%LOCALAPPDATA%\Programs\JCD`.
- **User PATH** — adds that directory to your user `PATH` (never the system
  `PATH`), skipping the change if it is already present.
- **PowerShell profile** — adds a `jcd` function to `$PROFILE` that runs the
  installed wrapper, skipping the change if it is already present.
- **Unblocked files** — removes the "Mark of the Web" from the installed
  `jcd.ps1` and `jcd_bin.exe` so the wrapper runs without a per-invocation
  `-ExecutionPolicy Bypass`.

The installer is idempotent: run it again to upgrade. It replaces the installed
binary and wrapper and never touches your configuration or history
(`%AppData%\jcd\config.json`, `%AppData%\jcd\tracked.json`).

If your execution policy blocks scripts, run:

```powershell
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

The installer itself may need `-ExecutionPolicy Bypass` when the ZIP was
downloaded. The files it installs are unblocked automatically, so the installed
`jcd.ps1` wrapper runs under a normal execution policy such as `RemoteSigned`.

> The profile integration targets `$PROFILE` for the PowerShell host that runs
> the installer, so installing from Windows PowerShell 5.1 and from PowerShell 7
> configure different profiles.

## Use

A child process cannot change its parent shell's working directory, so a thin
wrapper runs `jcd`, captures the banner and final path from **stdout**, and
applies the path with `Set-Location`. The interface itself renders to
**stderr**. On exit the wrapper prints the product banner and a
"Changing current directory to …" confirmation.

The easiest way to set up JCD is to
[install it from a release](#installing-from-a-release). To configure it
manually instead, put `jcd_bin.exe` on `PATH` (or set `JCD_BIN` to its full
path), copy `jcd.ps1` somewhere permanent, and add one of these to your
PowerShell profile:

```powershell
# Option A: a function (recommended; passes arguments reliably)
function jcd { & "C:\tools\jcd\scripts\jcd.ps1" @args }

# Option B: an alias
Set-Alias jcd "C:\tools\jcd\scripts\jcd.ps1"
```

Then:

```powershell
jcd                      # browse the current directory
jcd C:\some\directory    # start somewhere else
jcd --help               # usage
jcd --version            # version
jcd -- -dashdir          # a directory whose name begins with '-'
```

The wrapper must run in your current PowerShell session (via the function or
alias above), not in a separate process. `JCD_BIN` overrides the executable it
invokes; it defaults to `jcd_bin`.

## Configuration

An optional JSON file can select the theme, initial sort, and mouse support.
It is read from `--config <path>`, `JCD_CONFIG`, or the user config directory
(`%AppData%\jcd\config.json` on Windows):

```json
{
  "theme": "default",
  "sort": "name_asc",
  "mouse": false
}
```

- `theme`: `default` (colors) or `mono` (no colors; bold/underline/reverse).
- `sort`: `name_asc`, `name_desc`, `date_asc`, `date_desc`, `size_asc`, or
  `size_desc`.
- `mouse`: `true` enables wheel scrolling and click-to-select. Off by default
  so native terminal text selection keeps working.

A missing config file is fine; an invalid one is a usage error. Sizes are shown
in IEC units (`B`, `KiB`, `MiB`, `GiB`, …).

JCD also records the directories you select so **Quick Access** (`F4`) can offer
favorites, recent, and frequent paths. Favorites appear only in the Favorites
section, while Recent and Frequent list the remaining directories. This history
is stored separately in `%AppData%\jcd\tracked.json`, next to the config file,
so the user-authored configuration is never rewritten. Up to 100 directories
are kept, favoring favorites, then frequently used, then recently used entries.

## Keys

| Key | Action |
| --- | --- |
| Up / Down | Move the cursor |
| PgUp / PgDn | Move one page |
| Home / End | First / last entry |
| Enter | Enter a directory or open a file |
| Backspace | Go to the parent directory (drive list at a drive root) |
| F1 / ? | Toggle the help screen |
| F2 | Cycle sort mode (Name/Date/Size × Asc/Desc) |
| F3 | Enter search mode (typing selects the first match) |
| F4 | Open Quick Access |
| Type a character | Enter search mode seeded with that character |
| Enter (search) | Open the selected entry, then leave search |
| Tab / Shift+Tab (search) | Next / previous match |
| Esc (search) | Cancel search |
| Up / Down (Quick Access) | Move the selection |
| Enter (Quick Access) | Jump to the directory and quit |
| F (Quick Access) | Toggle the selected entry's favorite status |
| Esc (Quick Access) | Close the popup |
| Esc | Quit and report the current directory |

## Test

```sh
go test ./...
go vet ./...
```

## Project layout

```
cmd/jcd/            command entry point, argument handling, output protocol
internal/cli/       start-directory resolution, terminal size, exit codes
internal/config/    optional JSON configuration (theme, sort, mouse)
internal/fsentry/   portable filesystem model and pure logic (scan, sort, search, size)
internal/history/   tracked-directory store for Quick Access (favorites, recent, frequent)
internal/platform/  OS-specific behaviour (attributes, drives, open-file)
internal/tui/       Bubble Tea model, update, view, and styles
internal/integration/ end-to-end wrapper tests
scripts/            shell wrappers, installer, and release tooling
specs/              living product specifications
legacy-app/         original PowerBASIC sources — reference only, never modified
```

## Creating a Release Package

Maintainers create the GitHub Release archive with:

```powershell
.\scripts\build_release.ps1
```

The script reads the version from `internal/version/version.go` (bump that
constant before releasing), builds the stripped production binary, verifies the
binary reports that version, and stages `install.ps1`, `jcd.ps1`, `jcd_bin.exe`,
`readme.txt`, and `LICENSE`. It writes:

```
dist\jcd-<version>-windows-amd64.zip
```

Attach that ZIP to the GitHub Release. It contains everything a user needs to
install JCD: extract it, run `.\install.ps1`, restart PowerShell, and run
`jcd`.

## Specifications

The product is spec-driven. See `specs/` for the overview, requirements,
behaviour, architecture, and legacy-compatibility documents. Specs are updated
alongside the code.

## Legacy app

The original PowerBASIC sources live in `legacy-app/` and are provided purely
as reference material. They are **not** part of the build and must not be
modified.

## Versioning

The project follows [semantic versioning](https://semver.org/). The current
version is `1.3.1`.
