package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSave_insertsProjectHeading(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "active.md"), []byte("- [ ] A\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "backlog.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "done.md"), []byte(""), 0o644)

	p, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	projName := filepath.Base(dir)
	if err := SaveProject(p); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"active.md", "backlog.md", "done.md"} {
		b, _ := os.ReadFile(filepath.Join(dir, name))
		if !strings.HasPrefix(string(b), "# "+projName+"\n") {
			t.Errorf("%s: missing project heading:\n%s", name, string(b))
		}
	}
}

func TestSave_preservesCustomHeading(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "active.md"),
		[]byte("# Custom Title\n\n- [ ] A\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "backlog.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "done.md"), []byte(""), 0o644)

	p, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveProject(p); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "active.md"))
	if !strings.HasPrefix(string(b), "# Custom Title\n") {
		t.Errorf("custom heading overwritten:\n%s", string(b))
	}
}

func TestSave_updatesHeadingOnDirRename(t *testing.T) {
	// The file has "# old-project" in its H1 but the directory name is
	// "new-project". The plain-name rule rewrites the H1 to match.
	parent := t.TempDir()
	newDir := filepath.Join(parent, "new-project")
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(newDir, "active.md"),
		[]byte("# old-project\n\n- [ ] A\n"), 0o644)
	os.WriteFile(filepath.Join(newDir, "backlog.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(newDir, "done.md"), []byte(""), 0o644)

	p, err := LoadProject(newDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveProject(p); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(newDir, "active.md"))
	if !strings.HasPrefix(string(b), "# new-project\n") {
		t.Errorf("heading not rewritten to match dir:\n%s", string(b))
	}
}
