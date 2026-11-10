package fsentry

import "testing"

func TestAttributesMask(t *testing.T) {
	tests := []struct {
		name  string
		attrs Attributes
		want  string
	}{
		{"none", Attributes{}, "---------"},
		{"directory", Attributes{Directory: true}, "d--------"},
		{"symlink", Attributes{Symlink: true}, "-l-------"},
		{"read-only", Attributes{ReadOnly: true}, "--r------"},
		{"hidden", Attributes{Hidden: true}, "---h-----"},
		{"system", Attributes{System: true}, "----s----"},
		{"archive", Attributes{Archive: true}, "-----a---"},
		{"compressed", Attributes{Compressed: true}, "------c--"},
		{"not-indexed", Attributes{NotIndexed: true}, "-------i-"},
		{"encrypted", Attributes{Encrypted: true}, "--------e"},
		{
			"read-only and hidden",
			Attributes{ReadOnly: true, Hidden: true},
			"--rh-----",
		},
		{
			"all",
			Attributes{
				Directory:  true,
				Symlink:    true,
				ReadOnly:   true,
				Hidden:     true,
				System:     true,
				Archive:    true,
				Compressed: true,
				NotIndexed: true,
				Encrypted:  true,
			},
			"dlrhsacie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.attrs.Mask(); got != tt.want {
				t.Errorf("Mask() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDriveTypeLabel(t *testing.T) {
	tests := []struct {
		drive DriveType
		want  string
	}{
		{DriveUnknown, " (Unknown)"},
		{DriveRemovable, " (Removable drive)"},
		{DriveFixed, " (Hard drive)"},
		{DriveRemote, " (Network drive)"},
		{DriveCDROM, " (CD/DVD)"},
		{DriveRAMDisk, " (RAM drive)"},
		{DriveType(99), " (Unknown)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.drive.Label(); got != tt.want {
				t.Errorf("Label() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKindString(t *testing.T) {
	tests := map[Kind]string{
		KindFile:    "file",
		KindDir:     "dir",
		KindSymlink: "symlink",
		KindDrive:   "drive",
		KindOther:   "other",
	}
	for kind, want := range tests {
		if got := kind.String(); got != want {
			t.Errorf("Kind(%d).String() = %q, want %q", kind, got, want)
		}
	}
}

func TestEntryIsDirLike(t *testing.T) {
	tests := []struct {
		kind Kind
		want bool
	}{
		{KindFile, false},
		{KindDir, true},
		{KindSymlink, false},
		{KindDrive, true},
		{KindOther, false},
	}
	for _, tt := range tests {
		e := Entry{Kind: tt.kind}
		if got := e.IsDirLike(); got != tt.want {
			t.Errorf("Kind %v IsDirLike() = %v, want %v", tt.kind, got, tt.want)
		}
	}
}
