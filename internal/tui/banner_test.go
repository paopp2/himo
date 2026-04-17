package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestSetStatus_saveConflictSetsBanner simulates an external edit between load
// and save: pressing x returns ErrConflict, which must surface via m.banner.
func TestSetStatus_saveConflictSetsBanner(t *testing.T) {
	m := NewModel(testProject(t))
	dir := m.project.Dir
	// Bump active.md's mtime to invalidate the recorded value.
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(filepath.Join(dir, "active.md"), []byte("- [ ] external\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	got := m2.(Model).banner
	if got == "" {
		t.Fatalf("banner empty after save conflict")
	}
	if !strings.Contains(got, "blocked") {
		t.Errorf("banner = %q, want to mention blocked", got)
	}
}
