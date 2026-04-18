package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
)

const helpText = `Keybindings:

Navigation
  j/k        move cursor down/up
  g/G        top/bottom
  Ctrl+d/u   half page down/up
  /          search
  q          quit

Scope
  Tab/S-Tab  prev/next project
  P          project picker
  A          all-projects view

Filters
  0          all
  1          backlog
  2          pending
  3          active
  4          blocked
  5          done
  6          cancelled
  Esc        default (pending+active+blocked)

Actions
  Enter      open task notes in $EDITOR
  e          open current-filter file in $EDITOR
  Space      cycle status forward
  b/p/a      backlog / pending / active
  !/x/-      blocked / done / cancelled
  o/O        new task below / above
  d          delete (y/n confirm)
  v          toggle preview pane
  ?          toggle this help
`

func renderView(m Model) string {
	if m.showingHelp {
		return helpText
	}
	locs := m.visibleTaskLocations()
	tasks := make([]model.Task, len(locs))
	for i, loc := range locs {
		tasks[i] = loc.doc.Items[loc.idx].(store.TaskItem).Task
	}
	listStr := renderList(m, locs, tasks)
	var view string
	if m.width < 100 || m.hidePreview {
		view = listStr
	} else {
		view = sideBySide(listStr, renderPreview(m, tasks), m.width)
	}
	if m.prompting {
		view += "> new task: " + m.promptBuf + "_\n"
	}
	if m.searching {
		view += "/ search: " + m.searchBuf + "_\n"
	}
	if m.confirmingDelete {
		if m.cursor < len(tasks) {
			view += `Delete "` + tasks[m.cursor].Title + `"? y/n` + "\n"
		}
	}
	if m.pickerOpen {
		view += renderPicker(m)
	}
	if m.banner != "" {
		view += "! " + m.banner + "\n"
	}
	return view
}

func renderList(m Model, locs []taskLoc, tasks []model.Task) string {
	var b strings.Builder
	scope := m.project.Name
	if m.allProjects {
		scope = "all"
	}
	header := fmt.Sprintf("himo  %s  %s  %d tasks", scope, filterLabel(m.filter), len(tasks))
	if m.searchActive != "" {
		header += "  [search: " + m.searchActive + "]"
	}
	fmt.Fprintf(&b, "%s\n", header)
	b.WriteString(strings.Repeat("-", 60))
	b.WriteByte('\n')
	for i, t := range tasks {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		marker := t.Status.Marker()
		if marker == "" {
			marker = "-  "
		}
		note := "   "
		if t.HasNotes() {
			note = " N "
		}
		title := t.Title
		if m.allProjects && i < len(locs) {
			title = "[" + locs[i].proj.Name + "] " + title
		}
		fmt.Fprintf(&b, "%s%s %s%s\n", prefix, marker, title, note)
	}
	return b.String()
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

func renderPicker(m Model) string {
	var b strings.Builder
	b.WriteString("[picker] filter: " + m.pickerFilter + "_\n")
	for i, n := range m.filteredProjects() {
		prefix := "  "
		if i == m.pickerCursor {
			prefix = "> "
		}
		b.WriteString(prefix + n + "\n")
	}
	return b.String()
}

func renderPreview(m Model, tasks []model.Task) string {
	if len(tasks) == 0 || m.cursor >= len(tasks) {
		return "Notes:\n(no task selected)"
	}
	t := tasks[m.cursor]
	if !t.HasNotes() {
		return "Notes: " + t.Title + "\n\n(no notes - press Enter to add)"
	}
	lines := strings.Split(t.Notes, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimPrefix(ln, "    ")
	}
	return "Notes: " + t.Title + "\n\n" + strings.Join(lines, "\n")
}

func sideBySide(left, right string, width int) string {
	leftW := width * 60 / 100
	rightW := width - leftW - 3
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")
	n := maxInt(len(leftLines), len(rightLines))
	var b strings.Builder
	for i := 0; i < n; i++ {
		var l, r string
		if i < len(leftLines) {
			l = leftLines[i]
		}
		if i < len(rightLines) {
			r = rightLines[i]
		}
		if len(l) > leftW {
			l = l[:leftW]
		} else {
			l = l + strings.Repeat(" ", leftW-len(l))
		}
		if len(r) > rightW {
			r = r[:rightW]
		}
		fmt.Fprintf(&b, "%s | %s\n", l, r)
	}
	return b.String()
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
