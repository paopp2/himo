package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewTask_oPromptsAndInserts(t *testing.T) {
	m := NewModel(testProject(t))
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	if !m2.(Model).prompting {
		t.Fatalf("after o: prompting = false, want true")
	}
	mid := m2
	for _, r := range "Review PR" {
		mid, _ = mid.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	final, _ := mid.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	found := false
	for _, task := range final.(Model).project.AllTasks() {
		if task.Title == "Review PR" {
			found = true
		}
	}
	if !found {
		t.Errorf("task 'Review PR' not created after o+input+Enter")
	}
}
