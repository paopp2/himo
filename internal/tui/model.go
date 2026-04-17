package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/npaolopepito/himo/internal/store"
)

// Model is the top-level Bubble Tea model.
type Model struct {
	project *store.Project
	width   int
	height  int
	quit    bool
}

// NewModel builds a fresh Model for the given project.
func NewModel(p *store.Project) Model {
	return Model{project: p}
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
	return "himo (scaffold)\n"
}
