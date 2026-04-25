package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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
