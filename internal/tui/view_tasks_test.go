package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
)

func testStyles(t *testing.T) *Styles {
	t.Helper()
	// Deterministic: Ascii profile -> no ANSI sequences, just raw text.
	r := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	r.SetColorProfile(termenv.Ascii)
	return NewStylesWithRenderer(r, StyleOptions{})
}

func TestRenderTaskLine_defaultRow(t *testing.T) {
	st := testStyles(t)
	line := renderTaskLine(st, model.Task{
		Status: model.StatusPending, Title: "Buy groceries",
	}, renderTaskOpts{Width: 40})

	if !strings.Contains(line, "○") {
		t.Errorf("line missing pending glyph: %q", line)
	}
	if !strings.Contains(line, "Buy groceries") {
		t.Errorf("line missing title: %q", line)
	}
}

func TestRenderTaskLine_cursorRowHasBar(t *testing.T) {
	st := testStyles(t)
	line := renderTaskLine(st, model.Task{Status: model.StatusActive, Title: "X"},
		renderTaskOpts{Width: 40, Cursor: true})
	if !strings.HasPrefix(line, "▌") {
		t.Errorf("cursor row must start with cursor bar, got: %q", line)
	}
}

func TestRenderTaskLine_notesDot(t *testing.T) {
	st := testStyles(t)
	line := renderTaskLine(st, model.Task{
		Status: model.StatusPending, Title: "X", Notes: "    hi",
	}, renderTaskOpts{Width: 40})
	if !strings.Contains(line, "•") {
		t.Errorf("notes dot missing: %q", line)
	}
}

func TestRenderTaskLine_allProjectsChip(t *testing.T) {
	st := testStyles(t)
	line := renderTaskLine(st, model.Task{Status: model.StatusActive, Title: "X"},
		renderTaskOpts{Width: 60, AllProjects: true, ProjectName: "work"})
	if !strings.Contains(line, "[work]") {
		t.Errorf("project chip missing: %q", line)
	}
}

func TestView_narrowLayoutCollapses(t *testing.T) {
	m := NewModel(testProject(t))
	m.width, m.height = 60, 20
	m.styles = testStyles(t)
	out := renderView(m)

	// Filter bar keys (0-6 chips) disappear below 80 cols.
	if strings.Contains(out, "[0] All") && strings.Contains(out, "[6] Cancelled") {
		t.Errorf("filter bar still fully rendered at 60 cols:\n%s", out)
	}
	// Top bar collapses: scope shortcuts should be absent.
	if strings.Contains(out, "[A] all") || strings.Contains(out, "[P] picker") {
		t.Errorf("top bar still showing scope shortcuts at 60 cols:\n%s", out)
	}
}

func TestRenderList_usesStyledRows(t *testing.T) {
	m := NewModel(testProject(t))
	m.width = 120
	m.height = 30
	m.styles = testStyles(t)

	locs := m.visibleTaskLocations()
	tasks := make([]model.Task, len(locs))
	for i, loc := range locs {
		tasks[i] = loc.doc.Items[loc.idx].(store.TaskItem).Task
	}

	out := renderList(m, locs, tasks)
	if !strings.Contains(out, "○") {
		t.Errorf("list should use Unicode glyph, got:\n%s", out)
	}
	if strings.Contains(out, "[ ]") || strings.Contains(out, "[/]") {
		t.Errorf("list still has raw markers:\n%s", out)
	}
}
