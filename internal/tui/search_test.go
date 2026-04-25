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

func TestSearch_doesNotFilterList(t *testing.T) {
	m := NewModel(testProject(t))
	var cur tea.Model = m
	before := len(cur.(Model).visibleTasks())
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "groc" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	after := len(cur.(Model).visibleTasks())
	if after != before {
		t.Errorf("after committing search, visible task count = %d, want unchanged %d", after, before)
	}
	if cur.(Model).searchActive != "groc" {
		t.Errorf("searchActive = %q, want %q", cur.(Model).searchActive, "groc")
	}
}

func titles(ts []model.Task) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.Title
	}
	return out
}

func TestSearch_escDuringTypingRestoresCursor(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 1 // testProject has 2 tasks; start cursor on the second.
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "groc" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if got := cur.(Model).cursor; got != 1 {
		t.Errorf("after Esc, cursor = %d, want 1 (preSearchCursor)", got)
	}
	if got := cur.(Model).searchActive; got != "" {
		t.Errorf("after Esc, searchActive = %q, want empty", got)
	}
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
