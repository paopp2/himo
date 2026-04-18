package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	want := &State{LastProject: "work", LastFilter: "active", LastAllProjects: true}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if *got != *want {
		t.Errorf("round-trip: got %+v, want %+v", got, want)
	}

	path := filepath.Join(dir, "himo", "state.toml")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("state file missing at %s: %v", path, err)
	}
}

func TestLoad_missingFileReturnsZero(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.LastProject != "" || got.LastFilter != "" {
		t.Errorf("expected zero State, got %+v", got)
	}
}
