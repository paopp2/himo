package store

import (
	"bytes"

	"github.com/npaolopepito/himo/internal/model"
)

// RenderTaskLine returns the column-0 line for a task as it should appear in
// active.md/done.md (with marker) or backlog.md (without).
func RenderTaskLine(t model.Task) string {
	if t.Status == model.StatusBacklog {
		return "- " + t.Title
	}
	return "- " + t.Status.Marker() + " " + t.Title
}

// Render serializes a Document back to markdown bytes. It round-trips any
// Document produced by the parsers byte-for-byte (for unmodified input).
func Render(doc *Document) []byte {
	var buf bytes.Buffer
	for _, item := range doc.Items {
		switch it := item.(type) {
		case TaskItem:
			for _, line := range it.RawLines {
				buf.WriteString(line)
				buf.WriteByte('\n')
			}
		case DateHeading:
			buf.WriteString("## " + it.Date)
			buf.WriteByte('\n')
		case ProjectHeading:
			buf.WriteString(it.RawLine)
			buf.WriteByte('\n')
			buf.WriteByte('\n') // one blank line after, by convention.
		case OpaqueLines:
			for _, line := range it.Lines {
				buf.WriteString(line)
				buf.WriteByte('\n')
			}
		}
	}
	return buf.Bytes()
}
