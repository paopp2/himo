package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveProject_atomic(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "active.md"), []byte("- [ ] A\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "backlog.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "done.md"), []byte(""), 0o644)

	p, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveProject(p); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	b, _ := os.ReadFile(filepath.Join(dir, "active.md"))
	if string(b) != "- [ ] A\n" {
		t.Errorf("active.md after save = %q, want unchanged", string(b))
	}

	// No .tmp files left behind.
	matches, _ := filepath.Glob(filepath.Join(dir, "*.tmp"))
	if len(matches) != 0 {
		t.Errorf("leftover .tmp files: %v", matches)
	}
}

func TestSaveProject_mtimeConflict(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "active.md"), []byte("- [ ] A\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "backlog.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "done.md"), []byte(""), 0o644)

	p, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate an external edit changing the mtime.
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "active.md"), []byte("- [ ] external\n"), 0o644)

	err = SaveProject(p)
	if !IsConflict(err) {
		t.Errorf("SaveProject on stale project err = %v, want ErrConflict", err)
	}
}
