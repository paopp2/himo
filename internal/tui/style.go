package tui

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/paopp2/himo/internal/model"
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

	// ModePill* are per-mode pill styles. The pill base is shared (bold,
	// black foreground, padding); only the background changes per mode.
	ModePillNormal lipgloss.Style
	ModePillInsert lipgloss.Style
	ModePillSearch lipgloss.Style
	ModePillDelete lipgloss.Style
	ModePillPicker lipgloss.Style
	ModePillHelp   lipgloss.Style

	SearchHighlight lipgloss.Style

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
	accent := lipgloss.AdaptiveColor{Light: "#008787", Dark: "#008787"}
	errc := lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#ef4444"}
	ok := lipgloss.AdaptiveColor{Light: "#15803d", Dark: "#22c55e"}
	// Pill-only accents -- not promoted to top-level palette colors because
	// they're scoped to the mode pill and shouldn't drift into chips/titles.
	pillBlue := lipgloss.AdaptiveColor{Light: "#2563eb", Dark: "#3b82f6"}
	pillPurple := lipgloss.AdaptiveColor{Light: "#7c3aed", Dark: "#a855f7"}
	// Cursor row tint: a dark/light teal washed toward the pane background,
	// roughly the accent (#008787) blended ~15% over black/white. Subtle but
	// reads as part of the project palette instead of neutral gray.
	subtle := lipgloss.AdaptiveColor{Light: "#daeded", Dark: "#0f2222"}

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

		ModePillNormal: pillBase(r).Background(accent),
		ModePillInsert: pillBase(r).Background(ok),
		ModePillSearch: pillBase(r).Background(pillBlue),
		ModePillDelete: pillBase(r).Background(errc),
		ModePillPicker: pillBase(r).Background(pillPurple),
		ModePillHelp:   pillBase(r).Background(muted),

		SearchHighlight: r.NewStyle().Foreground(lipgloss.Color("0")).Background(accent),

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

// pillBase is the shared mode-pill style. Per-mode pills only override
// the background.
func pillBase(r *lipgloss.Renderer) lipgloss.Style {
	return r.NewStyle().Foreground(lipgloss.Color("0")).Bold(true).Padding(0, 1)
}

// ModePillStyle returns the pill style for the given mode. ModePrompt
// and ModeEdit share the INSERT pill since they render the same label.
func (s *Styles) ModePillStyle(m Mode) lipgloss.Style {
	switch m {
	case ModeNormal:
		return s.ModePillNormal
	case ModePrompt, ModeEdit:
		return s.ModePillInsert
	case ModeSearch:
		return s.ModePillSearch
	case ModeDelete:
		return s.ModePillDelete
	case ModePicker:
		return s.ModePillPicker
	case ModeHelp:
		return s.ModePillHelp
	}
	return s.ModePillNormal
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

// PaintCursorRow paints the cursor-row background across an already-styled
// row. Lipgloss' Render wraps content with SGR-open + reset-close, but any
// ANSI reset inside the content (from nested chip/glyph/title styles) also
// clears the outer background, so the highlight goes dark mid-row. Re-emit
// the background open-code after every reset to keep it lit edge to edge.
func (s *Styles) PaintCursorRow(row string) string {
	marker := s.CursorRowBG.Render("\x00")
	parts := strings.SplitN(marker, "\x00", 2)
	if len(parts) != 2 || parts[0] == "" {
		// ASCII / no-color renderer: nothing to paint.
		return row
	}
	open, closer := parts[0], parts[1]
	return open + strings.ReplaceAll(row, closer, closer+open) + closer
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
