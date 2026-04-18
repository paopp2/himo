package tui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/paopp2/himo/internal/model"
)

func TestFileForFilter(t *testing.T) {
	cases := []struct {
		name    string
		filter  Filter
		wantErr bool
		wantEnd string
	}{
		{"default", DefaultFilter(), false, "active.md"},
		{"backlog", Filter{Statuses: []model.Status{model.StatusBacklog}}, false, "backlog.md"},
		{"pending", Filter{Statuses: []model.Status{model.StatusPending}}, false, "active.md"},
		{"active", Filter{Statuses: []model.Status{model.StatusActive}}, false, "active.md"},
		{"blocked", Filter{Statuses: []model.Status{model.StatusBlocked}}, false, "active.md"},
		{"done", Filter{Statuses: []model.Status{model.StatusDone}}, false, "done.md"},
		{"cancelled", Filter{Statuses: []model.Status{model.StatusCancelled}}, false, "done.md"},
		{"all", Filter{All: true}, true, ""},
		{"ambiguous", Filter{Statuses: []model.Status{model.StatusActive, model.StatusDone}}, true, ""},
	}
	for _, tc := range cases {
		proj := testProject(t)
		got, err := fileForFilter(tc.filter, proj)
		if tc.wantErr {
			if err == nil {
				t.Errorf("%s: want error, got %q", tc.name, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}
		if !strings.HasSuffix(got, string(filepath.Separator)+tc.wantEnd) {
			t.Errorf("%s: path = %q, want suffix /%s", tc.name, got, tc.wantEnd)
		}
	}
}
