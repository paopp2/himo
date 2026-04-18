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

// snapshotOf captures the current state of proj plus the given cursor index.
// Slices are shallow-copied; store.Item values and their string contents are
// value-typed or immutable, so a shallow copy is safe.
func snapshotOf(proj *store.Project, cursor int) historyEntry {
	return historyEntry{
		projectDir: proj.Dir,
		active:     append([]store.Item(nil), proj.Active.Items...),
		backlog:    append([]store.Item(nil), proj.Backlog.Items...),
		done:       append([]store.Item(nil), proj.Done.Items...),
		cursor:     cursor,
	}
}

// pushUndo records a snapshot of proj before a mutation. The redo stack is
// cleared because any new mutation invalidates forward history.
func (m *Model) pushUndo(proj *store.Project) {
	m.undoStack = append(m.undoStack, snapshotOf(proj, m.cursor))
	if len(m.undoStack) > historyLimit {
		m.undoStack = m.undoStack[len(m.undoStack)-historyLimit:]
	}
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

// undo pops the top undo entry, snapshots current state to redoStack, applies
// the entry, saves, and restores the cursor. On save failure the in-memory
// state is reverted and the entry stays on the undo stack. On a
// project-not-loaded miss, the entry also stays on the stack.
func (m *Model) undo() {
	if len(m.undoStack) == 0 {
		return
	}
	entry := m.undoStack[len(m.undoStack)-1]
	proj := m.projectByDir(entry.projectDir)
	if proj == nil {
		m.banner = "undo: project not loaded"
		return
	}

	current := snapshotOf(proj, m.cursor)
	applyEntry(proj, entry)
	if err := m.saveWithBanner(proj, "undo"); err != nil {
		applyEntry(proj, current)
		return
	}
	m.undoStack = m.undoStack[:len(m.undoStack)-1]
	m.redoStack = append(m.redoStack, current)
	if len(m.redoStack) > historyLimit {
		m.redoStack = m.redoStack[len(m.redoStack)-historyLimit:]
	}
	m.cursor = entry.cursor
	if n := len(m.visibleTasks()); m.cursor >= n && n > 0 {
		m.cursor = n - 1
	}
	m.banner = "undone"
}

// redo pops the top redo entry, snapshots current state onto undoStack,
// applies the entry, saves, and restores the cursor. Symmetrical to undo.
// The redo stack is NOT cleared here; only pushUndo (a new mutation) clears
// it. redo must NOT go through pushUndo, because pushUndo would clear the
// redoStack.
func (m *Model) redo() {
	if len(m.redoStack) == 0 {
		return
	}
	entry := m.redoStack[len(m.redoStack)-1]
	proj := m.projectByDir(entry.projectDir)
	if proj == nil {
		m.banner = "redo: project not loaded"
		return
	}

	current := snapshotOf(proj, m.cursor)
	applyEntry(proj, entry)
	if err := m.saveWithBanner(proj, "redo"); err != nil {
		applyEntry(proj, current)
		return
	}
	m.redoStack = m.redoStack[:len(m.redoStack)-1]
	m.undoStack = append(m.undoStack, current)
	if len(m.undoStack) > historyLimit {
		m.undoStack = m.undoStack[len(m.undoStack)-historyLimit:]
	}
	m.cursor = entry.cursor
	if n := len(m.visibleTasks()); m.cursor >= n && n > 0 {
		m.cursor = n - 1
	}
	m.banner = "redone"
}
