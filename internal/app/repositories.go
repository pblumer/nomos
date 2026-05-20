package app

import (
	"net"
	"net/http"
	"strings"

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

// BuildExplorerTree builds the Cosmos Explorer tree as a DNS-shaped namespace
// (ADR-0022/0026): used TLDs at the top, domain labels beneath, and a server
// shown as a marker at the domain where it runs (e.g. cloud → blumer → nomos).
// Servers without a DNS domain (localhost / bare hostname) appear under a
// synthetic "local" branch (local → hostname). A server's content hangs
// directly under it when it has a single repository; with several repositories
// a repository level is inserted (Variant 3). Unreachable remote servers degrade
// to an "offline" marker.
func BuildExplorerTree(path string) (NamespaceTreeDTO, error) {
	mounts, err := ListMounts(path)
	if err != nil {
		return NamespaceTreeDTO{}, err
	}
	cosmos, _ := GetCosmos(path)
	root := NamespaceTreeNodeDTO{Label: fallback(cosmos.Name, "Local Cosmos"), Kind: "cosmos", CanOpenDetails: true}
	for _, m := range mounts.Mounts {
		srv, content := serverContent(path, m)
		insertServer(&root, dnsPlacement(m), srv, content)
	}
	return NamespaceTreeDTO{Root: root}, nil
}

// dnsPlacement returns the tree labels (TLD first) where a server is shown.
// A DNS host like nomos.blumer.cloud → [cloud, blumer, nomos]; a server without
// a DNS domain → [local, <host>].
func dnsPlacement(m MountDTO) []string {
	host := hostOnly(m.Endpoint)
	if host == "" {
		host = "localhost"
	}
	// localhost, an IP, or a bare hostname (e.g. a container ID) has no DNS
	// hierarchy: the local server is shown under a synthetic "local" branch
	// (local → host), a remote as a single leaf.
	if host == "localhost" || net.ParseIP(host) != nil || !strings.Contains(host, ".") {
		if m.Local {
			return []string{"local", host}
		}
		return []string{host}
	}
	// A real DNS host (e.g. nomos.blumer.cloud) is placed in the DNS hierarchy
	// regardless of whether it is the local server.
	labels := strings.Split(host, ".")
	out := make([]string, len(labels))
	for i, l := range labels {
		out[len(labels)-1-i] = l
	}
	return out
}

// hostOnly extracts the bare host from a mount endpoint (drops scheme, path, port).
func hostOnly(endpoint string) string {
	e := endpoint
	if i := strings.Index(e, "://"); i >= 0 {
		e = e[i+3:]
	}
	if i := strings.IndexByte(e, '/'); i >= 0 {
		e = e[:i]
	}
	if i := strings.LastIndexByte(e, ':'); i >= 0 {
		e = e[:i]
	}
	return e
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

// serverContent returns the server marker DTO and the nodes shown beneath it:
// the content directly for a single repository, or one node per repository when
// there are several.
func serverContent(path string, m MountDTO) (*ServerDTO, []NamespaceTreeNodeDTO) {
	if m.Local {
		repos, _ := ListRepositories(path)
		s := &ServerDTO{MountID: m.ID, Endpoint: m.Endpoint, Label: m.Label, Local: true, Authenticated: m.Authenticated, Status: "online", RepositoryCount: len(repos.Repositories)}
		return s, repoChildren(repos.Repositories, func(r RepositoryDTO) []NamespaceTreeNodeDTO {
			loc := r.Location
			if loc == "" {
				loc = path
			}
			ns, _ := BuildNamespaceTree(loc)
			return ns.Root.Children
		})
	}
	repos, err := fetchRemoteRepositories(m.Endpoint)
	if err != nil {
		return &ServerDTO{MountID: m.ID, Endpoint: m.Endpoint, Label: m.Label, Authenticated: m.Authenticated, Status: "unreachable"}, nil
	}
	s := &ServerDTO{MountID: m.ID, Endpoint: m.Endpoint, Label: m.Label, Authenticated: m.Authenticated, Status: "online", RepositoryCount: len(repos.Repositories)}
	return s, repoChildren(repos.Repositories, func(r RepositoryDTO) []NamespaceTreeNodeDTO {
		ns, nerr := fetchRemoteNamespaces(m.Endpoint, r.ID)
		if nerr != nil {
			return nil
		}
		return ns.Root.Children
	})
}

// repoChildren collapses a single repository (content shown directly) or wraps
// each repository when there are several (Variant 3).
func repoChildren(repos []RepositoryDTO, content func(RepositoryDTO) []NamespaceTreeNodeDTO) []NamespaceTreeNodeDTO {
	if len(repos) == 1 {
		return content(repos[0])
	}
	out := []NamespaceTreeNodeDTO{}
	for _, r := range repos {
		out = append(out, repositoryNode(r, content(r)))
	}
	return out
}

// insertServer places a server marker at its DNS path, creating neutral "dns"
// positioning nodes for the TLD and intermediate labels.
func insertServer(root *NamespaceTreeNodeDTO, labels []string, s *ServerDTO, content []NamespaceTreeNodeDTO) {
	node := root
	for i, label := range labels {
		idx := findChildByLabel(node, label)
		if idx == -1 {
			node.Children = append(node.Children, NamespaceTreeNodeDTO{Label: label, Kind: "dns"})
			idx = len(node.Children) - 1
		}
		if i == len(labels)-1 {
			node.Children[idx].Kind = "server"
			node.Children[idx].Server = s
			node.Children[idx].CanOpenDetails = true
			node.Children[idx].Children = content
		}
		node = &node.Children[idx]
	}
}

func findChildByLabel(n *NamespaceTreeNodeDTO, label string) int {
	for i := range n.Children {
		if n.Children[i].Label == label {
			return i
		}
	}
	return -1
}

// CreateRepository creates a new local filesystem repository on this server and
// records it in the server's repository config (ADR-0022 §2).
func CreateRepository(path, name string) (RepositoryDTO, error) {
	r, err := repo.NewLocalRegistry(path).CreateFilesystem("", name)
	if err != nil {
		return RepositoryDTO{}, Error(CodeInvalidInput, err.Error(), http.StatusBadRequest, err)
	}
	return RepositoryDTO{ID: r.ID, Name: r.Name, Kind: string(r.Kind), Location: r.Location, DefaultBranch: r.DefaultBranch, Status: r.Status, Head: r.Head}, nil
}
