package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/paopp2/himo/internal/model"
	"github.com/paopp2/himo/internal/store"
)

const (
	// narrowThreshold collapses the top bar into a single line and hides
	// the filter bar. The filter chips alone need ~95 cols to fit, and the
	// preview pane also hides at this width, so everything collapses together.
	narrowThreshold  = 100
	previewThreshold = 100 // below this, the preview pane hides.
	paneSeparator    = 2   // "  " between list and preview panes.
	defaultWidth     = 80
	defaultHeight    = 10
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
		{"n/N", "next / prev match"},
		{"Tab", "next project"},
		{"S-Tab", "prev project"},
		{"P", "project picker"},
		{"A", "all projects"},
		{"q", "quit"},
	})
	filters := col("Filters", [][2]string{
		{"`", "default filter"},
		{"1", "backlog"},
		{"2", "pending"},
		{"3", "active"},
		{"4", "blocked"},
		{"5", "done"},
		{"6", "cancelled"},
		{"0", "all"},
		{"s", "toggle sort"},
	})
	actions := col("Actions", [][2]string{
		{"Enter", "notes in $EDITOR"},
		{"Ctrl+o", "open URL"},
		{"e", "edit title inline"},
		{"Space", "cycle status"},
		{"b/p/a", "backlog / pending / active"},
		{"!/x/-", "blocked / done / cancelled"},
		{"o/O", "new below / above"},
		{"d", "delete"},
		{"u", "undo"},
		{"Ctrl+R", "redo"},
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
	modalHeight := m.height
	if modalHeight < defaultHeight {
		modalHeight = defaultHeight
	}
	if m.showingHelp {
		boxW := width - 8
		if boxW < 60 {
			boxW = 60
		}
		// renderHelp lays out 3 columns inside the box; pass the inner
		// width so the columns fit after the border + horizontal padding.
		innerW := boxW - 4
		return centeredBox(m.styles, modalInput{
			Title:    "Keybindings",
			Body:     renderHelp(m.styles, innerW),
			Hints:    "? close",
			Width:    width,
			Height:   modalHeight,
			BoxWidth: boxW,
		})
	}

	locs := m.visibleTaskLocations()
	tasks := make([]model.Task, len(locs))
	for i, loc := range locs {
		tasks[i] = loc.doc.Items[loc.idx].(store.TaskItem).Task
	}

	switch m.currentMode() {
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
		Sort:     m.sort,
	})
	fbar := renderFilterBar(m.styles, m.filter, m.statusCounts(), width)

	// Non-pane chrome rows: top bar + [filter bar] + blank + hint bar.
	chromeRows := 3
	if fbar != "" {
		chromeRows = 4
	}
	paneHeight := m.height - chromeRows

	var body string
	if width < previewThreshold || m.hidePreview {
		body = renderListPane(m, locs, tasks, width, paneHeight)
	} else {
		listW := width * 60 / 100
		previewW := width - listW - paneSeparator
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

	query := m.activeSearchQuery()
	matches := matchIndices(locs, query, m.allProjects)
	matchIdx := -1
	for i, p := range matches {
		if p == m.cursor {
			matchIdx = i + 1
			break
		}
	}

	hint := renderHintBar(m.styles, hintInput{
		Mode:           m.currentMode(),
		Width:          width,
		SearchBuf:      m.searchInput.View(),
		DeleteTitle:    deleteTitle(m, tasks),
		Banner:         m.banner,
		SearchActive:   query,
		SearchMatchIdx: matchIdx,
		SearchTotal:    len(matches),
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
		opts := taskLineInput{
			Width:       width,
			Cursor:      i == m.cursor,
			SearchQuery: m.activeSearchQuery(),
		}
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
// the cursor. When the user is prompting for a new task, a ghost row is
// spliced in at the insertion position (above the cursor for O, at the
// end of the list for o) and stays visible while the buffer is edited.
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
		opts := taskLineInput{
			Width:       width - 2,
			Cursor:      i == m.cursor && !m.prompting,
			SearchQuery: m.activeSearchQuery(),
		}
		if i == m.cursor && m.editing {
			opts.Editing = true
			opts.EditView = m.editInput.View()
		}
		if m.allProjects && i < len(locs) {
			opts.AllProjects = true
			opts.ProjectName = locs[i].proj.Name
		}
		rows[i] = renderTaskLine(m.styles, t, opts)
	}

	cursorRow := m.cursor
	if m.prompting {
		ghostIdx := len(rows)
		if len(rows) > 0 {
			if m.promptAbove {
				ghostIdx = m.cursor
			} else {
				ghostIdx = m.cursor + 1
			}
			if ghostIdx < 0 {
				ghostIdx = 0
			}
			if ghostIdx > len(rows) {
				ghostIdx = len(rows)
			}
		}
		ghost := renderGhostRow(m.styles, m.promptInput.View(), width-2)
		rows = append(rows[:ghostIdx], append([]string{ghost}, rows[ghostIdx:]...)...)
		cursorRow = ghostIdx
	}

	start := 0
	if len(rows) > contentH {
		start = cursorRow - contentH/2
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
	if visible == "" {
		visible = m.styles.Muted.Render(
			"No tasks match the current filter.\n" +
				"Press ` to reset filter, o to add one.")
	}
	// Subtract 2 from both dims so the rendered box (inner + border) is
	// exactly width x height cells. Lipgloss adds the border outside of
	// Width/Height.
	return m.styles.PaneBorderFocused.Width(width - 2).Height(height - 2).Render(visible)
}

func renderGhostRow(st *Styles, body string, width int) string {
	bar := st.CursorBar.Render("▌")
	marker := st.Accent.Render("+")
	left := bar + " " + marker + " " + body
	padding := width - lipgloss.Width(left)
	if padding < 0 {
		padding = 0
	}
	row := left + strings.Repeat(" ", padding)
	return st.PaintCursorRow(row)
}

type taskLineInput struct {
	Width       int
	Cursor      bool
	AllProjects bool
	ProjectName string
	Editing     bool
	EditView    string
	SearchQuery string
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

	// Render chip and title as independent segments. Concatenating a
	// pre-rendered (ANSI-laden) chip into the title and then re-rendering
	// with TitleStyle triggers lipgloss's char-by-char Strikethrough path,
	// which wraps each inner ESC byte individually and mangles the stream.
	var chip string
	if o.AllProjects && o.ProjectName != "" {
		chipStyle := st.Muted
		switch t.Status {
		case model.StatusDone, model.StatusCancelled:
			chipStyle = chipStyle.Strikethrough(true)
		}
		chipText := "[" + o.ProjectName + "] "
		if o.SearchQuery == "" {
			chip = chipStyle.Render(chipText)
		} else {
			chip = highlightMatch(chipText, o.SearchQuery, chipStyle, st.SearchHighlight)
		}
	}

	// Reserve cells for the row's fixed columns so a long title gets
	// truncated rather than wrapping. Layout: bar(1) " "(1) glyph(1) " "(1)
	// chip title " "(1) dot(1) -> 7 + chipWidth fixed cells. Leaving the
	// trailing separator space inside the budget keeps the dot column
	// aligned with rows whose titles fit naturally.
	titleMax := o.Width - 7 - lipgloss.Width(chip)
	if titleMax < 1 {
		titleMax = 1
	}
	var title string
	if o.Editing {
		title = o.EditView
	} else {
		truncated := runewidth.Truncate(t.Title, titleMax, "…")
		if o.SearchQuery == "" {
			title = st.TitleStyle(t.Status).Render(truncated)
		} else {
			title = highlightMatch(truncated, o.SearchQuery, st.TitleStyle(t.Status), st.SearchHighlight)
		}
	}
	title = chip + title

	dot := " "
	if t.HasNotes() {
		if t.URL() != "" {
			dot = st.Accent.Render("•")
		} else {
			dot = st.Muted.Render("•")
		}
	}

	left := bar + " " + glyph + " " + title
	padding := o.Width - lipgloss.Width(left) - 2
	if padding < 1 {
		padding = 1
	}
	row := left + strings.Repeat(" ", padding) + dot

	if o.Cursor {
		row = st.PaintCursorRow(row)
	}
	return row
}

func renderPickerBody(m Model) string {
	var b strings.Builder
	b.WriteString("/ " + m.pickerInput.View() + "\n\n")
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
		body = in.Styles.Muted.Render(
			"No tasks match the current filter.\n" +
				"Press ` to reset filter, o to add one.")
	case !in.Task.HasNotes():
		glyph := in.Styles.GlyphStyle(in.Task.Status).Render(in.Styles.StatusGlyph(in.Task.Status))
		header = glyph + " " + in.Task.Title
		body = in.Styles.Muted.Render(
			"No notes yet.\nPress Enter to open this task in your editor.")
	default:
		glyph := in.Styles.GlyphStyle(in.Task.Status).Render(in.Styles.StatusGlyph(in.Task.Status))
		header = glyph + " " + in.Task.Title
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

	content := body
	if header != "" {
		content = header + "\n\n" + body
	}
	// lipgloss.Height is a minimum, so long notes would otherwise grow the
	// box past its allotted paneHeight and overflow the terminal. Clamp
	// here, marking truncation with a muted ellipsis.
	content = clampLines(content, height-2, in.Styles.Muted.Render("…"))
	return in.Styles.PaneBorderFocused.Width(width - 2).Height(height - 2).Render(content)
}

// clampLines returns s trimmed to at most n lines. When truncated, the
// last line is replaced with marker so the reader sees content was cut.
func clampLines(s string, n int, marker string) string {
	if n <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	lines = lines[:n]
	lines[n-1] = marker
	return strings.Join(lines, "\n")
}

func stripNotesIndent(notes string) string {
	lines := strings.Split(notes, "\n")
	// Strip the minimum common leading-space indent so relative indentation
	// (e.g. nested list items) survives. Hardcoding a 4-space strip flattens
	// mixed-indent notes into a single level.
	minIndent := -1
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		n := 0
		for n < len(ln) && ln[n] == ' ' {
			n++
		}
		if minIndent == -1 || n < minIndent {
			minIndent = n
		}
	}
	if minIndent < 0 {
		minIndent = 0
	}
	for i, ln := range lines {
		if len(ln) >= minIndent {
			ln = ln[minIndent:]
		}
		// Markdown treats adjacent non-blank lines as soft-wrapped text in
		// the same paragraph; our users type each note line expecting it
		// to render on its own row. Append CommonMark's two-space hard
		// break so Glamour honors the line break.
		if strings.TrimSpace(ln) != "" {
			ln += "  "
		}
		lines[i] = ln
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
