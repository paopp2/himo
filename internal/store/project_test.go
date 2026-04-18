package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paopp2/himo/internal/model"
)

func TestLoadProject(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("active.md", "- [ ] A\n- [/] B\n")
	write("backlog.md", "- C\n")
	write("done.md", "# 2026-04-18\n- [x] D\n")

	p, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != filepath.Base(dir) {
		t.Errorf("Name = %q, want %q", p.Name, filepath.Base(dir))
	}
	if n := len(p.AllTasks()); n != 4 {
		t.Errorf("AllTasks: got %d, want 4", n)
	}
}

func TestNormalize_markDone(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	}
	write("active.md", "- [/] Ship RFC\n- [ ] Buy groceries\n")
	write("backlog.md", "")
	write("done.md", "")

	p, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate user editing: mark first task as done.
	for _, it := range p.Active.Items {
		if ti, ok := it.(TaskItem); ok && ti.Task.Title == "Ship RFC" {
			ti.Task.Status = model.StatusDone
			ti.RawLines[0] = "- [x] Ship RFC"
			p.Active = replaceTask(p.Active, ti)
			break
		}
	}
	if err := Normalize(p, "2026-04-18"); err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if err := SaveProject(p); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	active, _ := os.ReadFile(filepath.Join(dir, "active.md"))
	if strings.Contains(string(active), "Ship RFC") {
		t.Errorf("Ship RFC still in active.md:\n%s", active)
	}
	done, _ := os.ReadFile(filepath.Join(dir, "done.md"))
	if !strings.Contains(string(done), "# 2026-04-18") || !strings.Contains(string(done), "[x] Ship RFC") {
		t.Errorf("done.md missing Ship RFC under today:\n%s", done)
	}
}

// replaceTask returns a new Document with the given TaskItem replacing the
// one with the same title. Helper for tests.
func replaceTask(doc *Document, ti TaskItem) *Document {
	out := &Document{Items: make([]Item, len(doc.Items))}
	copy(out.Items, doc.Items)
	for i, it := range out.Items {
		if existing, ok := it.(TaskItem); ok && existing.Task.Title == ti.Task.Title {
			out.Items[i] = ti
			return out
		}
	}
	return out
}
