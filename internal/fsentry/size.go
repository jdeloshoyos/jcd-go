package fsentry

import (
	"fmt"
	"math"
)

// sizeUnits are standard IEC binary prefixes.
var sizeUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}

// HumanFileSize converts a byte count into standard IEC units.
//
// Values below 1 KiB are rendered as whole bytes ("512 B"). Larger values are
// divided by 1024 and rendered with one decimal place and an IEC unit
// ("1.5 MiB"). A value that would round up to 1024.0 is promoted to the next
// unit, so 1 MiB - 1 byte renders as "1.0 MiB" rather than "1024.0 KiB".
func HumanFileSize(bytes int64) string {
	if bytes > -1024 && bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	value := float64(bytes)
	unit := 0
	for unit < len(sizeUnits)-1 {
		if math.Abs(math.Round(value*10)/10) < 1024 {
			break
		}
		value /= 1024
		unit++
	}
	return fmt.Sprintf("%.1f %s", value, sizeUnits[unit])
}
