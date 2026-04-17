package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/npaolopepito/himo/internal/model"
)

func TestHelp_isModal(t *testing.T) {
	m := NewModel(testProject(t))
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m2.(Model).showingHelp {
		t.Fatalf("help did not open")
	}
	// x would normally mark the task done; while help is shown it must be ignored.
	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	for _, task := range m3.(Model).project.AllTasks() {
		if task.Status == model.StatusDone {
			t.Errorf("x changed status while help was shown: got %v", task.Status)
		}
	}
	if !m3.(Model).showingHelp {
		t.Errorf("x closed the help overlay; want it still showing")
	}
}

func TestHelp_toggle(t *testing.T) {
	m := NewModel(testProject(t))
	if m.showingHelp {
		t.Fatalf("showingHelp = true at start, want false")
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m2.(Model).showingHelp {
		t.Errorf("after ?: showingHelp = false, want true")
	}
	view := renderView(m2.(Model))
	if !strings.Contains(view, "Keybindings") {
		t.Errorf("help view missing 'Keybindings' header:\n%s", view)
	}
	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if m3.(Model).showingHelp {
		t.Errorf("after ??: showingHelp = true, want false")
	}
}
