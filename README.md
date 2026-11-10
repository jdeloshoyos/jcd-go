# JCD — Terminal Directory Navigator

A modern Go port of the PowerBASIC **JVA Directory Navigator** (v1.2, 2010).

JCD shows a directory listing in the terminal, lets you navigate the
filesystem with the keyboard, and — on exit — reports the final directory back
to a thin shell wrapper so the wrapper can change the shell's working
directory.

> **Status:** pre-release (`0.2.0`). The interactive TUI is implemented:
> browsing, navigation, sort cycling, search, file opening, a scrollable help
> screen, drive selection, themes, optional mouse support, configuration, and
> the PowerShell `cd` integration.

## Requirements

- Go 1.27 or newer
- PowerShell 5.1+ or PowerShell 7+ (for the wrapper)

## Build

```sh
go build -o jcd ./cmd/jcd
```

On Windows this produces `jcd.exe`.

## Use

A child process cannot change its parent shell's working directory, so a thin
wrapper runs `jcd`, captures the final path from **stdout**, and applies it
with `Set-Location`. The interface itself renders to **stderr**, keeping stdout
clean for the path.

Put the binary on `PATH` (or set `JCD_BIN` to its full path), then add one of
these to your PowerShell profile:

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
invokes; it defaults to `jcd`.

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
| `/` | Enter search mode |
| Enter (search) | Jump to the first match |
| Tab / Shift+Tab (search) | Next / previous match |
| Esc (search) | Cancel search |
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
internal/fsentry/   portable filesystem model and pure logic (scan, sort, search, size)
internal/platform/  OS-specific behaviour (attributes, drives, open-file)
internal/tui/       Bubble Tea model, update, view, and styles
internal/integration/ end-to-end wrapper tests
scripts/            shell wrappers (PowerShell first)
specs/              living product specifications
legacy-app/         original PowerBASIC sources — reference only, never modified
```

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
version is `0.2.0`; the first release candidate will be tagged `1.0.0-rc.1`.
