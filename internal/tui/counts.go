package tui

import (
	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
)

// statusCounts returns a map of Status -> task count across the current
// scope (single project or all-projects mode). Tasks hidden by the active
// search filter are included; search narrows the list view, not the
// filter-bar counts.
func (m Model) statusCounts() map[model.Status]int {
	out := make(map[model.Status]int, 6)
	add := func(p *store.Project) {
		for _, t := range p.AllTasks() {
			out[t.Status]++
		}
	}
	if m.allProjects {
		for _, p := range m.allProjectsCache {
			add(p)
		}
	} else {
		add(m.project)
	}
	return out
}
