package tui

import (
	"strings"
	"testing"

	"github.com/paopp2/himo/internal/model"
)

func TestRenderPreview_noSelection(t *testing.T) {
	st := testStyles(t)
	out := renderPreview(previewInput{Styles: st, Width: 40, Height: 10})
	if !strings.Contains(out, "No tasks match") {
		t.Errorf("no-selection state missing expected text:\n%s", out)
	}
}

func TestRenderPreview_noNotes(t *testing.T) {
	st := testStyles(t)
	out := renderPreview(previewInput{
		Styles: st, Width: 50, Height: 10,
		Task: &model.Task{Status: model.StatusPending, Title: "Buy milk"},
	})
	if !strings.Contains(out, "No notes yet") {
		t.Errorf("no-notes state missing text:\n%s", out)
	}
	if !strings.Contains(out, "Buy milk") {
		t.Errorf("title missing from header:\n%s", out)
	}
}

func TestRenderPreview_hasNotes(t *testing.T) {
	st := testStyles(t)
	out := renderPreview(previewInput{
		Styles: st, Width: 60, Height: 10,
		Task: &model.Task{
			Status: model.StatusActive,
			Title:  "Write doc",
			Notes:  "    Due Friday.\n    \n    - check storage",
		},
	})
	if !strings.Contains(out, "Write doc") {
		t.Errorf("title missing:\n%s", out)
	}
	if !strings.Contains(out, "Due Friday") {
		t.Errorf("notes body missing:\n%s", out)
	}
}
