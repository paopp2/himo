package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
)

func renderHelp(st *Styles, width int) string {
	col := func(heading string, rows [][2]string) string {
		var b strings.Builder
		b.WriteString(st.Accent.Render(heading))
		b.WriteString("\n")
		for _, r := range rows {
			b.WriteString(st.Base.Render(fmt.Sprintf("  %-10s  ", r[0])))
			b.WriteString(st.Muted.Render(r[1]))
			b.WriteString("\n")
		}
		return b.String()
	}

	nav := col("Navigation", [][2]string{
		{"j/k", "down / up"},
		{"g/G", "top / bottom"},
		{"Ctrl+d/u", "half page"},
		{"/", "search"},
		{"Tab", "next project"},
		{"S-Tab", "prev project"},
		{"P", "project picker"},
		{"A", "all projects"},
		{"q", "quit"},
	})
	filters := col("Filters", [][2]string{
		{"0", "all"},
		{"1", "backlog"},
		{"2", "pending"},
		{"3", "active"},
		{"4", "blocked"},
		{"5", "done"},
		{"6", "cancelled"},
		{"Esc", "default filter"},
	})
	actions := col("Actions", [][2]string{
		{"Enter", "notes in $EDITOR"},
		{"e", "edit current file"},
		{"Space", "cycle status"},
		{"b/p/a", "backlog / pending / active"},
		{"!/x/-", "blocked / done / cancelled"},
		{"o/O", "new below / above"},
		{"d", "delete"},
		{"v", "toggle preview"},
		{"?", "toggle this help"},
	})

	colW := (width - 4) / 3
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(colW).Render(nav),
		lipgloss.NewStyle().Width(colW).Render(filters),
		lipgloss.NewStyle().Width(colW).Render(actions),
	)
}

func renderView(m Model) string {
	if m.showingHelp {
		width := m.width
		if width <= 0 {
			width = 80
		}
		return renderHelp(m.styles, width) + "\n" +
			renderHintBar(m.styles, hintInput{Mode: ModeHelp, Width: width})
	}
	width := m.width
	if width <= 0 {
		width = 80
	}
	height := m.height
	if height < 10 {
		height = 10
	}

	locs := m.visibleTaskLocations()
	tasks := make([]model.Task, len(locs))
	for i, loc := range locs {
		tasks[i] = loc.doc.Items[loc.idx].(store.TaskItem).Task
	}

	// Modal overlays (prompt / delete / picker) replace the main view.
	switch m.currentMode() {
	case ModePrompt:
		title := "New task"
		if m.promptAbove {
			title = "New task (above)"
		}
		return centeredBox(m.styles, modalInput{
			Title:  title,
			Body:   "> " + m.promptBuf + "_",
			Hints:  "Enter create   Esc cancel",
			Width:  width,
			Height: height,
		})
	case ModeDelete:
		return centeredBox(m.styles, modalInput{
			Title:  "Delete task?",
			Body:   deleteTitle(m, tasks),
			Hints:  "y delete   n cancel",
			Width:  width,
			Height: height,
			Error:  true,
		})
	case ModePicker:
		return centeredBox(m.styles, modalInput{
			Title:  "Switch project",
			Body:   renderPickerBody(m),
			Hints:  "up/down move   Enter switch   Esc cancel",
			Width:  width,
			Height: height,
		})
	}

	top := renderTopBar(m.styles, topBarInput{
		Projects: m.projects,
		Current:  m.project.Name,
		Width:    width,
		AllMode:  m.allProjects,
	})
	fbar := renderFilterBar(m.styles, m.filter, m.statusCounts(), width)

	var body string
	if width < 100 || m.hidePreview {
		body = renderListPane(m, locs, tasks, width, m.height-4, true)
	} else {
		listW := width * 60 / 100
		previewW := width - listW - 3
		var previewTask *model.Task
		if len(tasks) > 0 && m.cursor < len(tasks) {
			previewTask = &tasks[m.cursor]
		}
		listPane := renderListPane(m, locs, tasks, listW, m.height-4, true)
		previewPane := renderPreview(previewInput{
			Styles: m.styles,
			Width:  previewW,
			Height: m.height - 4,
			Task:   previewTask,
		})
		body = lipgloss.JoinHorizontal(lipgloss.Top, listPane, "  ", previewPane)
	}

	hint := renderHintBar(m.styles, hintInput{
		Mode:        m.currentMode(),
		Width:       width,
		SearchBuf:   m.searchBuf,
		PromptBuf:   m.promptBuf,
		PromptAbove: m.promptAbove,
		DeleteTitle: deleteTitle(m, tasks),
		Banner:      m.banner,
	})
	view := top + "\n"
	if fbar != "" {
		view += fbar + "\n\n"
	} else {
		view += "\n"
	}
	view += body + "\n" + hint
	return view
}

func deleteTitle(m Model, tasks []model.Task) string {
	if !m.confirmingDelete || m.cursor >= len(tasks) {
		return ""
	}
	return tasks[m.cursor].Title
}

func renderList(m Model, locs []taskLoc, tasks []model.Task) string {
	var b strings.Builder
	width := m.width
	if width <= 0 {
		width = 80
	}
	for i, t := range tasks {
		opts := renderTaskOpts{Width: width, Cursor: i == m.cursor}
		if m.allProjects && i < len(locs) {
			opts.AllProjects = true
			opts.ProjectName = locs[i].proj.Name
		}
		b.WriteString(renderTaskLine(m.styles, t, opts))
		b.WriteByte('\n')
	}
	return b.String()
}

// renderListPane wraps the visible slice of task rows in a rounded-border
// box. When the list is taller than the pane, rows scroll so the cursor
// row is always visible.
func renderListPane(m Model, locs []taskLoc, tasks []model.Task, width, height int, focused bool) string {
	border := m.styles.PaneBorder
	if focused {
		border = m.styles.PaneBorderFocused
	}
	if height < 5 {
		height = 5
	}

	// Inner content height is pane height minus border top+bottom.
	contentH := height - 2
	if contentH < 1 {
		contentH = 1
	}

	rows := make([]string, len(tasks))
	for i, t := range tasks {
		opts := renderTaskOpts{Width: width - 2, Cursor: i == m.cursor}
		if m.allProjects && i < len(locs) {
			opts.AllProjects = true
			opts.ProjectName = locs[i].proj.Name
		}
		rows[i] = renderTaskLine(m.styles, t, opts)
	}

	// Compute the window so m.cursor is visible.
	start := 0
	if len(rows) > contentH {
		start = m.cursor - contentH/2
		if start < 0 {
			start = 0
		}
		if start > len(rows)-contentH {
			start = len(rows) - contentH
		}
	}
	end := start + contentH
	if end > len(rows) {
		end = len(rows)
	}

	visible := strings.Join(rows[start:end], "\n")
	return border.Width(width).Height(height).Render(visible)
}

type renderTaskOpts struct {
	Width       int
	Cursor      bool
	AllProjects bool
	ProjectName string
}

// renderTaskLine returns a single styled row for t.
// Layout: "▌ ● [work] Buy groceries                 •"
// - leftmost col: cursor bar or space
// - glyph
// - optional [project] chip
// - title (styled per status)
// - right-aligned notes dot or space
func renderTaskLine(st *Styles, t model.Task, o renderTaskOpts) string {
	bar := " "
	if o.Cursor {
		bar = st.CursorBar.Render("▌")
	}
	glyph := st.GlyphStyle(t.Status).Render(st.StatusGlyph(t.Status))

	title := t.Title
	if o.AllProjects && o.ProjectName != "" {
		title = st.Muted.Render("["+o.ProjectName+"] ") + title
	}
	title = st.TitleStyle(t.Status).Render(title)

	dot := " "
	if t.HasNotes() {
		dot = st.Muted.Render("•")
	}

	// Build the row with fixed slots: bar + glyph + title, right-align dot.
	left := bar + " " + glyph + " " + title
	padding := o.Width - lipgloss.Width(left) - 2
	if padding < 1 {
		padding = 1
	}
	row := left + strings.Repeat(" ", padding) + dot

	if o.Cursor {
		row = st.CursorRowBG.Render(row)
	}
	return row
}

func renderPickerBody(m Model) string {
	var b strings.Builder
	b.WriteString("/ " + m.pickerFilter + "_\n")
	b.WriteString(strings.Repeat("─", 30) + "\n")
	for i, n := range m.filteredProjects() {
		prefix := "  "
		if i == m.pickerCursor {
			prefix = m.styles.Accent.Render("◆ ")
		}
		b.WriteString(prefix + n + "\n")
	}
	return b.String()
}

type previewInput struct {
	Styles  *Styles
	Width   int
	Height  int
	Task    *model.Task
	Focused bool
}

func renderPreview(in previewInput) string {
	border := in.Styles.PaneBorder
	if in.Focused {
		border = in.Styles.PaneBorderFocused
	}
	width := in.Width
	if width < 20 {
		width = 20
	}
	height := in.Height
	if height < 5 {
		height = 5
	}

	var header, body string
	switch {
	case in.Task == nil:
		header = "Notes"
		body = in.Styles.Muted.Render(
			"No tasks match the current filter.\n" +
				"Press Esc to reset filter, o to add one.")
	case !in.Task.HasNotes():
		glyph := in.Styles.GlyphStyle(in.Task.Status).Render(in.Styles.StatusGlyph(in.Task.Status))
		header = "Notes  " + glyph + " " + in.Task.Title
		body = in.Styles.Muted.Render(
			"No notes yet.\nPress Enter to open this task in your editor.")
	default:
		glyph := in.Styles.GlyphStyle(in.Task.Status).Render(in.Styles.StatusGlyph(in.Task.Status))
		header = "Notes  " + glyph + " " + in.Task.Title
		raw := stripNotesIndent(in.Task.Notes)
		r, err := newNotesRenderer(width - 4)
		if err == nil {
			rendered, rerr := r.Render(raw)
			if rerr == nil {
				body = strings.TrimRight(rendered, "\n")
			}
		}
		if body == "" {
			body = raw
		}
	}

	content := header + "\n\n" + body
	return border.Width(width).Height(height).Render(content)
}

func stripNotesIndent(notes string) string {
	lines := strings.Split(notes, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimPrefix(ln, "    ")
	}
	return strings.Join(lines, "\n")
}

func filterLabel(f Filter) string {
	if f.All {
		return "all"
	}
	parts := make([]string, 0, len(f.Statuses))
	for _, s := range f.Statuses {
		parts = append(parts, s.String())
	}
	return strings.Join(parts, "+")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
