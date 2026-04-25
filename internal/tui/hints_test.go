package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestHintBar_normalMode(t *testing.T) {
	st := testStyles(t)
	out := renderHintBar(st, hintInput{Mode: ModeNormal, Width: 100})
	if !strings.Contains(out, "NORMAL") {
		t.Errorf("missing mode pill:\n%s", out)
	}
	for _, k := range []string{"j/k", "Enter", "Space", "o", "/", "?"} {
		if !strings.Contains(out, k) {
			t.Errorf("missing hint %q:\n%s", k, out)
		}
	}
}

func TestHintBar_searchMode(t *testing.T) {
	st := testStyles(t)
	out := renderHintBar(st, hintInput{
		Mode: ModeSearch, Width: 100, SearchBuf: "design",
	})
	if !strings.Contains(out, "SEARCH") {
		t.Errorf("missing SEARCH pill:\n%s", out)
	}
	if !strings.Contains(out, "design") {
		t.Errorf("search buf missing:\n%s", out)
	}
}

func TestHintBar_insertPillForPrompt(t *testing.T) {
	st := testStyles(t)
	out := renderHintBar(st, hintInput{Mode: ModePrompt, Width: 100})
	if !strings.Contains(out, "INSERT") {
		t.Errorf("expected INSERT pill for prompt mode:\n%s", out)
	}
	if strings.Contains(out, "PROMPT") {
		t.Errorf("PROMPT pill should be replaced by INSERT:\n%s", out)
	}
	for _, k := range []string{"Enter", "apply", "Esc", "cancel", "C-w", "del word"} {
		if !strings.Contains(out, k) {
			t.Errorf("missing insert hint %q:\n%s", k, out)
		}
	}
	if strings.Contains(out, "> ") {
		t.Errorf("hint bar should not echo input:\n%s", out)
	}
}

func TestHintBar_insertPillForEdit(t *testing.T) {
	st := testStyles(t)
	out := renderHintBar(st, hintInput{Mode: ModeEdit, Width: 100})
	if !strings.Contains(out, "INSERT") {
		t.Errorf("expected INSERT pill for edit mode:\n%s", out)
	}
	if strings.Contains(out, " EDIT ") {
		t.Errorf("EDIT pill should be replaced by INSERT:\n%s", out)
	}
	for _, k := range []string{"Enter", "apply", "Esc", "cancel", "C-w", "del word"} {
		if !strings.Contains(out, k) {
			t.Errorf("missing insert hint %q:\n%s", k, out)
		}
	}
}

func TestMetaHints_searchMatchPosition(t *testing.T) {
	st := NewStyles(StyleOptions{NoColor: true})
	got := ansi.Strip(metaHints(st, hintInput{
		Mode:           ModeNormal,
		SearchActive:   "groc",
		SearchMatchPos: 1,
		SearchTotal:    3,
	}))
	if !strings.Contains(got, "match 1 / 3") {
		t.Errorf("meta hints missing 'match 1 / 3': %q", got)
	}
}

func TestMetaHints_searchTotalOnlyWhenOffMatch(t *testing.T) {
	st := NewStyles(StyleOptions{NoColor: true})
	got := ansi.Strip(metaHints(st, hintInput{
		Mode:           ModeNormal,
		SearchActive:   "groc",
		SearchMatchPos: 0,
		SearchTotal:    3,
	}))
	if !strings.Contains(got, "3 matches") {
		t.Errorf("meta hints missing '3 matches': %q", got)
	}
}

func TestMetaHints_searchNoMatches(t *testing.T) {
	st := NewStyles(StyleOptions{NoColor: true})
	got := ansi.Strip(metaHints(st, hintInput{
		Mode:           ModeNormal,
		SearchActive:   "xyz",
		SearchMatchPos: 0,
		SearchTotal:    0,
	}))
	if !strings.Contains(got, "no matches") {
		t.Errorf("meta hints missing 'no matches': %q", got)
	}
}

func TestMetaHints_noIndicatorWhenSearchInactive(t *testing.T) {
	st := NewStyles(StyleOptions{NoColor: true})
	got := ansi.Strip(metaHints(st, hintInput{
		Mode: ModeNormal,
	}))
	if strings.Contains(got, "match") || strings.Contains(got, "matches") {
		t.Errorf("inactive search showed match indicator: %q", got)
	}
}

func TestHintBar_banner(t *testing.T) {
	st := testStyles(t)
	out := renderHintBar(st, hintInput{
		Mode:   ModeNormal,
		Width:  120,
		Banner: "editor: vim not found",
	})
	if !strings.Contains(out, "editor: vim not found") {
		t.Errorf("banner not rendered:\n%s", out)
	}
}
