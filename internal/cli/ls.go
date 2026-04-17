package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
)

// Ls prints tasks in a project (or all projects if project == "") filtered
// by status (or all statuses if status == "").
func Ls(baseDir, project, status string, out io.Writer) error {
	var wanted model.Status
	filtered := false
	if status != "" {
		s, ok := parseStatus(status)
		if !ok {
			return fmt.Errorf("unknown status: %q", status)
		}
		wanted = s
		filtered = true
	}

	projects := []string{project}
	if project == "" {
		names, err := store.ListProjects(baseDir)
		if err != nil {
			return err
		}
		projects = names
	}

	for _, name := range projects {
		p, err := store.LoadProject(filepath.Join(baseDir, name))
		if err != nil {
			return err
		}
		for _, task := range p.AllTasks() {
			if filtered && task.Status != wanted {
				continue
			}
			marker := task.Status.Marker()
			if marker == "" {
				marker = "[ ]"
			}
			fmt.Fprintf(out, "%s  %-8s  %s  %s\n",
				marker, task.Status, name, task.Title)
		}
	}
	return nil
}

func parseStatus(s string) (model.Status, bool) {
	switch s {
	case "pending":
		return model.StatusPending, true
	case "active":
		return model.StatusActive, true
	case "blocked":
		return model.StatusBlocked, true
	case "backlog":
		return model.StatusBacklog, true
	case "done":
		return model.StatusDone, true
	case "cancelled":
		return model.StatusCancelled, true
	}
	return 0, false
}
