// Package state persists the user's last session (project + filter) so
// the next launch can pick up where they left off.
package state

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type State struct {
	LastProject     string `toml:"last_project"`
	LastFilter      string `toml:"last_filter"`
	LastAllProjects bool   `toml:"last_all_projects"`
}

// Path resolves to $XDG_STATE_HOME/himo/state.toml, defaulting to
// ~/.local/state/himo/state.toml.
func Path() (string, error) {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home: %w", err)
		}
		dir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(dir, "himo", "state.toml"), nil
}

// Load reads the state file. A missing file is not an error; the zero
// State is returned.
func Load() (*State, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return &State{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	var s State
	if err := toml.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	return &s, nil
}

func Save(s *State) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("mkdir state: %w", err)
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(s); err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	return os.WriteFile(p, buf.Bytes(), 0o644)
}
