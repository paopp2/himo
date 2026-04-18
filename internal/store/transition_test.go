package store

import (
	"testing"

	"github.com/paopp2/himo/internal/model"
)

func TestTargetFile(t *testing.T) {
	tests := []struct {
		status model.Status
		want   FileName
	}{
		{model.StatusPending, FileActive},
		{model.StatusActive, FileActive},
		{model.StatusBlocked, FileActive},
		{model.StatusBacklog, FileBacklog},
		{model.StatusDone, FileDone},
		{model.StatusCancelled, FileDone},
	}
	for _, tt := range tests {
		if got := TargetFile(tt.status); got != tt.want {
			t.Errorf("TargetFile(%v) = %v, want %v", tt.status, got, tt.want)
		}
	}
}
