//go:build !windows

package platform

import (
	"os"
	"path/filepath"

	"jcd-go/internal/fsentry"
)

// Attrs derives the portable attribute model from Unix file metadata.
func Attrs(path string, info os.FileInfo) fsentry.Attributes {
	base := filepath.Base(path)
	return fsentry.Attributes{
		Directory: info.IsDir(),
		Symlink:   info.Mode()&os.ModeSymlink != 0,
		ReadOnly:  info.Mode().Perm()&0o200 == 0,
		Hidden:    len(base) > 0 && base[0] == '.',
	}
}
