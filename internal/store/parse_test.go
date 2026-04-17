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

func TestParseActive_notes(t *testing.T) {
	b, err := os.ReadFile("testdata/active_with_notes.md")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ParseActive(b)
	if err != nil {
		t.Fatalf("ParseActive: %v", err)
	}
	tasks := doc.Tasks()
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(tasks))
	}

	wantNotes := []string{
		"    Due Friday. Talk to Sam first.\n\n    - Storage model concerns\n    - Parser round-tripping",
		"    Check fridge.",
		"",
	}
	for i, want := range wantNotes {
		if tasks[i].Notes != want {
			t.Errorf("task %d notes:\n got: %q\nwant: %q", i, tasks[i].Notes, want)
		}
	}
	if !tasks[0].HasNotes() {
		t.Errorf("task 0 HasNotes = false, want true")
	}
	if tasks[2].HasNotes() {
		t.Errorf("task 2 HasNotes = true, want false")
	}
}

func TestParseBacklog_basic(t *testing.T) {
	b, err := os.ReadFile("testdata/backlog_basic.md")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ParseBacklog(b)
	if err != nil {
		t.Fatalf("ParseBacklog: %v", err)
	}
	tasks := doc.Tasks()
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(tasks))
	}
	for i, task := range tasks {
		if task.Status != model.StatusBacklog {
			t.Errorf("task %d status = %v, want backlog", i, task.Status)
		}
	}
	wantTitles := []string{"Refactor the auth module", "Explore alternative color schemes", "Revive the plugin idea"}
	for i, want := range wantTitles {
		if tasks[i].Title != want {
			t.Errorf("task %d title = %q, want %q", i, tasks[i].Title, want)
		}
	}
}

func TestParseDone_basic(t *testing.T) {
	b, err := os.ReadFile("testdata/done_basic.md")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ParseDone(b)
	if err != nil {
		t.Fatalf("ParseDone: %v", err)
	}
	tasks := doc.Tasks()
	if len(tasks) != 4 {
		t.Fatalf("got %d tasks, want 4", len(tasks))
	}
	want := []struct {
		status model.Status
		title  string
		date   string
	}{
		{model.StatusDone, "File expenses", "2026-04-18"},
		{model.StatusCancelled, "Revamp onboarding slides", "2026-04-18"},
		{model.StatusDone, "Ship RFC", "2026-04-17"},
		{model.StatusDone, "Review PRs", "2026-04-17"},
	}
	for i, w := range want {
		if tasks[i].Status != w.status || tasks[i].Title != w.title || tasks[i].Date != w.date {
			t.Errorf("task %d: got (%v,%q,%q), want (%v,%q,%q)",
				i, tasks[i].Status, tasks[i].Title, tasks[i].Date,
				w.status, w.title, w.date)
		}
	}
}
