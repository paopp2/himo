package store

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	cases := []struct {
		file   string
		parser func([]byte) (*Document, error)
	}{
		{"active_basic.md", ParseActive},
		{"active_with_notes.md", ParseActive},
		{"backlog_basic.md", ParseBacklog},
		{"done_basic.md", ParseDone},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			path := filepath.Join("testdata", tc.file)
			orig, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			doc, err := tc.parser(orig)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			got := Render(doc)
			if !bytes.Equal(got, orig) {
				t.Errorf("round-trip mismatch for %s:\ngot:\n%s\nwant:\n%s", tc.file, got, orig)
			}
		})
	}
}
