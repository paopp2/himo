package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Mode is the current interaction mode. Drives the hint bar + help.
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModePrompt
	ModeDelete
	ModePicker
	ModeHelp
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeSearch:
		return "SEARCH"
	case ModePrompt:
		return "PROMPT"
	case ModeDelete:
		return "DELETE?"
	case ModePicker:
		return "PICKER"
	case ModeHelp:
		return "HELP"
	}
	return "?"
}

type hintInput struct {
	Mode        Mode
	Width       int
	SearchBuf   string
	PromptBuf   string
	PromptAbove bool
	DeleteTitle string
	Banner      string
}

func renderHintBar(st *Styles, in hintInput) string {
	pill := st.ModePill.Render(in.Mode.String())
	mid := middleHints(st, in)
	meta := metaHints(st, in)

	width := in.Width
	if width <= 0 {
		width = 80
	}
	pad := width - lipgloss.Width(pill) - lipgloss.Width(mid) - lipgloss.Width(meta) - 4
	if pad < 1 {
		pad = 1
	}
	return pill + "  " + mid + strings.Repeat(" ", pad) + meta
}

func middleHints(st *Styles, in hintInput) string {
	switch in.Mode {
	case ModeNormal:
		return hintList(st, [][2]string{
			{"j/k", "move"}, {"Enter", "notes"}, {"Space", "cycle"},
			{"o", "new"}, {"/", "search"},
		})
	case ModeSearch:
		return st.Muted.Render("/ " + in.SearchBuf + "_")
	case ModePrompt:
		label := "New task"
		if in.PromptAbove {
			label = "New task (above)"
		}
		return st.Muted.Render(label+" > ") + in.PromptBuf + "_"
	case ModeDelete:
		return st.Err.Render("Delete: ") + in.DeleteTitle
	case ModePicker:
		return st.Muted.Render("Switch project")
	case ModeHelp:
		return st.Muted.Render("Full cheat sheet below")
	}
	return ""
}

func metaHints(st *Styles, in hintInput) string {
	var parts []string
	if in.Banner != "" {
		parts = append(parts, st.Err.Render("! ")+in.Banner)
	}
	switch in.Mode {
	case ModeNormal, ModeHelp:
		parts = append(parts, st.Muted.Render("? help"))
	case ModeSearch, ModePrompt, ModePicker:
		parts = append(parts, st.Muted.Render("Enter apply  Esc cancel"))
	case ModeDelete:
		parts = append(parts, st.Muted.Render("y delete  n cancel"))
	}
	return strings.Join(parts, "  |  ")
}

func hintList(st *Styles, kvs [][2]string) string {
	var parts []string
	for _, kv := range kvs {
		parts = append(parts, fmt.Sprintf("%s %s", kv[0], st.Muted.Render(kv[1])))
	}
	return strings.Join(parts, "  ")
}
