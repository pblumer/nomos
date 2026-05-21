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
	res, err := classifyWalk(path)
	if err != nil {
		return Tree{}, err
	}
	tree.Services = res.services
	tree.Decisions = res.decisions
	tree.Blueprints = res.blueprints
	tree.Instances = res.instances
	tree.Servicegraphs = res.servicegraphs
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

// ScanDecisions reads all Decision artifacts under the given root, recognizing
// them by content anywhere in the tree (folders are a free organization layer).
func ScanDecisions(root string) ([]DecisionNode, error) {
	res, err := classifyWalk(root)
	if err != nil {
		return nil, err
	}
	return res.decisions, nil
}

type walkResult struct {
	services      []ServiceNode
	decisions     []DecisionNode
	blueprints    []BlueprintNode
	instances     []InstanceNode
	servicegraphs []ServicegraphNode
}

func isBlueprintType(t string) bool {
	switch t {
	case "product_blueprint", "service_blueprint", "product", "service":
		return true
	}
	return false
}

func isInstanceType(t string) bool {
	return t == "product_instance" || t == "service_instance"
}

// configFiles are workspace metadata, never catalog artifacts.
var configFiles = map[string]bool{
	"cosmos.yaml":        true,
	"repositories.yaml":  true,
	"mounts.yaml":        true,
	"keys.yaml":          true,
	".nomos.folder.yaml": true,
}

// plumbingDirs lists derived directories that never hold authored artifacts.
func plumbingDirs(root string) map[string]bool {
	return map[string]bool{
		filepath.Clean(storage.ReposDir(root)): true,
		filepath.Clean(storage.CacheDir(root)): true,
		filepath.Clean(storage.IndexDir(root)): true,
	}
}

// classifyWalk walks root once and recognizes artifacts by content anywhere in
// the tree (ADR: below the repository the git/filesystem layout is the source
// of truth; identity follows artifact content, not a fixed directory). A
// directory holding service.yaml is a service and a directory holding
// decision.yaml is a decision — descent stops there so their internal layout
// (capabilities/, version snapshots, …) is never rescanned. Remaining YAML
// files are classified by their `type` field. The git database and derived
// Nomos directories are skipped.
func classifyWalk(root string) (walkResult, error) {
	var res walkResult
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return res, nil
		}
		return res, err
	}
	skip := plumbingDirs(root)
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return filepath.SkipAll
			}
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".git" || skip[filepath.Clean(p)] {
				return filepath.SkipDir
			}
			if _, err := os.Stat(filepath.Join(p, "service.yaml")); err == nil {
				var s model.Service
				if err := fsx.ReadYAML(filepath.Join(p, "service.yaml"), &s); err != nil {
					return err
				}
				res.services = append(res.services, ServiceNode{Path: p, Metadata: s, Name: firstNonEmpty(s.Name, filepath.Base(p))})
				return filepath.SkipDir
			}
			if _, err := os.Stat(filepath.Join(p, "decision.yaml")); err == nil {
				var dec model.Decision
				if err := fsx.ReadYAML(filepath.Join(p, "decision.yaml"), &dec); err != nil {
					return err
				}
				dn := DecisionNode{Path: p, Metadata: dec}
				if _, err := os.Stat(filepath.Join(p, "decision.dmn")); err == nil {
					dn.DMNPath = filepath.Join(p, "decision.dmn")
				}
				res.decisions = append(res.decisions, dn)
				return filepath.SkipDir
			}
			return nil
		}
		if !isYAML(p) || configFiles[d.Name()] {
			return nil
		}
		var probe struct {
			Type string `yaml:"type"`
		}
		if err := fsx.ReadYAML(p, &probe); err != nil {
			return nil // unreadable / non-conforming YAML is not an artifact
		}
		switch {
		case isBlueprintType(probe.Type):
			var b model.Blueprint
			if err := fsx.ReadYAML(p, &b); err != nil {
				return err
			}
			res.blueprints = append(res.blueprints, BlueprintNode{Path: p, Metadata: b})
		case isInstanceType(probe.Type):
			var inst model.Instance
			if err := fsx.ReadYAML(p, &inst); err != nil {
				return err
			}
			res.instances = append(res.instances, InstanceNode{Path: p, Metadata: inst})
		case probe.Type == "servicegraph":
			var sg model.Servicegraph
			if err := fsx.ReadYAML(p, &sg); err != nil {
				return err
			}
			res.servicegraphs = append(res.servicegraphs, ServicegraphNode{Path: p, Metadata: sg})
		}
		return nil
	})
	if err != nil {
		return res, err
	}
	sort.Slice(res.services, func(i, j int) bool { return res.services[i].Name < res.services[j].Name })
	sort.Slice(res.decisions, func(i, j int) bool { return res.decisions[i].Metadata.ID < res.decisions[j].Metadata.ID })
	sort.Slice(res.blueprints, func(i, j int) bool { return res.blueprints[i].Metadata.ID < res.blueprints[j].Metadata.ID })
	sort.Slice(res.instances, func(i, j int) bool { return res.instances[i].Metadata.ID < res.instances[j].Metadata.ID })
	sort.Slice(res.servicegraphs, func(i, j int) bool { return res.servicegraphs[i].Metadata.ID < res.servicegraphs[j].Metadata.ID })
	return res, nil
}

func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
