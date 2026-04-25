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

func TestRenderTopBar_sortIndicator(t *testing.T) {
	st := testStyles(t)
	natural := renderTopBar(st, topBarInput{
		Projects: []string{"work"},
		Current:  "work",
		Width:    100,
		Sort:     SortNatural,
	})
	if !strings.Contains(natural, "sort:natural") {
		t.Errorf("natural mode topbar missing 'sort:natural':\n%s", natural)
	}
	status := renderTopBar(st, topBarInput{
		Projects: []string{"work"},
		Current:  "work",
		Width:    100,
		Sort:     SortStatus,
	})
	if !strings.Contains(status, "sort:status") {
		t.Errorf("status mode topbar missing 'sort:status':\n%s", status)
	}
	if !strings.Contains(status, "[s]") {
		t.Errorf("status mode topbar missing [s] hint:\n%s", status)
	}
}
