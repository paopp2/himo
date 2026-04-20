package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/paopp2/himo/internal/model"
)

// renderFilterBar returns "" below narrowThreshold so narrow layouts drop
// the row entirely.
func renderFilterBar(st *Styles, f Filter, counts map[model.Status]int, width int) string {
	if width < narrowThreshold {
		return ""
	}
	type chip struct {
		key       string
		label     string
		st        model.Status
		isAll     bool
		isDefault bool
	}
	chips := []chip{
		{"`", "Default", 0, false, true},
		{"1", "Backlog", model.StatusBacklog, false, false},
		{"2", "Pending", model.StatusPending, false, false},
		{"3", "Active", model.StatusActive, false, false},
		{"4", "Blocked", model.StatusBlocked, false, false},
		{"5", "Done", model.StatusDone, false, false},
		{"6", "Cancelled", model.StatusCancelled, false, false},
		{"0", "All", 0, true, false},
	}
	isActive := func(c chip) bool {
		switch {
		case c.isAll:
			return f.All
		case c.isDefault:
			return isDefaultFilter(f)
		}
		for _, s := range f.Statuses {
			if s == c.st {
				return true
			}
		}
		return false
	}
	var parts []string
	for i, c := range chips {
		if i > 0 {
			parts = append(parts, "  ")
		}
		text := fmt.Sprintf("[%s] %s", c.key, c.label)
		if !c.isAll && !c.isDefault {
			text = fmt.Sprintf("%s %d", text, counts[c.st])
		}
		if isActive(c) {
			parts = append(parts, st.ChipActive.Render(text))
		} else {
			parts = append(parts, st.ChipMuted.Render(text))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, parts...)
}
