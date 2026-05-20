// Package mount manages the server mounts shown in the Cosmos Explorer
// (ADR-0022 §5). PR 4 implements a manual, config-backed store of remote server
// endpoints. The Resolver interface keeps the mount source pluggable so a
// .well-known/nomos federation source can be added later without changing the
// tree builder.
package mount

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nomos/nomos/internal/storage"
	"gopkg.in/yaml.v3"
)

// LocalID is the id of the implicit, non-removable local server mount.
const LocalID = "local"

var (
	ErrNotFound      = errors.New("mount not found")
	ErrExists        = errors.New("mount already exists")
	ErrLocalReserved = errors.New("the local server mount is reserved")
)

// Mount is one Nomos server reachable in the Explorer tree.
type Mount struct {
	ID       string `yaml:"id" json:"id"`
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	Label    string `yaml:"label,omitempty" json:"label,omitempty"`
	Local    bool   `yaml:"-" json:"local"`
}

// Resolver yields the configured server mounts. Implementations may read a
// config file (Store) or, in the future, a federation discovery source.
type Resolver interface {
	Mounts() ([]Mount, error)
}

// Store is a config-backed Resolver persisting remote mounts in
// .nomos/mounts.yaml. The local mount is implicit and not stored.
type Store struct {
	workspace string
}

func NewStore(workspace string) *Store { return &Store{workspace: workspace} }

type mountsFile struct {
	Mounts []Mount `yaml:"mounts"`
}

func (s *Store) load() (mountsFile, error) {
	var mf mountsFile
	data, err := os.ReadFile(storage.MountsFile(s.workspace))
	if err != nil {
		if os.IsNotExist(err) {
			return mf, nil
		}
		return mf, err
	}
	if err := yaml.Unmarshal(data, &mf); err != nil {
		return mountsFile{}, err
	}
	return mf, nil
}

func (s *Store) save(mf mountsFile) error {
	if err := os.MkdirAll(storage.NomosDir(s.workspace), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(mf)
	if err != nil {
		return err
	}
	return os.WriteFile(storage.MountsFile(s.workspace), data, 0o644)
}

// Mounts returns the persisted remote mounts (Resolver implementation).
func (s *Store) Mounts() ([]Mount, error) {
	mf, err := s.load()
	if err != nil {
		return nil, err
	}
	return mf.Mounts, nil
}

// Add registers a remote mount. The endpoint is required and must be unique;
// the id is derived from the endpoint when not given.
func (s *Store) Add(m Mount) (Mount, error) {
	m.Endpoint = strings.TrimSpace(m.Endpoint)
	if m.Endpoint == "" {
		return Mount{}, fmt.Errorf("endpoint is required")
	}
	if m.ID == "" {
		m.ID = idFromEndpoint(m.Endpoint)
	}
	if m.ID == LocalID {
		return Mount{}, ErrLocalReserved
	}
	mf, err := s.load()
	if err != nil {
		return Mount{}, err
	}
	for _, existing := range mf.Mounts {
		if existing.ID == m.ID || existing.Endpoint == m.Endpoint {
			return Mount{}, fmt.Errorf("%w: %s", ErrExists, m.Endpoint)
		}
	}
	mf.Mounts = append(mf.Mounts, m)
	if err := s.save(mf); err != nil {
		return Mount{}, err
	}
	return m, nil
}

// Remove deletes a remote mount by id. The local mount cannot be removed.
func (s *Store) Remove(id string) error {
	if id == LocalID {
		return ErrLocalReserved
	}
	mf, err := s.load()
	if err != nil {
		return err
	}
	out := mf.Mounts[:0]
	found := false
	for _, m := range mf.Mounts {
		if m.ID == id {
			found = true
			continue
		}
		out = append(out, m)
	}
	if !found {
		return fmt.Errorf("%w: %s", ErrNotFound, id)
	}
	mf.Mounts = out
	return s.save(mf)
}

func idFromEndpoint(endpoint string) string {
	id := strings.NewReplacer(":", "-", "/", "-", ".", "-").Replace(endpoint)
	return strings.Trim(id, "-")
}
