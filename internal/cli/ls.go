package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/paopp2/himo/internal/model"
	"github.com/paopp2/himo/internal/store"
)

// Ls prints tasks in a project (or all projects if project == "") filtered
// by status (or all statuses if status == "").
func Ls(baseDir, project, status string, out io.Writer) error {
	var wanted model.Status
	filtered := false
	if status != "" {
		s, ok := model.ParseStatusName(status)
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
