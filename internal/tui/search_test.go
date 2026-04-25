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

func TestSearch_incsearchJumpsCursorToFirstMatch(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 0
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "design" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if got := cur.(Model).cursor; got != 1 {
		t.Errorf("incsearch 'design' from cursor 0: got cursor %d, want 1", got)
	}
}

func TestSearch_incsearchNoMatchKeepsCursorAtPreSearch(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 1
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "xyzzy" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if got := cur.(Model).cursor; got != 1 {
		t.Errorf("incsearch with no match: got cursor %d, want 1 (preSearchCursor)", got)
	}
}

func TestSearch_incsearchEmptyBufferRestoresCursor(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 0
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if got := cur.(Model).cursor; got != 1 {
		t.Fatalf("expected cursor at 1 after typing 'd', got %d", got)
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if got := cur.(Model).cursor; got != 0 {
		t.Errorf("after backspace to empty buffer, cursor = %d, want 0 (preSearchCursor)", got)
	}
}

func TestSearch_enterCommitsAtIncsearchCursor(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 0
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "design" {
		cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if got := cur.(Model).cursor; got != 1 {
		t.Errorf("after Enter, cursor = %d, want 1", got)
	}
	if got := cur.(Model).searchActive; got != "design" {
		t.Errorf("searchActive = %q, want %q", got, "design")
	}
}

func TestSearch_nAdvancesToNextMatch(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 0
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if got := cur.(Model).cursor; got != 0 {
		t.Fatalf("after commit, cursor = %d, want 0 (Buy groceries matches 'e')", got)
	}
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if got := cur.(Model).cursor; got != 1 {
		t.Errorf("after n, cursor = %d, want 1 (Write design)", got)
	}
}

func TestSearch_nWrapsAtEndAndSetsBanner(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 1
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if got := cur.(Model).cursor; got != 0 {
		t.Errorf("n past last match: cursor = %d, want 0 (wrapped to Buy groceries)", got)
	}
	if got := cur.(Model).banner; got != "search hit BOTTOM, continuing at TOP" {
		t.Errorf("wrap banner = %q, want %q", got, "search hit BOTTOM, continuing at TOP")
	}
}

func TestSearch_NRetreatsToPreviousMatch(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 1
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	if got := cur.(Model).cursor; got != 0 {
		t.Errorf("after N, cursor = %d, want 0", got)
	}
}

func TestSearch_NWrapsAtStartAndSetsBanner(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 0
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	if got := cur.(Model).cursor; got != 1 {
		t.Errorf("N past first match: cursor = %d, want 1 (wrapped)", got)
	}
	if got := cur.(Model).banner; got != "search hit TOP, continuing at BOTTOM" {
		t.Errorf("wrap banner = %q, want %q", got, "search hit TOP, continuing at BOTTOM")
	}
}

func TestSearch_nWithNoMatchesSetsBanner(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 0
	m.searchActive = "xyzzy"
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if got := cur.(Model).cursor; got != 0 {
		t.Errorf("n with no matches moved cursor to %d, want 0", got)
	}
	if got := cur.(Model).banner; got != "no matches" {
		t.Errorf("no-match banner = %q, want %q", got, "no matches")
	}
}

func TestSearch_nWithNoActiveSearchIsNoop(t *testing.T) {
	m := NewModel(testProject(t))
	m.cursor = 0
	var cur tea.Model = m
	cur, _ = cur.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if got := cur.(Model).cursor; got != 0 {
		t.Errorf("n with no searchActive should not move cursor, got %d", got)
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
