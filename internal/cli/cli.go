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
	Action Action
	Dir    string
}

// ErrTooManyArgs is returned when more than one directory is given.
var ErrTooManyArgs = errors.New("too many arguments")

// ParseArgs parses the command line (without the program name).
//
// The first positional argument is the start directory. A leading "--"
// terminates option parsing, so directories whose names begin with "-" can be
// reached with `jcd -- -dirname`.
func ParseArgs(args []string) (Options, error) {
	if len(args) > 0 {
		switch args[0] {
		case "--version", "-v":
			return Options{Action: ActionVersion}, nil
		case "--help", "-h":
			return Options{Action: ActionHelp}, nil
		}
	}

	forcePath := false
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
		forcePath = true
	}

	switch len(args) {
	case 0:
		return Options{Action: ActionRun}, nil
	case 1:
		if !forcePath && strings.HasPrefix(args[0], "-") && args[0] != "-" {
			return Options{}, fmt.Errorf("unknown option %q", args[0])
		}
		return Options{Action: ActionRun, Dir: args[0]}, nil
	default:
		return Options{}, ErrTooManyArgs
	}
}

// Usage writes the help text to w.
func Usage(w io.Writer, version string) {
	fmt.Fprintf(w, `jcd %s - terminal directory navigator

Usage:
  jcd [directory]
  jcd --version
  jcd --help

Navigate with the arrow keys, Home/End, and PgUp/PgDn. Press Enter to open a
directory or file, Backspace to go up (the drive list at a drive root), F2 to
cycle sort modes, / to search, and F1 (or ?) for help.

Esc quits and prints the final directory to stdout so a shell wrapper (for
example scripts/jcd.ps1) can change into it.
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
