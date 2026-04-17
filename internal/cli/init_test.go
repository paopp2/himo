package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_writesConfigAndBaseDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	in := strings.NewReader("\n") // accept default
	var out bytes.Buffer
	err := Init(in, &out)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}

	cfgPath := filepath.Join(home, ".config", "himo", "config.toml")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("config file not created: %v", err)
	}
	baseDir := filepath.Join(home, "todos")
	if _, err := os.Stat(baseDir); err != nil {
		t.Errorf("base dir not created: %v", err)
	}
}
