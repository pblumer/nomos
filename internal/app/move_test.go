package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func moveTestCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "svc-a"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "svc-b"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: c\nname: C\nversion: 0.1.0\nstatus: draft\nowner: t\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "svc-a/service.yaml"), []byte("name: svc-a\ncapabilities:\n  - id: CAP-1\n    name: Cap One\nmethods:\n  - name: doThing\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "svc-b/service.yaml"), []byte("name: svc-b\n"), 0o644))
	return p
}

func TestMoveServiceElementBetweenServices(t *testing.T) {
	p := moveTestCosmos(t)
	dto, err := MoveServiceElement(p, "svc-a", "capability", "CAP-1", "svc-b")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range dto.Capabilities {
		if c == "Cap One" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Cap One should be in target service, got %+v", dto.Capabilities)
	}
	src, _ := GetService(p, "svc-a")
	for _, c := range src.Capabilities {
		if c == "Cap One" {
			t.Fatal("Cap One should be removed from source service")
		}
	}
}
