package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderView_goldenNormal(t *testing.T) {
	m := NewModel(testProject(t))
	m.width, m.height = 120, 30
	m.styles = testStyles(t)

	got := renderView(m)
	path := filepath.Join("testdata", "golden", "normal_120x30.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run UPDATE_GOLDEN=1 to generate): %v", err)
	}
	if string(want) != got {
		t.Errorf("golden mismatch. Run UPDATE_GOLDEN=1 to update.\n--- got ---\n%s\n--- want ---\n%s",
			got, string(want))
	}
}

func TestRenderView_goldenEdit(t *testing.T) {
	m := NewModel(testProject(t))
	m.width, m.height = 120, 30
	m.styles = testStyles(t)
	m = keypress(t, m, keyRune('e'))
	for range m.editBuf {
		m = keypress(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	m = typeString(t, m, "Buy gro")

	got := renderView(m)
	path := filepath.Join("testdata", "golden", "edit_120x30.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run UPDATE_GOLDEN=1 to generate): %v", err)
	}
	if string(want) != got {
		t.Errorf("golden mismatch. Run UPDATE_GOLDEN=1 to update.\n--- got ---\n%s\n--- want ---\n%s",
			got, string(want))
	}
}
