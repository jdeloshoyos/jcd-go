package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jcd-go/internal/fsentry"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Theme != "default" {
		t.Errorf("Theme = %q, want default", cfg.Theme)
	}
	if cfg.Sort != "name_asc" {
		t.Errorf("Sort = %q, want name_asc", cfg.Sort)
	}
	if cfg.Mouse {
		t.Error("Mouse = true, want false")
	}
}

func TestLoadEmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != Default() {
		t.Errorf("Load(\"\") = %+v, want defaults", cfg)
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != Default() {
		t.Errorf("Load(missing) = %+v, want defaults", cfg)
	}
}

func TestLoadValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	body := `{"theme":"mono","sort":"size_desc","mouse":true}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Theme != "mono" || cfg.Sort != "size_desc" || !cfg.Mouse {
		t.Errorf("Load = %+v, want mono/size_desc/true", cfg)
	}
}

func TestLoadPartialKeepsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"mouse":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Theme != "default" || cfg.Sort != "name_asc" || !cfg.Mouse {
		t.Errorf("Load = %+v, want defaults with mouse=true", cfg)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"theme":`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load returned nil error for invalid JSON")
	}
}

func TestSortMode(t *testing.T) {
	tests := []struct {
		sort string
		want fsentry.SortMode
	}{
		{"", fsentry.SortNameAsc},
		{"name_asc", fsentry.SortNameAsc},
		{"NAME_ASC", fsentry.SortNameAsc},
		{"name_desc", fsentry.SortNameDesc},
		{"date_asc", fsentry.SortDateAsc},
		{"date_desc", fsentry.SortDateDesc},
		{"size_asc", fsentry.SortSizeAsc},
		{"size_desc", fsentry.SortSizeDesc},
	}

	for _, tt := range tests {
		t.Run(tt.sort, func(t *testing.T) {
			got, err := Config{Sort: tt.sort}.SortMode()
			if err != nil {
				t.Fatalf("SortMode(%q): %v", tt.sort, err)
			}
			if got != tt.want {
				t.Errorf("SortMode(%q) = %v, want %v", tt.sort, got, tt.want)
			}
		})
	}

	if _, err := (Config{Sort: "bogus"}).SortMode(); err == nil {
		t.Error("SortMode(bogus) returned nil error")
	}
}

func TestDefaultPath(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Skip("user config directory is unavailable")
	}
	if !strings.HasSuffix(path, filepath.Join("jcd", "config.json")) {
		t.Errorf("DefaultPath() = %q, want a jcd/config.json suffix", path)
	}
}

func TestHistoryPath(t *testing.T) {
	path := HistoryPath()
	if path == "" {
		t.Skip("user config directory is unavailable")
	}
	if !strings.HasSuffix(path, filepath.Join("jcd", "tracked.json")) {
		t.Errorf("HistoryPath() = %q, want a jcd/tracked.json suffix", path)
	}
	if path == DefaultPath() {
		t.Error("HistoryPath() must not equal DefaultPath()")
	}
}
