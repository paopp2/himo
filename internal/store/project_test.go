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

// TestNormalize_singleProjectHeadingAfterMarkDone guards against a regression
// where each new-day "mark done" prepended the new date block before an
// existing leading ProjectHeading, which then caused the save step to add
// another PH and leave the original orphaned mid-document.
func TestNormalize_singleProjectHeadingAfterMarkDone(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, content string) {
		os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	}
	// done.md already has the leading "# work" heading from a prior save.
	write("active.md", "# work\n\n- [ ] A\n- [ ] B\n")
	write("backlog.md", "# work\n\n")
	write("done.md", "# work\n\n")

	toggleFirstPending := func(day string) {
		p, err := LoadProject(dir)
		if err != nil {
			t.Fatal(err)
		}
		for i, it := range p.Active.Items {
			if ti, ok := it.(TaskItem); ok && ti.Task.Status == model.StatusPending {
				ti.Task.Status = model.StatusDone
				ti.RawLines[0] = RenderTaskLine(ti.Task)
				p.Active.Items[i] = ti
				break
			}
		}
		if err := Normalize(p, day); err != nil {
			t.Fatalf("Normalize: %v", err)
		}
		if err := SaveProject(p); err != nil {
			t.Fatalf("SaveProject: %v", err)
		}
	}

	toggleFirstPending("2026-04-18") // first-ever entry into done.md
	b, _ := os.ReadFile(filepath.Join(dir, "done.md"))
	if got := strings.Count(string(b), "# work\n"); got != 1 {
		t.Fatalf("after 1 toggle, want 1 leading heading, got %d:\n%s", got, b)
	}

	toggleFirstPending("2026-04-19") // new day — the scenario that used to add an orphan PH
	b, _ = os.ReadFile(filepath.Join(dir, "done.md"))
	if got := strings.Count(string(b), "# work\n"); got != 1 {
		t.Fatalf("after new-day toggle, want 1 leading heading, got %d:\n%s", got, b)
	}
	if !strings.HasPrefix(string(b), "# work\n\n## 2026-04-19\n") {
		t.Errorf("newest date block must follow the leading heading, got:\n%s", b)
	}
}

// TestInsertDone_preservesLeadingProjectHeading pins the insertDone contract:
// when the doc starts with a ProjectHeading, a new date block is inserted
// after it, not before it.
func TestInsertDone_preservesLeadingProjectHeading(t *testing.T) {
	doc := &Document{Items: []Item{
		ProjectHeading{Name: "work", RawLine: "# work"},
	}}
	incoming := []TaskItem{
		{Task: model.Task{Status: model.StatusDone, Title: "Ship it"},
			RawLines: []string{"- [x] Ship it"}},
	}
	out := insertDone(doc, incoming, "2026-04-18")
	if len(out.Items) == 0 {
		t.Fatalf("empty result")
	}
	if _, ok := out.Items[0].(ProjectHeading); !ok {
		t.Fatalf("items[0] = %T, want ProjectHeading", out.Items[0])
	}
	if dh, ok := out.Items[1].(DateHeading); !ok || dh.Date != "2026-04-18" {
		t.Fatalf("items[1] = %#v, want DateHeading(2026-04-18)", out.Items[1])
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
