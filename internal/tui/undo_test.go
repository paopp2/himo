package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/paopp2/himo/internal/model"
)

// Press x to mark current task done, then u to undo. Status and cursor must
// return to their pre-mutation values, and the banner reads "undone".
func TestUndo_statusChange(t *testing.T) {
	m := NewModel(testProject(t))
	// Baseline: capture the task and status at cursor 0.
	before := m.visibleTasks()
	if len(before) == 0 {
		t.Fatal("no visible tasks in test project")
	}
	title := before[0].Title
	origStatus := before[0].Status

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	// Confirm the mutation actually happened.
	for _, task := range m2.(Model).project.AllTasks() {
		if task.Title == title && task.Status != model.StatusDone {
			t.Fatalf("pre-undo: %q status = %v, want done", title, task.Status)
		}
	}

	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	um := m3.(Model)
	for _, task := range um.project.AllTasks() {
		if task.Title == title && task.Status != origStatus {
			t.Errorf("after undo: %q status = %v, want %v", title, task.Status, origStatus)
		}
	}
	if um.cursor != 0 {
		t.Errorf("after undo: cursor = %d, want 0", um.cursor)
	}
	if um.banner != "undone" {
		t.Errorf("after undo: banner = %q, want \"undone\"", um.banner)
	}
}
