package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNav_jAndK(t *testing.T) {
	m := NewModel(testProject(t))
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m2.(Model).cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", m2.(Model).cursor)
	}
	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m3.(Model).cursor != 1 {
		t.Errorf("after jj (only 2 tasks): cursor = %d, want 1 (clamped)", m3.(Model).cursor)
	}
	m4, _ := m3.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m4.(Model).cursor != 0 {
		t.Errorf("after k: cursor = %d, want 0", m4.(Model).cursor)
	}
}

func TestNav_gAndG(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 1
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if m2.(Model).cursor != 0 {
		t.Errorf("after g: cursor = %d, want 0", m2.(Model).cursor)
	}
	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if m3.(Model).cursor != 1 {
		t.Errorf("after G: cursor = %d, want 1 (last)", m3.(Model).cursor)
	}
}
