# JCD — Product Overview

- **Status:** Draft
- **Version:** 1.3.0
- **Last updated:** 2026-11-11

## 1. Vision

JCD is a fast, keyboard-driven terminal directory navigator. It displays the
contents of a directory, lets the user move through the filesystem, and — on
exit — hands the final directory back to the calling shell so the shell can
change its working directory.

It is a modern Go port of the PowerBASIC **JVA Directory Navigator** (v1.2,
2010). The port preserves the original interaction model while taking
advantage of a cross-platform runtime and a modern TUI stack.

## 2. Goals

- **Functional parity** with the legacy navigator: listing, navigation,
  sorting, search, help, drive selection, and file opening.
- **Quick Access**: jump to frequently or recently used directories from
  anywhere in the filesystem with a single key (`F4`).
- **Single static binary** for simple deployment.
- **Simple Windows installation**: a self-contained ZIP release and a one-step,
  per-user installer (`install.ps1`), distributed through GitHub Releases.
- **Windows first**, with architecture that keeps Linux and macOS support
  cheap to add.
- **Spec-driven and well tested**: behaviour is described in `specs/` and
  protected by unit, integration, and regression tests.
- **Semantic versioning**, currently at the `1.3.0` release.

## 3. Non-goals (initial)

- File management operations (copy, move, delete, rename).
- File content preview or editing.
- A graphical configuration UI.
- MSI, winget, or Scoop packaging.
- Machine-wide or administrator installation.
- Installing directly from a source checkout.

## 4. Target platforms

| Platform | Status          | Notes                                    |
| -------- | --------------- | ---------------------------------------- |
| Windows  | Primary         | Drive selection, Windows attribute mask. |
| Linux    | Best effort     | Root as the single "drive".              |
| macOS    | Best effort     | Root as the single "drive".              |

## 5. Target users and usage

A developer or power user who wants to browse directories quickly from an
existing terminal session and then continue working in the chosen directory.
The app is launched through a shell wrapper (initially PowerShell) so the
parent shell can apply the final working directory.

```
PS> jcd [directory]
```

## 6. Related documents

- [`requirements.md`](requirements.md) — functional and non-functional
  requirements.
- [`behavior.md`](behavior.md) — interaction semantics and the output protocol.
- [`architecture.md`](architecture.md) — package layout, stack, and testing.
- [`compatibility.md`](compatibility.md) — legacy mapping and deliberate deltas.
