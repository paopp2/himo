package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/npaolopepito/himo/internal/model"
)

func TestAction_markDone(t *testing.T) {
	m := NewModel(testProject(t))
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	tasks := m2.(Model).visibleTasks()
	for _, task := range tasks {
		if task.Title == "Buy groceries" {
			t.Errorf("Buy groceries still visible under default filter after x")
		}
	}
	found := false
	for _, task := range m2.(Model).project.AllTasks() {
		if task.Title == "Buy groceries" && task.Status == model.StatusDone {
			found = true
		}
	}
	if !found {
		t.Errorf("Buy groceries not found with status Done after x")
	}
}

func TestAction_cycleStatusWithSpace(t *testing.T) {
	m := NewModel(testProject(t))
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	for _, task := range m2.(Model).project.AllTasks() {
		if task.Title == "Buy groceries" && task.Status != model.StatusActive {
			t.Errorf("after Space: Buy groceries status = %v, want active", task.Status)
		}
	}
}
