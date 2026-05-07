package cosmosfs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

type Tree struct {
	Path          string
	Cosmos        model.Cosmos
	Domains       []DomainNode
	Blueprints    []BlueprintNode
	Instances     []InstanceNode
	Servicegraphs []ServicegraphNode
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
type BlueprintNode struct {
	Path     string
	Metadata model.Blueprint
}
type InstanceNode struct {
	Path     string
	Metadata model.Instance
}
type ServicegraphNode struct {
	Path     string
	Metadata model.Servicegraph
}

func LoadTree(path string) (Tree, error) {
	var co model.Cosmos
	cosmosFile, _ := storage.CosmosFileForRead(path)
	if err := fsx.ReadYAML(cosmosFile, &co); err != nil {
		return Tree{}, err
	}
	tree := Tree{Path: path, Cosmos: co}
	ents, err := os.ReadDir(storage.DomainsDirForRead(path))
	if err != nil && !os.IsNotExist(err) {
		return Tree{}, err
	}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(storage.DomainsDirForRead(path), e.Name())
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
	blueprints, err := scanBlueprints(path)
	if err != nil {
		return Tree{}, err
	}
	instances, err := scanInstances(path)
	if err != nil {
		return Tree{}, err
	}
	tree.Blueprints = blueprints
	tree.Instances = instances
	servicegraphs, err := scanServicegraphs(path)
	if err != nil {
		return Tree{}, err
	}
	tree.Servicegraphs = servicegraphs
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

func scanBlueprints(path string) ([]BlueprintNode, error) {
	return scanBlueprintArtifacts(filepath.Join(storage.CatalogDirForRead(path), "blueprints"))
}

func scanBlueprintArtifacts(root string) ([]BlueprintNode, error) {
	var nodes []BlueprintNode
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nodes, nil
		}
		return nil, err
	}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isYAML(path) {
			return nil
		}
		var b model.Blueprint
		if err := fsx.ReadYAML(path, &b); err != nil {
			return err
		}
		if b.Type == "product_blueprint" || b.Type == "service_blueprint" {
			nodes = append(nodes, BlueprintNode{Path: path, Metadata: b})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Metadata.ID < nodes[j].Metadata.ID })
	return nodes, nil
}

func scanInstances(path string) ([]InstanceNode, error) {
	return scanInstanceArtifacts(filepath.Join(storage.CatalogDirForRead(path), "instances"))
}

func scanInstanceArtifacts(root string) ([]InstanceNode, error) {
	var nodes []InstanceNode
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nodes, nil
		}
		return nil, err
	}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isYAML(path) {
			return nil
		}
		var inst model.Instance
		if err := fsx.ReadYAML(path, &inst); err != nil {
			return err
		}
		if inst.Type == "product_instance" || inst.Type == "service_instance" {
			nodes = append(nodes, InstanceNode{Path: path, Metadata: inst})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Metadata.ID < nodes[j].Metadata.ID })
	return nodes, nil
}

func scanServicegraphs(path string) ([]ServicegraphNode, error) {
	root := storage.CatalogServicegraphsDirForRead(path)
	var nodes []ServicegraphNode
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nodes, nil
		}
		return nil, err
	}
	ents, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, e := range ents {
		if e.IsDir() || !isYAML(e.Name()) {
			continue
		}
		full := filepath.Join(root, e.Name())
		var sg model.Servicegraph
		if err := fsx.ReadYAML(full, &sg); err != nil {
			return nil, err
		}
		if sg.Type == "servicegraph" {
			nodes = append(nodes, ServicegraphNode{Path: full, Metadata: sg})
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Metadata.ID < nodes[j].Metadata.ID })
	return nodes, nil
}

func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
