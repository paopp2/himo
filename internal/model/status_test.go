package model

import "testing"

func TestTaskURL(t *testing.T) {
	tests := []struct {
		name  string
		notes string
		want  string
	}{
		{"empty notes", "", ""},
		{"no URL", "    just some notes", ""},
		{"bare https", "    https://github.com/paopp2/himo/issues/1", "https://github.com/paopp2/himo/issues/1"},
		{"bare http", "    http://localhost:3000/debug", "http://localhost:3000/debug"},
		{"URL mid-line", "    See https://example.com/foo for details", "https://example.com/foo"},
		{"multiple URLs first wins", "    https://first.com\n    https://second.com", "https://first.com"},
		{"URL with path and query", "    https://example.com/path?q=1&r=2#frag", "https://example.com/path?q=1&r=2#frag"},
		{"URL in parens", "    (https://example.com/foo)", "https://example.com/foo"},
		{"URL in brackets", "    [https://example.com/foo]", "https://example.com/foo"},
		{"URL followed by angle bracket", "    <https://example.com/foo>", "https://example.com/foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{Notes: tt.notes}
			if got := task.URL(); got != tt.want {
				t.Errorf("URL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStatusMarker(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusPending, "[ ]"},
		{StatusActive, "[/]"},
		{StatusBlocked, "[!]"},
		{StatusDone, "[x]"},
		{StatusCancelled, "[-]"},
		{StatusBacklog, ""},
	}
	for _, tt := range tests {
		if got := tt.status.Marker(); got != tt.want {
			t.Errorf("%v.Marker() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestParseMarker(t *testing.T) {
	tests := []struct {
		in   string
		want Status
		ok   bool
	}{
		{"[ ]", StatusPending, true},
		{"[/]", StatusActive, true},
		{"[!]", StatusBlocked, true},
		{"[x]", StatusDone, true},
		{"[X]", StatusDone, true},
		{"[-]", StatusCancelled, true},
		{"[~]", 0, false},
		{"", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseMarker(tt.in)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("ParseMarker(%q) = (%v,%v), want (%v,%v)", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}
