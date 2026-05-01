package store

import (
	"bufio"
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

// render returns the on-disk byte representation of p.
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
