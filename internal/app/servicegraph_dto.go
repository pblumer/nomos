package app

import "github.com/nomos/nomos/internal/model"

type ServicegraphDTO struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	Name           string            `json:"name"`
	Version        string            `json:"version"`
	Status         string            `json:"status"`
	Owner          string            `json:"owner"`
	Summary        string            `json:"summary"`
	RelatedProduct string            `json:"related_product"`
	Path           string            `json:"path"`
	NodeCount      int               `json:"node_count"`
	EdgeCount      int               `json:"edge_count"`
	RuleCount      int               `json:"rule_count"`
	Variants       []GraphVariantDTO `json:"variants,omitempty"`
	Nodes          []model.GraphNode `json:"nodes"`
	Edges          []model.GraphEdge `json:"edges"`
	Rules          []model.GraphRule `json:"rules,omitempty"`
}

type ServicegraphsDTO struct {
	Servicegraphs []ServicegraphDTO `json:"servicegraphs"`
	Count         int               `json:"count"`
}

type GraphVariantDTO struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Context string `json:"context,omitempty"`
}

type ExecutionOrderDTO struct {
	ServicegraphID string     `json:"servicegraph_id"`
	Steps          [][]string `json:"steps"`
}
