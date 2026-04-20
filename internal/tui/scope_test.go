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

func threeProjectBase(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	for _, name := range []string{"a", "b", "c"} {
		os.MkdirAll(filepath.Join(base, name), 0o755)
		os.WriteFile(filepath.Join(base, name, "active.md"),
			[]byte("- [ ] "+name+" task\n"), 0o644)
	}
	return base
}

// Tab in all-projects mode exits All and advances one project in the cycle.
func TestScope_tabInAllExitsAndCyclesNext(t *testing.T) {
	base := threeProjectBase(t)
	m, err := NewModelFromBase(base, "a", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cur, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if !cur.(Model).allProjects {
		t.Fatalf("A did not enter all-projects mode")
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyTab})
	got := cur.(Model)
	if got.allProjects {
		t.Errorf("after Tab: allProjects = true, want false")
	}
	if got.project.Name != "b" {
		t.Errorf("after Tab: project = %q, want b", got.project.Name)
	}
}

// Shift+Tab in all-projects mode exits All and steps one project backward.
func TestScope_shiftTabInAllExitsAndCyclesPrev(t *testing.T) {
	base := threeProjectBase(t)
	m, err := NewModelFromBase(base, "a", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cur, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if !cur.(Model).allProjects {
		t.Fatalf("A did not enter all-projects mode")
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	got := cur.(Model)
	if got.allProjects {
		t.Errorf("after Shift+Tab: allProjects = true, want false")
	}
	if got.project.Name != "c" {
		t.Errorf("after Shift+Tab: project = %q, want c", got.project.Name)
	}
}

func TestSessionAllProjects_reflectsMode(t *testing.T) {
	base := twoProjectBase(t)
	m, err := NewModelFromBase(base, "work", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if m.SessionAllProjects() {
		t.Fatalf("fresh model: SessionAllProjects = true, want false")
	}
	cur, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if !cur.(Model).SessionAllProjects() {
		t.Errorf("after A: SessionAllProjects = false, want true")
	}
}

func TestWithAllProjects_restoresMode(t *testing.T) {
	base := twoProjectBase(t)
	m, err := NewModelFromBase(base, "work", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m = m.WithAllProjects()
	if !m.allProjects {
		t.Errorf("WithAllProjects: allProjects = false, want true")
	}
	if len(m.allProjectsCache) == 0 {
		t.Errorf("WithAllProjects: allProjectsCache empty, want populated")
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
	// Save also seeds a ProjectHeading at the top of the doc, so the post-save
	// count is before + 1 (new task) + 1 (heading).
	if got != before+2 {
		t.Errorf("personal.Active.Items = %d, want %d", got, before+2)
	}
}
