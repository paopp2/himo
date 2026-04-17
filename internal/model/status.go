package model

type Status int

const (
	StatusPending Status = iota
	StatusActive
	StatusBlocked
	StatusBacklog
	StatusDone
	StatusCancelled
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusActive:
		return "active"
	case StatusBlocked:
		return "blocked"
	case StatusBacklog:
		return "backlog"
	case StatusDone:
		return "done"
	case StatusCancelled:
		return "cancelled"
	}
	return "unknown"
}

func (s Status) Marker() string {
	switch s {
	case StatusPending:
		return "[ ]"
	case StatusActive:
		return "[/]"
	case StatusBlocked:
		return "[!]"
	case StatusDone:
		return "[x]"
	case StatusCancelled:
		return "[-]"
	}
	return ""
}

func ParseMarker(s string) (Status, bool) {
	switch s {
	case "[ ]":
		return StatusPending, true
	case "[/]":
		return StatusActive, true
	case "[!]":
		return StatusBlocked, true
	case "[x]", "[X]":
		return StatusDone, true
	case "[-]":
		return StatusCancelled, true
	}
	return 0, false
}
