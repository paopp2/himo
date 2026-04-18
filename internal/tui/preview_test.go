package tui

import (
	"strings"
	"testing"

	"github.com/paopp2/himo/internal/store"
)

func TestPreview_showsNotesForHighlightedTask(t *testing.T) {
	m := NewModel(testProject(t))
	m.width, m.height = 120, 30
	// Add notes to the first task.
	m.project.Active.Items[0] = editTaskNotes(m.project.Active.Items[0], "Check the fridge first.")
	view := renderView(m)
	if !strings.Contains(view, "Check the fridge") {
		t.Errorf("preview missing notes:\n%s", view)
	}
}

func TestPreview_hiddenOnNarrow(t *testing.T) {
	m := NewModel(testProject(t))
	m.width = 60
	view := renderView(m)
	// narrow = no "Notes:" header in the view
	if strings.Contains(view, "Notes:") {
		t.Errorf("narrow view should not include preview header:\n%s", view)
	}
}

func editTaskNotes(it store.Item, notes string) store.Item {
	ti := it.(store.TaskItem)
	ti.Task.Notes = "    " + notes
	ti.RawLines = append(ti.RawLines, "    "+notes)
	return ti
}
