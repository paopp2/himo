package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paopp2/himo/internal/model"
)

func TestLoadPriority_missingFileReturnsEmpty(t *testing.T) {
	base := t.TempDir()
	p, err := LoadPriority(base)
	if err != nil {
		t.Fatalf("LoadPriority on missing file: %v", err)
	}
	if len(p.Entries) != 0 {
		t.Errorf("entries on missing file = %d, want 0", len(p.Entries))
	}
	wantPath := filepath.Join(base, ".himo", "active-priority")
	if p.Path != wantPath {
		t.Errorf("path = %q, want %q", p.Path, wantPath)
	}
}

func TestLoadPriority_parsesTabSeparatedLines(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, ".himo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "marketpush\tCreate April invoice\nmp-engine\tToken expired?\n"
	if err := os.WriteFile(filepath.Join(dir, "active-priority"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadPriority(base)
	if err != nil {
		t.Fatal(err)
	}
	want := []PriorityEntry{
		{Project: "marketpush", Title: "Create April invoice"},
		{Project: "mp-engine", Title: "Token expired?"},
	}
	if len(p.Entries) != len(want) {
		t.Fatalf("entries len = %d, want %d (%+v)", len(p.Entries), len(want), p.Entries)
	}
	for i := range want {
		if p.Entries[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, p.Entries[i], want[i])
		}
	}
}

func TestLoadPriority_skipsBlankAndMalformedLines(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, ".himo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// blank line, line without tab, valid line, trailing newline.
	body := "\nnotabhere\nproj\ttitle\n"
	if err := os.WriteFile(filepath.Join(dir, "active-priority"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadPriority(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Entries) != 1 || p.Entries[0] != (PriorityEntry{Project: "proj", Title: "title"}) {
		t.Errorf("entries = %+v, want one (proj, title)", p.Entries)
	}
}

func TestPrioritySave_writesAtomically(t *testing.T) {
	base := t.TempDir()
	p, err := LoadPriority(base)
	if err != nil {
		t.Fatal(err)
	}
	p.Entries = []PriorityEntry{
		{Project: "a", Title: "first"},
		{Project: "b", Title: "second with\ttabbed?? nope just a title"},
	}
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(p.Path)
	if err != nil {
		t.Fatal(err)
	}
	want := "a\tfirst\nb\tsecond with\ttabbed?? nope just a title\n"
	if string(got) != want {
		t.Errorf("file body = %q, want %q", string(got), want)
	}
	// .himo dir must exist.
	if _, err := os.Stat(filepath.Join(base, ".himo")); err != nil {
		t.Errorf("dir not created: %v", err)
	}
	// Verify there is no leftover .tmp.
	entries, _ := os.ReadDir(filepath.Join(base, ".himo"))
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("leftover tmp file: %s", e.Name())
		}
	}
}

func TestPrioritySave_emptyEntriesWritesEmptyFile(t *testing.T) {
	base := t.TempDir()
	p, _ := LoadPriority(base)
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(p.Path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("empty save body = %q, want empty", string(got))
	}
}

func TestPrioritySave_thenLoadRoundTrips(t *testing.T) {
	base := t.TempDir()
	p, _ := LoadPriority(base)
	p.Entries = []PriorityEntry{
		{Project: "p1", Title: "t1"},
		{Project: "p2", Title: "t2"},
	}
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	q, err := LoadPriority(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(q.Entries) != 2 || q.Entries[0] != p.Entries[0] || q.Entries[1] != p.Entries[1] {
		t.Errorf("round-trip mismatch: got %+v, want %+v", q.Entries, p.Entries)
	}
}

func TestPriorityReconcile_dropsOrphansAndAppendsNewcomers(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{
		{Project: "alpha", Title: "kept"},
		{Project: "alpha", Title: "orphan"},   // gone from actives
		{Project: "bravo", Title: "alsokept"}, // still active
	}}
	actives := []PriorityEntry{
		{Project: "alpha", Title: "kept"},
		{Project: "bravo", Title: "alsokept"},
		{Project: "alpha", Title: "newcomer"},
	}
	p.Reconcile(actives)
	want := []PriorityEntry{
		{Project: "alpha", Title: "kept"},
		{Project: "bravo", Title: "alsokept"},
		{Project: "alpha", Title: "newcomer"},
	}
	if len(p.Entries) != len(want) {
		t.Fatalf("entries = %+v, want %+v", p.Entries, want)
	}
	for i := range want {
		if p.Entries[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, p.Entries[i], want[i])
		}
	}
}

func TestPriorityReconcile_emptyIndexAppendsAll(t *testing.T) {
	p := &Priority{}
	actives := []PriorityEntry{
		{Project: "a", Title: "1"},
		{Project: "a", Title: "2"},
	}
	p.Reconcile(actives)
	if len(p.Entries) != 2 {
		t.Fatalf("entries = %+v, want 2", p.Entries)
	}
}

func TestPriorityReconcile_preservesOrderOfSurvivors(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{
		{Project: "p", Title: "third"},
		{Project: "p", Title: "first"},
		{Project: "p", Title: "second"},
	}}
	// Actives universe in some unrelated order — survivors keep index order.
	actives := []PriorityEntry{
		{Project: "p", Title: "first"},
		{Project: "p", Title: "second"},
		{Project: "p", Title: "third"},
	}
	p.Reconcile(actives)
	if p.Entries[0].Title != "third" || p.Entries[1].Title != "first" || p.Entries[2].Title != "second" {
		t.Errorf("survivor order lost: %+v", p.Entries)
	}
}

func TestPriorityIndexOf(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{
		{Project: "a", Title: "x"},
		{Project: "b", Title: "y"},
	}}
	if got := p.IndexOf("b", "y"); got != 1 {
		t.Errorf("IndexOf(b,y) = %d, want 1", got)
	}
	if got := p.IndexOf("nope", "missing"); got != -1 {
		t.Errorf("IndexOf(nope,missing) = %d, want -1", got)
	}
}

func TestPrioritySwapUp(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{
		{Project: "a", Title: "x"},
		{Project: "b", Title: "y"},
		{Project: "c", Title: "z"},
	}}
	if !p.SwapUp("b", "y") {
		t.Fatal("SwapUp returned false on movable entry")
	}
	if p.Entries[0].Title != "y" || p.Entries[1].Title != "x" {
		t.Errorf("after swap up: %+v", p.Entries)
	}
}

func TestPrioritySwapUp_atTopIsNoOp(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{
		{Project: "a", Title: "x"},
		{Project: "b", Title: "y"},
	}}
	if p.SwapUp("a", "x") {
		t.Error("SwapUp at top returned true, want false")
	}
	if p.Entries[0].Title != "x" {
		t.Errorf("entries mutated: %+v", p.Entries)
	}
}

func TestPrioritySwapUp_absentReturnsFalse(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{{Project: "a", Title: "x"}}}
	if p.SwapUp("nope", "missing") {
		t.Error("SwapUp on absent returned true")
	}
}

func TestPrioritySwapDown(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{
		{Project: "a", Title: "x"},
		{Project: "b", Title: "y"},
		{Project: "c", Title: "z"},
	}}
	if !p.SwapDown("a", "x") {
		t.Fatal("SwapDown returned false on movable entry")
	}
	if p.Entries[0].Title != "y" || p.Entries[1].Title != "x" {
		t.Errorf("after swap down: %+v", p.Entries)
	}
}

func TestPrioritySwapDown_atBottomIsNoOp(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{
		{Project: "a", Title: "x"},
		{Project: "b", Title: "y"},
	}}
	if p.SwapDown("b", "y") {
		t.Error("SwapDown at bottom returned true")
	}
}

func TestPriorityRename_preservesPosition(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{
		{Project: "a", Title: "old"},
		{Project: "a", Title: "stay"},
	}}
	p.Rename("a", "old", "new")
	if p.Entries[0].Title != "new" || p.Entries[1].Title != "stay" {
		t.Errorf("after rename: %+v", p.Entries)
	}
}

func TestPriorityRename_absentIsNoOp(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{{Project: "a", Title: "x"}}}
	p.Rename("a", "nope", "new")
	if p.Entries[0].Title != "x" {
		t.Errorf("mutated on absent: %+v", p.Entries)
	}
}

func TestPriorityRemove(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{
		{Project: "a", Title: "x"},
		{Project: "b", Title: "y"},
		{Project: "c", Title: "z"},
	}}
	p.Remove("b", "y")
	if len(p.Entries) != 2 || p.Entries[0].Title != "x" || p.Entries[1].Title != "z" {
		t.Errorf("after remove: %+v", p.Entries)
	}
}

func TestPriorityRemove_absentIsNoOp(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{{Project: "a", Title: "x"}}}
	p.Remove("nope", "missing")
	if len(p.Entries) != 1 {
		t.Errorf("mutated on absent: %+v", p.Entries)
	}
}

func TestPriorityAppend(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{{Project: "a", Title: "x"}}}
	p.Append("b", "y")
	if len(p.Entries) != 2 || p.Entries[1] != (PriorityEntry{Project: "b", Title: "y"}) {
		t.Errorf("after append: %+v", p.Entries)
	}
}

func TestPriorityAppend_duplicateIsNoOp(t *testing.T) {
	p := &Priority{Entries: []PriorityEntry{{Project: "a", Title: "x"}}}
	p.Append("a", "x")
	if len(p.Entries) != 1 {
		t.Errorf("duplicate appended: %+v", p.Entries)
	}
}

func TestActiveEntries_returnsOnlyActiveTasksInFileOrder(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("- [ ] pending one\n- [/] active one\n- [/] active two\n- [!] blocked one\n"),
		0o644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := ActiveEntries([]*Project{p})
	want := []PriorityEntry{
		{Project: filepath.Base(dir), Title: "active one"},
		{Project: filepath.Base(dir), Title: "active two"},
	}
	if len(got) != len(want) {
		t.Fatalf("ActiveEntries = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	_ = model.StatusActive // keep import used even if model unused above.
}

func TestActiveEntries_acrossMultipleProjects_alphabetical(t *testing.T) {
	base := t.TempDir()
	mk := func(name, body string) *Project {
		dir := filepath.Join(base, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "active.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		p, err := LoadProject(dir)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	// Pass projects in NON-alphabetical order; ActiveEntries must NOT resort.
	pZ := mk("zproj", "- [/] z1\n")
	pA := mk("aproj", "- [/] a1\n")
	got := ActiveEntries([]*Project{pZ, pA})
	want := []PriorityEntry{
		{Project: "zproj", Title: "z1"},
		{Project: "aproj", Title: "a1"},
	}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("ActiveEntries respects caller's project order: got %+v, want %+v", got, want)
	}
}
