package tui

import (
	"strings"
	"testing"
)

func TestCenteredBox_containsTitleAndBody(t *testing.T) {
	st := testStyles(t)
	out := centeredBox(st, modalInput{
		Title: "Delete task?",
		Body:  "Write design doc",
		Hints: "y delete   n cancel",
		Width: 80, Height: 20,
	})
	for _, want := range []string{"Delete task?", "Write design doc", "y delete"} {
		if !strings.Contains(out, want) {
			t.Errorf("centered box missing %q:\n%s", want, out)
		}
	}
}
