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
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "a.cloud/services/svc-a"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "b.cloud/services/svc-b"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: c\nname: C\nversion: 0.1.0\nstatus: draft\nowner: t\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "a.cloud/domain.yaml"), []byte("name: a.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "b.cloud/domain.yaml"), []byte("name: b.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "a.cloud/services/svc-a/service.yaml"), []byte("name: svc-a\nowned_by: a.cloud\ncapabilities:\n  - id: CAP-1\n    name: Cap One\nmethods:\n  - name: doThing\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "b.cloud/services/svc-b/service.yaml"), []byte("name: svc-b\nowned_by: b.cloud\n"), 0o644))
	return p
}

func TestMoveServiceBetweenDomains(t *testing.T) {
	p := moveTestCosmos(t)
	dto, err := MoveService(p, "a.cloud", "svc-a", "b.cloud")
	if err != nil {
		t.Fatal(err)
	}
	if dto.Domain != "b.cloud" || dto.OwnedBy != "b.cloud" {
		t.Fatalf("expected service under b.cloud owned_by b.cloud, got %+v", dto)
	}
	if _, err := GetService(p, "a.cloud", "svc-a"); err == nil {
		t.Fatal("service should no longer exist under a.cloud")
	}
}

func TestMoveDecisionBetweenDomains(t *testing.T) {
	p := moveTestCosmos(t)
	if _, err := CreateDecision(p, "a.cloud", CreateDecisionRequest{Name: "Rule X", ID: "DEC-1"}); err != nil {
		t.Fatal(err)
	}
	dto, err := MoveDecision(p, "a.cloud", "DEC-1", "b.cloud")
	if err != nil {
		t.Fatal(err)
	}
	if dto.ID != "DEC-1" {
		t.Fatalf("expected DEC-1, got %+v", dto)
	}
	if _, err := GetDecision(p, "a.cloud", "DEC-1"); err == nil {
		t.Fatal("decision should no longer exist under a.cloud")
	}
	if _, err := GetDecision(p, "b.cloud", "DEC-1"); err != nil {
		t.Fatalf("decision should exist under b.cloud: %v", err)
	}
}

func TestMoveServiceElementBetweenServices(t *testing.T) {
	p := moveTestCosmos(t)
	dto, err := MoveServiceElement(p, "a.cloud", "svc-a", "capability", "CAP-1", "b.cloud", "svc-b")
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
	src, _ := GetService(p, "a.cloud", "svc-a")
	for _, c := range src.Capabilities {
		if c == "Cap One" {
			t.Fatal("Cap One should be removed from source service")
		}
	}
}
