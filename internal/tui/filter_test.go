package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/paopp2/himo/internal/model"
)

func TestFilter_numberKeys(t *testing.T) {
	cases := []struct {
		key  rune
		want []model.Status
	}{
		{'1', []model.Status{model.StatusBacklog}},
		{'2', []model.Status{model.StatusPending}},
		{'3', []model.Status{model.StatusActive}},
		{'4', []model.Status{model.StatusBlocked}},
		{'5', []model.Status{model.StatusDone}},
		{'6', []model.Status{model.StatusCancelled}},
	}
	for _, tt := range cases {
		m := NewModel(testProject(t))
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
		got := m2.(Model).filter.Statuses
		if len(got) != 1 || got[0] != tt.want[0] {
			t.Errorf("key %c: filter = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestFilter_zero(t *testing.T) {
	m := NewModel(testProject(t))
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	if !m2.(Model).filter.All {
		t.Errorf("filter.All = false after 0, want true")
	}
}

// Backtick resets to the default Pending+Active+Blocked view from any filter.
func TestFilter_backtickResetsToDefault(t *testing.T) {
	m := NewModel(testProject(t))
	m.filter = Filter{Statuses: []model.Status{model.StatusDone}}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'`'}})
	f := m2.(Model).filter
	if f.All {
		t.Errorf("after `, filter.All = true, want default")
	}
	want := map[model.Status]bool{
		model.StatusPending: true, model.StatusActive: true, model.StatusBlocked: true,
	}
	if len(f.Statuses) != 3 {
		t.Fatalf("after `, filter.Statuses has %d, want 3", len(f.Statuses))
	}
	for _, s := range f.Statuses {
		if !want[s] {
			t.Errorf("after `, unexpected status %v", s)
		}
	}
}

// Esc on a single-status filter is a no-op for the filter (Esc is for All /
// search exits only). The filter must survive unchanged.
func TestFilter_escDoesNotResetCustomFilter(t *testing.T) {
	m := NewModel(testProject(t))
	m.filter = Filter{Statuses: []model.Status{model.StatusDone}}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	f := m2.(Model).filter
	if len(f.Statuses) != 1 || f.Statuses[0] != model.StatusDone {
		t.Errorf("after Esc, filter = %+v, want Done only", f)
	}
}
