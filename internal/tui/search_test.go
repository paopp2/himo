package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/paopp2/himo/internal/model"
)

func TestSearch_escClearsActiveSearch(t *testing.T) {
	m := NewModel(testProject(t))
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "groc" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cur.(Model).searchActive == "" {
		t.Fatalf("searchActive not set after search commit")
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cur.(Model).searchActive != "" {
		t.Errorf("searchActive = %q after Esc, want empty", cur.(Model).searchActive)
	}
}

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

func TestSearch_ctrlWDeletesLastWord(t *testing.T) {
	m := NewModel(testProject(t))
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "two words" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	got := cur.(Model).searchInput.Value()
	if got != "two " {
		t.Errorf("after Ctrl+W, searchInput.Value() = %q, want %q", got, "two ")
	}
}
