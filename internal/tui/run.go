package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/npaolopepito/himo/internal/store"
)

// Run opens the TUI on the given project.
func Run(p *store.Project) error {
	prog := tea.NewProgram(NewModel(p), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
