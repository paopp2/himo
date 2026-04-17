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
	task := model.Task{Status: model.StatusPending, Title: title}
	ti := store.TaskItem{
		Task:     task,
		RawLines: []string{store.RenderTaskLine(task)},
	}
	p.Active.Items = append(p.Active.Items, ti)
	return store.SaveProject(p)
}
