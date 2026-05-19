package fsx_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/nomos/nomos/internal/fsx"
)

func TestReadYAMLFromFS(t *testing.T) {
	fsys := fstest.MapFS{
		"config.yaml": {Data: []byte("id: test\nname: Test Cosmos\n")},
	}
	var out struct {
		ID   string `yaml:"id"`
		Name string `yaml:"name"`
	}
	if err := fsx.ReadYAMLFromFS(fsys, "config.yaml", &out); err != nil {
		t.Fatalf("ReadYAMLFromFS: %v", err)
	}
	if out.ID != "test" || out.Name != "Test Cosmos" {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestReadYAMLFromFS_NotFound(t *testing.T) {
	fsys := fstest.MapFS{}
	var out struct{ ID string }
	if err := fsx.ReadYAMLFromFS(fsys, "missing.yaml", &out); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadYAML_NotFound(t *testing.T) {
	var out struct{ ID string }
	if err := fsx.ReadYAML("/nonexistent/path/file.yaml", &out); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestWriteYAML_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.yaml")
	data := struct {
		ID   string `yaml:"id"`
		Name string `yaml:"name"`
	}{ID: "write-test", Name: "Write Test"}
	if err := fsx.WriteYAML(path, data); err != nil {
		t.Fatalf("WriteYAML: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !containsStr(string(b), "write-test") {
		t.Errorf("file content missing id: %s", b)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s[1:], sub) || s[:len(sub)] == sub)
}
