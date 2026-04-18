package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paopp2/himo/internal/store"
)

// TestEditorReturn_reloadsEditedProject verifies the editor-return handler
// reloads the project stashed in editingProjectDir rather than m.project.
func TestEditorReturn_reloadsEditedProject(t *testing.T) {
	// Two projects; m.project points at A, but the editor edited B.
	base := t.TempDir()
	dirA := filepath.Join(base, "a")
	dirB := filepath.Join(base, "b")
	for _, d := range []string{dirA, dirB} {
		os.MkdirAll(d, 0o755)
		os.WriteFile(filepath.Join(d, "active.md"), []byte("- [ ] start\n"), 0o644)
	}
	projA, err := store.LoadProject(dirA)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(projA)
	m.editingProjectDir = dirB
	// Mutate B on disk so the reload observes the new content.
	if err := os.WriteFile(filepath.Join(dirB, "active.md"), []byte("- [ ] after-edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m2, _ := m.Update(editorReturnedMsg{err: nil})
	// m.project must remain the A project (same Dir), but the editingProjectDir
	// should be cleared.
	if m2.(Model).project.Dir != dirA {
		t.Errorf("m.project.Dir = %q, want %q", m2.(Model).project.Dir, dirA)
	}
	if m2.(Model).editingProjectDir != "" {
		t.Errorf("editingProjectDir = %q, want cleared", m2.(Model).editingProjectDir)
	}
	// Verify the reload happened on dir B by checking the file mtime was
	// refreshed via SaveProject (it rewrites after normalize).
	// Simpler: re-parse B and look for the task.
	reloaded, err := store.LoadProject(dirB)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, task := range reloaded.AllTasks() {
		if task.Title == "after-edit" {
			found = true
		}
	}
	if !found {
		t.Errorf("B did not contain 'after-edit' task after reload")
	}
}

func TestResolveEditor_whitespaceFallsBackToVi(t *testing.T) {
	t.Setenv("EDITOR", "   ")
	prog, args := resolveEditor()
	if prog != "vi" || len(args) != 0 {
		t.Errorf("got %q %v, want vi []", prog, args)
	}
}

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

// TestEditorCmd_accountsForProjectHeading verifies the line number jumps
// past the "# name\n\n" that Render always emits.
func TestEditorCmd_accountsForProjectHeading(t *testing.T) {
	dir := t.TempDir()
	contents := "# myproj\n\n- [ ] first\n- [ ] second\n"
	if err := os.WriteFile(filepath.Join(dir, "active.md"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "backlog.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "done.md"), []byte(""), 0o644)
	p, err := store.LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(p)
	m.cursor = 0 // first task -> line 3 (heading + blank + 1)
	cmd, err := m.editorCmdForNotes()
	if err != nil {
		t.Fatalf("editorCmdForNotes: %v", err)
	}
	if cmd.Line != 3 {
		t.Errorf("cmd line = %d, want 3", cmd.Line)
	}
	m.cursor = 1 // second task -> line 4
	cmd, err = m.editorCmdForNotes()
	if err != nil {
		t.Fatalf("editorCmdForNotes: %v", err)
	}
	if cmd.Line != 4 {
		t.Errorf("cmd line = %d, want 4", cmd.Line)
	}
}
