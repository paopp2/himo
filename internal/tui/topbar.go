package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type topBarInput struct {
	Projects []string
	Current  string
	Width    int
	AllMode  bool
	Sort     Sort
}

func renderTopBar(st *Styles, in topBarInput) string {
	if in.Width < narrowThreshold {
		name := in.Current
		if in.AllMode {
			name = "all"
		}
		return st.Muted.Render("himo · " + name)
	}
	var left string
	if in.AllMode {
		left = st.Accent.Render("◆ all projects")
	} else {
		parts := make([]string, 0, len(in.Projects))
		for _, p := range in.Projects {
			if p == in.Current {
				parts = append(parts, st.Accent.Render("◆ "+p))
			} else {
				parts = append(parts, st.Muted.Render(p))
			}
		}
		left = strings.Join(parts, "    ")
	}
	right := st.Muted.Render("[A] all") + "  " +
		st.Muted.Render("[P] picker") + "  " +
		st.Muted.Render("[s] sort:"+sortName(in.Sort))

	padWidth := in.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if padWidth < 2 {
		padWidth = 2
	}
	return left + strings.Repeat(" ", padWidth) + right
}
