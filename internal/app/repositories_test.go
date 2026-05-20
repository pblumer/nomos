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

func TestBuildExplorerTreeWrapsNamespacesUnderServerRepository(t *testing.T) {
	tree, err := BuildExplorerTree(createAppTestCosmos(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Root.Children) != 1 {
		t.Fatalf("expected single server child, got %d", len(tree.Root.Children))
	}
	server := tree.Root.Children[0]
	if server.Kind != "server" || server.Server == nil || !server.Server.Local {
		t.Fatalf("expected local server node, got %+v", server)
	}
	if len(server.Children) != 1 {
		t.Fatalf("expected single repository child, got %d", len(server.Children))
	}
	repository := server.Children[0]
	if repository.Kind != "repository" || repository.Repository == nil || repository.Repository.ID != "default" {
		t.Fatalf("expected default repository node, got %+v", repository)
	}
	if findTreeNode(repository, "namespace-parent", "Namespaces") == nil {
		t.Fatal("namespace tree should hang under the repository node")
	}
	if findTreeNode(repository, "service", "user-account") == nil {
		t.Fatal("repository content should include domain services")
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
	server := tree.Root.Children[0]
	if len(server.Children) != 2 {
		t.Fatalf("expected 2 repository nodes under local server, got %d", len(server.Children))
	}
}
