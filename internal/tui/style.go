package tui

import (
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/npaolopepito/himo/internal/model"
)

// StyleOptions toggles the visual degradations. Both default to false.
type StyleOptions struct {
	AsciiGlyphs bool
	NoColor     bool
}

// Styles holds every Lip Gloss style and glyph lookup used by the TUI.
// One instance is built at Model construction; views never allocate styles
// inline.
type Styles struct {
	opts     StyleOptions
	renderer *lipgloss.Renderer

	// Text styles.
	Base   lipgloss.Style
	Muted  lipgloss.Style
	Accent lipgloss.Style
	Err    lipgloss.Style
	Ok     lipgloss.Style
	Dim    lipgloss.Style
	Strike lipgloss.Style

	// Cursor row.
	CursorBar   lipgloss.Style
	CursorRowBG lipgloss.Style

	// Borders.
	PaneBorder        lipgloss.Style
	PaneBorderFocused lipgloss.Style

	// Chips + pills (filter bar, mode pill).
	ChipMuted  lipgloss.Style
	ChipActive lipgloss.Style
	ModePill   lipgloss.Style

	// Task title variants by status.
	TitleBacklog   lipgloss.Style
	TitlePending   lipgloss.Style
	TitleActive    lipgloss.Style
	TitleBlocked   lipgloss.Style
	TitleDone      lipgloss.Style
	TitleCancelled lipgloss.Style

	// Status glyphs colored.
	GlyphBacklog   lipgloss.Style
	GlyphPending   lipgloss.Style
	GlyphActive    lipgloss.Style
	GlyphBlocked   lipgloss.Style
	GlyphDone      lipgloss.Style
	GlyphCancelled lipgloss.Style
}

// NewStyles builds a Styles for the default renderer.
func NewStyles(opts StyleOptions) *Styles {
	return NewStylesWithRenderer(lipgloss.DefaultRenderer(), opts)
}

// NewStylesWithRenderer builds a Styles against a given renderer. Tests use
// this to pin a termenv profile for stable output.
func NewStylesWithRenderer(r *lipgloss.Renderer, opts StyleOptions) *Styles {
	if opts.NoColor {
		r = lipgloss.NewRenderer(io.Discard, termenv.WithProfile(termenv.Ascii))
	}
	st := &Styles{opts: opts, renderer: r}

	muted := lipgloss.AdaptiveColor{Light: "#9ca3af", Dark: "#6b7280"}
	accent := lipgloss.AdaptiveColor{Light: "#a21caf", Dark: "#d946ef"}
	errc := lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#ef4444"}
	ok := lipgloss.AdaptiveColor{Light: "#15803d", Dark: "#22c55e"}
	subtle := lipgloss.AdaptiveColor{Light: "#e5e7eb", Dark: "#374151"}

	st.Base = r.NewStyle()
	st.Muted = r.NewStyle().Foreground(muted)
	st.Accent = r.NewStyle().Foreground(accent)
	st.Err = r.NewStyle().Foreground(errc)
	st.Ok = r.NewStyle().Foreground(ok)
	st.Dim = r.NewStyle().Foreground(muted)
	st.Strike = r.NewStyle().Strikethrough(true).Foreground(muted)

	st.CursorBar = r.NewStyle().Foreground(accent).Bold(true)
	st.CursorRowBG = r.NewStyle().Background(subtle)

	st.PaneBorder = r.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(muted)
	st.PaneBorderFocused = r.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(accent)

	st.ChipMuted = r.NewStyle().Foreground(muted)
	st.ChipActive = r.NewStyle().
		Foreground(accent).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(accent)

	// Mode pill: filled accent background, bright foreground.
	st.ModePill = r.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(accent).
		Bold(true).
		Padding(0, 1)

	// Title variants.
	st.TitleBacklog = r.NewStyle().Italic(true).Foreground(muted)
	st.TitlePending = r.NewStyle()
	st.TitleActive = r.NewStyle().Bold(true)
	st.TitleBlocked = r.NewStyle()
	st.TitleDone = r.NewStyle().Strikethrough(true).Foreground(muted)
	st.TitleCancelled = r.NewStyle().Strikethrough(true).Foreground(muted)

	// Glyph variants.
	st.GlyphBacklog = r.NewStyle().Foreground(muted)
	st.GlyphPending = r.NewStyle()
	st.GlyphActive = r.NewStyle().Foreground(ok).Bold(true)
	st.GlyphBlocked = r.NewStyle().Foreground(errc).Bold(true)
	st.GlyphDone = r.NewStyle().Foreground(ok)
	st.GlyphCancelled = r.NewStyle().Foreground(errc).Faint(true)

	return st
}

// StatusGlyph returns the glyph to render for s, respecting AsciiGlyphs.
func (s *Styles) StatusGlyph(st model.Status) string {
	if s.opts.AsciiGlyphs {
		switch st {
		case model.StatusBacklog, model.StatusPending:
			return "o"
		case model.StatusActive, model.StatusBlocked:
			return "*"
		case model.StatusDone:
			return "x"
		case model.StatusCancelled:
			return "-"
		}
		return "?"
	}
	switch st {
	case model.StatusBacklog, model.StatusPending:
		return "○"
	case model.StatusActive, model.StatusBlocked:
		return "●"
	case model.StatusDone:
		return "✓"
	case model.StatusCancelled:
		return "✗"
	}
	return "?"
}

// GlyphStyle returns the style to use when rendering the glyph for s.
func (s *Styles) GlyphStyle(st model.Status) lipgloss.Style {
	switch st {
	case model.StatusBacklog:
		return s.GlyphBacklog
	case model.StatusPending:
		return s.GlyphPending
	case model.StatusActive:
		return s.GlyphActive
	case model.StatusBlocked:
		return s.GlyphBlocked
	case model.StatusDone:
		return s.GlyphDone
	case model.StatusCancelled:
		return s.GlyphCancelled
	}
	return s.Base
}

// TitleStyle returns the style to use when rendering a task title with
// status st.
func (s *Styles) TitleStyle(st model.Status) lipgloss.Style {
	switch st {
	case model.StatusBacklog:
		return s.TitleBacklog
	case model.StatusPending:
		return s.TitlePending
	case model.StatusActive:
		return s.TitleActive
	case model.StatusBlocked:
		return s.TitleBlocked
	case model.StatusDone:
		return s.TitleDone
	case model.StatusCancelled:
		return s.TitleCancelled
	}
	return s.Base
}
