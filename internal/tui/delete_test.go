package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDelete_requiresConfirm(t *testing.T) {
	m := NewModel(testProject(t))
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !m2.(Model).confirmingDelete {
		t.Fatalf("d: confirmingDelete = false, want true")
	}
	before := len(m2.(Model).project.AllTasks())
	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if len(m3.(Model).project.AllTasks()) != before-1 {
		t.Errorf("task count after y: %d, want %d", len(m3.(Model).project.AllTasks()), before-1)
	}
}

func TestDelete_cancel(t *testing.T) {
	m := NewModel(testProject(t))
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	before := len(m2.(Model).project.AllTasks())
	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m3.(Model).confirmingDelete {
		t.Errorf("after n: still confirming")
	}
	if len(m3.(Model).project.AllTasks()) != before {
		t.Errorf("task count changed on cancel")
	}
}
