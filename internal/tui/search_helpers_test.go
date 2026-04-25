package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
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

// stripANSI returns s with ANSI escape sequences removed. Tests use it to
// assert on rendered text without coupling to terminal styling.
func stripANSI(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		if r == 0x1b {
			in = true
			continue
		}
		if in {
			if r == 'm' {
				in = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func TestHighlightMatch_emptyNeedle(t *testing.T) {
	base, hl := hlStyles()
	got := stripANSI(highlightMatch("Buy groceries", "", base, hl))
	if got != "Buy groceries" {
		t.Errorf("empty needle: got %q, want %q", got, "Buy groceries")
	}
}

func TestHighlightMatch_singleMatch(t *testing.T) {
	base, hl := hlStyles()
	got := highlightMatch("Buy groceries", "groc", base, hl)
	if stripANSI(got) != "Buy groceries" {
		t.Errorf("plain text mismatch: got %q", stripANSI(got))
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
	if stripANSI(got) != "foo bar foo" {
		t.Errorf("plain text mismatch: got %q", stripANSI(got))
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
	if stripANSI(got) != "Buy Groceries" {
		t.Errorf("plain text mismatch: got %q", stripANSI(got))
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
	if stripANSI(got) != "Buy groceries" {
		t.Errorf("plain text mismatch: got %q", stripANSI(got))
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
	if stripANSI(got) != "café noir" {
		t.Errorf("plain text mismatch: got %q", stripANSI(got))
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
	if !equalInts(got, want) {
		t.Errorf("matchIndices: got %v, want %v", got, want)
	}
}

func TestMatchIndices_caseInsensitive(t *testing.T) {
	p := &store.Project{Name: "work"}
	locs := []taskLoc{locFromTitle(p, "Buy GROCERIES")}
	got := matchIndices(locs, "groc", false)
	if !equalInts(got, []int{0}) {
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
	if !equalInts(got, []int{1}) {
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

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
