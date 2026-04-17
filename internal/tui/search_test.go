package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/npaolopepito/himo/internal/model"
)

func TestSearch_filtersByTitle(t *testing.T) {
	m := NewModel(testProject(t))
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "groc" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	visible := cur.(Model).visibleTasks()
	if len(visible) != 1 || visible[0].Title != "Buy groceries" {
		t.Errorf("search 'groc': visible = %v, want [Buy groceries]", titles(visible))
	}
}

func titles(ts []model.Task) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.Title
	}
	return out
}
