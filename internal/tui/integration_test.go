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

func TestIntegration_search(t *testing.T) {
	tm, _ := newIntegrationTestModel(t)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "groc" {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		s := string(b)
		return strings.Contains(s, "Buy groceries") && !strings.Contains(s, "Write design")
	}, teatest.WithDuration(2*time.Second))
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t)
}

func TestIntegration_newTaskInBacklog(t *testing.T) {
	tm, p := newIntegrationTestModel(t)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	for _, r := range "Refactor auth" {
		if r == ' ' {
			tm.Send(tea.KeyMsg{Type: tea.KeySpace})
			continue
		}
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
