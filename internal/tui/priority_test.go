package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paopp2/himo/internal/store"
)

func TestNewModelFromBase_loadsAndReconcilesPriority(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "p")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Two active tasks in active.md.
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [/] alpha\n- [/] bravo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Pre-existing priority file ranks bravo first AND has an orphan.
	prDir := filepath.Join(base, ".himo")
	if err := os.MkdirAll(prDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prDir, "active-priority"),
		[]byte("p\tbravo\np\torphan\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := NewModelFromBase(base, "p", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	pr := m.priorityForTest()
	want := []store.PriorityEntry{
		{Project: "p", Title: "bravo"},
		{Project: "p", Title: "alpha"},
	}
	if len(pr.Entries) != len(want) {
		t.Fatalf("priority entries = %+v, want %+v", pr.Entries, want)
	}
	for i := range want {
		if pr.Entries[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, pr.Entries[i], want[i])
		}
	}
}
