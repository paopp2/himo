package store

import (
	"os"
	"path/filepath"
	"testing"
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
