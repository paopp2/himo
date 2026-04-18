package store

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/npaolopepito/himo/internal/model"
)

// allFiles lists the three filenames that make up a project, in a fixed order.
var allFiles = [...]string{"active.md", "backlog.md", "done.md"}

// Project is a loaded project (three parsed documents).
type Project struct {
	Name    string
	Dir     string
	Active  *Document
	Backlog *Document
	Done    *Document

	mtimes map[string]time.Time // filename -> mtime at load
}

// LoadProject parses the three files in a project directory. Missing files
// are treated as empty.
func LoadProject(dir string) (*Project, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("stat project: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", dir)
	}
	load := func(name string, parse func([]byte) (*Document, error)) (*Document, error) {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if os.IsNotExist(err) {
			return &Document{}, nil
		}
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		return parse(b)
	}
	active, err := load("active.md", ParseActive)
	if err != nil {
		return nil, err
	}
	backlog, err := load("backlog.md", ParseBacklog)
	if err != nil {
		return nil, err
	}
	done, err := load("done.md", ParseDone)
	if err != nil {
		return nil, err
	}
	p := &Project{
		Name:    filepath.Base(dir),
		Dir:     dir,
		Active:  active,
		Backlog: backlog,
		Done:    done,
		mtimes:  make(map[string]time.Time),
	}
	for _, name := range allFiles {
		info, err := os.Stat(filepath.Join(dir, name))
		if err == nil {
			p.mtimes[name] = info.ModTime()
		}
	}
	return p, nil
}

// AllTasks returns every task in the project across all three files.
func (p *Project) AllTasks() []model.Task {
	var out []model.Task
	out = append(out, p.Active.Tasks()...)
	out = append(out, p.Backlog.Tasks()...)
	out = append(out, p.Done.Tasks()...)
	return out
}

// Normalize moves task items that are in the wrong file for their status.
// `today` is the YYYY-MM-DD string used when inserting newly-done tasks into
// done.md under today's heading.
func Normalize(p *Project, today string) error {
	var activeOut, backlogOut, doneOut []Item
	movers := map[FileName][]TaskItem{}

	partition := func(src *Document, self FileName) []Item {
		var kept []Item
		for _, it := range src.Items {
			if ti, ok := it.(TaskItem); ok {
				target := TargetFile(ti.Task.Status)
				if target != self {
					movers[target] = append(movers[target], ti)
					continue
				}
			}
			kept = append(kept, it)
		}
		return kept
	}
	activeOut = partition(p.Active, FileActive)
	backlogOut = partition(p.Backlog, FileBacklog)
	doneOut = partition(p.Done, FileDone)

	// Apply incoming tasks to each destination.
	for _, ti := range movers[FileActive] {
		ti = canonicalizeOutgoing(ti)
		activeOut = append(activeOut, ti)
	}
	for _, ti := range movers[FileBacklog] {
		ti = canonicalizeOutgoing(ti)
		backlogOut = append(backlogOut, ti)
	}

	p.Active = &Document{Items: activeOut}
	p.Backlog = &Document{Items: backlogOut}
	p.Done = insertDone(&Document{Items: doneOut}, movers[FileDone], today)
	p.Done = pruneEmptyDateHeadings(p.Done)
	return nil
}

func canonicalizeOutgoing(ti TaskItem) TaskItem {
	ti.RawLines[0] = RenderTaskLine(ti.Task)
	ti.Task.Date = ""
	return ti
}

// insertDone places newly-done tasks at the top of done.md, under a heading
// for today (creating it if necessary).
func insertDone(doc *Document, incoming []TaskItem, today string) *Document {
	if len(incoming) == 0 {
		return doc
	}
	stamped := make([]Item, 0, len(incoming))
	for _, ti := range incoming {
		ti.Task.Date = today
		ti.RawLines[0] = RenderTaskLine(ti.Task)
		stamped = append(stamped, ti)
	}
	// If a heading for `today` already exists, insert directly after it.
	for i, it := range doc.Items {
		if h, ok := it.(DateHeading); ok && h.Date == today {
			out := &Document{Items: make([]Item, 0, len(doc.Items)+len(stamped))}
			out.Items = append(out.Items, doc.Items[:i+1]...)
			out.Items = append(out.Items, stamped...)
			out.Items = append(out.Items, doc.Items[i+1:]...)
			return out
		}
	}
	// Else prepend a new heading and the tasks.
	out := &Document{Items: make([]Item, 0, len(doc.Items)+len(stamped)+2)}
	out.Items = append(out.Items, DateHeading{Date: today})
	out.Items = append(out.Items, stamped...)
	if len(doc.Items) > 0 {
		out.Items = append(out.Items, OpaqueLines{Lines: []string{""}}) // blank separator
		out.Items = append(out.Items, doc.Items...)
	}
	return out
}

// pruneEmptyDateHeadings removes DateHeading items that have no TaskItems
// between them and the next heading (or end of doc).
func pruneEmptyDateHeadings(doc *Document) *Document {
	out := &Document{Items: make([]Item, 0, len(doc.Items))}
	for i := 0; i < len(doc.Items); i++ {
		if _, ok := doc.Items[i].(DateHeading); ok {
			hasTasks := false
			for j := i + 1; j < len(doc.Items); j++ {
				if _, isH := doc.Items[j].(DateHeading); isH {
					break
				}
				if _, isT := doc.Items[j].(TaskItem); isT {
					hasTasks = true
					break
				}
			}
			if !hasTasks {
				continue
			}
		}
		out.Items = append(out.Items, doc.Items[i])
	}
	return out
}

// ListProjects returns the names of every immediate subdirectory under base
// that contains an active.md file.
func ListProjects(baseDir string) ([]string, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("read base dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(baseDir, e.Name(), "active.md")); err == nil {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
