package tui

import (
	"strings"
	"testing"
)

func TestGlamourRender_basicMarkdown(t *testing.T) {
	r, err := newNotesRenderer(60)
	if err != nil {
		t.Fatalf("newNotesRenderer: %v", err)
	}
	out, err := r.Render("Due Friday.\n\n- check storage\n- check parser\n")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(out, "Due Friday.") {
		t.Errorf("body missing: %q", out)
	}
}

func TestStripNotesIndent_preservesRelativeIndent(t *testing.T) {
	// Notes with mixed 2-space and 6-space indent (a nested list). The
	// minimum common indent is 2; stripping it must leave the 4-space gap
	// so Glamour sees a nested list, not a flat one.
	notes := "  - parent\n      - child A\n      - child B"
	got := stripNotesIndent(notes)
	want := "- parent  \n    - child A  \n    - child B  "
	if got != want {
		t.Errorf("stripNotesIndent mismatch:\ngot:\n%q\nwant:\n%q", got, want)
	}
}
