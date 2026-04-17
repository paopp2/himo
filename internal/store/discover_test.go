package store

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestListProjects(t *testing.T) {
	base := t.TempDir()
	os.Mkdir(filepath.Join(base, "work"), 0o755)
	os.WriteFile(filepath.Join(base, "work", "active.md"), []byte(""), 0o644)
	os.Mkdir(filepath.Join(base, "personal"), 0o755)
	os.WriteFile(filepath.Join(base, "personal", "active.md"), []byte(""), 0o644)
	os.Mkdir(filepath.Join(base, "not-a-project"), 0o755) // missing active.md

	names, err := ListProjects(base)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(names)
	want := []string{"personal", "work"}
	if len(names) != len(want) || names[0] != want[0] || names[1] != want[1] {
		t.Errorf("ListProjects = %v, want %v", names, want)
	}
}
