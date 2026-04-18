package tui

import (
	"strings"
	"testing"
)

func TestRenderHelp_threeColumns(t *testing.T) {
	st := testStyles(t)
	out := renderHelp(st, 120)
	for _, want := range []string{
		"Navigation", "Filters", "Actions",
		"j", "Enter", "Space",
		"0", "1", "2", "3", "4", "5", "6",
		"o", "O", "d", "!", "x", "-",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q in:\n%s", want, out)
		}
	}
}
