package tui

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/paopp2/himo/internal/model"
	"github.com/paopp2/himo/internal/store"
)

// keypress sends a single tea.KeyMsg through Update and returns the next Model.
func keypress(t *testing.T, m Model, key tea.KeyMsg) Model {
	t.Helper()
	next, _ := m.Update(key)
	return next.(Model)
}

func keyRune(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func typeString(t *testing.T, m Model, s string) Model {
	t.Helper()
	for _, r := range s {
		m = keypress(t, m, keyRune(r))
	}
	return m
}

func firstTaskTitle(t *testing.T, m Model) string {
	t.Helper()
	_, doc, idx, ok := m.currentTaskItem()
	if !ok {
		t.Fatal("no current task")
	}
	return doc.Items[idx].(store.TaskItem).Task.Title
}

func TestEdit_e_entersEditMode(t *testing.T) {
	m := NewModel(testProject(t))
	original := firstTaskTitle(t, m)
	m = keypress(t, m, keyRune('e'))
	if !m.editing {
		t.Fatal("after e: m.editing = false, want true")
	}
	if m.currentMode() != ModeEdit {
		t.Errorf("currentMode = %v, want ModeEdit", m.currentMode())
	}
	if got := m.editInput.Value(); got != original {
		t.Errorf("editInput.Value() = %q, want %q", got, original)
	}
	if m.editOrig != original {
		t.Errorf("editOrig = %q, want %q", m.editOrig, original)
	}
}

func TestEdit_typesAndCommits(t *testing.T) {
	m := NewModel(testProject(t))
	m = keypress(t, m, keyRune('e'))
	for range m.editInput.Value() {
		m = keypress(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	m = typeString(t, m, "Renamed")
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.editing {
		t.Error("editing flag still set after Enter")
	}
	if got := firstTaskTitle(t, m); got != "Renamed" {
		t.Errorf("title after edit = %q, want %q", got, "Renamed")
	}
	reloaded, err := store.LoadProject(m.project.Dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, task := range reloaded.AllTasks() {
		if task.Title == "Renamed" {
			found = true
			break
		}
	}
	if !found {
		t.Error(`"Renamed" not persisted to disk`)
	}
}

func TestEdit_escCancels(t *testing.T) {
	m := NewModel(testProject(t))
	original := firstTaskTitle(t, m)
	m = keypress(t, m, keyRune('e'))
	m = typeString(t, m, "noise")
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.editing {
		t.Error("editing flag still set after Esc")
	}
	if got := firstTaskTitle(t, m); got != original {
		t.Errorf("title after Esc = %q, want unchanged %q", got, original)
	}
}

func TestEdit_ctrlCCancels(t *testing.T) {
	m := NewModel(testProject(t))
	original := firstTaskTitle(t, m)
	m = keypress(t, m, keyRune('e'))
	m = typeString(t, m, "noise")
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if m.editing {
		t.Error("editing flag still set after Ctrl+C")
	}
	if got := firstTaskTitle(t, m); got != original {
		t.Errorf("title after Ctrl+C = %q, want unchanged %q", got, original)
	}
}

func TestEdit_emptyBufferIsCancel(t *testing.T) {
	m := NewModel(testProject(t))
	original := firstTaskTitle(t, m)
	m = keypress(t, m, keyRune('e'))
	for range m.editInput.Value() {
		m = keypress(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.editing {
		t.Error("editing flag still set after empty Enter")
	}
	if got := firstTaskTitle(t, m); got != original {
		t.Errorf("title after empty Enter = %q, want unchanged %q", got, original)
	}
}

func TestEdit_unchangedTitleIsCancel(t *testing.T) {
	m := NewModel(testProject(t))
	before := append([]store.Item(nil), m.project.Active.Items...)
	m = keypress(t, m, keyRune('e'))
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.editing {
		t.Error("editing flag still set after no-op Enter")
	}
	if !reflect.DeepEqual(m.project.Active.Items, before) {
		t.Errorf("Active.Items mutated by no-op Enter")
	}
}

func TestEdit_preservesStatus(t *testing.T) {
	m := NewModel(testProject(t))
	// Cursor sits on the first active-file task; it's pending. Cycle to done
	// via the existing 'x' keybind so we can verify status is preserved across
	// an inline edit.
	m = keypress(t, m, keyRune('x'))
	// After 'x' the task moves to done.md; switch filter to done so the cursor
	// lands on it again.
	m = keypress(t, m, keyRune('5'))
	if got := m.currentMode(); got != ModeNormal {
		t.Fatalf("mode after filter switch = %v, want ModeNormal", got)
	}
	_, doc, idx, ok := m.currentTaskItem()
	if !ok {
		t.Fatal("no done task to edit")
	}
	if doc.Items[idx].(store.TaskItem).Task.Status != model.StatusDone {
		t.Fatal("expected cursor on a done task")
	}
	m = keypress(t, m, keyRune('e'))
	for range m.editInput.Value() {
		m = keypress(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	m = typeString(t, m, "Edited")
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	_, doc, idx, ok = m.currentTaskItem()
	if !ok {
		t.Fatal("task vanished after edit")
	}
	ti := doc.Items[idx].(store.TaskItem)
	if ti.Task.Status != model.StatusDone {
		t.Errorf("status after edit = %v, want Done", ti.Task.Status)
	}
	if ti.Task.Title != "Edited" {
		t.Errorf("title after edit = %q, want %q", ti.Task.Title, "Edited")
	}
	if len(ti.RawLines) == 0 || ti.RawLines[0][:5] != "- [x]" {
		t.Errorf("RawLines[0] = %q, want prefix %q", ti.RawLines[0], "- [x]")
	}
}

func TestEdit_preservesNotes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [ ] Task with notes\n    line one\n    line two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "backlog.md"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "done.md"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	proj, err := store.LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(proj)
	m = keypress(t, m, keyRune('e'))
	for range m.editInput.Value() {
		m = keypress(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	m = typeString(t, m, "Renamed")
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	_, doc, idx, ok := m.currentTaskItem()
	if !ok {
		t.Fatal("task missing after edit")
	}
	ti := doc.Items[idx].(store.TaskItem)
	if ti.Task.Notes == "" {
		t.Errorf("notes lost after edit; got empty")
	}
	if !strings.Contains(ti.Task.Notes, "line one") || !strings.Contains(ti.Task.Notes, "line two") {
		t.Errorf("notes mangled: %q", ti.Task.Notes)
	}
}

func TestEdit_emptyListNoOp(t *testing.T) {
	m := NewModel(testProject(t))
	// Switch to a status with no tasks (cancelled).
	m = keypress(t, m, keyRune('6'))
	m = keypress(t, m, keyRune('e'))
	if m.editing {
		t.Errorf("editing flag set on empty list; want no-op")
	}
}

func TestEdit_allProjectsTargetsCorrectProject(t *testing.T) {
	base := t.TempDir()
	makeProj := func(name, body string) {
		dir := filepath.Join(base, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "active.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, f := range []string{"backlog.md", "done.md"} {
			if err := os.WriteFile(filepath.Join(dir, f), nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	makeProj("a", "- [ ] A-task\n")
	makeProj("b", "- [ ] B-task\n")

	m, err := NewModelFromBase(base, "a", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m = keypress(t, m, keyRune('A')) // enter all-projects mode
	// Find the cursor index for "B-task" by scanning visible tasks.
	target := -1
	for i, task := range m.visibleTasks() {
		if task.Title == "B-task" {
			target = i
			break
		}
	}
	if target < 0 {
		t.Fatal("B-task not visible in all-projects mode")
	}
	for m.cursor < target {
		m = keypress(t, m, keyRune('j'))
	}
	m = keypress(t, m, keyRune('e'))
	for range m.editInput.Value() {
		m = keypress(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	m = typeString(t, m, "B-edited")
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	// Reload both projects from disk and verify only B changed.
	a, err := store.LoadProject(filepath.Join(base, "a"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := store.LoadProject(filepath.Join(base, "b"))
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range a.AllTasks() {
		if task.Title != "A-task" {
			t.Errorf("project a touched: title %q", task.Title)
		}
	}
	bTitles := make([]string, 0)
	for _, task := range b.AllTasks() {
		bTitles = append(bTitles, task.Title)
	}
	wantB := []string{"B-edited"}
	if !reflect.DeepEqual(bTitles, wantB) {
		t.Errorf("project b titles = %v, want %v", bTitles, wantB)
	}
}

func TestEdit_undoRevertsTitleChange(t *testing.T) {
	m := NewModel(testProject(t))
	original := firstTaskTitle(t, m)
	m = keypress(t, m, keyRune('e'))
	for range m.editInput.Value() {
		m = keypress(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	m = typeString(t, m, "TempTitle")
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if got := firstTaskTitle(t, m); got != "TempTitle" {
		t.Fatalf("pre-undo title = %q, want %q", got, "TempTitle")
	}
	m = keypress(t, m, keyRune('u'))
	if got := firstTaskTitle(t, m); got != original {
		t.Errorf("post-undo title = %q, want %q", got, original)
	}
}

func TestEdit_ctrlWDeletesLastWord(t *testing.T) {
	m := NewModel(testProject(t))
	m = keypress(t, m, keyRune('e'))
	for range m.editInput.Value() {
		m = keypress(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	m = typeString(t, m, "two words")
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyCtrlW})
	if got := m.editInput.Value(); got != "two " {
		t.Errorf("after Ctrl+W, editInput.Value() = %q, want %q", got, "two ")
	}
}

// TestEdit_saveConflictRollsBack mirrors TestSetStatus_saveConflictSetsBanner:
// an external write between load and commit must surface a banner, leave the
// in-memory title unchanged, and not record an undo entry.
func TestEdit_saveConflictRollsBack(t *testing.T) {
	m := NewModel(testProject(t))
	original := firstTaskTitle(t, m)
	dir := m.project.Dir
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(filepath.Join(dir, "active.md"), []byte("- [ ] external\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m = keypress(t, m, keyRune('e'))
	for range m.editInput.Value() {
		m = keypress(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	m = typeString(t, m, "Renamed")
	m = keypress(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.banner == "" || !strings.Contains(m.banner, "blocked") {
		t.Errorf("banner = %q, want to mention blocked", m.banner)
	}
	if got := firstTaskTitle(t, m); got != original {
		t.Errorf("title after conflict = %q, want unchanged %q", got, original)
	}
	if len(m.undoStack) != 0 {
		t.Errorf("undoStack len = %d, want 0 (rollback should drop the entry)", len(m.undoStack))
	}
}
