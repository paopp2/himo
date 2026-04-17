package store

import (
	"os"
	"testing"

	"github.com/npaolopepito/himo/internal/model"
)

func TestParseActive_basic(t *testing.T) {
	b, err := os.ReadFile("testdata/active_basic.md")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ParseActive(b)
	if err != nil {
		t.Fatalf("ParseActive: %v", err)
	}
	tasks := doc.Tasks()
	if len(tasks) != 4 {
		t.Fatalf("got %d tasks, want 4", len(tasks))
	}
	want := []struct {
		status model.Status
		title  string
	}{
		{model.StatusActive, "Write design doc"},
		{model.StatusPending, "Buy groceries"},
		{model.StatusBlocked, "Migrate payments table"},
		{model.StatusPending, "Reply to Alex"},
	}
	for i, w := range want {
		if tasks[i].Status != w.status || tasks[i].Title != w.title {
			t.Errorf("task %d: got (%v,%q), want (%v,%q)",
				i, tasks[i].Status, tasks[i].Title, w.status, w.title)
		}
	}
}
