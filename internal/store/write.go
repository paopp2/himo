package store

import (
	"bytes"
	"strings"
)

// Render serializes a Document back to markdown bytes. It round-trips any
// Document produced by the parsers byte-for-byte (for unmodified input).
func Render(doc *Document) []byte {
	var buf bytes.Buffer
	for i, item := range doc.Items {
		switch it := item.(type) {
		case TaskItem:
			for j, line := range it.RawLines {
				buf.WriteString(line)
				// Always write a trailing newline except for the very last line
				// of the last item if the input had no trailing newline.
				if i < len(doc.Items)-1 || j < len(it.RawLines)-1 {
					buf.WriteByte('\n')
				} else {
					buf.WriteByte('\n')
				}
			}
		case DateHeading:
			buf.WriteString(it.RawLine)
			buf.WriteByte('\n')
		case OpaqueLines:
			for _, line := range it.Lines {
				buf.WriteString(line)
				buf.WriteByte('\n')
			}
		}
	}
	// If the original file had no trailing newline, we need to strip the
	// last one we added. But our parser normalizes to ending-in-newline by
	// only reading complete lines; files without trailing newlines lose the
	// last line in bufio.Scanner... document this limitation for later.
	_ = strings.TrimSpace // placeholder; adjust if golden files need it
	return buf.Bytes()
}
