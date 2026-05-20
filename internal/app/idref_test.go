package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func TestServiceRefByIDSurvivesMove(t *testing.T) {
	p := t.TempDir()
	must := func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "a.cloud/services/svc-a"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "b.cloud"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints/products"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: c\nname: C\nversion: 0.1.0\nstatus: draft\nowner: t\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "a.cloud/domain.yaml"), []byte("name: a.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "b.cloud/domain.yaml"), []byte("name: b.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "a.cloud/services/svc-a/service.yaml"), []byte("id: SVC-A\nname: svc-a\nowned_by: a.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints/products/prod.yaml"), []byte("id: PROD-1\ntype: product_blueprint\nname: Product\noffered_by: a.cloud\nfulfillment:\n  required_services:\n    - service_ref: SVC-A\n      role: primary\n      required: true\n"), 0o644))

	check := func(wantDomain string) {
		bp, err := GetBlueprint(p, "PROD-1")
		if err != nil {
			t.Fatal(err)
		}
		rs := bp.Fulfillment.RequiredServices
		if len(rs) != 1 || rs[0].ResolutionStatus != "resolved" || rs[0].ResolvedDomain != wantDomain {
			t.Fatalf("want resolved in %s, got %+v", wantDomain, rs)
		}
	}
	check("a.cloud")
	if _, err := MoveService(p, "a.cloud", "svc-a", "b.cloud"); err != nil {
		t.Fatal(err)
	}
	check("b.cloud") // ID ref still resolves after the service moved
}
