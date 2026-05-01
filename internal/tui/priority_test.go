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

func TestVisibleTasks_SortStatus_ActiveGroupOrdersByPriority(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "p")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// File order: alpha (active), pending, bravo (active).
	// SortStatus puts the two actives first; priority decides their order.
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [/] alpha\n- [ ] pen\n- [/] bravo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	prDir := filepath.Join(base, ".himo")
	if err := os.MkdirAll(prDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prDir, "active-priority"),
		[]byte("p\tbravo\np\talpha\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewModelFromBase(base, "p", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m.filter = Filter{All: true}
	m.sort = SortStatus
	got := titles(m.visibleTasks())
	// First two are actives in priority order; third is pending.
	if got[0] != "bravo" || got[1] != "alpha" || got[2] != "pen" {
		t.Errorf("got %v, want [bravo alpha pen]", got)
	}
}

func TestReorder_ShiftJ_movesTaskDownInActiveFilter(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "p")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [/] one\n- [/] two\n- [/] three\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewModelFromBase(base, "p", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m.filter = Filter{Statuses: []model.Status{model.StatusActive}}
	m.cursor = 0 // on "one"

	out, _ := m.Update(keyRune('J'))
	got := titles(out.(Model).visibleTasks())
	want := []string{"two", "one", "three"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d = %q, want %q (full %v)", i, got[i], want[i], got)
		}
	}
	// Cursor follows the moved task.
	if out.(Model).cursor != 1 {
		t.Errorf("cursor = %d, want 1", out.(Model).cursor)
	}
}

func TestReorder_ShiftK_movesTaskUpInActiveFilter(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "p")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [/] one\n- [/] two\n- [/] three\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewModelFromBase(base, "p", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m.filter = Filter{Statuses: []model.Status{model.StatusActive}}
	m.cursor = 2 // on "three"

	out, _ := m.Update(keyRune('K'))
	got := titles(out.(Model).visibleTasks())
	want := []string{"one", "three", "two"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d = %q, want %q (full %v)", i, got[i], want[i], got)
		}
	}
	if out.(Model).cursor != 1 {
		t.Errorf("cursor = %d, want 1", out.(Model).cursor)
	}
}

func TestReorder_ShiftJ_atBottom_isNoOp(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "p")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [/] one\n- [/] two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewModelFromBase(base, "p", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m.filter = Filter{Statuses: []model.Status{model.StatusActive}}
	m.cursor = 1

	out, _ := m.Update(keyRune('J'))
	got := titles(out.(Model).visibleTasks())
	if got[0] != "one" || got[1] != "two" {
		t.Errorf("got %v, want unchanged", got)
	}
	if out.(Model).cursor != 1 {
		t.Errorf("cursor moved on no-op")
	}
}

func TestReorder_persistsToDisk(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "p")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [/] one\n- [/] two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewModelFromBase(base, "p", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m.filter = Filter{Statuses: []model.Status{model.StatusActive}}
	m.cursor = 0

	_, _ = m.Update(keyRune('J'))

	pr, err := store.LoadPriority(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(pr.Entries) != 2 || pr.Entries[0].Title != "two" || pr.Entries[1].Title != "one" {
		t.Errorf("disk state = %+v, want [two, one]", pr.Entries)
	}
}

func TestReorder_inDefaultFilter_setsHint(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "p")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [/] one\n- [ ] pen\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewModelFromBase(base, "p", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// Default filter (pending+active+blocked) — reorder NOT enabled.
	m.cursor = 0

	out, _ := m.Update(keyRune('J'))
	got := out.(Model)
	if got.banner == "" {
		t.Errorf("expected hint banner, got empty")
	}
	// And tasks did not move.
	titles := titles(got.visibleTasks())
	if titles[0] != "one" || titles[1] != "pen" {
		t.Errorf("tasks moved on no-op gesture: %v", titles)
	}
}

func TestReorder_onNonActiveTask_inSortStatus_setsHint(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "p")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [/] act\n- [ ] pen\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewModelFromBase(base, "p", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m.filter = Filter{All: true}
	m.sort = SortStatus
	// SortStatus puts active first, so cursor=1 lands on the pending task.
	m.cursor = 1

	out, _ := m.Update(keyRune('J'))
	got := out.(Model)
	if got.banner == "" {
		t.Errorf("expected hint banner on non-active reorder")
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
