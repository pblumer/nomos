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
