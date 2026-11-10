package fsentry

import "fmt"

var sizeUnits = []string{"bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

// HumanFileSize converts a byte count into the legacy human-readable form,
// e.g. "1.5 MB".
//
// The boundary behaviour intentionally mirrors the legacy app: the value is
// divided only while it is strictly greater than 1024. As a result, exactly
// 1024 bytes renders as "1024 bytes" and 1 MiB renders as "1024.0 KB". This
// quirk is preserved for parity and can be revisited during release polish.
func HumanFileSize(bytes int64) string {
	value := float64(bytes)
	level := 0
	for value > 1024 && level < len(sizeUnits)-1 {
		value /= 1024
		level++
	}
	if level == 0 {
		return fmt.Sprintf("%d bytes", bytes)
	}
	return fmt.Sprintf("%.1f %s", value, sizeUnits[level])
}
