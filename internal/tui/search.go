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

// nextMatch returns the closest matching position in locs in the requested
// direction starting at from. Scans up to len(locs) positions, wrapping if
// needed. `from` may be out of [0,len) range; this signals the cursor
// stepped off the end. When it does, any match found is by definition a wrap.
// Returns ok=false when there are zero matches (or empty list/needle).
func nextMatch(locs []taskLoc, needle string, allProjects bool, from int, forward bool) (idx int, wrapped bool, ok bool) {
	n := len(locs)
	if n == 0 || needle == "" {
		return 0, false, false
	}
	needleLower := strings.ToLower(needle)
	step := 1
	if !forward {
		step = -1
	}
	fromOOB := from < 0 || from >= n
	start := ((from % n) + n) % n
	for i := 0; i < n; i++ {
		pos := ((start+step*i)%n + n) % n
		if !taskLocMatches(locs[pos], needleLower, allProjects) {
			continue
		}
		w := fromOOB
		if !w {
			if forward && pos < start {
				w = true
			} else if !forward && pos > start {
				w = true
			}
		}
		return pos, w, true
	}
	return 0, false, false
}
