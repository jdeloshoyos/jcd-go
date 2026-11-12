# JCD — Requirements

- **Status:** Draft
- **Version:** 1.2.0
- **Last updated:** 2026-11-11

Requirement IDs are stable references used by the test suite and by
[`compatibility.md`](compatibility.md).

## 1. Functional requirements

### 1.1 Startup and arguments

| ID            | Requirement |
| ------------- | ----------- |
| `FR-START-1`  | The app accepts an optional target directory argument. With no argument it starts in the current working directory. |
| `FR-START-2`  | Relative arguments are resolved against the current working directory. |
| `FR-START-3`  | An argument that is not an existing directory produces a clear error on stderr and a non-zero exit code. |
| `FR-START-4`  | Trailing path separators are tolerated. |
| `FR-START-5`  | The app requires a terminal of at least 80 columns × 20 rows; otherwise it exits with a clear error. |
| `FR-START-6`  | The app supports `--version` / `-v` and `--help` / `-h`. |
| `FR-START-7`  | A leading `--` ends option parsing, so directories whose names begin with `-` can be passed. Unknown options are usage errors. |
| `FR-START-8`  | Paths containing spaces, Unicode, and shell-special characters are supported. |

### 1.2 Listing

| ID           | Requirement |
| ------------ | ----------- |
| `FR-LIST-1`  | The listing shows all files and subdirectories, including hidden and system entries. |
| `FR-LIST-2`  | The listing reports the number of directories, the number of files, and the total size of regular files. |
| `FR-LIST-3`  | An empty directory is displayed with zero entries and no errors. |
| `FR-LIST-4`  | Each entry exposes name, kind (file/directory/symlink/drive), size, modification time, and attributes. |

### 1.3 Navigation

| ID          | Requirement |
| ----------- | ----------- |
| `FR-NAV-1`  | Up/Down moves the cursor and scrolls the view when needed. |
| `FR-NAV-2`  | PgUp/PgDn move the cursor by one visible page. |
| `FR-NAV-3`  | Home/End jump to the first/last entry. |
| `FR-NAV-4`  | Enter descends into the selected directory. |
| `FR-NAV-5`  | Backspace moves to the parent directory and selects its first entry. |
| `FR-NAV-6`  | Backspace at a Windows drive root opens the drive-selection screen. |
| `FR-NAV-7`  | Enter on the drive-selection screen navigates to the selected drive root. |

### 1.4 Sorting

| ID           | Requirement |
| ------------ | ----------- |
| `FR-SORT-1`  | F2 cycles through six sort modes: Name/Date/Size × Ascending/Descending. |
| `FR-SORT-2`  | Directories (and drives) always sort before files, in every mode. |
| `FR-SORT-3`  | Name comparisons are case-insensitive. |
| `FR-SORT-4`  | Date and size ties are broken by ascending name. |
| `FR-SORT-5`  | The active sort mode is shown in the status bar. |

### 1.5 Search

| ID           | Requirement |
| ------------ | ----------- |
| `FR-SRCH-1`  | F3 enters an explicit search-input mode; typing any visible, printable character not bound to another action also enters it, seeded with that character. |
| `FR-SRCH-2`  | Matching is a case- and diacritic-insensitive substring match anywhere in the entry name: `ima` matches `imágenes`, `imá`, `imà`, and `ïma`. |
| `FR-SRCH-3`  | Typing moves the selection to the first match; Tab cycles through matches, wrapping around. |
| `FR-SRCH-4`  | Esc cancels search mode without changing the selection. |
| `FR-SRCH-5`  | A search with no matches leaves the selection unchanged and shows a zero match count. |
| `FR-SRCH-6`  | The current search pattern is shown in the footer while searching. |
| `FR-SRCH-7`  | Enter opens the selected entry (entering a directory or launching a file) and exits search mode. |
| `FR-SRCH-8`  | The matched substring is highlighted in every matching row while searching. |

### 1.6 Display

| ID           | Requirement |
| ------------ | ----------- |
| `FR-DISP-1`  | Row 1 shows the current path, truncated when it exceeds the width. |
| `FR-DISP-2`  | Row 2 shows directory/file counts and total size, plus the position and sort mode. |
| `FR-DISP-3`  | Entry rows show the attribute mask, modification time, name, and (for files) human-readable size. |
| `FR-DISP-4`  | Directories, files, symlinks, and drives are visually distinguishable. |
| `FR-DISP-5`  | The selected entry is clearly highlighted. |
| `FR-DISP-6`  | The footer shows the primary key hints. |
| `FR-DISP-7`  | Sizes use standard IEC units (`B`, `KiB`, `MiB`, `GiB`, …). |

### 1.7 Help

| ID           | Requirement |
| ------------ | ----------- |
| `FR-HELP-1`  | F1 (or `?`) opens a scrollable help screen. |
| `FR-HELP-2`  | The help screen documents keys, search, sorting, and the attribute mask. |
| `FR-HELP-3`  | Esc or F1 closes help and restores the previous view. |

### 1.8 File opening

| ID           | Requirement |
| ------------ | ----------- |
| `FR-OPEN-1`  | Enter on a regular file opens it with the operating system's default handler. |
| `FR-OPEN-2`  | A failure to open a file is reported without terminating the app. |

### 1.9 Exit and shell integration

| ID           | Requirement |
| ------------ | ----------- |
| `FR-EXIT-1`  | Esc quits the app and restores the terminal. |
| `FR-EXIT-2`  | On success the app prints a banner line (`JCD Terminal Directory Navigator v<version>`) followed by the resolved target path to stdout and exits `0`. |
| `FR-EXIT-3`  | On failure the app prints a diagnostic to stderr and exits non-zero. |
| `FR-EXIT-4`  | The output contains no UI escape sequences or decorations. |
| `FR-EXIT-5`  | A PowerShell wrapper prints the banner and a "Changing current directory to …" line, then applies the returned path with `Set-Location`. |

### 1.10 Configuration and mouse

| ID            | Requirement |
| ------------- | ----------- |
| `FR-CONF-1`   | The app reads an optional JSON configuration from `--config`, `JCD_CONFIG`, or `<UserConfigDir>/jcd/config.json`. |
| `FR-CONF-2`   | The theme is selectable (`default` or `mono`). |
| `FR-CONF-3`   | The initial sort mode is configurable. |
| `FR-CONF-4`   | Mouse support is configurable and off by default. |
| `FR-CONF-5`   | An invalid configuration file, theme, or sort value is a usage error. |
| `FR-MOUSE-1`  | With mouse enabled, the wheel moves the selection and scrolls the help screen. |
| `FR-MOUSE-2`  | With mouse enabled, left-clicking an entry selects it. |

### 1.11 Quick Access

| ID         | Requirement |
| ---------- | ----------- |
| `FR-QA-1`  | F4 opens a single scrollable Quick Access popup from browsing, grouped into Favorites, Recent, and Frequent sections. |
| `FR-QA-2`  | Every normal exit that reports a directory records a visit: existing entries increment their visit count and timestamp; new entries start at `1`. |
| `FR-QA-3`  | The tracked collection is capped at 100 directories, evicting the lowest-ranked entries and favoring favorites, then frequent, then recent entries. |
| `FR-QA-4`  | Favorites shows all favorites; Recent shows the 10 most recently visited; Frequent shows the 10 most visited. A directory may appear in several sections but not twice in one. |
| `FR-QA-5`  | Up/Down move between directory entries, skipping section headers; PgUp/PgDn and Home/End also move the selection. |
| `FR-QA-6`  | Enter jumps to the selected directory and quits, so the wrapper changes the shell directory. |
| `FR-QA-7`  | Esc closes the popup without navigating; F toggles the selected entry's favorite status and refreshes the list immediately. |
| `FR-QA-8`  | Favorite entries show a `★` indicator; non-favorites do not. |
| `FR-QA-9`  | Long paths are middle-truncated at segment boundaries, preserving the beginning and end and replacing removed middle segments with `...`. |
| `FR-QA-10` | The tracked history is persisted separately from the configuration in `<UserConfigDir>/jcd/tracked.json`. |

### 1.12 Installation and distribution

| ID            | Requirement |
| ------------- | ----------- |
| `FR-INST-1`   | Releases are distributed as a ZIP archive containing `install.ps1`, `jcd_bin.exe`, and `jcd.ps1` at the archive root. |
| `FR-INST-2`   | `install.ps1` locates `jcd_bin.exe` and `jcd.ps1` next to itself via `$PSScriptRoot` and fails with a clear error if either is missing. |
| `FR-INST-3`   | The installer installs for the current user only, into `%LOCALAPPDATA%\Programs\JCD`, creating the directory if needed; it never requires administrator privileges. |
| `FR-INST-4`   | The installer copies `jcd_bin.exe` and `jcd.ps1` into the installation directory, replacing existing files on upgrade. |
| `FR-INST-5`   | The installer adds the installation directory to the **User** `PATH` only, never the system `PATH`, without creating duplicates, and is safe to run repeatedly. |
| `FR-INST-6`   | The installer adds the `jcd` integration function to the current user's `$PROFILE`, creating missing directories and files, preserving existing contents, avoiding duplicates, and is safe to run repeatedly. |
| `FR-INST-7`   | The installer never removes user data; `%AppData%\jcd\config.json` and `%AppData%\jcd\tracked.json` are preserved across upgrades. |
| `FR-INST-8`   | Already-configured conditions (existing installation directory, PATH entry, or profile block) are reported as informational, not as errors. |
| `FR-REL-1`    | `build_release.ps1` builds the production binary, stages the distribution files, and produces `dist\jcd-<version>-windows-amd64.zip` whose contents match the layout required by the installer. |
| `FR-REL-2`    | The release build fails immediately if the production build fails. |
| `FR-REL-3`    | The release version is read from `internal/version/version.go`, so the archive name always matches the version the shipped binary reports. |

## 2. Non-functional requirements

| ID            | Requirement |
| ------------- | ----------- |
| `NFR-1`       | Single self-contained binary; no runtime dependencies beyond the OS. |
| `NFR-2`       | Portable core: filesystem logic is platform-agnostic; OS specifics are isolated in `internal/platform`. |
| `NFR-3`       | The core package has no terminal or UI dependencies so it can be unit-tested directly. |
| `NFR-4`       | Listings of directories with thousands of entries remain responsive. |
| `NFR-5`       | All keys behave consistently on Windows terminals (including Windows Terminal and conhost). |
| `NFR-6`       | The build and test suite run on Windows, Linux, and macOS in CI. |
| `NFR-7`       | Behaviour is covered by unit tests, TUI-level regression tests, and golden snapshots. |
| `NFR-8`       | Specs are updated whenever behaviour changes. |
| `NFR-9`       | Benchmarks cover sorting, searching, and scanning large listings. |

## 3. Acceptance criteria (initial milestone set)

- `go build ./...`, `go vet ./...`, and `go test ./...` pass on all three CI
  platforms.
- `internal/fsentry` has table-driven tests for scan, sort, search, size
  conversion, and attribute masks.
- The interactive TUI (later milestone) reproduces the full key set above.
- A PowerShell integration test verifies that the wrapper changes directory.
