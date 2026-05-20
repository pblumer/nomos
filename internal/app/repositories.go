package app

import (
	"net/http"

	"github.com/nomos/nomos/internal/repo"
)

// ListRepositories returns the repositories the server manages (ADR-0022 §2).
// PR 1 covers the local server: a single default repository for the active
// workspace, its name taken from the workspace cosmos when present.
func ListRepositories(path string) (RepositoriesDTO, error) {
	cosmos, _ := GetCosmos(path)
	out := RepositoriesDTO{Repositories: []RepositoryDTO{}}
	for _, r := range repo.NewLocalRegistry(path).List() {
		name := r.Name
		if name == "" {
			name = fallback(cosmos.Name, "Local Repository")
		}
		out.Repositories = append(out.Repositories, RepositoryDTO{
			ID:            r.ID,
			Name:          name,
			Kind:          string(r.Kind),
			Location:      r.Location,
			DefaultBranch: r.DefaultBranch,
			Status:        r.Status,
			Head:          r.Head,
		})
	}
	return out, nil
}

// GetRepository resolves one repository by id on this server (ADR-0022 §2).
// The returned Location is the workspace path that repository-scoped reads use.
func GetRepository(path, id string) (RepositoryDTO, error) {
	repos, err := ListRepositories(path)
	if err != nil {
		return RepositoryDTO{}, err
	}
	for _, r := range repos.Repositories {
		if r.ID == id {
			return r, nil
		}
	}
	return RepositoryDTO{}, Error(CodeRepositoryNotFound, "Repository not found: "+id, http.StatusNotFound, nil)
}

// LocalServerEndpoint is the conventional endpoint of the local Nomos server
// (ADR-0015 default port 7373).
const LocalServerEndpoint = "localhost:7373"

// BuildExplorerTree wraps the namespace tree under the local server and its
// repositories for the Cosmos Explorer (ADR-0022 §1/§6). The plain namespace
// tree (BuildNamespaceTree) is left unchanged for the namespace/domain views.
// The server and repository levels are always shown.
func BuildExplorerTree(path string) (NamespaceTreeDTO, error) {
	ns, err := BuildNamespaceTree(path)
	if err != nil {
		return NamespaceTreeDTO{}, err
	}
	repos, err := ListRepositories(path)
	if err != nil {
		return NamespaceTreeDTO{}, err
	}
	content := ns.Root.Children
	server := NamespaceTreeNodeDTO{
		Label:          LocalServerEndpoint,
		Kind:           "server",
		CanOpenDetails: true,
		Server:         &ServerDTO{Endpoint: LocalServerEndpoint, Local: true, Status: "online", RepositoryCount: len(repos.Repositories)},
	}
	for i := range repos.Repositories {
		r := repos.Repositories[i]
		server.Children = append(server.Children, NamespaceTreeNodeDTO{
			Label:          r.Name,
			Kind:           "repository",
			CanOpenDetails: true,
			Repository:     &r,
			Children:       content,
		})
	}
	root := ns.Root
	root.Children = []NamespaceTreeNodeDTO{server}
	return NamespaceTreeDTO{Root: root}, nil
}
