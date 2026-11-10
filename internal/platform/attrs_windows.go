//go:build windows

package platform

import (
	"os"
	"syscall"

	"jcd-go/internal/fsentry"
)

// Windows file attribute bits (winnt.h).
const (
	fileAttributeReadOnly   = 0x00000001
	fileAttributeHidden     = 0x00000002
	fileAttributeSystem     = 0x00000004
	fileAttributeDirectory  = 0x00000010
	fileAttributeArchive    = 0x00000020
	fileAttributeReparse    = 0x00000400
	fileAttributeCompressed = 0x00000800
	fileAttributeNotIndexed = 0x00002000
	fileAttributeEncrypted  = 0x00004000
)

// Attrs extracts the portable attribute model from Windows file attributes.
func Attrs(_ string, info os.FileInfo) fsentry.Attributes {
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		// Fall back to the generic mode information.
		return fsentry.Attributes{
			Directory: info.IsDir(),
			Symlink:   info.Mode()&os.ModeSymlink != 0,
		}
	}

	f := data.FileAttributes
	return fsentry.Attributes{
		Directory:  f&fileAttributeDirectory != 0,
		Symlink:    f&fileAttributeReparse != 0,
		ReadOnly:   f&fileAttributeReadOnly != 0,
		Hidden:     f&fileAttributeHidden != 0,
		System:     f&fileAttributeSystem != 0,
		Archive:    f&fileAttributeArchive != 0,
		Compressed: f&fileAttributeCompressed != 0,
		NotIndexed: f&fileAttributeNotIndexed != 0,
		Encrypted:  f&fileAttributeEncrypted != 0,
	}
}
