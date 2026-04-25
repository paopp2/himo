package tui

import (
	"github.com/charmbracelet/lipgloss"
)

type modalInput struct {
	Title    string
	Body     string
	Hints    string
	Width    int
	Height   int
	Error    bool
	BoxWidth int // overrides default sizing when > 0
}

// centeredBox renders body with title + hints in a rounded-border box,
// sized to min(50, 60% of terminal), centered within (Width, Height).
// Error=true styles the title in red instead of accent. BoxWidth, when
// set, overrides the default sizing -- used by wide content like the
// help screen that doesn't fit the 50-cell cap.
func centeredBox(st *Styles, in modalInput) string {
	boxW := in.Width * 60 / 100
	if boxW > 50 {
		boxW = 50
	}
	if boxW < 30 {
		boxW = 30
	}
	if in.BoxWidth > 0 {
		boxW = in.BoxWidth
		if boxW > in.Width-2 {
			boxW = in.Width - 2
		}
	}

	titleStyle := st.Accent
	if in.Error {
		titleStyle = st.Err
	}

	content := titleStyle.Render(in.Title) + "\n\n" +
		in.Body + "\n\n" +
		st.Muted.Render(in.Hints)

	box := st.PaneBorderFocused.
		Padding(1, 2).
		Width(boxW).
		Render(content)

	return lipgloss.Place(in.Width, in.Height,
		lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
	)
}
