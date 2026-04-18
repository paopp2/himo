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
	m, err := NewModelFromBase(base, "work", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m2.(Model).project.Name != "personal" {
		t.Errorf("after Tab: project = %q, want personal", m2.(Model).project.Name)
	}
}

// TestAllProjects_insertTargetsCursorProject verifies o in all-projects mode
// inserts the new task into the cursor's owning project rather than m.project.
func TestAllProjects_insertTargetsCursorProject(t *testing.T) {
	base := twoProjectBase(t)
	m, err := NewModelFromBase(base, "work", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// Enter all-projects mode, then move cursor to the "personal" task.
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if !cur.(Model).allProjects {
		t.Fatalf("A did not enter all-projects mode")
	}
	// Move down until the cursor's project is "personal". Project order is not
	// guaranteed across filesystems, so walk the visible list and stop when it
	// matches.
	for i := 0; i < 4; i++ {
		p, _, _, ok := cur.(Model).currentTaskItem()
		if ok && p != nil && p.Name == "personal" {
			break
		}
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	p, _, _, ok := cur.(Model).currentTaskItem()
	if !ok || p == nil || p.Name != "personal" {
		t.Fatalf("could not move cursor to personal project task")
	}
	before := len(p.Active.Items)
	// o, type title, Enter.
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	for _, r := range "newp" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Look up "personal" in the model's cache and verify the insert landed.
	var got int
	for _, proj := range cur.(Model).allProjectsCache {
		if proj.Name == "personal" {
			got = len(proj.Active.Items)
		}
	}
	if got != before+1 {
		t.Errorf("personal.Active.Items = %d, want %d", got, before+1)
	}
}
