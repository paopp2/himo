package tui

import (
	"path/filepath"

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
	project  *store.Project
	filter   Filter
	cursor   int
	width    int
	height   int
	quit     bool
	baseDir  string
	projects []string
}

// NewModel builds a fresh Model for the given project.
func NewModel(p *store.Project) Model {
	return Model{project: p, filter: DefaultFilter()}
}

// NewModelFromBase loads the named project from baseDir and returns a Model
// seeded with the list of sibling projects for Tab cycling.
func NewModelFromBase(baseDir, name string) (Model, error) {
	p, err := store.LoadProject(filepath.Join(baseDir, name))
	if err != nil {
		return Model{}, err
	}
	projects, err := store.ListProjects(baseDir)
	if err != nil {
		return Model{}, err
	}
	return Model{
		project:  p,
		filter:   DefaultFilter(),
		baseDir:  baseDir,
		projects: projects,
	}, nil
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor+1 < len(m.visibleTasks()) {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "g":
			m.cursor = 0
		case "G":
			if n := len(m.visibleTasks()); n > 0 {
				m.cursor = n - 1
			}
		case "ctrl+d":
			half := maxInt(m.height/2, 1)
			if n := len(m.visibleTasks()); n > 0 {
				m.cursor = minInt(m.cursor+half, n-1)
			}
		case "ctrl+u":
			half := maxInt(m.height/2, 1)
			m.cursor = maxInt(m.cursor-half, 0)
		case "0":
			m.filter = Filter{All: true}
			m.cursor = 0
		case "1":
			m.filter = Filter{Statuses: []model.Status{model.StatusBacklog}}
			m.cursor = 0
		case "2":
			m.filter = Filter{Statuses: []model.Status{model.StatusPending}}
			m.cursor = 0
		case "3":
			m.filter = Filter{Statuses: []model.Status{model.StatusActive}}
			m.cursor = 0
		case "4":
			m.filter = Filter{Statuses: []model.Status{model.StatusBlocked}}
			m.cursor = 0
		case "5":
			m.filter = Filter{Statuses: []model.Status{model.StatusDone}}
			m.cursor = 0
		case "6":
			m.filter = Filter{Statuses: []model.Status{model.StatusCancelled}}
			m.cursor = 0
		case "esc":
			m.filter = DefaultFilter()
			m.cursor = 0
		case "tab":
			m.switchProject(+1)
		case "shift+tab":
			m.switchProject(-1)
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

// switchProject cycles m.project by delta through m.projects, in place.
// Delta of +1 is next, -1 is previous; wraps around.
func (m *Model) switchProject(delta int) {
	if len(m.projects) == 0 {
		return
	}
	idx := -1
	for i, n := range m.projects {
		if n == m.project.Name {
			idx = i
			break
		}
	}
	if idx < 0 {
		idx = 0
	}
	next := (idx + delta + len(m.projects)) % len(m.projects)
	p, err := store.LoadProject(filepath.Join(m.baseDir, m.projects[next]))
	if err != nil {
		return
	}
	m.project = p
	m.cursor = 0
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
