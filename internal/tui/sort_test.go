package tui

import (
	"testing"

	"github.com/paopp2/himo/internal/model"
)

func TestSortFromName_roundTrip(t *testing.T) {
	cases := []struct {
		name string
		want Sort
	}{
		{"natural", SortNatural},
		{"status", SortStatus},
		{"", SortNatural},
		{"bogus", SortNatural},
	}
	for _, tt := range cases {
		got := SortFromName(tt.name)
		if got != tt.want {
			t.Errorf("SortFromName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
	if got := sortName(SortNatural); got != "natural" {
		t.Errorf("sortName(SortNatural) = %q, want %q", got, "natural")
	}
	if got := sortName(SortStatus); got != "status" {
		t.Errorf("sortName(SortStatus) = %q, want %q", got, "status")
	}
}

func TestStatusSortRank_attentionFirst(t *testing.T) {
	want := []model.Status{
		model.StatusActive,
		model.StatusBlocked,
		model.StatusPending,
		model.StatusBacklog,
		model.StatusDone,
		model.StatusCancelled,
	}
	for i, s := range want {
		if got := statusSortRank(s); got != i {
			t.Errorf("statusSortRank(%v) = %d, want %d", s, got, i)
		}
	}
}
