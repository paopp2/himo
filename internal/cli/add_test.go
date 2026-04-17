package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddTask(t *testing.T) {
	base := t.TempDir()
	os.MkdirAll(filepath.Join(base, "work"), 0o755)
	os.WriteFile(filepath.Join(base, "work", "active.md"), []byte(""), 0o644)

	if err := AddTask(base, "work", "Buy groceries"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(base, "work", "active.md"))
	if !strings.Contains(string(b), "- [ ] Buy groceries") {
		t.Errorf("active.md missing task:\n%s", b)
	}
}

func TestAddTask_appendsBelowExisting(t *testing.T) {
	base := t.TempDir()
	os.MkdirAll(filepath.Join(base, "work"), 0o755)
	os.WriteFile(filepath.Join(base, "work", "active.md"), []byte("- [/] A\n"), 0o644)

	if err := AddTask(base, "work", "B"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(base, "work", "active.md"))
	want := "- [/] A\n- [ ] B\n"
	if string(b) != want {
		t.Errorf("active.md = %q, want %q", b, want)
	}
}

func TestAddTask_missingProject(t *testing.T) {
	base := t.TempDir()
	if err := AddTask(base, "nope", "anything"); err == nil {
		t.Errorf("AddTask on missing project: want error")
	}
}
