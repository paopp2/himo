package store

import "bytes"

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
			buf.WriteString(it.RawLine)
			buf.WriteByte('\n')
		case OpaqueLines:
			for _, line := range it.Lines {
				buf.WriteString(line)
				buf.WriteByte('\n')
			}
		}
	}
	return buf.Bytes()
}
