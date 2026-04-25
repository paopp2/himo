package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/paopp2/himo/internal/model"
	"github.com/paopp2/himo/internal/store"
)

// hlStyles returns base/highlight styles bound to a TrueColor renderer so
// lipgloss emits ANSI escapes deterministically in tests (the default
// renderer drops ANSI when stdout is not a TTY).
func hlStyles() (lipgloss.Style, lipgloss.Style) {
	r := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.TrueColor))
	r.SetColorProfile(termenv.TrueColor)
	return r.NewStyle(), r.NewStyle().Reverse(true)
}

func TestHighlightMatch_emptyNeedle(t *testing.T) {
	base, hl := hlStyles()
	got := ansi.Strip(highlightMatch("Buy groceries", "", base, hl))
	if got != "Buy groceries" {
		t.Errorf("empty needle: got %q, want %q", got, "Buy groceries")
	}
}

func TestHighlightMatch_singleMatch(t *testing.T) {
	base, hl := hlStyles()
	got := highlightMatch("Buy groceries", "groc", base, hl)
	if ansi.Strip(got) != "Buy groceries" {
		t.Errorf("plain text mismatch: got %q", ansi.Strip(got))
	}
	if !strings.Contains(got, "groc") {
		t.Errorf("expected 'groc' in output, got %q", got)
	}
	hlOpen := hl.Render("X")
	hlOpen = hlOpen[:strings.Index(hlOpen, "X")]
	if !strings.Contains(got, hlOpen+"groc") {
		t.Errorf("expected highlight open-code immediately before 'groc'\noutput: %q", got)
	}
}

func TestHighlightMatch_multipleMatches(t *testing.T) {
	base, hl := hlStyles()
	got := highlightMatch("foo bar foo", "foo", base, hl)
	if ansi.Strip(got) != "foo bar foo" {
		t.Errorf("plain text mismatch: got %q", ansi.Strip(got))
	}
	hlOpen := hl.Render("X")
	hlOpen = hlOpen[:strings.Index(hlOpen, "X")]
	if strings.Count(got, hlOpen+"foo") != 2 {
		t.Errorf("expected 2 highlighted occurrences, got %q", got)
	}
}

func TestHighlightMatch_caseInsensitive(t *testing.T) {
	base, hl := hlStyles()
	got := highlightMatch("Buy Groceries", "groc", base, hl)
	if ansi.Strip(got) != "Buy Groceries" {
		t.Errorf("plain text mismatch: got %q", ansi.Strip(got))
	}
	hlOpen := hl.Render("X")
	hlOpen = hlOpen[:strings.Index(hlOpen, "X")]
	if !strings.Contains(got, hlOpen+"Groc") {
		t.Errorf("expected highlight to preserve original case 'Groc', got %q", got)
	}
}

func TestHighlightMatch_noMatch(t *testing.T) {
	base, hl := hlStyles()
	got := highlightMatch("Buy groceries", "xyz", base, hl)
	if ansi.Strip(got) != "Buy groceries" {
		t.Errorf("plain text mismatch: got %q", ansi.Strip(got))
	}
	hlOpen := hl.Render("X")
	hlOpen = hlOpen[:strings.Index(hlOpen, "X")]
	if strings.Contains(got, hlOpen) {
		t.Errorf("expected no highlight open-code, got %q", got)
	}
}

func TestHighlightMatch_multibyte(t *testing.T) {
	base, hl := hlStyles()
	got := highlightMatch("café noir", "fé", base, hl)
	if ansi.Strip(got) != "café noir" {
		t.Errorf("plain text mismatch: got %q", ansi.Strip(got))
	}
}

func locFromTitle(p *store.Project, title string) taskLoc {
	doc := &store.Document{}
	t := model.Task{Status: model.StatusPending, Title: title}
	doc.Items = append(doc.Items, store.TaskItem{
		Task:     t,
		RawLines: []string{store.RenderTaskLine(t)},
	})
	return taskLoc{proj: p, doc: doc, idx: len(doc.Items) - 1}
}

func TestMatchIndices_emptyNeedleReturnsNil(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{locFromTitle(p, "Buy groceries"), locFromTitle(p, "Write design")}
	if got := matchIndices(locs, "", false); got != nil {
		t.Errorf("empty needle: got %v, want nil", got)
	}
}

func TestMatchIndices_titleSubstring(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{
		locFromTitle(p, "Buy groceries"),
		locFromTitle(p, "Write design"),
		locFromTitle(p, "Design review"),
	}
	got := matchIndices(locs, "design", false)
	want := []int{1, 2}
	if !slices.Equal(got, want) {
		t.Errorf("matchIndices: got %v, want %v", got, want)
	}
}

func TestMatchIndices_caseInsensitive(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{locFromTitle(p, "Buy GROCERIES")}
	got := matchIndices(locs, "groc", false)
	if !slices.Equal(got, []int{0}) {
		t.Errorf("case-insensitive: got %v, want [0]", got)
	}
}

func TestMatchIndices_projectNameInAllProjectsMode(t *testing.T) {
	work := &store.Project{Name: "work"}
	personal := &store.Project{Name: "personal"}
	locs := []taskLoc{
		locFromTitle(work, "Buy groceries"),
		locFromTitle(personal, "Read book"),
	}
	got := matchIndices(locs, "person", true)
	if !slices.Equal(got, []int{1}) {
		t.Errorf("all-projects project match: got %v, want [1]", got)
	}
}

func TestMatchIndices_projectNameIgnoredInSingleProjectMode(t *testing.T) {
	work := &store.Project{Name: "work"}
	locs := []taskLoc{locFromTitle(work, "Buy groceries")}
	if got := matchIndices(locs, "work", false); got != nil {
		t.Errorf("single-project mode should ignore project name: got %v, want nil", got)
	}
}

func TestNextMatch_forwardFindsAtOrAfter(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{
		locFromTitle(p, "Buy groceries"),
		locFromTitle(p, "Write design"),
		locFromTitle(p, "Design review"),
	}
	idx, wrapped, ok := nextMatch(locs, "design", false, 1, true)
	if !ok || idx != 1 || wrapped {
		t.Errorf("forward from 1: idx=%d wrapped=%v ok=%v, want 1 false true", idx, wrapped, ok)
	}
}

func TestNextMatch_forwardWrapsToEarlierMatch(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{
		locFromTitle(p, "Design A"),
		locFromTitle(p, "Buy groceries"),
		locFromTitle(p, "Read book"),
	}
	idx, wrapped, ok := nextMatch(locs, "design", false, 2, true)
	if !ok || idx != 0 || !wrapped {
		t.Errorf("forward wrap from 2: idx=%d wrapped=%v ok=%v, want 0 true true", idx, wrapped, ok)
	}
}

func TestNextMatch_backwardFindsAtOrBefore(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{
		locFromTitle(p, "Design A"),
		locFromTitle(p, "Buy groceries"),
		locFromTitle(p, "Design B"),
	}
	idx, wrapped, ok := nextMatch(locs, "design", false, 2, false)
	if !ok || idx != 2 || wrapped {
		t.Errorf("backward from 2: idx=%d wrapped=%v ok=%v, want 2 false true", idx, wrapped, ok)
	}
}

func TestNextMatch_backwardWrapsToLaterMatch(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{
		locFromTitle(p, "Buy groceries"),
		locFromTitle(p, "Read book"),
		locFromTitle(p, "Design A"),
	}
	idx, wrapped, ok := nextMatch(locs, "design", false, 0, false)
	if !ok || idx != 2 || !wrapped {
		t.Errorf("backward wrap from 0: idx=%d wrapped=%v ok=%v, want 2 true true", idx, wrapped, ok)
	}
}

func TestNextMatch_noMatches(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{locFromTitle(p, "Buy groceries"), locFromTitle(p, "Read book")}
	idx, wrapped, ok := nextMatch(locs, "xyz", false, 0, true)
	if ok || idx != 0 || wrapped {
		t.Errorf("no matches: idx=%d wrapped=%v ok=%v, want 0 false false", idx, wrapped, ok)
	}
}

func TestNextMatch_emptyList(t *testing.T) {
	idx, wrapped, ok := nextMatch(nil, "design", false, 0, true)
	if ok || idx != 0 || wrapped {
		t.Errorf("empty: idx=%d wrapped=%v ok=%v, want 0 false false", idx, wrapped, ok)
	}
}

func TestNextMatch_singleMatchSelfFromCursor(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{
		locFromTitle(p, "Buy groceries"),
		locFromTitle(p, "Design"),
		locFromTitle(p, "Read book"),
	}
	idx, wrapped, ok := nextMatch(locs, "design", false, 1, true)
	if !ok || idx != 1 || wrapped {
		t.Errorf("self-match: idx=%d wrapped=%v ok=%v, want 1 false true", idx, wrapped, ok)
	}
}

func TestNextMatch_fromBeyondEndForwardWraps(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{
		locFromTitle(p, "Design A"),
		locFromTitle(p, "Buy groceries"),
	}
	idx, wrapped, ok := nextMatch(locs, "design", false, 2, true)
	if !ok || idx != 0 || !wrapped {
		t.Errorf("from beyond end: idx=%d wrapped=%v ok=%v, want 0 true true", idx, wrapped, ok)
	}
}

func TestNextMatch_fromNegativeBackwardWraps(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{
		locFromTitle(p, "Buy groceries"),
		locFromTitle(p, "Design A"),
	}
	idx, wrapped, ok := nextMatch(locs, "design", false, -1, false)
	if !ok || idx != 1 || !wrapped {
		t.Errorf("from negative: idx=%d wrapped=%v ok=%v, want 1 true true", idx, wrapped, ok)
	}
}
