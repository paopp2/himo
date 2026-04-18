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
