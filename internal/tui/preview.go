package tui

import (
	"github.com/charmbracelet/glamour"
)

// newNotesRenderer builds a Glamour renderer scaled to width. The style is
// "notty" (low-contrast, no background blocks) which matches the muted
// aesthetic of the rest of the TUI.
func newNotesRenderer(width int) (*glamour.TermRenderer, error) {
	if width < 20 {
		width = 20
	}
	return glamour.NewTermRenderer(
		glamour.WithStylePath("notty"),
		glamour.WithWordWrap(width),
	)
}
