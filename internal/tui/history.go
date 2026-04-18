package tui

import (
	"github.com/paopp2/himo/internal/store"
)

// historyLimit caps the number of entries kept in undoStack and redoStack.
const historyLimit = 50

// historyEntry is a snapshot of one project's three document item slices and
// the cursor position, taken immediately before a mutation. Restoring this
// entry reverses the mutation.
type historyEntry struct {
	projectDir string
	active     []store.Item
	backlog    []store.Item
	done       []store.Item
	cursor     int
}

// snapshotOf captures proj's three document item slices plus the cursor.
// Slices are copied and TaskItem.RawLines is cloned so later in-place edits
// (e.g. setStatus rewriting RawLines[0]) do not corrupt the snapshot.
func snapshotOf(proj *store.Project, cursor int) historyEntry {
	return historyEntry{
		projectDir: proj.Dir,
		active:     cloneItems(proj.Active.Items),
		backlog:    cloneItems(proj.Backlog.Items),
		done:       cloneItems(proj.Done.Items),
		cursor:     cursor,
	}
}

// cloneItems returns a copy of items safe for long-term retention. TaskItem
// values carry a RawLines []string that setStatus mutates in place; shallow
// slice copying would alias that backing array and corrupt the snapshot.
func cloneItems(items []store.Item) []store.Item {
	out := make([]store.Item, len(items))
	for i, it := range items {
		if ti, ok := it.(store.TaskItem); ok {
			ti.RawLines = append([]string(nil), ti.RawLines...)
			out[i] = ti
		} else {
			out[i] = it
		}
	}
	return out
}

// pushUndo records a pre-mutation snapshot of proj. redoStack is NOT cleared
// here — clearing is deferred to commitUndo, so a mutation that fails its
// normalize or save step does not destroy forward history.
func (m *Model) pushUndo(proj *store.Project) {
	m.undoStack = append(m.undoStack, snapshotOf(proj, m.cursor))
	if len(m.undoStack) > historyLimit {
		m.undoStack = m.undoStack[len(m.undoStack)-historyLimit:]
	}
}

// commitUndo marks the last pushUndo as final: the mutation succeeded, so
// any pending redo history is invalidated.
func (m *Model) commitUndo() {
	m.redoStack = nil
}

// popUndo removes the top of the undo stack. Used for rollback when a
// mutation's save step fails after pushUndo was called.
func (m *Model) popUndo() {
	if n := len(m.undoStack); n > 0 {
		m.undoStack = m.undoStack[:n-1]
	}
}

// projectByDir finds a loaded project whose Dir matches. In single-project
// mode only m.project is considered; in all-projects mode, m.allProjectsCache
// is also searched. Returns nil if the project is no longer loaded.
func (m *Model) projectByDir(dir string) *store.Project {
	if m.project != nil && m.project.Dir == dir {
		return m.project
	}
	for _, p := range m.allProjectsCache {
		if p != nil && p.Dir == dir {
			return p
		}
	}
	return nil
}

// applyEntry writes entry's snapshot into proj. Does not save.
func applyEntry(proj *store.Project, entry historyEntry) {
	proj.Active.Items = entry.active
	proj.Backlog.Items = entry.backlog
	proj.Done.Items = entry.done
}

// clampCursor keeps m.cursor inside the current visible list.
func (m *Model) clampCursor() {
	if n := len(m.visibleTasks()); m.cursor >= n && n > 0 {
		m.cursor = n - 1
	}
}

// applyHistory moves the top entry from 'from' to 'to', applying it to the
// referenced project and restoring the cursor. op is the banner verb used
// on errors (e.g. "undo"); okBanner is the success banner. Used by undo
// and redo, which differ only by which stack they pop from.
func (m *Model) applyHistory(from, to *[]historyEntry, op, okBanner string) {
	if len(*from) == 0 {
		return
	}
	entry := (*from)[len(*from)-1]
	proj := m.projectByDir(entry.projectDir)
	if proj == nil {
		m.banner = op + ": project not loaded"
		return
	}

	current := snapshotOf(proj, m.cursor)
	applyEntry(proj, entry)
	if err := m.saveWithBanner(proj, op); err != nil {
		applyEntry(proj, current)
		return
	}
	*from = (*from)[:len(*from)-1]
	*to = append(*to, current)
	if len(*to) > historyLimit {
		*to = (*to)[len(*to)-historyLimit:]
	}
	m.cursor = entry.cursor
	m.clampCursor()
	m.banner = okBanner
}

func (m *Model) undo() { m.applyHistory(&m.undoStack, &m.redoStack, "undo", "undone") }
func (m *Model) redo() { m.applyHistory(&m.redoStack, &m.undoStack, "redo", "redone") }
