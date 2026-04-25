package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/paopp2/himo/internal/store"
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

// taskLocMatches reports whether loc's task title (or project name when
// allProjects is true) contains needleLower (already lowercased).
func taskLocMatches(loc taskLoc, needleLower string, allProjects bool) bool {
	ti, ok := loc.doc.Items[loc.idx].(store.TaskItem)
	if !ok {
		return false
	}
	if strings.Contains(strings.ToLower(ti.Task.Title), needleLower) {
		return true
	}
	if allProjects && loc.proj != nil &&
		strings.Contains(strings.ToLower(loc.proj.Name), needleLower) {
		return true
	}
	return false
}

// matchIndices returns positions in locs whose task matches needle,
// case-insensitive. In all-projects mode the project name is also matched.
// Empty needle returns nil.
func matchIndices(locs []taskLoc, needle string, allProjects bool) []int {
	if needle == "" {
		return nil
	}
	needleLower := strings.ToLower(needle)
	var out []int
	for i, loc := range locs {
		if taskLocMatches(loc, needleLower, allProjects) {
			out = append(out, i)
		}
	}
	return out
}
