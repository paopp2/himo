package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/paopp2/himo/internal/model"
)

// Press j to move the cursor to task 1, x to mark it done, then u to undo.
// Status and cursor must return to their pre-mutation values, the banner
// reads "undone", and the undo stack must have been popped.
func TestUndo_statusChange(t *testing.T) {
	m := NewModel(testProject(t))
	// Baseline: capture the task and status at cursor 1 so the restore has
	// real distance to cover (starting cursor is 0).
	before := m.visibleTasks()
	if len(before) < 2 {
		t.Fatalf("need at least 2 visible tasks, got %d", len(before))
	}
	title := before[1].Title
	origStatus := before[1].Status

	// Move cursor to index 1 before the mutation.
	m1, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if got := m1.(Model).cursor; got != 1 {
		t.Fatalf("after j: cursor = %d, want 1", got)
	}

	m2, _ := m1.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
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
	if um.cursor != 1 {
		t.Errorf("after undo: cursor = %d, want 1", um.cursor)
	}
	if um.banner != "undone" {
		t.Errorf("after undo: banner = %q, want \"undone\"", um.banner)
	}
	if n := len(um.undoStack); n != 0 {
		t.Errorf("after undo: len(undoStack) = %d, want 0", n)
	}
}
