package store

import (
	"bufio"
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/paopp2/himo/internal/model"
)

// PriorityEntry identifies one active task in the priority index.
type PriorityEntry struct {
	Project string
	Title   string
}

// Priority is an ordered list of active task entries plus the on-disk path.
// Order in Entries is the priority order; index 0 is highest priority.
type Priority struct {
	Path    string
	Entries []PriorityEntry
}

// LoadPriority reads <baseDir>/.himo/active-priority. A missing file returns
// an empty Priority with the path filled in (so Save will create the file
// at the right location).
func LoadPriority(baseDir string) (*Priority, error) {
	path := filepath.Join(baseDir, ".himo", "active-priority")
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &Priority{Path: path}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	p := &Priority{Path: path}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		tab := strings.IndexByte(line, '\t')
		if tab <= 0 || tab == len(line)-1 {
			continue
		}
		p.Entries = append(p.Entries, PriorityEntry{
			Project: line[:tab],
			Title:   line[tab+1:],
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Priority) render() []byte {
	var buf bytes.Buffer
	for _, e := range p.Entries {
		buf.WriteString(e.Project)
		buf.WriteByte('\t')
		buf.WriteString(e.Title)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

// Save writes the priority list atomically. The parent directory is created
// if missing. An empty entry list yields an empty file.
func (p *Priority) Save() error {
	if err := os.MkdirAll(filepath.Dir(p.Path), 0o755); err != nil {
		return err
	}
	return writeAtomic(p.Path, p.render())
}

// Reconcile aligns p.Entries with the universe of currently-active tasks.
// 1. Any entry whose (project, title) is not in `actives` is dropped.
// 2. Any (project, title) in `actives` not in p.Entries is appended in
//    the order it appears in `actives`.
// 3. Duplicate entries within p.Entries are de-duplicated to the first
//    occurrence (defensive against externally-edited indexes).
// Order of surviving entries is preserved.
func (p *Priority) Reconcile(actives []PriorityEntry) {
	have := make(map[PriorityEntry]bool, len(actives))
	for _, e := range actives {
		have[e] = true
	}
	kept := make([]PriorityEntry, 0, len(actives))
	keptSet := make(map[PriorityEntry]bool, len(p.Entries))
	for _, e := range p.Entries {
		if have[e] && !keptSet[e] {
			kept = append(kept, e)
			keptSet[e] = true
		}
	}
	for _, e := range actives {
		if !keptSet[e] {
			kept = append(kept, e)
			keptSet[e] = true
		}
	}
	p.Entries = kept
}

func (p *Priority) IndexOf(project, title string) int {
	for i, e := range p.Entries {
		if e.Project == project && e.Title == title {
			return i
		}
	}
	return -1
}

// Mutators below are no-ops when the entry is absent.

func (p *Priority) SwapUp(project, title string) bool {
	i := p.IndexOf(project, title)
	if i <= 0 {
		return false
	}
	p.Entries[i-1], p.Entries[i] = p.Entries[i], p.Entries[i-1]
	return true
}

func (p *Priority) SwapDown(project, title string) bool {
	i := p.IndexOf(project, title)
	if i < 0 || i >= len(p.Entries)-1 {
		return false
	}
	p.Entries[i+1], p.Entries[i] = p.Entries[i], p.Entries[i+1]
	return true
}

// Rename rewrites the title of the entry matching (project, oldTitle).
// Position is preserved.
func (p *Priority) Rename(project, oldTitle, newTitle string) {
	if i := p.IndexOf(project, oldTitle); i >= 0 {
		p.Entries[i].Title = newTitle
	}
}

func (p *Priority) Remove(project, title string) {
	if i := p.IndexOf(project, title); i >= 0 {
		p.Entries = append(p.Entries[:i], p.Entries[i+1:]...)
	}
}

func (p *Priority) Append(project, title string) {
	if p.IndexOf(project, title) >= 0 {
		return
	}
	p.Entries = append(p.Entries, PriorityEntry{Project: project, Title: title})
}

// ActiveEntries returns one PriorityEntry per active task across the given
// projects, in caller order (project order, then file order within each
// project's active.md).
func ActiveEntries(projects []*Project) []PriorityEntry {
	var out []PriorityEntry
	for _, p := range projects {
		if p == nil || p.Active == nil {
			continue
		}
		for _, it := range p.Active.Items {
			ti, ok := it.(TaskItem)
			if !ok {
				continue
			}
			if ti.Task.Status != model.StatusActive {
				continue
			}
			out = append(out, PriorityEntry{Project: p.Name, Title: ti.Task.Title})
		}
	}
	return out
}
