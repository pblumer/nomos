package app

import (
	"net/http"
	"sort"

	"github.com/nomos/nomos/internal/namespace"
)

// IndexEntryDTO maps a stable artifact ID to its current location/address
// (ADR-0028). The address is derived and may change when content moves; the ID
// is the stable reference.
type IndexEntryDTO struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Name    string `json:"name,omitempty"`
	Address string `json:"address"`
	Domain  string `json:"domain,omitempty"`
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
	for _, d := range tree.Domains {
		canonical := namespace.Canonical(d.Name)
		for _, s := range d.Services {
			svcAddr := canonical + "/" + s.Name
			add(IndexEntryDTO{ID: s.Metadata.ID, Kind: "service", Name: s.Name, Address: svcAddr, Domain: canonical, Path: s.Path})
			for _, c := range s.Metadata.Capabilities {
				add(IndexEntryDTO{ID: c.ID, Kind: "capability", Name: c.Name, Address: svcAddr + "/" + c.ID, Domain: canonical, Path: s.Path})
			}
			for _, o := range s.Metadata.DataObjects {
				add(IndexEntryDTO{ID: o.ID, Kind: "data-object", Name: o.Name, Address: svcAddr + "/" + o.ID, Domain: canonical, Path: s.Path})
			}
			for _, u := range s.Metadata.UserInterfaces {
				add(IndexEntryDTO{ID: u.ID, Kind: "user-interface", Name: u.Name, Address: svcAddr + "/" + u.ID, Domain: canonical, Path: s.Path})
			}
		}
		for _, dec := range d.Decisions {
			add(IndexEntryDTO{ID: dec.Metadata.ID, Kind: "decision", Name: dec.Metadata.Name, Address: canonical + "/decisions/" + dec.Metadata.ID, Domain: canonical, Path: dec.Path})
		}
	}
	for _, b := range tree.Blueprints {
		kind := "blueprint"
		if b.Metadata.Type == "product_blueprint" {
			kind = "product"
		}
		add(IndexEntryDTO{ID: b.Metadata.ID, Kind: kind, Name: b.Metadata.Name, Address: b.Metadata.ID, Domain: b.Metadata.OfferedBy, Path: b.Path})
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
