package cosmosfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/storage"
)

func makeCosmosDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	must(t, os.MkdirAll(storage.NomosDir(dir), 0o755))
	must(t, os.WriteFile(storage.CosmosFile(dir), []byte("id: test-cosmos\nname: Test\nversion: 0.1.0\nstatus: draft\nowner: Team\n"), 0o644))
	return dir
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadTree_EmptyCosmos(t *testing.T) {
	dir := makeCosmosDir(t)
	tree, err := cosmosfs.LoadTree(dir)
	if err != nil {
		t.Fatalf("LoadTree: %v", err)
	}
	if tree.Cosmos.ID != "test-cosmos" {
		t.Errorf("unexpected cosmos ID: %q", tree.Cosmos.ID)
	}
	if len(tree.Domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(tree.Domains))
	}
}

func TestLoadTree_WithDomain(t *testing.T) {
	dir := makeCosmosDir(t)
	domainDir := filepath.Join(storage.DomainsDir(dir), "identity.blumer.cloud")
	must(t, os.MkdirAll(domainDir, 0o755))
	must(t, os.WriteFile(filepath.Join(domainDir, "domain.yaml"), []byte("name: identity.blumer.cloud\nowner: Identity Team\nstatus: draft\n"), 0o644))

	tree, err := cosmosfs.LoadTree(dir)
	if err != nil {
		t.Fatalf("LoadTree: %v", err)
	}
	if len(tree.Domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(tree.Domains))
	}
}

func TestLoadTree_WithService(t *testing.T) {
	dir := makeCosmosDir(t)
	domainDir := filepath.Join(storage.DomainsDir(dir), "identity.blumer.cloud")
	svcDir := filepath.Join(domainDir, "services", "user-account")
	must(t, os.MkdirAll(svcDir, 0o755))
	must(t, os.WriteFile(filepath.Join(domainDir, "domain.yaml"), []byte("name: identity.blumer.cloud\nstatus: draft\n"), 0o644))
	must(t, os.WriteFile(filepath.Join(svcDir, "service.yaml"), []byte("name: user-account\nstatus: draft\n"), 0o644))

	tree, err := cosmosfs.LoadTree(dir)
	if err != nil {
		t.Fatalf("LoadTree: %v", err)
	}
	if len(tree.Domains) != 1 || len(tree.Domains[0].Services) != 1 {
		t.Errorf("expected 1 domain with 1 service: %+v", tree.Domains)
	}
}

func TestLoadTree_WithDecision(t *testing.T) {
	dir := makeCosmosDir(t)
	domainDir := filepath.Join(storage.DomainsDir(dir), "identity.blumer.cloud")
	decDir := filepath.Join(domainDir, "decisions", "DEC-001")
	must(t, os.MkdirAll(decDir, 0o755))
	must(t, os.WriteFile(filepath.Join(domainDir, "domain.yaml"), []byte("name: identity.blumer.cloud\nstatus: draft\n"), 0o644))
	must(t, os.WriteFile(filepath.Join(decDir, "decision.yaml"), []byte("id: DEC-001\ntype: decision\nname: Test Decision\nstatus: draft\n"), 0o644))

	tree, err := cosmosfs.LoadTree(dir)
	if err != nil {
		t.Fatalf("LoadTree: %v", err)
	}
	if len(tree.Domains[0].Decisions) != 1 {
		t.Errorf("expected 1 decision, got %d", len(tree.Domains[0].Decisions))
	}
}

func TestScanDecisions_NoDMN(t *testing.T) {
	dir := t.TempDir()
	decDir := filepath.Join(dir, "decisions", "DEC-002")
	must(t, os.MkdirAll(decDir, 0o755))
	must(t, os.WriteFile(filepath.Join(decDir, "decision.yaml"), []byte("id: DEC-002\ntype: decision\nname: No DMN\nstatus: draft\n"), 0o644))

	nodes, err := cosmosfs.ScanDecisions(dir)
	if err != nil {
		t.Fatalf("ScanDecisions: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].DMNPath != "" {
		t.Errorf("expected empty DMNPath, got %q", nodes[0].DMNPath)
	}
}

func TestScanDecisions_WithDMN(t *testing.T) {
	dir := t.TempDir()
	decDir := filepath.Join(dir, "decisions", "DEC-003")
	must(t, os.MkdirAll(decDir, 0o755))
	must(t, os.WriteFile(filepath.Join(decDir, "decision.yaml"), []byte("id: DEC-003\ntype: decision\nname: With DMN\nstatus: draft\n"), 0o644))
	must(t, os.WriteFile(filepath.Join(decDir, "decision.dmn"), []byte(`<?xml version="1.0"?><definitions/>`), 0o644))

	nodes, err := cosmosfs.ScanDecisions(dir)
	if err != nil {
		t.Fatalf("ScanDecisions: %v", err)
	}
	if len(nodes) != 1 || nodes[0].DMNPath == "" {
		t.Errorf("expected DMNPath to be set: %+v", nodes)
	}
}

func TestScanDecisions_Empty(t *testing.T) {
	dir := t.TempDir()
	nodes, err := cosmosfs.ScanDecisions(dir)
	if err != nil {
		t.Fatalf("ScanDecisions on empty dir: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected empty, got %d", len(nodes))
	}
}

func TestLoadTree_WithBlueprint(t *testing.T) {
	dir := makeCosmosDir(t)
	bpDir := filepath.Join(storage.CatalogDir(dir), "blueprints")
	must(t, os.MkdirAll(bpDir, 0o755))
	must(t, os.WriteFile(filepath.Join(bpDir, "my-product.yaml"), []byte("id: PROD-001\ntype: product_blueprint\nname: My Product\nstatus: draft\n"), 0o644))

	tree, err := cosmosfs.LoadTree(dir)
	if err != nil {
		t.Fatalf("LoadTree: %v", err)
	}
	if len(tree.Blueprints) != 1 {
		t.Errorf("expected 1 blueprint, got %d", len(tree.Blueprints))
	}
}

func TestLoadTree_NoCosmosFile(t *testing.T) {
	dir := t.TempDir()
	_, err := cosmosfs.LoadTree(dir)
	if err == nil {
		t.Error("expected error for missing cosmos.yaml")
	}
}

func TestLoadTree_WithInstance(t *testing.T) {
	dir := makeCosmosDir(t)
	instDir := filepath.Join(storage.CatalogDir(dir), "instances", "products")
	must(t, os.MkdirAll(instDir, 0o755))
	must(t, os.WriteFile(filepath.Join(instDir, "inst-001.yaml"), []byte("id: INST-001\ntype: product_instance\nblueprint_ref: PB-1\nblueprint_version: 0.1.0\ncompliance_status: compliant\n"), 0o644))

	tree, err := cosmosfs.LoadTree(dir)
	if err != nil {
		t.Fatalf("LoadTree: %v", err)
	}
	if len(tree.Instances) != 1 {
		t.Errorf("expected 1 instance, got %d", len(tree.Instances))
	}
}

func TestLoadTree_WithServicegraph(t *testing.T) {
	dir := makeCosmosDir(t)
	sgDir := filepath.Join(storage.CatalogDir(dir), "servicegraphs")
	must(t, os.MkdirAll(sgDir, 0o755))
	must(t, os.WriteFile(filepath.Join(sgDir, "sg-001.yaml"), []byte("id: SG-001\ntype: servicegraph\nname: Test Graph\n"), 0o644))

	tree, err := cosmosfs.LoadTree(dir)
	if err != nil {
		t.Fatalf("LoadTree: %v", err)
	}
	if len(tree.Servicegraphs) != 1 {
		t.Errorf("expected 1 servicegraph, got %d", len(tree.Servicegraphs))
	}
}
