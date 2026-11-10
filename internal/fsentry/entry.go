// Package fsentry provides the filesystem model and pure logic used by the
// directory navigator: entries, attribute masks, sorting, searching, and
// directory scanning.
//
// The package has no terminal or UI dependencies. OS-specific attribute
// extraction is injected via an AttrFunc so that this package stays portable
// and unit-testable on every platform.
package fsentry

import "time"

// Kind classifies a filesystem entry.
type Kind int

const (
	KindFile Kind = iota
	KindDir
	KindSymlink
	KindDrive
	KindOther
)

// String returns a human-readable label for the kind.
func (k Kind) String() string {
	switch k {
	case KindFile:
		return "file"
	case KindDir:
		return "dir"
	case KindSymlink:
		return "symlink"
	case KindDrive:
		return "drive"
	default:
		return "other"
	}
}

// Attributes is the portable attribute model rendered as the legacy
// nine-character mask.
type Attributes struct {
	Directory  bool
	Symlink    bool
	ReadOnly   bool
	Hidden     bool
	System     bool
	Archive    bool
	Compressed bool
	NotIndexed bool
	Encrypted  bool
}

// Mask renders the legacy attribute mask in the order d l r h s a c i e.
// Attributes that are not set are rendered as '-'.
func (a Attributes) Mask() string {
	mask := []byte("---------")
	set := func(pos int, on bool, ch byte) {
		if on {
			mask[pos] = ch
		}
	}
	set(0, a.Directory, 'd')
	set(1, a.Symlink, 'l')
	set(2, a.ReadOnly, 'r')
	set(3, a.Hidden, 'h')
	set(4, a.System, 's')
	set(5, a.Archive, 'a')
	set(6, a.Compressed, 'c')
	set(7, a.NotIndexed, 'i')
	set(8, a.Encrypted, 'e')
	return string(mask)
}

// DriveType mirrors the Windows GetDriveType values used by the legacy app.
type DriveType int

const (
	DriveUnknown   DriveType = 0
	DriveInvalid   DriveType = 1
	DriveRemovable DriveType = 2
	DriveFixed     DriveType = 3
	DriveRemote    DriveType = 4
	DriveCDROM     DriveType = 5
	DriveRAMDisk   DriveType = 6
)

// Label returns the descriptive suffix shown next to a drive.
func (d DriveType) Label() string {
	switch d {
	case DriveRemovable:
		return " (Removable drive)"
	case DriveFixed:
		return " (Hard drive)"
	case DriveRemote:
		return " (Network drive)"
	case DriveCDROM:
		return " (CD/DVD)"
	case DriveRAMDisk:
		return " (RAM drive)"
	default:
		return " (Unknown)"
	}
}

// Entry is a single item in a directory listing.
type Entry struct {
	Name      string
	Path      string
	Kind      Kind
	Size      int64
	ModTime   time.Time
	Attrs     Attributes
	DriveType DriveType
}

// IsDirLike reports whether the entry behaves like a directory for sorting and
// navigation (directories and drives).
func (e Entry) IsDirLike() bool {
	return e.Kind == KindDir || e.Kind == KindDrive
}
