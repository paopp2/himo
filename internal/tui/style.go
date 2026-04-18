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

// Styles is the single home for every Lip Gloss style and the glyph table.
type Styles struct {
	asciiGlyphs bool

	Base   lipgloss.Style
	Muted  lipgloss.Style
	Accent lipgloss.Style
	Err    lipgloss.Style

	CursorBar   lipgloss.Style
	CursorRowBG lipgloss.Style

	PaneBorder        lipgloss.Style
	PaneBorderFocused lipgloss.Style

	ChipMuted  lipgloss.Style
	ChipActive lipgloss.Style
	ModePill   lipgloss.Style

	TitleBacklog   lipgloss.Style
	TitlePending   lipgloss.Style
	TitleActive    lipgloss.Style
	TitleBlocked   lipgloss.Style
	TitleDone      lipgloss.Style
	TitleCancelled lipgloss.Style

	GlyphBacklog   lipgloss.Style
	GlyphPending   lipgloss.Style
	GlyphActive    lipgloss.Style
	GlyphBlocked   lipgloss.Style
	GlyphDone      lipgloss.Style
	GlyphCancelled lipgloss.Style
}

func NewStyles(opts StyleOptions) *Styles {
	return NewStylesWithRenderer(lipgloss.DefaultRenderer(), opts)
}

// NewStylesWithRenderer builds a Styles against a given renderer. Tests use
// this to pin a termenv profile for stable output.
func NewStylesWithRenderer(r *lipgloss.Renderer, opts StyleOptions) *Styles {
	if opts.NoColor {
		r = lipgloss.NewRenderer(io.Discard, termenv.WithProfile(termenv.Ascii))
	}

	muted := lipgloss.AdaptiveColor{Light: "#9ca3af", Dark: "#6b7280"}
	accent := lipgloss.AdaptiveColor{Light: "#0d9488", Dark: "#2dd4bf"}
	errc := lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#ef4444"}
	ok := lipgloss.AdaptiveColor{Light: "#15803d", Dark: "#22c55e"}
	subtle := lipgloss.AdaptiveColor{Light: "#e5e7eb", Dark: "#374151"}

	return &Styles{
		asciiGlyphs: opts.AsciiGlyphs,

		Base:   r.NewStyle(),
		Muted:  r.NewStyle().Foreground(muted),
		Accent: r.NewStyle().Foreground(accent),
		Err:    r.NewStyle().Foreground(errc),

		CursorBar:   r.NewStyle().Foreground(accent).Bold(true),
		CursorRowBG: r.NewStyle().Background(subtle),

		PaneBorder:        r.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(muted),
		PaneBorderFocused: r.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(accent),

		ChipMuted:  r.NewStyle().Foreground(muted),
		ChipActive: r.NewStyle().Foreground(accent).Bold(true),
		ModePill:   r.NewStyle().Foreground(lipgloss.Color("0")).Background(accent).Bold(true).Padding(0, 1),

		TitleBacklog:   r.NewStyle().Italic(true).Foreground(muted),
		TitlePending:   r.NewStyle(),
		TitleActive:    r.NewStyle().Bold(true),
		TitleBlocked:   r.NewStyle(),
		TitleDone:      r.NewStyle().Strikethrough(true).Foreground(muted),
		TitleCancelled: r.NewStyle().Strikethrough(true).Foreground(muted),

		GlyphBacklog:   r.NewStyle().Foreground(muted),
		GlyphPending:   r.NewStyle(),
		GlyphActive:    r.NewStyle().Foreground(ok).Bold(true),
		GlyphBlocked:   r.NewStyle().Foreground(errc).Bold(true),
		GlyphDone:      r.NewStyle().Foreground(ok),
		GlyphCancelled: r.NewStyle().Foreground(errc).Faint(true),
	}
}

// StatusGlyph returns the glyph for s, using ASCII when AsciiGlyphs is set.
func (s *Styles) StatusGlyph(st model.Status) string {
	if s.asciiGlyphs {
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
