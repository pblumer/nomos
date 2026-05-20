package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListAndAddMounts(t *testing.T) {
	p := createAppTestCosmos(t)
	mounts, err := ListMounts(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(mounts.Mounts) != 1 || !mounts.Mounts[0].Local {
		t.Fatalf("expected only local mount, got %+v", mounts.Mounts)
	}

	if _, err := AddMount(p, "nomos.blumer.cloud:7373", "Prod", ""); err != nil {
		t.Fatal(err)
	}
	mounts, _ = ListMounts(p)
	if len(mounts.Mounts) != 2 || mounts.Mounts[1].Endpoint != "nomos.blumer.cloud:7373" {
		t.Fatalf("expected local + remote mount, got %+v", mounts.Mounts)
	}
}

func TestBuildExplorerTreeAggregatesRemoteMount(t *testing.T) {
	p := createAppTestCosmos(t)
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/repositories":
			w.Write([]byte(`{"repositories":[{"id":"default","name":"Remote Cosmos","kind":"filesystem"}]}`))
		case "/api/v1/repositories/default/namespaces":
			w.Write([]byte(`{"root":{"label":"Remote Cosmos","kind":"cosmos","children":[{"label":"Namespaces","kind":"namespace-parent","children":[{"label":"example","kind":"namespace"}]}]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer remote.Close()
	endpoint := strings.TrimPrefix(remote.URL, "http://")

	if _, err := AddMount(p, endpoint, "Remote", ""); err != nil {
		t.Fatal(err)
	}
	tree, err := BuildExplorerTree(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Root.Children) != 2 {
		t.Fatalf("expected local + remote server, got %d", len(tree.Root.Children))
	}
	remoteServer := tree.Root.Children[1]
	if remoteServer.Kind != "server" || remoteServer.Server == nil || remoteServer.Server.Status != "online" {
		t.Fatalf("expected online remote server, got %+v", remoteServer)
	}
	if findTreeNode(remoteServer, "namespace", "example") == nil {
		t.Fatal("remote repository content should be aggregated into the tree")
	}
}

func TestBuildExplorerTreeDegradesUnreachableMount(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddMount(p, "127.0.0.1:9", "Down", ""); err != nil {
		t.Fatal(err)
	}
	tree, err := BuildExplorerTree(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Root.Children) != 2 {
		t.Fatalf("expected local + remote server, got %d", len(tree.Root.Children))
	}
	down := tree.Root.Children[1]
	if down.Server == nil || down.Server.Status != "unreachable" || len(down.Children) != 0 {
		t.Fatalf("expected degraded unreachable server with no children, got %+v", down)
	}
}

func TestPingReturnsIdentityAndPeers(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddMount(p, "nomos.blumer.cloud:7373", "Prod", ""); err != nil {
		t.Fatal(err)
	}
	ping, err := Ping(p)
	if err != nil {
		t.Fatal(err)
	}
	if ping.Server.Name == "" {
		t.Fatal("expected server name in ping")
	}
	if len(ping.Peers) != 1 || ping.Peers[0].Endpoint != "nomos.blumer.cloud:7373" {
		t.Fatalf("expected one advertised peer, got %+v", ping.Peers)
	}
}

func TestDiscoverServersAggregatesPeers(t *testing.T) {
	p := createAppTestCosmos(t)
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/ping" {
			w.Write([]byte(`{"server":{"name":"Remote","version":"0.1.0","repositoryCount":1},"peers":[{"endpoint":"newserver:7373","label":"New"}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer remote.Close()
	endpoint := strings.TrimPrefix(remote.URL, "http://")
	if _, err := AddMount(p, endpoint, "Remote", ""); err != nil {
		t.Fatal(err)
	}
	disc, err := DiscoverServers(p)
	if err != nil {
		t.Fatal(err)
	}
	var found *DiscoveredServerDTO
	for i := range disc.Servers {
		if disc.Servers[i].Endpoint == "newserver:7373" {
			found = &disc.Servers[i]
		}
		if disc.Servers[i].Endpoint == endpoint {
			t.Fatal("already-mounted server should not be a candidate")
		}
	}
	if found == nil || found.Mounted || found.Via != endpoint {
		t.Fatalf("expected newserver candidate via %s, got %+v", endpoint, disc.Servers)
	}
}

func TestRemoteBaseURL(t *testing.T) {
	cases := map[string]string{
		"host:7373":                  "http://host:7373",
		"https://nomos.blumer.cloud": "https://nomos.blumer.cloud",
		"http://x:7373/":             "http://x:7373",
	}
	for in, want := range cases {
		if got := RemoteBaseURL(in); got != want {
			t.Errorf("RemoteBaseURL(%q) = %q, want %q", in, got, want)
		}
	}
}
