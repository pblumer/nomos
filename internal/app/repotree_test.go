package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func TestBuildRepoTreeMirrorsPhysicalLayout(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	// A freely organized folder holding a nested service.
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "team-a", "mailbox"), 0o755))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "team-a", "mailbox", "service.yaml"),
		[]byte("name: mailbox\nowner: Mailing Team\nstatus: draft\n"), 0o644))
	// A blueprint file under the catalog.
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints"), 0o755))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "account.yaml"),
		[]byte("id: PROD-ACC-001\ntype: product_blueprint\nname: Account\n"), 0o644))
	// Plumbing that must stay hidden.
	must(os.MkdirAll(storage.IndexDir(p), 0o755))
	must(os.MkdirAll(filepath.Join(p, ".git", "objects"), 0o755))

	nodes := BuildRepoTree(p)

	nomos := findChild(nodes, "folder", ".nomos")
	if nomos == nil {
		t.Fatal("expected the .nomos working directory to be mirrored as a folder")
	}
	services := findChild(nomos.Children, "folder", "services")
	if services == nil {
		t.Fatal("expected services/ folder under .nomos")
	}
	// Direct service leaf.
	if findChild(services.Children, "service", "user-account") == nil {
		t.Fatal("expected user-account service leaf")
	}
	// Nested folder holding a service leaf (free organization preserved).
	teamA := findChild(services.Children, "folder", "team-a")
	if teamA == nil {
		t.Fatal("expected team-a folder")
	}
	mailbox := findChild(teamA.Children, "service", "mailbox")
	if mailbox == nil {
		t.Fatal("expected mailbox service leaf inside team-a")
	}
	if mailbox.GitPath != filepath.Join(".nomos", "services", "team-a", "mailbox") {
		t.Fatalf("GitPath = %q, want mirrored relative path", mailbox.GitPath)
	}
	// Service dirs collapse to a single leaf: no service.yaml file node under them.
	if len(mailbox.Children) != 0 {
		t.Fatalf("service leaf should not expose its internal files, got %d children", len(mailbox.Children))
	}
	// Blueprint file recognized as a typed node carrying its ID.
	catalog := findChild(nomos.Children, "folder", "catalog")
	if catalog == nil {
		t.Fatal("expected catalog/ folder")
	}
	bps := findChild(catalog.Children, "folder", "blueprints")
	if bps == nil {
		t.Fatal("expected blueprints/ folder")
	}
	bp := findChild(bps.Children, "blueprint", "Account")
	if bp == nil || bp.Canonical != "PROD-ACC-001" {
		t.Fatalf("expected blueprint node with id PROD-ACC-001, got %+v", bp)
	}
	// cosmos.yaml shows as a plain file node.
	if findChild(nomos.Children, "file", "cosmos.yaml") == nil {
		t.Fatal("expected cosmos.yaml file node")
	}
	// Hidden plumbing is not mirrored.
	if findChild(nomos.Children, "folder", "index") != nil {
		t.Fatal("index/ should be hidden")
	}
	if findChild(nodes, "folder", ".git") != nil {
		t.Fatal(".git should be hidden")
	}
}

func TestCreateRepoFolderAndMove(t *testing.T) {
	p := createAppTestCosmos(t)

	if err := CreateRepoFolder(p, ".nomos/services/team-a"); err != nil {
		t.Fatalf("CreateRepoFolder: %v", err)
	}
	if fi, err := os.Stat(filepath.Join(p, ".nomos", "services", "team-a")); err != nil || !fi.IsDir() {
		t.Fatalf("folder not created: %v", err)
	}
	// Duplicate creation is rejected.
	if err := CreateRepoFolder(p, ".nomos/services/team-a"); err == nil {
		t.Fatal("expected conflict creating an existing folder")
	}
	// Path traversal is rejected.
	if err := CreateRepoFolder(p, "../escape"); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(p), "escape")); err == nil {
		t.Fatal("traversal escaped the workspace")
	}

	// Move a service directory into the new folder; identity is path-independent.
	from := ".nomos/services/user-account"
	if err := MoveRepoNode(p, from, ".nomos/services/team-a"); err != nil {
		t.Fatalf("MoveRepoNode: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p, ".nomos", "services", "team-a", "user-account", "service.yaml")); err != nil {
		t.Fatalf("service not moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p, ".nomos", "services", "user-account")); err == nil {
		t.Fatal("source still present after move")
	}
	// Moving a folder into its own subtree is rejected.
	if err := MoveRepoNode(p, ".nomos/services", ".nomos/services/team-a"); err == nil {
		t.Fatal("expected move-into-own-subtree to be rejected")
	}
}

func findChild(nodes []NamespaceTreeNodeDTO, kind, label string) *NamespaceTreeNodeDTO {
	for i := range nodes {
		if nodes[i].Kind == kind && nodes[i].Label == label {
			return &nodes[i]
		}
	}
	return nil
}
