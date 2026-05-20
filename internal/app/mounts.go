package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/nomos/nomos/internal/mount"
)

// MountDTO is one server mount surfaced by the API (ADR-0022 §5).
type MountDTO struct {
	ID       string `json:"id"`
	Endpoint string `json:"endpoint"`
	Label    string `json:"label,omitempty"`
	Local    bool   `json:"local"`
}

type MountsDTO struct {
	Mounts []MountDTO `json:"mounts"`
}

func localMountDTO() MountDTO {
	return MountDTO{ID: mount.LocalID, Endpoint: LocalServerEndpoint, Label: "Local", Local: true}
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
		out.Mounts = append(out.Mounts, MountDTO{ID: m.ID, Endpoint: m.Endpoint, Label: m.Label})
	}
	return out, nil
}

// AddMount registers a remote server mount.
func AddMount(path, endpoint, label string) (MountDTO, error) {
	m, err := mount.NewStore(path).Add(mount.Mount{Endpoint: endpoint, Label: label})
	if err != nil {
		status := http.StatusBadRequest
		code := CodeInvalidInput
		if errors.Is(err, mount.ErrExists) {
			status, code = http.StatusConflict, CodeMountExists
		}
		return MountDTO{}, Error(code, err.Error(), status, err)
	}
	return MountDTO{ID: m.ID, Endpoint: m.Endpoint, Label: m.Label}, nil
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

var explorerHTTPClient = &http.Client{Timeout: 5 * time.Second}

func fetchRemoteRepositories(endpoint string) (RepositoriesDTO, error) {
	var out RepositoriesDTO
	err := getRemoteJSON("http://"+endpoint+"/api/v1/repositories", &out)
	return out, err
}

func fetchRemoteNamespaces(endpoint, repoID string) (NamespaceTreeDTO, error) {
	var out NamespaceTreeDTO
	err := getRemoteJSON("http://"+endpoint+"/api/v1/repositories/"+url.PathEscape(repoID)+"/namespaces", &out)
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
