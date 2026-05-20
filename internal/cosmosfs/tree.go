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
	Services      []ServiceNode
	Decisions     []DecisionNode
	Blueprints    []BlueprintNode
	Instances     []InstanceNode
	Servicegraphs []ServicegraphNode
}

type DecisionNode struct {
	Path     string // directory containing decision.yaml
	DMNPath  string // path to .dmn file, empty if not present
	Metadata model.Decision
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
	services, err := scanServices(path)
	if err != nil {
		return Tree{}, err
	}
	tree.Services = services
	decisions, err := scanDecisions(storage.DecisionsDir(path))
	if err != nil {
		return Tree{}, err
	}
	tree.Decisions = decisions
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
		if b.Type == "product_blueprint" || b.Type == "service_blueprint" || b.Type == "product" || b.Type == "service" {
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

// scanServices walks the services root recursively so that artifacts may be
// freely organized into nested folders (ADR: folders are a human/git-facing
// organization layer; identity is the stable service ID resolved via the index,
// the directory path is only the derived address). Any directory that holds a
// service.yaml is a service node; descent stops there so a service's own
// subdirectories (capabilities/, etc.) are never mistaken for nested services.
func scanServices(path string) ([]ServiceNode, error) {
	root := storage.ServicesDir(path)
	var nodes []ServiceNode
	err := filepath.WalkDir(root, func(dir string, de os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return filepath.SkipAll
			}
			return walkErr
		}
		if !de.IsDir() {
			return nil
		}
		yamlFile := filepath.Join(dir, "service.yaml")
		if _, err := os.Stat(yamlFile); err != nil {
			return nil
		}
		var s model.Service
		if err := fsx.ReadYAML(yamlFile, &s); err != nil {
			return err
		}
		nodes = append(nodes, ServiceNode{Path: dir, Metadata: s, Name: firstNonEmpty(s.Name, filepath.Base(dir))})
		return filepath.SkipDir
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Name < nodes[j].Name })
	return nodes, nil
}

// ScanDecisions reads all Decision artifacts under the given decisions root.
func ScanDecisions(root string) ([]DecisionNode, error) {
	return scanDecisions(root)
}

// scanDecisions walks the decisions root recursively (free folder nesting).
// A directory holding a decision.yaml is a decision node; descent stops there so
// per-version snapshot directories (versions/<v>/decision.yaml) are not picked
// up as separate decisions.
func scanDecisions(root string) ([]DecisionNode, error) {
	var nodes []DecisionNode
	err := filepath.WalkDir(root, func(dir string, de os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return filepath.SkipAll
			}
			return walkErr
		}
		if !de.IsDir() {
			return nil
		}
		yamlFile := filepath.Join(dir, "decision.yaml")
		if _, err := os.Stat(yamlFile); err != nil {
			return nil
		}
		var d model.Decision
		if err := fsx.ReadYAML(yamlFile, &d); err != nil {
			return err
		}
		dn := DecisionNode{Path: dir, Metadata: d}
		dmnFile := filepath.Join(dir, "decision.dmn")
		if _, err := os.Stat(dmnFile); err == nil {
			dn.DMNPath = dmnFile
		}
		nodes = append(nodes, dn)
		return filepath.SkipDir
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Metadata.ID < nodes[j].Metadata.ID })
	return nodes, nil
}

func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
