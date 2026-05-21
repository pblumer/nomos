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
	return BuildExplorerTreeForHost(path, "")
}

// BuildExplorerTreeForHost is BuildExplorerTree with an optional request Host
// header used to auto-detect the local server's public domain when NOMOS_DOMAIN
// is unset (see localDisplayEndpoint).
func BuildExplorerTreeForHost(path, hostHint string) (NamespaceTreeDTO, error) {
	mounts, err := ListMountsForHost(path, hostHint)
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
			return BuildRepoTree(loc)
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
	return repoToDTO(r), nil
}

// DeleteRepository removes a filesystem repository from this server (ADR-0022).
// The default repository cannot be removed.
func DeleteRepository(path, id string) error {
	if err := repo.NewLocalRegistry(path).Delete(id); err != nil {
		return Error(CodeInvalidInput, err.Error(), http.StatusBadRequest, err)
	}
	return nil
}

// RenameRepository updates a repository's display name (ADR-0022).
func RenameRepository(path, id, name string) (RepositoryDTO, error) {
	r, err := repo.NewLocalRegistry(path).Rename(id, name)
	if err != nil {
		return RepositoryDTO{}, Error(CodeInvalidInput, err.Error(), http.StatusBadRequest, err)
	}
	return repoToDTO(r), nil
}

func repoToDTO(r repo.Repository) RepositoryDTO {
	return RepositoryDTO{ID: r.ID, Name: r.Name, Kind: string(r.Kind), Location: r.Location, DefaultBranch: r.DefaultBranch, Status: r.Status, Head: r.Head}
}

// repoLocation resolves a repository id to its working directory on this server.
func repoLocation(path, id string) (string, error) {
	r, err := GetRepository(path, id)
	if err != nil {
		return "", err
	}
	if r.Location == "" {
		return path, nil
	}
	return r.Location, nil
}

// GitStatus returns the working-tree state of repository id (ADR-0022 §3).
func GitStatus(path, id string) (GitStatusDTO, error) {
	loc, err := repoLocation(path, id)
	if err != nil {
		return GitStatusDTO{}, err
	}
	st, err := repo.Status(loc)
	if err != nil {
		return GitStatusDTO{}, Error(CodeInternalError, err.Error(), http.StatusInternalServerError, err)
	}
	out := GitStatusDTO{Initialized: st.Initialized, Branch: st.Branch, Head: st.Head, Dirty: st.Dirty, Files: []GitFileChangeDTO{}}
	for _, f := range st.Files {
		out.Files = append(out.Files, GitFileChangeDTO{Code: f.Code, Path: f.Path})
	}
	return out, nil
}

// CommitRepo stages and commits all working-tree changes of repository id.
func CommitRepo(path, id, message string) (GitStatusDTO, error) {
	loc, err := repoLocation(path, id)
	if err != nil {
		return GitStatusDTO{}, err
	}
	if err := repo.Commit(loc, message); err != nil {
		return GitStatusDTO{}, Error(CodeInvalidInput, err.Error(), http.StatusBadRequest, err)
	}
	return GitStatus(path, id)
}

// GitBranches lists the local branches of repository id.
func GitBranches(path, id string) (GitBranchesDTO, error) {
	loc, err := repoLocation(path, id)
	if err != nil {
		return GitBranchesDTO{}, err
	}
	current, all, err := repo.Branches(loc)
	if err != nil {
		return GitBranchesDTO{}, Error(CodeInternalError, err.Error(), http.StatusInternalServerError, err)
	}
	if all == nil {
		all = []string{}
	}
	return GitBranchesDTO{Current: current, Branches: all}, nil
}

// CreateBranch creates a branch in repository id, optionally switching to it.
func CreateBranch(path, id, name string, checkout bool) (GitBranchesDTO, error) {
	loc, err := repoLocation(path, id)
	if err != nil {
		return GitBranchesDTO{}, err
	}
	if err := repo.CreateBranch(loc, name, checkout); err != nil {
		return GitBranchesDTO{}, Error(CodeInvalidInput, err.Error(), http.StatusBadRequest, err)
	}
	return GitBranches(path, id)
}

// CheckoutBranch switches repository id to an existing branch.
func CheckoutBranch(path, id, name string) (GitBranchesDTO, error) {
	loc, err := repoLocation(path, id)
	if err != nil {
		return GitBranchesDTO{}, err
	}
	if err := repo.Checkout(loc, name); err != nil {
		return GitBranchesDTO{}, Error(CodeInvalidInput, err.Error(), http.StatusBadRequest, err)
	}
	return GitBranches(path, id)
}

// GitTags lists the tags (releases) of repository id.
func GitTags(path, id string) (GitTagsDTO, error) {
	loc, err := repoLocation(path, id)
	if err != nil {
		return GitTagsDTO{}, err
	}
	tags, err := repo.Tags(loc)
	if err != nil {
		return GitTagsDTO{}, Error(CodeInternalError, err.Error(), http.StatusInternalServerError, err)
	}
	out := GitTagsDTO{Tags: []GitTagDTO{}}
	for _, t := range tags {
		out.Tags = append(out.Tags, GitTagDTO{Name: t.Name, Message: t.Message})
	}
	return out, nil
}

// CreateTag creates a tag (release) at HEAD of repository id.
func CreateTag(path, id, name, message string) (GitTagsDTO, error) {
	loc, err := repoLocation(path, id)
	if err != nil {
		return GitTagsDTO{}, err
	}
	if err := repo.CreateTag(loc, name, message); err != nil {
		return GitTagsDTO{}, Error(CodeInvalidInput, err.Error(), http.StatusBadRequest, err)
	}
	return GitTags(path, id)
}
