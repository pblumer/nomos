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

// BuildExplorerTree builds the Cosmos Explorer tree by aggregating over the
// mounted servers (ADR-0022 §1/§4/§6): the local server plus configured remote
// servers. Each server exposes its repositories, and each repository's content
// is the namespace tree. The plain namespace tree (BuildNamespaceTree) is left
// unchanged for the namespace/domain views. Server and repository levels are
// always shown; an unreachable remote server degrades to a status badge instead
// of breaking the whole tree.
func BuildExplorerTree(path string) (NamespaceTreeDTO, error) {
	mounts, err := ListMounts(path)
	if err != nil {
		return NamespaceTreeDTO{}, err
	}
	cosmos, _ := GetCosmos(path)
	root := NamespaceTreeNodeDTO{Label: fallback(cosmos.Name, "Local Cosmos"), Kind: "cosmos", CanOpenDetails: true}
	for _, m := range mounts.Mounts {
		if m.Local {
			root.Children = append(root.Children, localServerNode(path, m))
		} else {
			root.Children = append(root.Children, remoteServerNode(m))
		}
	}
	return NamespaceTreeDTO{Root: root}, nil
}

func serverNode(m MountDTO, repoCount int, status string) NamespaceTreeNodeDTO {
	return NamespaceTreeNodeDTO{
		Label:          m.Endpoint,
		Kind:           "server",
		CanOpenDetails: true,
		Server:         &ServerDTO{MountID: m.ID, Endpoint: m.Endpoint, Label: m.Label, Local: m.Local, Status: status, RepositoryCount: repoCount},
	}
}

func repositoryNode(r RepositoryDTO, content []NamespaceTreeNodeDTO) NamespaceTreeNodeDTO {
	rc := r
	return NamespaceTreeNodeDTO{
		Label:          r.Name,
		Kind:           "repository",
		CanOpenDetails: true,
		Repository:     &rc,
		Children:       content,
	}
}

func localServerNode(path string, m MountDTO) NamespaceTreeNodeDTO {
	repos, _ := ListRepositories(path)
	ns, _ := BuildNamespaceTree(path)
	server := serverNode(m, len(repos.Repositories), "online")
	for _, r := range repos.Repositories {
		server.Children = append(server.Children, repositoryNode(r, ns.Root.Children))
	}
	return server
}

func remoteServerNode(m MountDTO) NamespaceTreeNodeDTO {
	repos, err := fetchRemoteRepositories(m.Endpoint)
	if err != nil {
		return serverNode(m, 0, "unreachable")
	}
	server := serverNode(m, len(repos.Repositories), "online")
	for _, r := range repos.Repositories {
		var content []NamespaceTreeNodeDTO
		if ns, nerr := fetchRemoteNamespaces(m.Endpoint, r.ID); nerr == nil {
			content = ns.Root.Children
		}
		server.Children = append(server.Children, repositoryNode(r, content))
	}
	return server
}
