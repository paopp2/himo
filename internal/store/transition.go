package store

import "github.com/npaolopepito/himo/internal/model"

// FileName identifies which of the three per-project files a task belongs in.
type FileName int

const (
	FileActive FileName = iota
	FileBacklog
	FileDone
)

func (f FileName) String() string {
	switch f {
	case FileActive:
		return "active.md"
	case FileBacklog:
		return "backlog.md"
	case FileDone:
		return "done.md"
	}
	return "unknown"
}

// TargetFile returns the file a task with the given status must live in.
func TargetFile(s model.Status) FileName {
	switch s {
	case model.StatusPending, model.StatusActive, model.StatusBlocked:
		return FileActive
	case model.StatusBacklog:
		return FileBacklog
	case model.StatusDone, model.StatusCancelled:
		return FileDone
	}
	return FileActive
}
