package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nomos/nomos/internal/mount"
	versionpkg "github.com/nomos/nomos/internal/version"
)

// MountDTO is one server mount surfaced by the API (ADR-0022 §5). The token is
// never exposed; Authenticated reports whether one is configured (ADR-0023).
type MountDTO struct {
	ID            string `json:"id"`
	Endpoint      string `json:"endpoint"`
	Label         string `json:"label,omitempty"`
	Local         bool   `json:"local"`
	Authenticated bool   `json:"authenticated"`
}

type MountsDTO struct {
	Mounts []MountDTO `json:"mounts"`
}

func localMountDTO() MountDTO {
	return MountDTO{ID: mount.LocalID, Endpoint: LocalServerEndpoint, Label: "Local", Local: true, Authenticated: true}
}

// ListMounts returns the implicit local server mount followed by the configured
// remote mounts.
func ListMounts(path string) (MountsDTO, error) {
	out := MountsDTO{Mounts: []MountDTO{localMountDTO()}}
	remotes, err := mount.NewStore(path).Mounts()
	if err != nil {
		return MountsDTO{}, Error(CodeInternalError, err.Error(), http.StatusInternalServerError, err)
	}
	for _, m := range remotes {
		out.Mounts = append(out.Mounts, MountDTO{ID: m.ID, Endpoint: m.Endpoint, Label: m.Label, Authenticated: m.Token != ""})
	}
	return out, nil
}

// AddMount registers a remote server mount. An optional token is the remote
// server's API key used for proxied writes (ADR-0023).
func AddMount(path, endpoint, label, token string) (MountDTO, error) {
	m, err := mount.NewStore(path).Add(mount.Mount{Endpoint: endpoint, Label: label, Token: token})
	if err != nil {
		status := http.StatusBadRequest
		code := CodeInvalidInput
		if errors.Is(err, mount.ErrExists) {
			status, code = http.StatusConflict, CodeMountExists
		}
		return MountDTO{}, Error(code, err.Error(), status, err)
	}
	return MountDTO{ID: m.ID, Endpoint: m.Endpoint, Label: m.Label, Authenticated: m.Token != ""}, nil
}

// RemoveMount unmounts a remote server by id. The local mount cannot be removed.
func RemoveMount(path, id string) error {
	err := mount.NewStore(path).Remove(id)
	if err == nil {
		return nil
	}
	status, code := http.StatusBadRequest, CodeInvalidInput
	if errors.Is(err, mount.ErrNotFound) {
		status, code = http.StatusNotFound, CodeMountNotFound
	}
	return Error(code, err.Error(), status, err)
}

// MountTarget resolves a remote mount to its endpoint and token for proxying
// writes (ADR-0023). The local mount is not a proxy target.
func MountTarget(path, mountID string) (endpoint, token string, err error) {
	if mountID == mount.LocalID {
		return "", "", Error(CodeInvalidInput, "the local server is not a proxy target", http.StatusBadRequest, nil)
	}
	m, ferr := mount.NewStore(path).Find(mountID)
	if ferr != nil {
		if errors.Is(ferr, mount.ErrNotFound) {
			return "", "", Error(CodeMountNotFound, ferr.Error(), http.StatusNotFound, ferr)
		}
		return "", "", Error(CodeInternalError, ferr.Error(), http.StatusInternalServerError, ferr)
	}
	return m.Endpoint, m.Token, nil
}

var explorerHTTPClient = &http.Client{Timeout: 5 * time.Second}

// RemoteBaseURL returns the base URL for a mount endpoint. A bare host:port
// defaults to http; an endpoint that already includes a scheme is used as-is,
// so HTTPS endpoints (e.g. https://nomos.blumer.cloud) work behind a TLS proxy.
func RemoteBaseURL(endpoint string) string {
	if strings.Contains(endpoint, "://") {
		return strings.TrimRight(endpoint, "/")
	}
	return "http://" + endpoint
}

func fetchRemoteRepositories(endpoint string) (RepositoriesDTO, error) {
	var out RepositoriesDTO
	err := getRemoteJSON(RemoteBaseURL(endpoint)+"/api/v1/repositories", &out)
	return out, err
}

func fetchRemoteNamespaces(endpoint, repoID string) (NamespaceTreeDTO, error) {
	var out NamespaceTreeDTO
	err := getRemoteJSON(RemoteBaseURL(endpoint)+"/api/v1/repositories/"+url.PathEscape(repoID)+"/namespaces", &out)
	return out, err
}

func getRemoteJSON(u string, v any) error {
	resp, err := explorerHTTPClient.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote returned status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func getRemoteJSONAuth(u, token string, v any) error {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("X-API-Key", token)
	}
	resp, err := explorerHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote returned status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// PingPeerDTO is one peer a server advertises (ADR-0025): an endpoint only,
// never a token.
type PingPeerDTO struct {
	Endpoint string `json:"endpoint"`
	Label    string `json:"label,omitempty"`
}

type PingServerDTO struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	RepositoryCount int    `json:"repositoryCount"`
}

// PingDTO is the response of GET /api/v1/ping (ADR-0025): the server's identity
// plus the endpoints of its configured mounts.
type PingDTO struct {
	Server PingServerDTO `json:"server"`
	Peers  []PingPeerDTO `json:"peers"`
}

// Ping returns this server's identity and its advertised peers (its remote
// mounts), without exposing any tokens.
func Ping(path string) (PingDTO, error) {
	cosmos, _ := GetCosmos(path)
	repos, _ := ListRepositories(path)
	out := PingDTO{
		Server: PingServerDTO{Name: fallback(cosmos.Name, "Local Cosmos"), Version: versionpkg.Get().Version, RepositoryCount: len(repos.Repositories)},
		Peers:  []PingPeerDTO{},
	}
	remotes, _ := mount.NewStore(path).Mounts()
	for _, m := range remotes {
		out.Peers = append(out.Peers, PingPeerDTO{Endpoint: m.Endpoint, Label: m.Label})
	}
	return out, nil
}

// DiscoveredServerDTO is a discovery candidate (ADR-0025).
type DiscoveredServerDTO struct {
	Endpoint  string `json:"endpoint"`
	Label     string `json:"label,omitempty"`
	Name      string `json:"name,omitempty"`
	Reachable bool   `json:"reachable"`
	Mounted   bool   `json:"mounted"`
	Via       string `json:"via,omitempty"`
}

type DiscoveryDTO struct {
	Servers []DiscoveredServerDTO `json:"servers"`
}

// DiscoverServers performs 1-hop peer-gossip discovery (ADR-0025): it pings the
// configured remote mounts (authenticated with their tokens), aggregates their
// advertised peers, and returns the not-yet-mounted endpoints as candidates.
func DiscoverServers(path string) (DiscoveryDTO, error) {
	remotes, err := mount.NewStore(path).Mounts()
	if err != nil {
		return DiscoveryDTO{}, Error(CodeInternalError, err.Error(), http.StatusInternalServerError, err)
	}
	known := map[string]bool{LocalServerEndpoint: true}
	for _, m := range remotes {
		known[m.Endpoint] = true
	}
	out := DiscoveryDTO{Servers: []DiscoveredServerDTO{}}
	seen := map[string]bool{}
	for _, m := range remotes {
		ping, err := fetchRemotePing(m.Endpoint, m.Token)
		if err != nil {
			continue
		}
		for _, peer := range ping.Peers {
			if peer.Endpoint == "" || known[peer.Endpoint] || seen[peer.Endpoint] {
				continue
			}
			seen[peer.Endpoint] = true
			cand := DiscoveredServerDTO{Endpoint: peer.Endpoint, Label: peer.Label, Via: m.Endpoint}
			if cp, perr := fetchRemotePing(peer.Endpoint, ""); perr == nil {
				cand.Reachable = true
				cand.Name = cp.Server.Name
			}
			out.Servers = append(out.Servers, cand)
		}
	}
	return out, nil
}

func fetchRemotePing(endpoint, token string) (PingDTO, error) {
	var out PingDTO
	err := getRemoteJSONAuth(RemoteBaseURL(endpoint)+"/api/v1/ping", token, &out)
	return out, err
}
