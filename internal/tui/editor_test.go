package tui

import (
	"strings"
	"testing"
)

func TestEditorCmd_notesLocation(t *testing.T) {
	m := NewModel(testProject(t))
	cmd, err := m.editorCmdForNotes()
	if err != nil {
		t.Fatalf("editorCmdForNotes: %v", err)
	}
	if !strings.HasSuffix(cmd.Path, "active.md") {
		t.Errorf("cmd path = %q, want ends with active.md", cmd.Path)
	}
	if cmd.Line != 1 {
		t.Errorf("cmd line = %d, want 1", cmd.Line)
	}
}
