package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPicker_filterMatches(t *testing.T) {
	base := twoProjectBase(t)
	m, err := NewModelFromBase(base, "work", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	if !cur.(Model).pickerOpen {
		t.Fatalf("P did not open picker")
	}
	for _, r := range "pers" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	names := cur.(Model).filteredProjects()
	if len(names) != 1 || names[0] != "personal" {
		t.Errorf("filtered: %v, want [personal]", names)
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cur.(Model).pickerOpen {
		t.Errorf("picker still open after Enter")
	}
	if cur.(Model).project.Name != "personal" {
		t.Errorf("project = %q after pick, want personal", cur.(Model).project.Name)
	}
}

func TestPicker_escClosesWithoutSwitch(t *testing.T) {
	base := twoProjectBase(t)
	m, err := NewModelFromBase(base, "work", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cur.(Model).pickerOpen {
		t.Errorf("picker still open after Esc")
	}
	if cur.(Model).project.Name != "work" {
		t.Errorf("project = %q after Esc, want work", cur.(Model).project.Name)
	}
}

func TestAllProjects_aggregates(t *testing.T) {
	base := twoProjectBase(t)
	m, err := NewModelFromBase(base, "work", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if !m2.(Model).allProjects {
		t.Fatalf("A did not enter all-projects mode")
	}
	visible := m2.(Model).visibleTasks()
	if len(visible) != 2 {
		t.Errorf("all-projects visible = %v, want 2 tasks", titles(visible))
	}
	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if m3.(Model).allProjects {
		t.Errorf("second A did not exit all-projects mode")
	}
}
