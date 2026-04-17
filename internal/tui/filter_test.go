package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/npaolopepito/himo/internal/model"
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

func TestFilter_esc(t *testing.T) {
	m := NewModel(testProject(t))
	m.filter = Filter{All: true}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	f := m2.(Model).filter
	if f.All {
		t.Errorf("after Esc, filter.All = true, want default")
	}
	if len(f.Statuses) != 3 {
		t.Errorf("after Esc, filter.Statuses has %d, want 3", len(f.Statuses))
	}
}
