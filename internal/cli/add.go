package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
)

// AddTask appends a pending task to the project's active.md.
func AddTask(baseDir, project, title string) error {
	dir := filepath.Join(baseDir, project)
	if _, err := os.Stat(filepath.Join(dir, "active.md")); err != nil {
		return fmt.Errorf("project %q not found", project)
	}
	p, err := store.LoadProject(dir)
	if err != nil {
		return err
	}
	ti := store.TaskItem{
		Task:     model.Task{Status: model.StatusPending, Title: title},
		RawLines: []string{fmt.Sprintf("- [ ] %s", title)},
	}
	p.Active.Items = append(p.Active.Items, ti)
	return store.SaveProject(p)
}
