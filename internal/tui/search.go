package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// highlightMatch returns s with all case-insensitive occurrences of needle
// rendered through hl, and the rest rendered through base. Empty needle
// returns base.Render(s) unchanged.
func highlightMatch(s, needle string, base, hl lipgloss.Style) string {
	if needle == "" {
		return base.Render(s)
	}
	lowS := strings.ToLower(s)
	lowN := strings.ToLower(needle)
	if !strings.Contains(lowS, lowN) {
		return base.Render(s)
	}
	var b strings.Builder
	i := 0
	for i < len(lowS) {
		j := strings.Index(lowS[i:], lowN)
		if j < 0 {
			b.WriteString(base.Render(s[i:]))
			break
		}
		if j > 0 {
			b.WriteString(base.Render(s[i : i+j]))
		}
		b.WriteString(hl.Render(s[i+j : i+j+len(lowN)]))
		i += j + len(lowN)
	}
	return b.String()
}
