package tui

import (
	"fmt"
	"strings"
)

func renderView(m Model) string {
	tasks := m.visibleTasks()
	var b strings.Builder
	fmt.Fprintf(&b, "himo  %s  %s  %d tasks\n", m.project.Name, filterLabel(m.filter), len(tasks))
	b.WriteString(strings.Repeat("-", minInt(m.width, 60)))
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
