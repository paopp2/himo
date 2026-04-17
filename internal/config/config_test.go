package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_fromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`
base_dir = "~/todos"
editor = "nvim"
default_project = "work"
preview_pane = true
`), 0o644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	home, _ := os.UserHomeDir()
	if cfg.BaseDir != filepath.Join(home, "todos") {
		t.Errorf("BaseDir = %q, want expanded path", cfg.BaseDir)
	}
	if cfg.Editor != "nvim" {
		t.Errorf("Editor = %q, want nvim", cfg.Editor)
	}
	if cfg.DefaultProject != "work" {
		t.Errorf("DefaultProject = %q, want work", cfg.DefaultProject)
	}
	if !cfg.PreviewPane {
		t.Errorf("PreviewPane = false, want true")
	}
}

func TestLoad_envOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`base_dir = "~/todos"`), 0o644)
	t.Setenv("JOT_DIR", "/tmp/other")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BaseDir != "/tmp/other" {
		t.Errorf("BaseDir = %q, want /tmp/other (JOT_DIR override)", cfg.BaseDir)
	}
}

func TestLoad_missingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.toml")
	if !os.IsNotExist(err) {
		t.Errorf("Load(missing) err = %v, want os.IsNotExist", err)
	}
}
