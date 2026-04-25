package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paopp2/himo/internal/model"
	"github.com/paopp2/himo/internal/store"
)

func TestSortFromName_roundTrip(t *testing.T) {
	cases := []struct {
		name string
		want Sort
	}{
		{"natural", SortNatural},
		{"status", SortStatus},
		{"", SortNatural},
		{"bogus", SortNatural},
	}
	for _, tt := range cases {
		got := SortFromName(tt.name)
		if got != tt.want {
			t.Errorf("SortFromName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
	if got := sortName(SortNatural); got != "natural" {
		t.Errorf("sortName(SortNatural) = %q, want %q", got, "natural")
	}
	if got := sortName(SortStatus); got != "status" {
		t.Errorf("sortName(SortStatus) = %q, want %q", got, "status")
	}
}

func TestStatusSortRank_attentionFirst(t *testing.T) {
	want := []model.Status{
		model.StatusActive,
		model.StatusBlocked,
		model.StatusPending,
		model.StatusBacklog,
		model.StatusDone,
		model.StatusCancelled,
	}
	for i, s := range want {
		if got := statusSortRank(s); got != i {
			t.Errorf("statusSortRank(%v) = %d, want %d", s, got, i)
		}
	}
}

// sortFixtureProject writes a project with one task per status so the
// ordering effect is unambiguous. Source order in active.md is
// pending -> active -> blocked, which differs from SortStatus order.
func sortFixtureProject(t *testing.T) *store.Project {
	t.Helper()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [ ] pen\n- [/] act\n- [!] blk\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "backlog.md"),
		[]byte("- bak\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "done.md"),
		[]byte("## 2026-01-01\n- [x] don\n- [-] can\n"), 0o644)
	p, err := store.LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestVisibleTasks_SortNaturalUnchanged(t *testing.T) {
	m := NewModel(sortFixtureProject(t))
	m.filter = Filter{All: true}
	got := titles(m.visibleTasks())
	want := []string{"pen", "act", "blk", "bak", "don", "can"}
	assertSequence(t, "SortNatural", got, want)
}

func TestVisibleTasks_SortStatusReordersByRank(t *testing.T) {
	m := NewModel(sortFixtureProject(t))
	m.filter = Filter{All: true}
	m.sort = SortStatus
	got := titles(m.visibleTasks())
	want := []string{"act", "blk", "pen", "bak", "don", "can"}
	assertSequence(t, "SortStatus", got, want)
}

func assertSequence(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %d tasks (%v), want %d (%v)", label, len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s: row %d = %q, want %q (full got: %v)", label, i, got[i], want[i], got)
		}
	}
}

// sortFixtureMultiProject writes two projects, each with one active and
// one done task. The status interleave assertion is robust to project
// listing order: SortStatus must surface both actives before either done.
func sortFixtureMultiProject(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	for _, name := range []string{"alpha", "bravo"} {
		os.MkdirAll(filepath.Join(base, name), 0o755)
		os.WriteFile(filepath.Join(base, name, "active.md"),
			[]byte("- [/] "+name+"-act\n"), 0o644)
		os.WriteFile(filepath.Join(base, name, "done.md"),
			[]byte("## 2026-01-01\n- [x] "+name+"-don\n"), 0o644)
	}
	return base
}

func TestVisibleTasks_SortStatusInterleavesProjects(t *testing.T) {
	base := sortFixtureMultiProject(t)
	m, err := NewModelFromBase(base, "alpha", StyleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m = m.WithAllProjects()
	m.filter = Filter{All: true}
	m.sort = SortStatus
	got := titles(m.visibleTasks())
	if len(got) != 4 {
		t.Fatalf("got %d tasks (%v), want 4", len(got), got)
	}
	// First two rows are the two actives, last two are the dones; the
	// alpha/bravo order within each rank is whatever ListProjects yields.
	for i, suffix := range []string{"-act", "-act", "-don", "-don"} {
		if !strings.HasSuffix(got[i], suffix) {
			t.Errorf("row %d = %q, want suffix %q (full got: %v)", i, got[i], suffix, got)
		}
	}
}
