package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/paopp2/himo/internal/store"
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
	// Save seeds a ProjectHeading at Items[0]; the inserted task is next.
	idx := 0
	if _, isHeading := active[0].(store.ProjectHeading); isHeading {
		idx = 1
	}
	if len(active) <= idx {
		t.Fatal("no task items after O insert")
	}
	ti, ok := active[idx].(store.TaskItem)
	if !ok {
		t.Fatalf("active[%d] = %T, want TaskItem", idx, active[idx])
	}
	if ti.Task.Title != "First" {
		t.Errorf("inserted task title = %q, want First", ti.Task.Title)
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

func TestNewTask_ctrlCCancelsPrompt(t *testing.T) {
	m := NewModel(testProject(t))
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	mid := m2
	for _, r := range "junk" {
		mid, _ = mid.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	final, _ := mid.(Model).Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if final.(Model).prompting {
		t.Error("Ctrl+C did not cancel prompt; prompting still true")
	}
	if final.(Model).promptBuf != "" {
		t.Errorf("promptBuf after Ctrl+C = %q, want empty", final.(Model).promptBuf)
	}
	for _, task := range final.(Model).project.AllTasks() {
		if task.Title == "junk" {
			t.Errorf("Ctrl+C should not have committed buffer; found task %q", task.Title)
		}
	}
}
