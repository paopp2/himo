package store

import "github.com/npaolopepito/himo/internal/model"

// Document is the parsed form of a single markdown file. It holds items in
// the exact order they appear, so writing back produces a byte-compatible
// round-trip for unchanged input.
type Document struct {
	Items []Item
}

// Item is one of TaskItem, DateHeading, OpaqueLines.
type Item interface{ isItem() }

// TaskItem is a column-0 task line plus its indented notes block.
type TaskItem struct {
	Task     model.Task
	RawLines []string // original lines (task line + notes), for round-trip
}

func (TaskItem) isItem() {}

// DateHeading is a "# YYYY-MM-DD" line (done.md only).
type DateHeading struct {
	Date    string
	RawLine string
}

func (DateHeading) isItem() {}

// OpaqueLines preserves lines we don't recognize (blank lines, stray text,
// non-date headings, fenced code). Adjacent opaque lines are grouped.
type OpaqueLines struct {
	Lines []string
}

func (OpaqueLines) isItem() {}

// Tasks returns a read-only snapshot of the task items in document order.
// Mutations to the returned slice do not affect the Document.
func (d *Document) Tasks() []model.Task {
	out := make([]model.Task, 0, len(d.Items))
	for _, it := range d.Items {
		if ti, ok := it.(TaskItem); ok {
			out = append(out, ti.Task)
		}
	}
	return out
}
