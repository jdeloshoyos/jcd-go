// Package config loads the optional jcd configuration file.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"jcd-go/internal/fsentry"
)

// Config is the optional user configuration.
type Config struct {
	Theme string `json:"theme"`
	Sort  string `json:"sort"`
	Mouse bool   `json:"mouse"`
}

// Default returns the built-in configuration.
func Default() Config {
	return Config{Theme: "default", Sort: "name_asc", Mouse: false}
}

// DefaultPath returns the per-user configuration path, or "" when it cannot be
// determined.
func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "jcd", "config.json")
}

// HistoryPath returns the per-user tracked-directory path, or "" when it cannot
// be determined. It deliberately lives next to the config file but is separate
// from it, so the user-authored configuration is never rewritten.
func HistoryPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "jcd", "tracked.json")
}

// Load reads the configuration at path. A missing file (or an empty path)
// yields the defaults.
func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

// SortMode converts the configured sort name into a SortMode.
func (c Config) SortMode() (fsentry.SortMode, error) {
	switch strings.ToLower(strings.TrimSpace(c.Sort)) {
	case "", "name_asc":
		return fsentry.SortNameAsc, nil
	case "name_desc":
		return fsentry.SortNameDesc, nil
	case "date_asc":
		return fsentry.SortDateAsc, nil
	case "date_desc":
		return fsentry.SortDateDesc, nil
	case "size_asc":
		return fsentry.SortSizeAsc, nil
	case "size_desc":
		return fsentry.SortSizeDesc, nil
	default:
		return 0, fmt.Errorf("unknown sort %q", c.Sort)
	}
}
