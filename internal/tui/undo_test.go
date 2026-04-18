package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/paopp2/himo/internal/model"
	"github.com/paopp2/himo/internal/store"
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

// A status change followed by undo must restore the on-disk line exactly,
// not just Task.Status. Render the active document and compare to the
// pre-mutation bytes — catches any RawLines aliasing between snapshot and
// live items.
func TestUndo_statusChange_preservesRawLines(t *testing.T) {
	m := NewModel(testProject(t))
	// Save once up front so the "before" rendering includes any chrome
	// (leading project heading) that SaveProject injects. Without this
	// stabilization, the first mutation would introduce the heading and
	// the byte comparison would fail for reasons unrelated to aliasing.
	if err := store.SaveProject(m.project); err != nil {
		t.Fatal(err)
	}
	before := string(store.Render(m.project.Active))

	cur, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	after := string(store.Render(cur.(Model).project.Active))
	if after != before {
		t.Errorf("after undo: active document changed on disk:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// Press d then y to delete the current task, then u. The deleted task must
// return at its original index.
func TestUndo_delete(t *testing.T) {
	m := NewModel(testProject(t))
	before := m.visibleTasks()
	if len(before) < 2 {
		t.Fatal("need at least 2 visible tasks for this test")
	}
	title := before[0].Title

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	// Confirm deletion happened.
	for _, task := range m3.(Model).project.AllTasks() {
		if task.Title == title {
			t.Fatalf("pre-undo: %q still present after d+y", title)
		}
	}

	m4, _ := m3.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	um := m4.(Model)
	found := false
	for _, task := range um.project.AllTasks() {
		if task.Title == title {
			found = true
		}
	}
	if !found {
		t.Errorf("after undo: %q missing; want restored", title)
	}
	if um.cursor != 0 {
		t.Errorf("after undo: cursor = %d, want 0", um.cursor)
	}
}

// Press o, type a title, Enter. Then u. The newly inserted task must be gone
// and the cursor must be where it was before the insert.
func TestUndo_insert(t *testing.T) {
	m := NewModel(testProject(t))
	origCursor := m.cursor
	origCount := len(m.project.AllTasks())

	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	for _, r := range "ephemeral" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Confirm insert.
	if got := len(cur.(Model).project.AllTasks()); got != origCount+1 {
		t.Fatalf("pre-undo: task count = %d, want %d", got, origCount+1)
	}

	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	um := cur.(Model)
	if got := len(um.project.AllTasks()); got != origCount {
		t.Errorf("after undo: task count = %d, want %d", got, origCount)
	}
	for _, task := range um.project.AllTasks() {
		if task.Title == "ephemeral" {
			t.Errorf("after undo: \"ephemeral\" still present")
		}
	}
	if um.cursor != origCursor {
		t.Errorf("after undo: cursor = %d, want %d", um.cursor, origCursor)
	}
}

// After undoing a status change, Ctrl-R must re-apply the change.
func TestRedo_afterStatusUndo(t *testing.T) {
	m := NewModel(testProject(t))
	title := m.visibleTasks()[0].Title

	cur, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	um := cur.(Model)

	for _, task := range um.project.AllTasks() {
		if task.Title == title && task.Status != model.StatusDone {
			t.Errorf("after redo: %q status = %v, want done", title, task.Status)
		}
	}
	if um.banner != "redone" {
		t.Errorf("after redo: banner = %q, want \"redone\"", um.banner)
	}
}

// After deleting a task and undoing, Ctrl-R must re-delete it.
func TestRedo_afterDelete(t *testing.T) {
	m := NewModel(testProject(t))
	title := m.visibleTasks()[0].Title

	cur, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyCtrlR})

	for _, task := range cur.(Model).project.AllTasks() {
		if task.Title == title {
			t.Errorf("after redo: %q present; want deleted again", title)
		}
	}
}

// After inserting a task and undoing, Ctrl-R must re-insert it.
func TestRedo_afterInsert(t *testing.T) {
	m := NewModel(testProject(t))
	origCount := len(m.project.AllTasks())

	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	for _, r := range "redoable" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyCtrlR})

	if got := len(cur.(Model).project.AllTasks()); got != origCount+1 {
		t.Errorf("after redo: task count = %d, want %d", got, origCount+1)
	}
	found := false
	for _, task := range cur.(Model).project.AllTasks() {
		if task.Title == "redoable" {
			found = true
		}
	}
	if !found {
		t.Errorf("after redo: \"redoable\" missing; want re-inserted")
	}
}

// After u, making a new mutation must discard the redo stack: Ctrl-R becomes
// a no-op (banner empty, no status flip on a fresh task).
func TestRedo_clearedByNewMutation(t *testing.T) {
	m := NewModel(testProject(t))
	tasks := m.visibleTasks()
	if len(tasks) < 2 {
		t.Fatal("need >= 2 visible tasks")
	}

	// x on task 0, u to revert, j to move to task 1, x on task 1 (new mutation).
	cur, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	before := map[string]model.Status{}
	for _, task := range cur.(Model).project.AllTasks() {
		before[task.Title] = task.Status
	}

	// Ctrl-R should now do nothing.
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	after := map[string]model.Status{}
	for _, task := range cur.(Model).project.AllTasks() {
		after[task.Title] = task.Status
	}
	for title, s := range before {
		if after[title] != s {
			t.Errorf("Ctrl-R after new mutation changed %q: %v -> %v", title, s, after[title])
		}
	}
	if len(cur.(Model).redoStack) != 0 {
		t.Errorf("redoStack len = %d after new mutation, want 0", len(cur.(Model).redoStack))
	}
}

// Issuing 60 mutations must leave undoStack capped at historyLimit (50).
func TestUndo_stackCap(t *testing.T) {
	m := NewModel(testProject(t))
	var cur tea.Model = m
	// Switch to the All filter so Done tasks stay visible and every key
	// press finds a task under the cursor. Without this, DefaultFilter hides
	// tasks as soon as they flip to Done and setStatus early-returns on the
	// empty visible list.
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	// Alternate x and p to keep producing real mutations.
	for i := 0; i < 60; i++ {
		key := 'x'
		if i%2 == 1 {
			key = 'p'
		}
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
	}
	if got := len(cur.(Model).undoStack); got != historyLimit {
		t.Errorf("undoStack len = %d after 60 mutations, want %d", got, historyLimit)
	}
}

// In all-projects mode, mutate a task in one project, move to another, mutate
// that too, then u twice. Both mutations must be reverted in their respective
// projects.
func TestUndo_allProjectsCrossProject(t *testing.T) {
	base := twoProjectBase(t)
	m, err := NewModelFromBase(base, "work", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}

	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})

	// Mutate the task the cursor is on (some project).
	firstProj, _, _, _ := cur.(Model).currentTaskItem()
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Move until cursor's project differs, then mutate again.
	for i := 0; i < 5; i++ {
		p, _, _, ok := cur.(Model).currentTaskItem()
		if ok && p != nil && p.Name != firstProj.Name {
			break
		}
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	secondProj, _, _, ok := cur.(Model).currentTaskItem()
	if !ok || secondProj == nil || secondProj.Name == firstProj.Name {
		t.Fatalf("could not find a task in a different project")
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Two undos.
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Neither project should contain a Done task after both undos.
	for _, p := range cur.(Model).allProjectsCache {
		for _, task := range p.AllTasks() {
			if task.Status == model.StatusDone {
				t.Errorf("%s/%q still Done after two undos", p.Name, task.Title)
			}
		}
	}
}

// If the user mutates in all-projects mode, then exits to a single project
// that isn't the mutated one, u must show the "project not loaded" banner
// without modifying any state.
func TestUndo_projectUnloaded(t *testing.T) {
	base := twoProjectBase(t)
	m, err := NewModelFromBase(base, "work", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}

	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})

	// Move cursor until we're on a task from a project other than "work".
	for i := 0; i < 5; i++ {
		p, _, _, ok := cur.(Model).currentTaskItem()
		if ok && p != nil && p.Name != "work" {
			break
		}
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	p, _, _, ok := cur.(Model).currentTaskItem()
	if !ok || p == nil || p.Name == "work" {
		t.Fatalf("could not move cursor off 'work'")
	}
	// Mutate the non-work task.
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Exit all-projects mode. m.project is still "work".
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if cur.(Model).allProjects {
		t.Fatalf("expected exit from all-projects mode")
	}

	// Undo — target project is no longer loaded.
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	if cur.(Model).banner != "undo: project not loaded" {
		t.Errorf("banner = %q, want \"undo: project not loaded\"", cur.(Model).banner)
	}
	// Entry should remain on the stack.
	if len(cur.(Model).undoStack) == 0 {
		t.Errorf("undoStack empty after blocked undo; want entry retained")
	}
}

// After returning from $EDITOR, undo/redo history must be cleared — the
// external edit may have rewritten the file and existing snapshots are
// stale.
func TestUndo_clearedOnEditorReturn(t *testing.T) {
	m := NewModel(testProject(t))
	cur, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if len(cur.(Model).undoStack) == 0 {
		t.Fatal("precondition: expected undoStack to be non-empty after x")
	}

	// Simulate the editor returning without error.
	cur, _ = cur.(Model).Update(editorReturnedMsg{})
	um := cur.(Model)
	if len(um.undoStack) != 0 {
		t.Errorf("undoStack len = %d after editor return, want 0", len(um.undoStack))
	}
	if len(um.redoStack) != 0 {
		t.Errorf("redoStack len = %d after editor return, want 0", len(um.redoStack))
	}
}

// If undo's save step fails (mtime conflict), the in-memory state must be
// reverted and the entry must stay on the undo stack.
func TestUndo_saveFailure(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "active.md"), []byte("- [ ] Alpha\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "backlog.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "done.md"), []byte(""), 0o644)
	p, err := store.LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(p)

	// Mutate (this save succeeds).
	cur, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Simulate an external write: bump active.md's mtime to something else.
	future := time.Now().Add(5 * time.Second)
	if err := os.Chtimes(filepath.Join(dir, "active.md"), future, future); err != nil {
		t.Fatal(err)
	}

	// undo should now fail to save with ErrConflict.
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	um := cur.(Model)
	if !strings.Contains(um.banner, "blocked") {
		t.Errorf("banner = %q, want one containing \"blocked\"", um.banner)
	}
	// In-memory state must still be post-mutation (task is Done).
	foundDone := false
	for _, task := range um.project.AllTasks() {
		if task.Title == "Alpha" && task.Status == model.StatusDone {
			foundDone = true
		}
	}
	if !foundDone {
		t.Errorf("after failed undo: Alpha not Done; in-memory state was not reverted")
	}
	// Entry must remain on stack.
	if len(um.undoStack) == 0 {
		t.Errorf("undoStack empty after failed undo; want entry retained")
	}
}
