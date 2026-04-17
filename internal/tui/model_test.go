package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

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
	return &store.Project{
		Name:    "t",
		Dir:     t.TempDir(),
		Active:  &store.Document{},
		Backlog: &store.Document{},
		Done:    &store.Document{},
	}
}
