package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/paopp2/himo/internal/store"
)

// editorCmd is a (file, line) pair ready to hand to the user's $EDITOR.
// Line is 1-based; 0 means "no jump".
type editorCmd struct {
	Path string
	Line int
}

// resolveEditor picks the program + leading args from $EDITOR, defaulting to vi.
func resolveEditor() (prog string, args []string) {
	parts := strings.Fields(os.Getenv("EDITOR"))
	if len(parts) == 0 {
		return "vi", nil
	}
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
		case store.ProjectHeading:
			// Render always emits "# name\n\n" (heading + one trailing
			// blank), so the heading takes two lines in the on-disk layout.
			line += 2
		case store.OpaqueLines:
			line += len(v.Lines)
		}
	}
	return 0, fmt.Errorf("item %d not found", idx)
}

func (m Model) openEditor(ec editorCmd) tea.Cmd {
	c := buildEditorCmd(ec.Path, ec.Line)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorReturnedMsg{err: err}
	})
}

type editorReturnedMsg struct{ err error }

