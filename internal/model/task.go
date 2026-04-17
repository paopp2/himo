package model

// Task is a single todo item. It carries enough information to round-trip
// back to its source line without a stable ID. Identity during normalization
// is positional (source file + original line).
type Task struct {
	Status Status
	Title  string
	Notes  string // indented content verbatim, with the 4-space indent preserved
	Date   string // "YYYY-MM-DD", only set for done/cancelled in done.md
}

// HasNotes reports whether the task has any non-empty notes content.
func (t Task) HasNotes() bool {
	for _, r := range t.Notes {
		if r != ' ' && r != '\t' && r != '\n' {
			return true
		}
	}
	return false
}
