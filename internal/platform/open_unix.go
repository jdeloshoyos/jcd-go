//go:build !windows

package platform

import (
	"os/exec"
	"runtime"
)

// OpenFile opens path with the platform's default handler.
func OpenFile(path string) error {
	opener := "xdg-open"
	if runtime.GOOS == "darwin" {
		opener = "open"
	}
	return exec.Command(opener, path).Start()
}
