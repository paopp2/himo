package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
)

// Filter narrows the visible tasks by status.
type Filter struct {
	Statuses []model.Status
	All      bool
}

// DefaultFilter is the active+pending+blocked view shown on startup.
func DefaultFilter() Filter {
	return Filter{Statuses: []model.Status{model.StatusPending, model.StatusActive, model.StatusBlocked}}
}

// Model is the top-level Bubble Tea model.
type Model struct {
	project *store.Project
	filter  Filter
	cursor  int
	width   int
	height  int
	quit    bool
}

// NewModel builds a fresh Model for the given project.
func NewModel(p *store.Project) Model {
	return Model{project: p, filter: DefaultFilter()}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quit {
		return ""
	}
	return renderView(m)
}

func (m Model) visibleTasks() []model.Task {
	if m.filter.All {
		return m.project.AllTasks()
	}
	all := m.project.AllTasks()
	out := make([]model.Task, 0, len(all))
	for _, t := range all {
		for _, s := range m.filter.Statuses {
			if t.Status == s {
				out = append(out, t)
				break
			}
		}
	}
	return out
}
