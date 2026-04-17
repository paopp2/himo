package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewProject(t *testing.T) {
	base := t.TempDir()
	if err := NewProject(base, "work"); err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "work", "active.md")); err != nil {
		t.Errorf("active.md not created: %v", err)
	}
}

func TestNewProject_alreadyExists(t *testing.T) {
	base := t.TempDir()
	os.MkdirAll(filepath.Join(base, "work"), 0o755)
	err := NewProject(base, "work")
	if err == nil {
		t.Errorf("NewProject on existing dir: want error, got nil")
	}
}

func TestNewProject_invalidName(t *testing.T) {
	base := t.TempDir()
	for _, name := range []string{"", "with/slash", ".hidden", ".."} {
		if err := NewProject(base, name); err == nil {
			t.Errorf("NewProject(%q): want error", name)
		}
	}
}
