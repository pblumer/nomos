package app

import (
	"net/http"
	"path/filepath"
	"sort"

	"github.com/nomos/nomos/internal/storage"
)

// derivedAddress returns the artifact's location relative to the cosmos .nomos
// directory. Folders are a free human/git organization layer (ADR-0027); the
// address is derived from wherever the artifact currently lives, while the
// stable ID is the durable reference resolved through this index.
func derivedAddress(cosmosPath, artifactPath string) string {
	rel, err := filepath.Rel(storage.NomosDir(cosmosPath), artifactPath)
	if err != nil {
		return artifactPath
	}
	return filepath.ToSlash(rel)
}

// IndexEntryDTO maps a stable artifact ID to its current location/address
// (ADR-0028). The address is derived and may change when content moves; the ID
// is the stable reference.
type IndexEntryDTO struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Name    string `json:"name,omitempty"`
	Address string `json:"address"`
	Path    string `json:"path,omitempty"`
}

type IndexDTO struct {
	Entries []IndexEntryDTO `json:"entries"`
}

// BuildIDIndex scans the cosmos and indexes every artifact that carries a stable
// ID, mapping it to its current derived address (ADR-0028).
func BuildIDIndex(path string) (IndexDTO, error) {
	tree, err := load(path)
	if err != nil {
		return IndexDTO{}, err
	}
	out := IndexDTO{Entries: []IndexEntryDTO{}}
	add := func(e IndexEntryDTO) {
		if e.ID != "" {
			out.Entries = append(out.Entries, e)
		}
	}
	for _, s := range tree.Services {
		svcAddr := derivedAddress(path, s.Path)
		add(IndexEntryDTO{ID: s.Metadata.ID, Kind: "service", Name: s.Name, Address: svcAddr, Path: s.Path})
		for _, c := range s.Metadata.Capabilities {
			add(IndexEntryDTO{ID: c.ID, Kind: "capability", Name: c.Name, Address: svcAddr + "/" + c.ID, Path: s.Path})
		}
		for _, o := range s.Metadata.DataObjects {
			add(IndexEntryDTO{ID: o.ID, Kind: "data-object", Name: o.Name, Address: svcAddr + "/" + o.ID, Path: s.Path})
		}
		for _, u := range s.Metadata.UserInterfaces {
			add(IndexEntryDTO{ID: u.ID, Kind: "user-interface", Name: u.Name, Address: svcAddr + "/" + u.ID, Path: s.Path})
		}
	}
	for _, dec := range tree.Decisions {
		add(IndexEntryDTO{ID: dec.Metadata.ID, Kind: "decision", Name: dec.Metadata.Name, Address: derivedAddress(path, dec.Path), Path: dec.Path})
	}
	for _, b := range tree.Blueprints {
		kind := "blueprint"
		if b.Metadata.Type == "product_blueprint" {
			kind = "product"
		}
		add(IndexEntryDTO{ID: b.Metadata.ID, Kind: kind, Name: b.Metadata.Name, Address: b.Metadata.ID, Path: b.Path})
	}
	for _, i := range tree.Instances {
		add(IndexEntryDTO{ID: i.Metadata.ID, Kind: "instance", Name: i.Metadata.Name, Address: i.Metadata.ID, Path: i.Path})
	}
	sort.Slice(out.Entries, func(a, b int) bool { return out.Entries[a].ID < out.Entries[b].ID })
	return out, nil
}

// ResolveID looks up an artifact by its stable ID and returns its current
// address (ADR-0028).
func ResolveID(path, id string) (IndexEntryDTO, error) {
	idx, err := BuildIDIndex(path)
	if err != nil {
		return IndexEntryDTO{}, err
	}
	for _, e := range idx.Entries {
		if e.ID == id {
			return e, nil
		}
	}
	return IndexEntryDTO{}, Error(CodeInvalidInput, "ID not found: "+id, http.StatusNotFound, nil)
}
