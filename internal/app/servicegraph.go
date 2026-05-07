package app

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/graph"
	"github.com/nomos/nomos/internal/model"
)

func ListServicegraphs(path string) (ServicegraphsDTO, error) {
	tree, err := load(path)
	if err != nil {
		return ServicegraphsDTO{}, err
	}
	out := ServicegraphsDTO{Servicegraphs: []ServicegraphDTO{}}
	for _, sg := range tree.Servicegraphs {
		out.Servicegraphs = append(out.Servicegraphs, servicegraphDTO(sg))
	}
	out.Count = len(out.Servicegraphs)
	return out, nil
}

func GetServicegraph(path, id string) (ServicegraphDTO, error) {
	items, err := ListServicegraphs(path)
	if err != nil {
		return ServicegraphDTO{}, err
	}
	for _, sg := range items.Servicegraphs {
		if sg.ID == id {
			return sg, nil
		}
	}
	return ServicegraphDTO{}, Error(CodeServicegraphNotFound, "Servicegraph not found: "+id, http.StatusNotFound, nil)
}

func CreateServicegraph(path string, sg model.Servicegraph) error {
	if _, err := os.Stat(filepath.Join(path, "cosmos.yaml")); err != nil {
		return Error(CodeCosmosMissing, "cosmos.yaml not found", http.StatusNotFound, err)
	}
	if strings.TrimSpace(sg.ID) == "" {
		return Error(CodeInvalidInput, "Servicegraph ID is required", http.StatusBadRequest, nil)
	}
	if sg.Type != "servicegraph" {
		return Error(CodeInvalidInput, "type must be servicegraph", http.StatusBadRequest, nil)
	}

	existing, err := ListServicegraphs(path)
	if err != nil {
		return err
	}
	for _, s := range existing.Servicegraphs {
		if s.ID == sg.ID {
			return Error(CodeServicegraphAlreadyExists, "Servicegraph already exists: "+sg.ID, http.StatusConflict, nil)
		}
	}

	dir := filepath.Join(path, "catalog", "servicegraphs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Error(CodeInternalError, "Failed to create servicegraph directory: "+err.Error(), http.StatusInternalServerError, err)
	}

	slug := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(sg.ID, " ", "-"), "_", "-"))
	fn := filepath.Join(dir, slug+".yaml")
	return fsx.WriteYAML(fn, sg)
}

func DeleteServicegraph(path, id string) error {
	tree, err := load(path)
	if err != nil {
		return err
	}
	for _, sg := range tree.Servicegraphs {
		if sg.Metadata.ID == id {
			if err := os.Remove(sg.Path); err != nil {
				return Error(CodeInternalError, "Failed to delete servicegraph: "+err.Error(), http.StatusInternalServerError, err)
			}
			return nil
		}
	}
	return Error(CodeServicegraphNotFound, "Servicegraph not found: "+id, http.StatusNotFound, nil)
}

func GetServicegraphMermaid(path, id string) (GraphDTO, error) {
	tree, err := load(path)
	if err != nil {
		return GraphDTO{}, err
	}
	for _, sg := range tree.Servicegraphs {
		if sg.Metadata.ID == id {
			return GraphDTO{Format: "mermaid", Content: graph.ServicegraphMermaid(sg.Metadata)}, nil
		}
	}
	return GraphDTO{}, Error(CodeServicegraphNotFound, "Servicegraph not found: "+id, http.StatusNotFound, nil)
}

func GetServicegraphExecutionOrder(path, id string) (ExecutionOrderDTO, error) {
	tree, err := load(path)
	if err != nil {
		return ExecutionOrderDTO{}, err
	}
	for _, sg := range tree.Servicegraphs {
		if sg.Metadata.ID == id {
			groups := graph.ParallelGroups(sg.Metadata)
			return ExecutionOrderDTO{ServicegraphID: id, Steps: groups}, nil
		}
	}
	return ExecutionOrderDTO{}, Error(CodeServicegraphNotFound, "Servicegraph not found: "+id, http.StatusNotFound, nil)
}

func servicegraphDTO(sg cosmosfs.ServicegraphNode) ServicegraphDTO {
	variants := make([]GraphVariantDTO, 0, len(sg.Metadata.Variants))
	for _, v := range sg.Metadata.Variants {
		variants = append(variants, GraphVariantDTO{ID: v.ID, Name: v.Name, Context: v.Context})
	}
	return ServicegraphDTO{
		ID:             sg.Metadata.ID,
		Type:           sg.Metadata.Type,
		Name:           sg.Metadata.Name,
		Version:        sg.Metadata.Version,
		Status:         sg.Metadata.Status,
		Owner:          sg.Metadata.Owner,
		Summary:        sg.Metadata.Summary,
		RelatedProduct: sg.Metadata.RelatedProduct,
		Path:           sg.Path,
		NodeCount:      len(sg.Metadata.Nodes),
		EdgeCount:      len(sg.Metadata.Edges),
		RuleCount:      len(sg.Metadata.Rules),
		Variants:       variants,
		Nodes:          sg.Metadata.Nodes,
		Edges:          sg.Metadata.Edges,
		Rules:          sg.Metadata.Rules,
	}
}
