package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/paopp2/himo/internal/model"
)

func TestStatusGlyph_unicode(t *testing.T) {
	cases := []struct {
		s    model.Status
		want string
	}{
		{model.StatusBacklog, "○"},
		{model.StatusPending, "○"},
		{model.StatusActive, "●"},
		{model.StatusBlocked, "●"},
		{model.StatusDone, "✓"},
		{model.StatusCancelled, "✗"},
	}
	st := NewStyles(StyleOptions{})
	for _, tc := range cases {
		got := st.StatusGlyph(tc.s)
		if got != tc.want {
			t.Errorf("StatusGlyph(%v) = %q, want %q", tc.s, got, tc.want)
		}
	}
}

func TestStatusGlyph_ascii(t *testing.T) {
	cases := []struct {
		s    model.Status
		want string
	}{
		{model.StatusBacklog, "o"},
		{model.StatusPending, "o"},
		{model.StatusActive, "*"},
		{model.StatusBlocked, "*"},
		{model.StatusDone, "x"},
		{model.StatusCancelled, "-"},
	}
	st := NewStyles(StyleOptions{AsciiGlyphs: true})
	for _, tc := range cases {
		got := st.StatusGlyph(tc.s)
		if got != tc.want {
			t.Errorf("StatusGlyph(%v) = %q, want %q", tc.s, got, tc.want)
		}
	}
}

func TestStyles_noColor_rendersPlainText(t *testing.T) {
	st := NewStyles(StyleOptions{NoColor: true})
	out := st.Accent.Render("hello")
	if strings.Contains(out, "\x1b[38") || strings.Contains(out, "\x1b[48") {
		t.Errorf("no_color output contained color escape: %q", out)
	}
}

func TestStyles_trueColor_emitsColorEscape(t *testing.T) {
	r := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.TrueColor))
	r.SetColorProfile(termenv.TrueColor)
	st := NewStylesWithRenderer(r, StyleOptions{})
	out := st.Accent.Render("hello")
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("true-color accent render produced no escape: %q", out)
	}
}

func TestModePillStyle_distinctPerMode(t *testing.T) {
	r := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.TrueColor))
	r.SetColorProfile(termenv.TrueColor)
	st := NewStylesWithRenderer(r, StyleOptions{})
	modes := []Mode{ModeNormal, ModePrompt, ModeSearch, ModeDelete, ModePicker, ModeHelp}
	seen := make(map[string]Mode, len(modes))
	for _, m := range modes {
		out := st.ModePillStyle(m).Render(m.String())
		if other, dup := seen[out]; dup {
			t.Errorf("mode %v shares pill rendering with %v: %q", m, other, out)
		}
		seen[out] = m
	}
}

func TestModePillStyle_promptAndEditShareInsertColor(t *testing.T) {
	r := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.TrueColor))
	r.SetColorProfile(termenv.TrueColor)
	st := NewStylesWithRenderer(r, StyleOptions{})
	prompt := st.ModePillStyle(ModePrompt).Render("INSERT")
	edit := st.ModePillStyle(ModeEdit).Render("INSERT")
	if prompt != edit {
		t.Errorf("ModePrompt and ModeEdit must share INSERT styling:\nprompt=%q\nedit  =%q", prompt, edit)
	}
}

// TestStyledInput_caretRendersAsReverseBlock pins the visible-caret invariant:
// after Focus, the textinput's cursor cell must emit reverse-video so the
// accent foreground reads as a background block. ASCII-profile golden tests
// can't catch this because both visible and invisible states render as a bare
// space; this exercises the TrueColor path where the difference shows up.
func TestStyledInput_caretRendersAsReverseBlock(t *testing.T) {
	r := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.TrueColor))
	r.SetColorProfile(termenv.TrueColor)
	st := NewStylesWithRenderer(r, StyleOptions{})
	ti := newStyledInput(st)
	ti.SetValue("x")
	ti.CursorEnd()
	ti.Focus()
	out := ti.View()
	if !strings.Contains(out, "\x1b[7m") && !strings.Contains(out, "\x1b[7;") {
		t.Errorf("focused caret missing reverse-video escape: %q", out)
	}
}
