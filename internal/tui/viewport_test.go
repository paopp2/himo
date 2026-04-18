package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
)

// bigTestProject writes n pending tasks into a temp project and returns a
// loaded Project.
func bigTestProject(t *testing.T, n int) *store.Project {
	t.Helper()
	dir := t.TempDir()
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("- [ ] Task ")
		b.WriteString(strings.Repeat("a", 3))
		b.WriteByte('\n')
	}
	if err := os.WriteFile(filepath.Join(dir, "active.md"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := store.LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestList_cursorStaysVisible verifies that when the list is taller than
// the pane, the rendered pane is bounded in line count AND still contains
// the cursor bar for a cursor past the initial window.
func TestList_cursorStaysVisible(t *testing.T) {
	m := NewModel(bigTestProject(t, 40))
	m.width = 80
	m.height = 12
	m.styles = testStyles(t)
	m.cursor = 30

	locs := m.visibleTaskLocations()
	tasks := make([]model.Task, len(locs))
	for i, loc := range locs {
		tasks[i] = loc.doc.Items[loc.idx].(store.TaskItem).Task
	}
	paneHeight := 8
	pane := renderListPane(m, locs, tasks, 80, paneHeight)

	// With proper slicing, the pane's line count should be bounded by the
	// pane height plus a small border allowance. Without slicing, all 40
	// rows are emitted and the pane overflows.
	lines := strings.Count(pane, "\n") + 1
	if lines > paneHeight+4 {
		t.Errorf("list pane emitted %d lines, want <= %d (pane height + border slop)", lines, paneHeight+4)
	}

	// The cursor bar must survive the clipping.
	if !strings.Contains(pane, "▌") {
		t.Errorf("cursor bar missing in scrolled list:\n%s", pane)
	}
}
