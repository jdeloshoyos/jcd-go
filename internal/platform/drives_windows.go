//go:build windows

package platform

import (
	"syscall"
	"unsafe"

	"jcd-go/internal/fsentry"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetLogicalDrives = kernel32.NewProc("GetLogicalDrives")
	procGetDriveTypeW    = kernel32.NewProc("GetDriveTypeW")
)

// Drives enumerates available system drives C: through Z:, mirroring the
// legacy drive-selection screen.
func Drives() ([]fsentry.Entry, error) {
	mask, _, callErr := procGetLogicalDrives.Call()
	if mask == 0 {
		return nil, callErr
	}

	entries := make([]fsentry.Entry, 0, 24)
	for i := 2; i < 26; i++ { // C: .. Z:
		if mask&(1<<uint(i)) == 0 {
			continue
		}

		letter := byte('A' + i)
		root := string(letter) + `:\`
		ptr, err := syscall.UTF16PtrFromString(root)
		if err != nil {
			continue
		}

		driveType, _, _ := procGetDriveTypeW.Call(uintptr(unsafe.Pointer(ptr)))
		if fsentry.DriveType(driveType) == fsentry.DriveInvalid {
			continue
		}

		entries = append(entries, fsentry.Entry{
			Name:      string(letter) + ":",
			Path:      root,
			Kind:      fsentry.KindDrive,
			DriveType: fsentry.DriveType(driveType),
			Attrs:     fsentry.Attributes{Directory: true},
		})
	}
	return entries, nil
}
