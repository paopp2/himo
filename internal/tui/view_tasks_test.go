package tui

import (
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/paopp2/himo/internal/model"
	"github.com/paopp2/himo/internal/store"
)

func testStyles(t *testing.T) *Styles {
	t.Helper()
	// Deterministic: Ascii profile -> no ANSI sequences, just raw text.
	r := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	r.SetColorProfile(termenv.Ascii)
	return NewStylesWithRenderer(r, StyleOptions{})
}

func testStylesWithColor(t *testing.T) *Styles {
	t.Helper()
	r := lipgloss.NewRenderer(io.Discard, termenv.WithProfile(termenv.TrueColor))
	r.SetColorProfile(termenv.TrueColor)
	return NewStylesWithRenderer(r, StyleOptions{})
}

func TestRenderTaskLine_defaultRow(t *testing.T) {
	st := testStyles(t)
	line := renderTaskLine(st, model.Task{
		Status: model.StatusPending, Title: "Buy groceries",
	}, taskLineInput{Width: 40})

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
		taskLineInput{Width: 40, Cursor: true})
	if !strings.HasPrefix(line, "▌") {
		t.Errorf("cursor row must start with cursor bar, got: %q", line)
	}
}

func TestRenderTaskLine_notesDot(t *testing.T) {
	st := testStyles(t)
	line := renderTaskLine(st, model.Task{
		Status: model.StatusPending, Title: "X", Notes: "    hi",
	}, taskLineInput{Width: 40})
	if !strings.Contains(line, "•") {
		t.Errorf("notes dot missing: %q", line)
	}
}

func TestRenderTaskLine_notesDotAccentWhenURL(t *testing.T) {
	st := testStyles(t)

	// Task with notes but no URL: dot should still be present.
	noURL := renderTaskLine(st, model.Task{
		Status: model.StatusPending, Title: "X", Notes: "    just notes",
	}, taskLineInput{Width: 40})
	if !strings.Contains(noURL, "\u2022") {
		t.Errorf("notes-only task missing dot: %q", noURL)
	}

	// Task with a URL in notes: dot should still be present.
	withURL := renderTaskLine(st, model.Task{
		Status: model.StatusPending, Title: "X", Notes: "    https://example.com",
	}, taskLineInput{Width: 40})
	if !strings.Contains(withURL, "\u2022") {
		t.Errorf("URL task missing dot: %q", withURL)
	}
}

func TestRenderTaskLine_allProjectsChip(t *testing.T) {
	st := testStyles(t)
	line := renderTaskLine(st, model.Task{Status: model.StatusActive, Title: "X"},
		taskLineInput{Width: 60, AllProjects: true, ProjectName: "work"})
	if !strings.Contains(line, "[work]") {
		t.Errorf("project chip missing: %q", line)
	}
}

func TestRenderTaskLine_longTitleTruncatedWithEllipsis(t *testing.T) {
	st := testStyles(t)
	long := "This is a very long task title that absolutely will not fit"
	line := renderTaskLine(st, model.Task{
		Status: model.StatusPending, Title: long, Notes: "    hi",
	}, taskLineInput{Width: 30})

	if strings.Contains(line, "\n") {
		t.Errorf("row wrapped to multiple lines:\n%q", line)
	}
	if lipgloss.Width(line) > 30 {
		t.Errorf("row width %d exceeds Width=30: %q", lipgloss.Width(line), line)
	}
	if !strings.Contains(line, "…") {
		t.Errorf("expected ellipsis marker: %q", line)
	}
	// Notes dot must remain right-aligned even when title is truncated.
	if !strings.HasSuffix(line, "•") {
		t.Errorf("notes dot not at row end: %q", line)
	}
}

func TestRenderTaskLine_longTitleWithChipTruncates(t *testing.T) {
	st := testStyles(t)
	long := "This is a very long task title that does not fit alongside the chip"
	line := renderTaskLine(st, model.Task{Status: model.StatusActive, Title: long},
		taskLineInput{Width: 40, AllProjects: true, ProjectName: "work"})

	if strings.Contains(line, "\n") {
		t.Errorf("row wrapped to multiple lines:\n%q", line)
	}
	if lipgloss.Width(line) > 40 {
		t.Errorf("row width %d exceeds Width=40: %q", lipgloss.Width(line), line)
	}
	if !strings.Contains(line, "[work]") {
		t.Errorf("chip dropped during truncation: %q", line)
	}
	if !strings.Contains(line, "…") {
		t.Errorf("expected ellipsis marker: %q", line)
	}
}

// Indicators on truncated and non-truncated rows must land in the same
// column, otherwise the right-edge "notes/link" dots zigzag visually.
func TestRenderTaskLine_indicatorAlignmentMatchesShortRows(t *testing.T) {
	st := testStyles(t)
	short := renderTaskLine(st, model.Task{
		Status: model.StatusPending, Title: "Short", Notes: "    hi",
	}, taskLineInput{Width: 50})
	long := renderTaskLine(st, model.Task{
		Status: model.StatusPending,
		Title:  "A title long enough that it must be truncated to fit",
		Notes:  "    hi",
	}, taskLineInput{Width: 50})

	if lipgloss.Width(short) != lipgloss.Width(long) {
		t.Fatalf("row widths diverge: short=%d long=%d\nshort=%q\nlong=%q",
			lipgloss.Width(short), lipgloss.Width(long), short, long)
	}
}

// Regression: nesting a Muted-rendered chip inside a TitleStyle(Done) render
// triggers lipgloss's char-by-char Strikethrough path, which wraps each inner
// ESC byte individually and corrupts the stream. The signature is a raw
// ESC-ESC pair in the output; the visible symptom is literal "[38;2;...m"
// text with strikethrough drawn through it.
func TestRenderTaskLine_doneAllProjects_noMangledEscape(t *testing.T) {
	r := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.TrueColor))
	r.SetColorProfile(termenv.TrueColor)
	st := NewStylesWithRenderer(r, StyleOptions{})
	line := renderTaskLine(st, model.Task{Status: model.StatusDone, Title: "Do the thing"},
		taskLineInput{Width: 60, AllProjects: true, ProjectName: "work"})

	if strings.Contains(line, "\x1b\x1b") {
		t.Errorf("output contains adjacent ESC bytes (inner CSI got char-wrapped): %q", line)
	}
	if strings.Contains(line, "[38;2;") {
		// After a correct render, every "[38;2;" substring must be preceded by an
		// ESC. A bare "[38;2;" at display time is the user-visible bug.
		idx := 0
		for {
			j := strings.Index(line[idx:], "[38;2;")
			if j < 0 {
				break
			}
			pos := idx + j
			if pos == 0 || line[pos-1] != 0x1b {
				t.Errorf("bare CSI payload (not preceded by ESC) at %d: %q", pos, line)
				break
			}
			idx = pos + 1
		}
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

func TestRenderTaskLine_editingShowsBuffer(t *testing.T) {
	st := testStyles(t)
	line := renderTaskLine(st, model.Task{
		Status: model.StatusPending, Title: "Original",
	}, taskLineInput{Width: 60, Cursor: true, Editing: true, EditView: "Buffered█"})
	if !strings.Contains(line, "Buffered") {
		t.Errorf("editing row missing buffer text: %q", line)
	}
	if strings.Contains(line, "Original") {
		t.Errorf("editing row should hide original title; got: %q", line)
	}
	// EditView is passed verbatim; the caret is part of the rendered view.
	if !strings.Contains(line, "█") {
		t.Errorf("editing row missing caret: %q", line)
	}
}

func TestRenderTaskLine_editingDoneTaskNoStrikethrough(t *testing.T) {
	st := testStyles(t)
	line := renderTaskLine(st, model.Task{
		Status: model.StatusDone, Title: "Done task",
	}, taskLineInput{Width: 60, Cursor: true, Editing: true, EditView: "Live"})
	if !strings.Contains(line, "Live") {
		t.Errorf("editing row missing buffer text on done task: %q", line)
	}
}

func TestRenderTaskLine_highlightsTitleSubstring(t *testing.T) {
	st := testStylesWithColor(t)
	task := model.Task{Status: model.StatusPending, Title: "Buy groceries"}
	row := renderTaskLine(st, task, taskLineInput{
		Width:       60,
		SearchQuery: "groc",
	})
	plain := stripANSI(row)
	if !strings.Contains(plain, "Buy groceries") {
		t.Fatalf("rendered row missing title: %q", plain)
	}
	hlOpen := st.SearchHighlight.Render("X")
	hlOpen = hlOpen[:strings.Index(hlOpen, "X")]
	if hlOpen == "" {
		t.Fatalf("expected SearchHighlight to produce ANSI; got empty open-code")
	}
	if !strings.Contains(row, hlOpen+"groc") {
		t.Errorf("expected highlight open-code immediately before 'groc'\nrow: %q", row)
	}
}

func TestRenderTaskLine_noQueryNoHighlight(t *testing.T) {
	st := testStylesWithColor(t)
	task := model.Task{Status: model.StatusPending, Title: "Buy groceries"}
	row := renderTaskLine(st, task, taskLineInput{Width: 60})
	hlOpen := st.SearchHighlight.Render("X")
	hlOpen = hlOpen[:strings.Index(hlOpen, "X")]
	if hlOpen == "" {
		t.Fatalf("expected SearchHighlight to produce ANSI; got empty open-code")
	}
	if strings.Contains(row, hlOpen) {
		t.Errorf("unexpected SearchHighlight code in row: %q", row)
	}
}
