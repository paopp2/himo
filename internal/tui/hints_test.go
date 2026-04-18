package tui

import (
	"strings"
	"testing"
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
