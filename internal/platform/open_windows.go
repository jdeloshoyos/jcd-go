//go:build windows

package platform

import "os/exec"

// OpenFile opens path with the operating system's default handler.
func OpenFile(path string) error {
	// The empty string is the window title argument required by `start`; it
	// keeps paths containing spaces from being treated as the title.
	return exec.Command("cmd", "/c", "start", "", path).Start()
}
