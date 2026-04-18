package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
)

func TestModel_initAndQuit(t *testing.T) {
	m := NewModel(testProject(t))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func testProject(t *testing.T) *store.Project {
	t.Helper()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "active.md"), []byte("- [ ] Buy groceries\n- [/] Write design\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "backlog.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "done.md"), []byte(""), 0o644)
	p, err := store.LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestView_showsTasks(t *testing.T) {
	m := NewModel(testProject(t))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		s := string(b)
		return strings.Contains(s, "Buy groceries") && strings.Contains(s, "Write design")
	}, teatest.WithDuration(time.Second))
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t)
}

func TestNewModel_hasStyles(t *testing.T) {
	m := NewModel(testProject(t))
	if m.styles == nil {
		t.Fatal("m.styles is nil")
	}
	if got := m.styles.StatusGlyph(model.StatusActive); got != "●" {
		t.Errorf("default glyph for active = %q, want ●", got)
	}
}

func TestNewModel_asciiGlyphsOption(t *testing.T) {
	m := NewModelWithOptions(testProject(t), StyleOptions{AsciiGlyphs: true})
	if got := m.styles.StatusGlyph(model.StatusActive); got != "*" {
		t.Errorf("ascii glyph for active = %q, want *", got)
	}
}
