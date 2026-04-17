package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	BaseDir        string `toml:"base_dir"`
	Editor         string `toml:"editor"`
	DefaultProject string `toml:"default_project"`
	PreviewPane    bool   `toml:"preview_pane"`

	// Path of the config file this was loaded from (for save).
	Path string `toml:"-"`
}

// DefaultPath returns the canonical config file location.
func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(base, "himo", "config.toml"), nil
}

// Load reads config.toml and applies env overrides. Returns os.ErrNotExist
// if the file does not exist.
func Load(path string) (*Config, error) {
	cfg := &Config{Path: path, PreviewPane: true}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}
	if v := os.Getenv("JOT_DIR"); v != "" {
		cfg.BaseDir = v
	}
	if v := os.Getenv("EDITOR"); v != "" && cfg.Editor == "" {
		cfg.Editor = v
	}
	cfg.BaseDir = expandHome(cfg.BaseDir)
	return cfg, nil
}

// Save writes the config to its Path.
func Save(cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		return fmt.Errorf("mkdir config: %w", err)
	}
	f, err := os.Create(cfg.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
