package model

import "regexp"

// Task is a single todo item. It carries enough information to round-trip
// back to its source line without a stable ID. Identity during normalization
// is positional (source file + original line).
type Task struct {
	Status Status
	Title  string
	Notes  string // indented content verbatim, with the 4-space indent preserved
	Date   string // "YYYY-MM-DD", only set for done/cancelled in done.md
}

var urlRe = regexp.MustCompile(`https?://[^\s)>\]]+`)

// HasNotes reports whether the task has any non-empty notes content.
func (t Task) HasNotes() bool {
	for _, r := range t.Notes {
		if r != ' ' && r != '\t' && r != '\n' {
			return true
		}
	}
	return false
}

// URL returns the first HTTP(S) URL found in the task's notes, or "".
func (t Task) URL() string {
	return urlRe.FindString(t.Notes)
}
