package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/npaolopepito/himo/internal/store"
)

func TestNewTask_OInsertsAbove(t *testing.T) {
	m := NewModel(testProject(t))
	// Cursor is on the first visible task under the default filter.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}})
	mid := m2
	for _, r := range "First" {
		mid, _ = mid.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	final, _ := mid.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	active := final.(Model).project.Active.Items
	if len(active) == 0 {
		t.Fatal("active items empty after O insert")
	}
	ti, ok := active[0].(store.TaskItem)
	if !ok {
		t.Fatalf("active[0] = %T, want TaskItem", active[0])
	}
	if ti.Task.Title != "First" {
		t.Errorf("active[0].Title = %q, want First", ti.Task.Title)
	}
	if final.(Model).promptAbove {
		t.Errorf("promptAbove still true after Enter; want cleared")
	}
}

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
