package server

import (
	"path/filepath"
	"testing"
)

func TestResolveSourcePath(t *testing.T) {
	root := t.TempDir()
	h := &handler{cosmosPath: root}

	tests := []struct {
		name string
		raw  string
		ok   bool
	}{
		{"absolute inside root", filepath.Join(root, ".nomos", "cosmos.yaml"), true},
		{"repo-relative path", ".nomos/cosmos.yaml", true},
		{"repo-relative markdown", "README.md", true},
		{"unsupported extension", ".nomos/notes.txt", false},
		{"traversal escape", "../secrets.yaml", false},
		{"empty", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := h.resolveSourcePath(tc.raw)
			if ok != tc.ok {
				t.Fatalf("resolveSourcePath(%q) ok = %v, want %v", tc.raw, ok, tc.ok)
			}
			if ok {
				rel, err := filepath.Rel(root, got)
				if err != nil || rel == ".." || filepath.IsAbs(rel) && rel != got {
					t.Fatalf("resolved path %q escapes root %q", got, root)
				}
			}
		})
	}
}
