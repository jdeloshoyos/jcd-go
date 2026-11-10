//go:build !windows

package platform

import "jcd-go/internal/fsentry"

// Drives returns the filesystem root. The drive-selection screen is a Windows
// concept; on Unix the root is the single "drive".
func Drives() ([]fsentry.Entry, error) {
	return []fsentry.Entry{{
		Name:  "/",
		Path:  "/",
		Kind:  fsentry.KindDrive,
		Attrs: fsentry.Attributes{Directory: true},
	}}, nil
}
