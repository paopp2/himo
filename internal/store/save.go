package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrConflict indicates that a file was changed externally since load.
var ErrConflict = errors.New("file changed externally since load")

// IsConflict reports whether err indicates a mtime conflict.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// SaveProject writes all three files atomically. It first checks each file's
// mtime against the value recorded at load; if any has changed, it returns
// ErrConflict without writing anything.
func SaveProject(p *Project) error {
	// Conflict check phase.
	for _, name := range allFiles {
		info, err := os.Stat(filepath.Join(p.Dir, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("stat %s: %w", name, err)
		}
		if recorded, ok := p.mtimes[name]; ok && !info.ModTime().Equal(recorded) {
			return fmt.Errorf("%s: %w", name, ErrConflict)
		}
	}
	// Write phase.
	writes := []struct {
		name string
		doc  *Document
	}{
		{"active.md", p.Active},
		{"backlog.md", p.Backlog},
		{"done.md", p.Done},
	}
	for _, w := range writes {
		if err := writeAtomic(filepath.Join(p.Dir, w.name), Render(w.doc)); err != nil {
			return fmt.Errorf("write %s: %w", w.name, err)
		}
	}
	// Refresh mtimes so subsequent saves don't spuriously conflict.
	for _, name := range allFiles {
		if info, err := os.Stat(filepath.Join(p.Dir, name)); err == nil {
			p.mtimes[name] = info.ModTime()
		}
	}
	return nil
}

func writeAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
