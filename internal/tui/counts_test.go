package tui

import (
	"testing"

	"github.com/npaolopepito/himo/internal/model"
)

func TestStatusCounts(t *testing.T) {
	m := NewModel(testProject(t)) // active: 1 pending, 1 active
	c := m.statusCounts()
	if c[model.StatusPending] != 1 {
		t.Errorf("pending = %d, want 1", c[model.StatusPending])
	}
	if c[model.StatusActive] != 1 {
		t.Errorf("active = %d, want 1", c[model.StatusActive])
	}
	if c[model.StatusDone] != 0 {
		t.Errorf("done = %d, want 0", c[model.StatusDone])
	}
}
