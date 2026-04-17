package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// NewProject creates a new project directory under baseDir with an empty
// active.md.
func NewProject(baseDir, name string) error {
	if name == "" || strings.ContainsAny(name, "/\\") || strings.HasPrefix(name, ".") {
		return fmt.Errorf("invalid project name: %q", name)
	}
	dir := filepath.Join(baseDir, name)
	if err := os.Mkdir(dir, 0o755); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("project %q already exists", name)
		}
		return err
	}
	return os.WriteFile(filepath.Join(dir, "active.md"), nil, 0o644)
}
