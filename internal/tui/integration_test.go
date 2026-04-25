package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/paopp2/himo/internal/store"
)

func newIntegrationTestModel(t *testing.T) (*teatest.TestModel, *store.Project) {
	t.Helper()
	p := testProject(t)
	m := NewModel(p)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))
	return tm, p
}

func TestIntegration_cycleAndFilterDone(t *testing.T) {
	tm, _ := newIntegrationTestModel(t)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		s := string(b)
		return strings.Contains(s, "Write design") && strings.Contains(s, "✓")
	}, teatest.WithDuration(2*time.Second))
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t)
}

func TestIntegration_newTaskInBacklog(t *testing.T) {
	tm, p := newIntegrationTestModel(t)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	for _, r := range "Refactor auth" {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(string(b), "Refactor auth")
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t)

	data, err := os.ReadFile(filepath.Join(p.Dir, "backlog.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "- Refactor auth") {
		t.Errorf("backlog.md missing new task:\n%s", string(data))
	}
}

func TestSearch_endToEndSession(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 0

	cur := tea.Model(m)
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !cur.(Model).searching {
		t.Fatal("expected searching=true after '/'")
	}

	for _, r := range "design" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if got := cur.(Model).cursor; got != 1 {
		t.Fatalf("incsearch cursor = %d, want 1", got)
	}

	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cur.(Model).searching {
		t.Fatal("expected searching=false after Enter")
	}
	if got := cur.(Model).searchActive; got != "design" {
		t.Fatalf("searchActive = %q, want 'design'", got)
	}

	// 'n' wraps because "design" matches only the second task.
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if got := cur.(Model).cursor; got != 1 {
		t.Errorf("n with single match: cursor = %d, want 1 (stays)", got)
	}
	if got := cur.(Model).banner; got != "search hit BOTTOM, continuing at TOP" {
		t.Errorf("banner = %q", got)
	}

	// Esc in normal mode clears searchActive.
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if got := cur.(Model).searchActive; got != "" {
		t.Errorf("searchActive after Esc = %q, want empty", got)
	}
}
