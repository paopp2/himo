package store

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"

	"github.com/npaolopepito/himo/internal/model"
)

var (
	// Column-0 task line: "- [m] title" where m is one of our markers.
	// Captures: 1=marker (with brackets), 2=title.
	taskLineRe = regexp.MustCompile(`^- (\[[ /!xX\-]\]) (.+)$`)

	// Backlog line: "- title" at column 0, with no checkbox.
	// Captures: 1=title. Must not match task lines (order matters).
	backlogLineRe = regexp.MustCompile(`^- (.+)$`)

	// Date heading: "# YYYY-MM-DD".
	dateHeadingRe = regexp.MustCompile(`^# (\d{4}-\d{2}-\d{2})\s*$`)
)

// ParseActive parses the contents of active.md.
func ParseActive(b []byte) (*Document, error) {
	return parseDoc(b, parseActiveLine)
}

// lineParser classifies a single line as a task, date heading, or opaque.
type lineParser func(line string) (Item, bool)

func parseDoc(b []byte, lp lineParser) (*Document, error) {
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Buffer(nil, 1024*1024)

	var items []Item
	var opaqueBuf []string
	var currentTask *TaskItem

	flushOpaque := func() {
		if len(opaqueBuf) > 0 {
			items = append(items, OpaqueLines{Lines: append([]string(nil), opaqueBuf...)})
			opaqueBuf = opaqueBuf[:0]
		}
	}
	flushTask := func() {
		if currentTask != nil {
			items = append(items, *currentTask)
			currentTask = nil
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Indented content: belongs to the current task as notes, otherwise opaque.
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') && currentTask != nil {
			currentTask.RawLines = append(currentTask.RawLines, line)
			continue
		}

		// Blank line: might belong to current task's notes (loose list item rule)
		// if followed by more indented content. For simplicity we attach it to the
		// current task tentatively; if the next line is non-indented, we flush.
		if line == "" && currentTask != nil {
			currentTask.RawLines = append(currentTask.RawLines, line)
			continue
		}

		// Non-indented, non-blank: the current task ends here.
		flushTask()

		if item, ok := lp(line); ok {
			flushOpaque()
			if ti, isTask := item.(TaskItem); isTask {
				copy := ti
				currentTask = &copy
				currentTask.RawLines = []string{line}
				continue
			}
			items = append(items, item)
			continue
		}

		opaqueBuf = append(opaqueBuf, line)
	}
	flushTask()
	flushOpaque()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Trim trailing blank notes lines from each task (attached tentatively).
	for i := range items {
		if ti, ok := items[i].(TaskItem); ok {
			if len(ti.RawLines) > 1 {
				ti.Task.Notes = strings.TrimRight(strings.Join(ti.RawLines[1:], "\n"), "\n")
			}
			items[i] = ti
		}
	}

	return &Document{Items: items}, nil
}

func parseActiveLine(line string) (Item, bool) {
	if m := taskLineRe.FindStringSubmatch(line); m != nil {
		status, ok := model.ParseMarker(m[1])
		if !ok {
			return nil, false
		}
		return TaskItem{Task: model.Task{Status: status, Title: m[2]}}, true
	}
	return nil, false
}

// ParseBacklog parses the contents of backlog.md.
func ParseBacklog(b []byte) (*Document, error) {
	return parseDoc(b, parseBacklogLine)
}

func parseBacklogLine(line string) (Item, bool) {
	// A backlog line looks like "- title" at column 0.
	// Must NOT match a task line (which starts with "- [").
	if strings.HasPrefix(line, "- [") {
		return nil, false
	}
	if m := backlogLineRe.FindStringSubmatch(line); m != nil {
		return TaskItem{Task: model.Task{Status: model.StatusBacklog, Title: m[1]}}, true
	}
	return nil, false
}

// ParseDone parses the contents of done.md. Task Date fields are set to the
// current date heading the task appears under (empty if outside any heading).
func ParseDone(b []byte) (*Document, error) {
	doc, err := parseDoc(b, parseDoneLine)
	if err != nil {
		return nil, err
	}
	// Walk items and propagate the current date heading to each TaskItem.
	currentDate := ""
	for i := range doc.Items {
		switch it := doc.Items[i].(type) {
		case DateHeading:
			currentDate = it.Date
		case TaskItem:
			it.Task.Date = currentDate
			doc.Items[i] = it
		}
	}
	return doc, nil
}

func parseDoneLine(line string) (Item, bool) {
	if m := dateHeadingRe.FindStringSubmatch(line); m != nil {
		return DateHeading{Date: m[1], RawLine: line}, true
	}
	return parseActiveLine(line)
}
