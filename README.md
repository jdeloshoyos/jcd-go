# JCD — Terminal Directory Navigator

A modern Go port of the PowerBASIC **JVA Directory Navigator** (v1.2, 2010).

The application shows a directory listing in the terminal, lets you navigate
the filesystem with the keyboard, and — on exit — reports the final directory
back to a thin shell wrapper so the wrapper can change the shell's working
directory.

> **Status:** early development (`0.1.0`). The interactive TUI is not
> implemented yet; the current milestone provides the scaffold and the pure
> filesystem core (`internal/fsentry`).

## Requirements

- Go 1.27 or newer

## Build

```sh
go build ./...
```

A runnable binary will be produced in `cmd/jcd` once the TUI lands:

```sh
go build -o jcd ./cmd/jcd
```

## Test

```sh
go test ./...
go vet ./...
```

## Project layout

```
cmd/jcd/          command entry point
internal/fsentry/ portable filesystem model and pure logic (scan, sort, search, size)
internal/platform/ OS-specific behaviour (attributes, drives, open-file)
internal/tui/     Bubble Tea user interface (later milestone)
scripts/          shell wrappers (PowerShell first)
specs/            living product specifications
legacy-app/       original PowerBASIC sources — reference only, never modified
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

The project follows [semantic versioning](https://semver.org/). During initial
development the version stays at the `0.1.0` placeholder until the first
release candidate.
