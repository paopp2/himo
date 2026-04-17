package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func twoProjectBase(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	for _, name := range []string{"work", "personal"} {
		os.MkdirAll(filepath.Join(base, name), 0o755)
		os.WriteFile(filepath.Join(base, name, "active.md"),
			[]byte("- [ ] "+name+" task\n"), 0o644)
	}
	return base
}

func TestScope_tabSwitchesProject(t *testing.T) {
	base := twoProjectBase(t)
	m, err := NewModelFromBase(base, "work")
	if err != nil {
		t.Fatal(err)
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m2.(Model).project.Name != "personal" {
		t.Errorf("after Tab: project = %q, want personal", m2.(Model).project.Name)
	}
}
