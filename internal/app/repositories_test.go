package app

import "testing"

func TestListRepositoriesReturnsLocalDefault(t *testing.T) {
	p := createAppTestCosmos(t)
	dto, err := ListRepositories(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(dto.Repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(dto.Repositories))
	}
	r := dto.Repositories[0]
	if r.ID != "default" {
		t.Errorf("ID = %q, want default", r.ID)
	}
	if r.Kind != "filesystem" {
		t.Errorf("Kind = %q, want filesystem", r.Kind)
	}
	if r.Location != p {
		t.Errorf("Location = %q, want %q", r.Location, p)
	}
	if r.Name != "Local Cosmos" {
		t.Errorf("Name = %q, want Local Cosmos", r.Name)
	}
}

func TestBuildExplorerTreePlacesLocalServerUnderLocal(t *testing.T) {
	tree, err := BuildExplorerTree(createAppTestCosmos(t))
	if err != nil {
		t.Fatal(err)
	}
	// Local server lives under a synthetic "local" branch (local → hostname).
	local := findTreeNode(tree.Root, "dns", "local")
	if local == nil {
		t.Fatal("expected a 'local' DNS branch for the local server")
	}
	server := findLocalServerNode(tree.Root)
	if server == nil || server.Server == nil || !server.Server.Local {
		t.Fatalf("expected local server node, got %+v", server)
	}
	// Single repository → content hangs directly under the server (no repo level).
	if findTreeNode(*server, "namespace-parent", "Namespaces") == nil {
		t.Fatal("namespace tree should hang directly under the single-repo server")
	}
	if findTreeNode(*server, "service", "user-account") == nil {
		t.Fatal("server content should include domain services")
	}
}

func TestBuildExplorerTreePlacesLocalServerUnderDomain(t *testing.T) {
	t.Setenv("NOMOS_DOMAIN", "nomos.blumer.cloud")
	tree, err := BuildExplorerTree(createAppTestCosmos(t))
	if err != nil {
		t.Fatal(err)
	}
	// With a public domain configured, the local server is placed in the DNS
	// hierarchy (cloud → blumer → nomos) instead of under "local".
	if findTreeNode(tree.Root, "dns", "local") != nil {
		t.Fatal("local server should not appear under 'local' when NOMOS_DOMAIN is set")
	}
	cloud := findTreeNode(tree.Root, "dns", "cloud")
	if cloud == nil {
		t.Fatal("expected a 'cloud' TLD branch")
	}
	if findTreeNode(*cloud, "dns", "blumer") == nil {
		t.Fatal("expected a 'blumer' branch under 'cloud'")
	}
	server := findLocalServerNode(tree.Root)
	if server == nil || server.Server == nil || !server.Server.Local {
		t.Fatalf("expected local server node, got %+v", server)
	}
	if server.Label != "nomos" {
		t.Errorf("server label = %q, want nomos", server.Label)
	}
}

func TestListRepositoriesNameFallsBackWithoutCosmos(t *testing.T) {
	dto, err := ListRepositories(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(dto.Repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(dto.Repositories))
	}
	if dto.Repositories[0].Name != "Local Repository" {
		t.Errorf("Name = %q, want Local Repository", dto.Repositories[0].Name)
	}
}

func TestCreateRepositoryAndListAndScopedContent(t *testing.T) {
	p := createAppTestCosmos(t)
	created, err := CreateRepository(p, "Team Beta")
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "team-beta" || created.Kind != "filesystem" {
		t.Fatalf("unexpected created repo: %+v", created)
	}
	repos, _ := ListRepositories(p)
	var ids []string
	for _, r := range repos.Repositories {
		ids = append(ids, r.ID)
	}
	if len(repos.Repositories) != 2 {
		t.Fatalf("expected default + team-beta, got %v", ids)
	}
	// New repo resolves and has its own (empty) cosmos.
	got, err := GetRepository(p, "team-beta")
	if err != nil || got.Location == "" {
		t.Fatalf("GetRepository(team-beta) = %+v err %v", got, err)
	}
	co, err := GetCosmos(got.Location)
	if err != nil || co.Name != "Team Beta" {
		t.Fatalf("new repo cosmos = %+v err %v", co, err)
	}
	// Duplicate name → conflict/error.
	if _, err := CreateRepository(p, "Team Beta"); err == nil {
		t.Fatal("duplicate repository should fail")
	}
}

func TestExplorerTreeShowsMultipleRepositories(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := CreateRepository(p, "extra"); err != nil {
		t.Fatal(err)
	}
	tree, err := BuildExplorerTree(p)
	if err != nil {
		t.Fatal(err)
	}
	server := findLocalServerNode(tree.Root)
	if server == nil {
		t.Fatal("local server node not found")
	}
	repos := 0
	for _, c := range server.Children {
		if c.Kind == "repository" {
			repos++
		}
	}
	if repos != 2 {
		t.Fatalf("expected 2 repository nodes under local server, got %d", repos)
	}
}
