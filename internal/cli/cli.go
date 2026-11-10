// Package cli contains command-line helpers for the jcd entry point.
package cli

import (
	"errors"
	"fmt"
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
