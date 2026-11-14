// Package cli contains command-line helpers for the jcd entry point.
package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/x/term"
)

// Exit codes used by the command.
const (
	ExitOK       = 0
	ExitError    = 1
	ExitUsage    = 2
	ExitTooSmall = 3
)

// ModeInfo is the mode-selector line printed first on stdout for informational
// output (--help, --version). The PowerShell wrapper detects it to tell
// informational output apart from a directory-change result and relays the
// rest verbatim. Keep it in sync with scripts/jcd.ps1.
const ModeInfo = "Information"

// Action selects what the command should do after parsing arguments.
type Action int

const (
	// ActionRun starts the interactive navigator.
	ActionRun Action = iota
	// ActionVersion prints the version and exits.
	ActionVersion
	// ActionHelp prints usage and exits.
	ActionHelp
)

// Options is the result of parsing the command line.
type Options struct {
	Action     Action
	Dir        string
	ConfigPath string
}

// ErrTooManyArgs is returned when more than one directory is given.
var ErrTooManyArgs = errors.New("too many arguments")

// ParseArgs parses the command line (without the program name).
//
// The first positional argument is the start directory. A leading "--"
// terminates option parsing, so directories whose names begin with "-" can be
// reached with `jcd -- -dirname`.
func ParseArgs(args []string) (Options, error) {
	opts := Options{Action: ActionRun}
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			positional = append(positional, args[i+1:]...)
			i = len(args)

		case arg == "--version" || arg == "-v":
			return Options{Action: ActionVersion}, nil

		case arg == "--help" || arg == "-h":
			return Options{Action: ActionHelp}, nil

		case arg == "--config" || arg == "-c":
			if i+1 >= len(args) {
				return Options{}, fmt.Errorf("option %q requires a path", arg)
			}
			opts.ConfigPath = args[i+1]
			i++

		case strings.HasPrefix(arg, "--config="):
			opts.ConfigPath = strings.TrimPrefix(arg, "--config=")

		case strings.HasPrefix(arg, "-") && arg != "-":
			return Options{}, fmt.Errorf("unknown option %q", arg)

		default:
			positional = append(positional, arg)
		}
	}

	switch len(positional) {
	case 0:
		return opts, nil
	case 1:
		opts.Dir = positional[0]
		return opts, nil
	default:
		return Options{}, ErrTooManyArgs
	}
}

// Usage writes the help text to w.
func Usage(w io.Writer, version string) {
	fmt.Fprintf(w, `jcd %s - terminal directory navigator

Usage:
  jcd [directory]
  jcd --config <path> [directory]
  jcd --version
  jcd --help

Navigate with the arrow keys, Home/End, and PgUp/PgDn. Press Enter to open a
directory or file, Backspace to go up (the drive list at a drive root), F2 to
cycle sort modes, F3 to search (or just start typing), F4 for Quick Access,
and F1 (or ?) for help.

Esc quits and prints the final directory to stdout so a shell wrapper (for
example scripts/jcd.ps1) can change into it.

Configuration:
  An optional JSON file selects the theme, initial sort, and mouse support.
  It is read from --config <path>, JCD_CONFIG, or <UserConfigDir>/jcd/config.json
  (missing files are ignored).
`, version)
}

// ResolveStartDir resolves arg into an absolute, existing directory path.
// An empty arg yields cwd. Relative paths are resolved against cwd.
func ResolveStartDir(arg, cwd string) (string, error) {
	if arg == "" {
		return filepath.Clean(cwd), nil
	}

	// On Windows a path beginning with a single separator is rooted on the
	// current drive; leave UNC paths (\\server\share) untouched.
	if filepath.Separator == '\\' &&
		strings.HasPrefix(arg, `\`) &&
		!strings.HasPrefix(arg, `\\`) {
		if vol := filepath.VolumeName(cwd); vol != "" {
			arg = vol + arg
		}
	}

	if !filepath.IsAbs(arg) {
		arg = filepath.Join(cwd, arg)
	}
	abs, err := filepath.Abs(arg)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)

	info, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("%s is not a valid directory or does not exist", abs)
		}
		return "", fmt.Errorf("%s: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", abs)
	}
	return abs, nil
}

// TerminalSize returns the size of the terminal attached to stderr, which is
// where the interface is rendered.
func TerminalSize() (width, height int, err error) {
	return term.GetSize(os.Stderr.Fd())
}
