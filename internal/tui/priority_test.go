package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paopp2/himo/internal/model"
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

func TestVisibleTasks_ActiveFilter_OrdersByPriority(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "p")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// File order: alpha, bravo, charlie. Priority will rank charlie first.
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [/] alpha\n- [/] bravo\n- [/] charlie\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	prDir := filepath.Join(base, ".himo")
	if err := os.MkdirAll(prDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prDir, "active-priority"),
		[]byte("p\tcharlie\np\talpha\np\tbravo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := NewModelFromBase(base, "p", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m.filter = Filter{Statuses: []model.Status{model.StatusActive}}
	got := titles(m.visibleTasks())
	want := []string{"charlie", "alpha", "bravo"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d = %q, want %q (full %v)", i, got[i], want[i], got)
		}
	}
}

func TestVisibleTasks_AllProjectsActive_OrdersByGlobalPriority(t *testing.T) {
	base := t.TempDir()
	mkProj := func(name, body string) {
		dir := filepath.Join(base, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "active.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mkProj("alpha", "- [/] a1\n- [/] a2\n")
	mkProj("bravo", "- [/] b1\n")
	prDir := filepath.Join(base, ".himo")
	if err := os.MkdirAll(prDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Cross-project priority: b1 then a2 then a1.
	if err := os.WriteFile(filepath.Join(prDir, "active-priority"),
		[]byte("bravo\tb1\nalpha\ta2\nalpha\ta1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewModelFromBase(base, "alpha", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m = m.WithAllProjects()
	m.filter = Filter{Statuses: []model.Status{model.StatusActive}}
	got := titles(m.visibleTasks())
	want := []string{"b1", "a2", "a1"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d = %q, want %q (full %v)", i, got[i], want[i], got)
		}
	}
}
