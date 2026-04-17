package model

import "testing"

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
