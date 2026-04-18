package tui

import (
	"strings"
	"testing"
)

func TestRenderTopBar_projectsAndActive(t *testing.T) {
	st := testStyles(t)
	out := renderTopBar(st, topBarInput{
		Projects: []string{"work", "personal", "side"},
		Current:  "personal",
		Width:    100,
	})
	for _, want := range []string{"work", "personal", "side", "[A]", "[P]"} {
		if !strings.Contains(out, want) {
			t.Errorf("topbar missing %q:\n%s", want, out)
		}
	}
}

func TestRenderTopBar_allProjectsMode(t *testing.T) {
	st := testStyles(t)
	out := renderTopBar(st, topBarInput{
		Projects: []string{"work", "personal"},
		Current:  "",
		Width:    100,
		AllMode:  true,
	})
	if !strings.Contains(out, "all projects") {
		t.Errorf("all-mode top bar missing label:\n%s", out)
	}
}
