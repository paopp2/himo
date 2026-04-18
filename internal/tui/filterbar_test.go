package tui

import (
	"strings"
	"testing"

	"github.com/paopp2/himo/internal/model"
)

func TestRenderFilterBar_showsAllChipsWithCounts(t *testing.T) {
	st := testStyles(t)
	counts := map[model.Status]int{
		model.StatusBacklog:   3,
		model.StatusPending:   5,
		model.StatusActive:    1,
		model.StatusBlocked:   0,
		model.StatusDone:      12,
		model.StatusCancelled: 2,
	}
	bar := renderFilterBar(st, DefaultFilter(), counts, 120)
	for _, want := range []string{
		"[0] All", "[1] Backlog", "3",
		"[2] Pending", "5",
		"[3] Active", "1",
		"[4] Blocked", "0",
		"[5] Done", "12",
		"[6] Cancelled", "2",
	} {
		if !strings.Contains(bar, want) {
			t.Errorf("bar missing %q:\n%s", want, bar)
		}
	}
}

func TestRenderFilterBar_activeStatusIsHighlighted(t *testing.T) {
	st := testStyles(t)
	counts := map[model.Status]int{}
	f := Filter{Statuses: []model.Status{model.StatusBacklog}}
	bar := renderFilterBar(st, f, counts, 120)
	// In Ascii profile this is content-only; we check the surrounding
	// structure in a later golden test. For now: no error, non-empty.
	if bar == "" {
		t.Fatal("filter bar empty")
	}
}
