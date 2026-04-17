package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/npaolopepito/himo/internal/config"
)

// Init runs the first-run interactive setup: prompt for base_dir, create it,
// create the config file. Safe to run repeatedly (re-prompts and overwrites).
func Init(in io.Reader, out io.Writer) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home dir: %w", err)
	}
	defaultBase := filepath.Join(home, "todos")

	fmt.Fprintf(out, "Where should todos live? [%s]: ", defaultBase)
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		line = defaultBase
	}
	if strings.HasPrefix(line, "~/") {
		line = filepath.Join(home, line[2:])
	}

	if err := os.MkdirAll(line, 0o755); err != nil {
		return fmt.Errorf("create base dir: %w", err)
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}
	cfg := &config.Config{
		Path:        cfgPath,
		BaseDir:     line,
		PreviewPane: true,
	}
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Fprintf(out, "Wrote %s\n", cfgPath)
	fmt.Fprintf(out, "Base dir: %s\n", line)
	return nil
}
