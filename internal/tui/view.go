package tui

import (
	"fmt"
	"strings"

	"github.com/npaolopepito/himo/internal/model"
)

func renderView(m Model) string {
	tasks := m.visibleTasks()
	listStr := renderList(m, tasks)
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
		if tasks := m.visibleTasks(); m.cursor < len(tasks) {
			view += `Delete "` + tasks[m.cursor].Title + `"? y/n` + "\n"
		}
	}
	if m.banner != "" {
		view += "! " + m.banner + "\n"
	}
	return view
}

func renderList(m Model, tasks []model.Task) string {
	var b strings.Builder
	header := fmt.Sprintf("himo  %s  %s  %d tasks", m.project.Name, filterLabel(m.filter), len(tasks))
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
		fmt.Fprintf(&b, "%s%s %s%s\n", prefix, marker, t.Title, note)
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
