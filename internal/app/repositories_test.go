package app

import (
	"context"
	"errors"
	"testing"

	"github.com/nomos/nomos/internal/idgen"
	"github.com/nomos/nomos/internal/storage"
)

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
	// Single repository → the physical working directory is mirrored directly
	// under the server (no repo level): .nomos/ → services/ → service leaves.
	if findTreeNode(*server, "folder", "services") == nil {
		t.Fatal("server content should mirror the physical services/ folder")
	}
	if findTreeNode(*server, "service", "user-account") == nil {
		t.Fatal("server content should include service leaves")
	}
}

func TestBuildExplorerTreePlacesLocalServerUnderDomain(t *testing.T) {
	t.Setenv("NOMOS_DOMAIN", "nomos.blumer.cloud")
	stubDomainOwnership(t, "nomos.blumer.cloud")
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

func TestBuildExplorerTreeAutoDetectsDomainFromHost(t *testing.T) {
	// No NOMOS_DOMAIN: the public domain is taken from the request Host header
	// (with port) and honored once ownership is provable.
	t.Setenv("NOMOS_DOMAIN", "")
	stubDomainOwnership(t, "nomos.blumer.cloud")
	tree, err := BuildExplorerTreeForHost(createAppTestCosmos(t), "nomos.blumer.cloud:443")
	if err != nil {
		t.Fatal(err)
	}
	cloud := findTreeNode(tree.Root, "dns", "cloud")
	if cloud == nil || findTreeNode(*cloud, "dns", "blumer") == nil {
		t.Fatal("Host header should auto-place the verified domain in the DNS hierarchy")
	}
	if findTreeNode(tree.Root, "dns", "local") != nil {
		t.Fatal("server should not also appear under 'local'")
	}
}

func TestBuildExplorerTreeIgnoresUnverifiedHost(t *testing.T) {
	t.Setenv("NOMOS_DOMAIN", "")
	stubDomainOwnership(t, "") // host header not provable
	tree, err := BuildExplorerTreeForHost(createAppTestCosmos(t), "nomos.blumer.cloud")
	if err != nil {
		t.Fatal(err)
	}
	if findTreeNode(tree.Root, "dns", "cloud") != nil {
		t.Fatal("an unverified Host header must not be promoted")
	}
	if findTreeNode(tree.Root, "dns", "local") == nil {
		t.Fatal("expected fallback to the 'local' branch")
	}
}

func TestBuildExplorerTreeVerifiesViaParentZone(t *testing.T) {
	t.Setenv("NOMOS_DOMAIN", "nomos.blumer.cloud")
	// Only the registrable parent zone carries the record.
	stubDomainOwnership(t, "blumer.cloud")
	tree, err := BuildExplorerTree(createAppTestCosmos(t))
	if err != nil {
		t.Fatal(err)
	}
	cloud := findTreeNode(tree.Root, "dns", "cloud")
	if cloud == nil || findTreeNode(*cloud, "dns", "blumer") == nil {
		t.Fatal("a record on the parent zone should verify the subdomain")
	}
}

func TestBuildExplorerTreeIgnoresUnverifiedDomain(t *testing.T) {
	t.Setenv("NOMOS_DOMAIN", "nomos.blumer.cloud")
	stubDomainOwnership(t, "") // no record anywhere
	tree, err := BuildExplorerTree(createAppTestCosmos(t))
	if err != nil {
		t.Fatal(err)
	}
	// Unverified domain must not be promoted into the DNS hierarchy; the server
	// falls back under the synthetic "local" branch.
	if findTreeNode(tree.Root, "dns", "cloud") != nil {
		t.Fatal("unverified domain should not appear in the DNS hierarchy")
	}
	if findTreeNode(tree.Root, "dns", "local") == nil {
		t.Fatal("expected fallback to the 'local' branch for an unverified domain")
	}
}

func TestBuildExplorerTreeRejectsTLDOnlyRecord(t *testing.T) {
	t.Setenv("NOMOS_DOMAIN", "nomos.blumer.cloud")
	// A record on the bare TLD must never grant ownership of a subdomain.
	stubDomainOwnership(t, "cloud")
	tree, err := BuildExplorerTree(createAppTestCosmos(t))
	if err != nil {
		t.Fatal(err)
	}
	if findTreeNode(tree.Root, "dns", "cloud") != nil {
		t.Fatal("a TLD-only record must not verify the domain")
	}
}

// stubDomainOwnership overrides the TXT resolver and resets the verification
// cache so that exactly _nomos.<recordDomain> carries the proof for one test;
// pass "" for a domain that is provable nowhere.
func stubDomainOwnership(t *testing.T, recordDomain string) {
	t.Helper()
	prev := lookupTXT
	lookupTXT = func(_ context.Context, name string) ([]string, error) {
		if recordDomain != "" && name == "_nomos."+recordDomain {
			return []string{"nomos-domain=" + recordDomain}, nil
		}
		return nil, errors.New("no record")
	}
	domainVerifyMu.Lock()
	domainVerifyCache = map[string]domainVerifyResult{}
	domainVerifyMu.Unlock()
	t.Cleanup(func() {
		lookupTXT = prev
		domainVerifyMu.Lock()
		domainVerifyCache = map[string]domainVerifyResult{}
		domainVerifyMu.Unlock()
	})
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
	t.Setenv(storage.ReposDirEnv, t.TempDir()) // keep repos out of the real ~/.nomos-repos
	p := createAppTestCosmos(t)
	created, err := CreateRepository(p, "Team Beta")
	if err != nil {
		t.Fatal(err)
	}
	// The id is system-generated per ADR-0020 (REP_ prefix), not slug-derived.
	if !idgen.IsValidForType(created.ID, "repository") || created.Kind != "filesystem" {
		t.Fatalf("unexpected created repo: %+v", created)
	}
	repos, _ := ListRepositories(p)
	var ids []string
	for _, r := range repos.Repositories {
		ids = append(ids, r.ID)
	}
	if len(repos.Repositories) != 2 {
		t.Fatalf("expected default + new repo, got %v", ids)
	}
	// New repo resolves and has its own (empty) cosmos.
	got, err := GetRepository(p, created.ID)
	if err != nil || got.Location == "" {
		t.Fatalf("GetRepository(%s) = %+v err %v", created.ID, got, err)
	}
	co, err := GetCosmos(got.Location)
	if err != nil || co.Name != "Team Beta" {
		t.Fatalf("new repo cosmos = %+v err %v", co, err)
	}
	// Names are mutable display labels (ADR-0020), not identities: creating a
	// second repository with the same name succeeds with a distinct id.
	dup, err := CreateRepository(p, "Team Beta")
	if err != nil {
		t.Fatalf("second repository with a duplicate name should succeed: %v", err)
	}
	if dup.ID == created.ID {
		t.Fatalf("expected a distinct id for the duplicate-named repo, got %q twice", dup.ID)
	}
}

func TestExplorerTreeShowsMultipleRepositories(t *testing.T) {
	t.Setenv(storage.ReposDirEnv, t.TempDir())
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
