package cosmosfs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
)

type Tree struct {
	Path    string
	Cosmos  model.Cosmos
	Domains []DomainNode
}
type DomainNode struct {
	Path     string
	Name     string
	Metadata model.Domain
	Services []ServiceNode
}
type ServiceNode struct {
	Path     string
	Name     string
	Metadata model.Service
}

func LoadTree(path string) (Tree, error) {
	var co model.Cosmos
	if err := fsx.ReadYAML(filepath.Join(path, "cosmos.yaml"), &co); err != nil {
		return Tree{}, err
	}
	tree := Tree{Path: path, Cosmos: co}
	ents, err := os.ReadDir(filepath.Join(path, "domains"))
	if err != nil {
		if os.IsNotExist(err) {
			return tree, nil
		}
		return Tree{}, err
	}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(path, "domains", e.Name())
		if _, err := os.Stat(filepath.Join(dir, "domain.yaml")); err != nil {
			continue
		}
		var d model.Domain
		if err := fsx.ReadYAML(filepath.Join(dir, "domain.yaml"), &d); err != nil {
			return Tree{}, err
		}
		dn := DomainNode{Path: dir, Metadata: d, Name: firstNonEmpty(d.Name, d.DNSName, e.Name())}
		sents, err := os.ReadDir(filepath.Join(dir, "services"))
		if err != nil && !os.IsNotExist(err) {
			return Tree{}, err
		}
		for _, se := range sents {
			if !se.IsDir() {
				continue
			}
			sdir := filepath.Join(dir, "services", se.Name())
			if _, err := os.Stat(filepath.Join(sdir, "service.yaml")); err != nil {
				continue
			}
			var s model.Service
			if err := fsx.ReadYAML(filepath.Join(sdir, "service.yaml"), &s); err != nil {
				return Tree{}, err
			}
			dn.Services = append(dn.Services, ServiceNode{Path: sdir, Metadata: s, Name: firstNonEmpty(s.Name, se.Name())})
		}
		sort.Slice(dn.Services, func(i, j int) bool { return dn.Services[i].Name < dn.Services[j].Name })
		tree.Domains = append(tree.Domains, dn)
	}
	sort.Slice(tree.Domains, func(i, j int) bool { return tree.Domains[i].Name < tree.Domains[j].Name })
	return tree, nil
}
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
