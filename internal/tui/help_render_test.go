package tui

import (
	"strings"
	"testing"
)

func TestRenderHelp_showsCtrlO(t *testing.T) {
	st := testStyles(t)
	out := renderHelp(st, 120)
	if !strings.Contains(out, "Ctrl+o") {
		t.Errorf("help missing Ctrl+o:\n%s", out)
	}
	if !strings.Contains(out, "open URL") {
		t.Errorf("help missing 'open URL' label:\n%s", out)
	}
}

func TestRenderHelp_threeColumns(t *testing.T) {
	st := testStyles(t)
	out := renderHelp(st, 120)
	for _, want := range []string{
		"Navigation", "Filters", "Actions",
		"j", "Enter", "Space",
		"0", "1", "2", "3", "4", "5", "6",
		"o", "O", "d", "!", "x", "-",
		"u", "Ctrl+R",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q in:\n%s", want, out)
		}
	}
}

func TestRenderHelp_showsInlineEdit(t *testing.T) {
	st := testStyles(t)
	out := renderHelp(st, 120)
	if !strings.Contains(out, "edit title inline") {
		t.Errorf("help missing 'edit title inline' label:\n%s", out)
	}
	if strings.Contains(out, "edit current file") {
		t.Errorf("help still mentions removed 'edit current file' action:\n%s", out)
	}
}

func TestRenderHelp_showsSortToggle(t *testing.T) {
	st := testStyles(t)
	out := renderHelp(st, 120)
	if !strings.Contains(out, "toggle sort") {
		t.Errorf("help missing 'toggle sort' label:\n%s", out)
	}
}
