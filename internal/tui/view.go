package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
)

const (
	narrowThreshold    = 80  // below this, top/filter bars collapse.
	previewThreshold   = 100 // below this, the preview pane hides.
	previewGutter      = 3   // cells between list pane and preview pane.
	bodyVerticalChrome = 4   // top bar + filter bar + blank + hint bar rows.
	defaultWidth       = 80
	defaultHeight      = 10
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
	width := m.width
	if width <= 0 {
		width = defaultWidth
	}
	if m.showingHelp {
		return renderHelp(m.styles, width) + "\n" +
			renderHintBar(m.styles, hintInput{Mode: ModeHelp, Width: width})
	}

	modalHeight := m.height
	if modalHeight < defaultHeight {
		modalHeight = defaultHeight
	}

	locs := m.visibleTaskLocations()
	tasks := make([]model.Task, len(locs))
	for i, loc := range locs {
		tasks[i] = loc.doc.Items[loc.idx].(store.TaskItem).Task
	}

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
			Height: modalHeight,
		})
	case ModeDelete:
		return centeredBox(m.styles, modalInput{
			Title:  "Delete task?",
			Body:   deleteTitle(m, tasks),
			Hints:  "y delete   n cancel",
			Width:  width,
			Height: modalHeight,
			Error:  true,
		})
	case ModePicker:
		return centeredBox(m.styles, modalInput{
			Title:  "Switch project",
			Body:   renderPickerBody(m),
			Hints:  "up/down move   Enter switch   Esc cancel",
			Width:  width,
			Height: modalHeight,
		})
	}

	top := renderTopBar(m.styles, topBarInput{
		Projects: m.projects,
		Current:  m.project.Name,
		Width:    width,
		AllMode:  m.allProjects,
	})
	fbar := renderFilterBar(m.styles, m.filter, m.statusCounts(), width)

	paneHeight := m.height - bodyVerticalChrome
	var body string
	if width < previewThreshold || m.hidePreview {
		body = renderListPane(m, locs, tasks, width, paneHeight)
	} else {
		listW := width * 60 / 100
		previewW := width - listW - previewGutter
		var previewTask *model.Task
		if len(tasks) > 0 && m.cursor < len(tasks) {
			previewTask = &tasks[m.cursor]
		}
		listPane := renderListPane(m, locs, tasks, listW, paneHeight)
		previewPane := renderPreview(previewInput{
			Styles: m.styles,
			Width:  previewW,
			Height: paneHeight,
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
		width = defaultWidth
	}
	for i, t := range tasks {
		opts := taskLineInput{Width: width, Cursor: i == m.cursor}
		if m.allProjects && i < len(locs) {
			opts.AllProjects = true
			opts.ProjectName = locs[i].proj.Name
		}
		b.WriteString(renderTaskLine(m.styles, t, opts))
		b.WriteByte('\n')
	}
	return b.String()
}

// renderListPane renders a bordered, scrolled window of task rows around
// m.cursor so the cursor row is always visible.
func renderListPane(m Model, locs []taskLoc, tasks []model.Task, width, height int) string {
	if height < 5 {
		height = 5
	}
	contentH := height - 2 // border top + bottom
	if contentH < 1 {
		contentH = 1
	}

	rows := make([]string, len(tasks))
	for i, t := range tasks {
		opts := taskLineInput{Width: width - 2, Cursor: i == m.cursor}
		if m.allProjects && i < len(locs) {
			opts.AllProjects = true
			opts.ProjectName = locs[i].proj.Name
		}
		rows[i] = renderTaskLine(m.styles, t, opts)
	}

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
	return m.styles.PaneBorderFocused.Width(width).Height(height).Render(visible)
}

type taskLineInput struct {
	Width       int
	Cursor      bool
	AllProjects bool
	ProjectName string
}

// renderTaskLine returns a single styled row:
//
//	▌ ● [work] Buy groceries                 •
//	^  ^ ^     ^                              ^
//	|  | |     title                          notes dot
//	|  | project chip
//	|  glyph
//	cursor bar
func renderTaskLine(st *Styles, t model.Task, o taskLineInput) string {
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
	b.WriteString("/ " + m.pickerFilter + "_\n\n")
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
	Styles *Styles
	Width  int
	Height int
	Task   *model.Task
}

func renderPreview(in previewInput) string {
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
		// Glamour errors fall back to the raw indented notes. Acceptable:
		// a malformed renderer config or transient resource issue would
		// still show the notes, just unstyled. Surfacing via m.banner would
		// require threading state through render, which view funcs avoid.
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
	return in.Styles.PaneBorderFocused.Width(width).Height(height).Render(content)
}

func stripNotesIndent(notes string) string {
	lines := strings.Split(notes, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimPrefix(ln, "    ")
	}
	return strings.Join(lines, "\n")
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
