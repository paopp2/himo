package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLs_allProject(t *testing.T) {
	base := t.TempDir()
	os.MkdirAll(filepath.Join(base, "work"), 0o755)
	os.WriteFile(filepath.Join(base, "work", "active.md"), []byte("- [ ] A\n- [/] B\n"), 0o644)
	os.WriteFile(filepath.Join(base, "work", "backlog.md"), []byte("- C\n"), 0o644)
	os.WriteFile(filepath.Join(base, "work", "done.md"), []byte("# 2026-04-18\n- [x] D\n"), 0o644)

	var out bytes.Buffer
	if err := Ls(base, "work", "", &out); err != nil {
		t.Fatalf("Ls: %v", err)
	}
	got := out.String()
	for _, want := range []string{"A", "B", "C", "D"} {
		if !strings.Contains(got, want) {
			t.Errorf("Ls output missing %q:\n%s", want, got)
		}
	}
}

func TestLs_filterByStatus(t *testing.T) {
	base := t.TempDir()
	os.MkdirAll(filepath.Join(base, "work"), 0o755)
	os.WriteFile(filepath.Join(base, "work", "active.md"), []byte("- [ ] A\n- [/] B\n"), 0o644)
	os.WriteFile(filepath.Join(base, "work", "backlog.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(base, "work", "done.md"), []byte(""), 0o644)

	var out bytes.Buffer
	if err := Ls(base, "work", "pending", &out); err != nil {
		t.Fatalf("Ls: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "A") {
		t.Errorf("expected A in output:\n%s", got)
	}
	if strings.Contains(got, "B") {
		t.Errorf("did not expect B in pending filter:\n%s", got)
	}
}
