package tui

import "github.com/paopp2/himo/internal/model"

// Sort selects the row order in the task list.
type Sort int

const (
	// SortNatural preserves the document iteration order
	// (per-project, then active.md -> backlog.md -> done.md, then source line).
	SortNatural Sort = iota
	// SortStatus groups rows by status using statusSortRank, then by
	// project order, then by original position within the natural order.
	SortStatus
)

// SortFromName is the inverse of sortName. Unknown names fall back to
// SortNatural so missing or stale state files do not change behavior.
func SortFromName(name string) Sort {
	if name == "status" {
		return SortStatus
	}
	return SortNatural
}

// sortName returns the stable string used by state persistence.
func sortName(s Sort) string {
	if s == SortStatus {
		return "status"
	}
	return "natural"
}

// statusSortRank ranks statuses for SortStatus. Lower rank renders first.
// Order is attention-first: active and blocked surface above pending,
// backlog, done, and cancelled.
func statusSortRank(s model.Status) int {
	switch s {
	case model.StatusActive:
		return 0
	case model.StatusBlocked:
		return 1
	case model.StatusPending:
		return 2
	case model.StatusBacklog:
		return 3
	case model.StatusDone:
		return 4
	case model.StatusCancelled:
		return 5
	}
	return 6
}
