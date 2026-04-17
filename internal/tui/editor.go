package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/npaolopepito/himo/internal/store"
)

// editorCmd is a (file, line) pair ready to hand to the user's $EDITOR.
// Line is 1-based; 0 means "no jump".
type editorCmd struct {
	Path string
	Line int
}

// resolveEditor picks the program + leading args from $EDITOR, defaulting to vi.
func resolveEditor() (prog string, args []string) {
	e := os.Getenv("EDITOR")
	if e == "" {
		e = "vi"
	}
	parts := strings.Fields(e)
	return parts[0], parts[1:]
}

// buildEditorCmd constructs the exec.Cmd that jumps to line in path.
// A line of 0 is treated as "open the file with no jump".
func buildEditorCmd(path string, line int) *exec.Cmd {
	prog, args := resolveEditor()
	if line > 0 {
		args = append(args, fmt.Sprintf("+%d", line))
	}
	args = append(args, path)
	return exec.Command(prog, args...)
}

// editorCmdForNotes resolves the file+line of the highlighted task.
func (m Model) editorCmdForNotes() (editorCmd, error) {
	proj, doc, idx, ok := m.currentTaskItem()
	if !ok {
		return editorCmd{}, fmt.Errorf("no task selected")
	}
	path := docFilename(proj, doc)
	if path == "" {
		return editorCmd{}, fmt.Errorf("unknown document")
	}
	line, err := taskLineInFile(doc, idx)
	if err != nil {
		return editorCmd{}, err
	}
	return editorCmd{Path: path, Line: line}, nil
}

// docFilename maps a Document pointer back to its on-disk filename within proj.
func docFilename(proj *store.Project, doc *store.Document) string {
	var f store.FileName
	switch doc {
	case proj.Active:
		f = store.FileActive
	case proj.Backlog:
		f = store.FileBacklog
	case proj.Done:
		f = store.FileDone
	default:
		return ""
	}
	return filepath.Join(proj.Dir, f.String())
}

// taskLineInFile counts lines of preceding items (including notes rawlines,
// date headings, and opaque blocks) to arrive at the 1-based line of items[idx].
func taskLineInFile(doc *store.Document, idx int) (int, error) {
	if idx < 0 || idx >= len(doc.Items) {
		return 0, fmt.Errorf("item %d out of range", idx)
	}
	line := 1
	for i, it := range doc.Items {
		if i == idx {
			return line, nil
		}
		switch v := it.(type) {
		case store.TaskItem:
			line += len(v.RawLines)
		case store.DateHeading:
			line += 1
		case store.OpaqueLines:
			line += len(v.Lines)
		}
	}
	return 0, fmt.Errorf("item %d not found", idx)
}

// openEditor returns a tea.Cmd that suspends the TUI, runs $EDITOR, and then
// delivers an editorReturnedMsg.
func (m Model) openEditor(ec editorCmd) tea.Cmd {
	c := buildEditorCmd(ec.Path, ec.Line)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorReturnedMsg{err: err}
	})
}

// editorReturnedMsg signals that the editor subprocess has exited.
type editorReturnedMsg struct{ err error }

// fileForFilter maps the filter to the file `e` should open within proj.
// Returns an error when the filter is "all" or spans multiple target files
// (ambiguous; user picks 1-6 first).
func fileForFilter(f Filter, proj *store.Project) (string, error) {
	if f.All {
		return "", fmt.Errorf("choose a filter first (1-6) to pick a file")
	}
	if len(f.Statuses) == 0 {
		return "", fmt.Errorf("no filter active")
	}
	target := store.TargetFile(f.Statuses[0])
	for _, s := range f.Statuses[1:] {
		if store.TargetFile(s) != target {
			return "", fmt.Errorf("filter spans multiple files; narrow with 1-6")
		}
	}
	return filepath.Join(proj.Dir, target.String()), nil
}
