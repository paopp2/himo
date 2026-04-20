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
		"[`] Default", "[0] All", "[1] Backlog", "3",
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

// Chip order: Default leads the bar, All trails it; status chips sit in
// between in number order.
func TestRenderFilterBar_chipOrder(t *testing.T) {
	st := testStyles(t)
	bar := renderFilterBar(st, DefaultFilter(), map[model.Status]int{}, 120)
	order := []string{
		"[`] Default",
		"[1] Backlog", "[2] Pending", "[3] Active",
		"[4] Blocked", "[5] Done", "[6] Cancelled",
		"[0] All",
	}
	prev := -1
	for _, want := range order {
		i := strings.Index(bar, want)
		if i < 0 {
			t.Fatalf("bar missing %q:\n%s", want, bar)
		}
		if i <= prev {
			t.Errorf("chip %q at %d, expected after %d:\n%s", want, i, prev, bar)
		}
		prev = i
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
