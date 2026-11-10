package fsentry

import "testing"

func TestHumanFileSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero", 0, "0 bytes"},
		{"small", 512, "512 bytes"},
		{"exactly one kibibyte (legacy boundary)", 1024, "1024 bytes"},
		{"just above one kibibyte", 1025, "1.0 KB"},
		{"two kibibytes", 2048, "2.0 KB"},
		{"one mebibyte (legacy boundary)", 1 << 20, "1024.0 KB"},
		{"just above one mebibyte", (1 << 20) + 1, "1.0 MB"},
		{"one and a half megabytes", 1_500_000, "1.4 MB"},
		{"one gibibyte", 1 << 30, "1024.0 MB"},
		{"one gibibyte plus", (1 << 30) + (1 << 20), "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanFileSize(tt.bytes); got != tt.want {
				t.Errorf("HumanFileSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}
