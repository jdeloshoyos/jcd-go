package fsentry

import "testing"

func TestHumanFileSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero", 0, "0 B"},
		{"small", 512, "512 B"},
		{"just below one kibibyte", 1023, "1023 B"},
		{"one kibibyte", 1024, "1.0 KiB"},
		{"just above one kibibyte", 1025, "1.0 KiB"},
		{"two kibibytes", 2048, "2.0 KiB"},
		{"one and a half kibibytes", 1536, "1.5 KiB"},
		{"1023 kibibytes", 1023 * 1024, "1023.0 KiB"},
		{"one mebibyte", 1 << 20, "1.0 MiB"},
		{"one mebibyte minus one", (1 << 20) - 1, "1.0 MiB"},
		{"one mebibyte plus one", (1 << 20) + 1, "1.0 MiB"},
		{"one and a half mebibytes", 1_572_864, "1.5 MiB"},
		{"one gibibyte", 1 << 30, "1.0 GiB"},
		{"one gibibyte plus one mebibyte", (1 << 30) + (1 << 20), "1.0 GiB"},
		{"one tebibyte", 1 << 40, "1.0 TiB"},
		{"negative kibibytes", -2048, "-2.0 KiB"},
		{"negative bytes", -512, "-512 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanFileSize(tt.bytes); got != tt.want {
				t.Errorf("HumanFileSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}
