package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
)

// decisionsDir returns the decisions directory for a domain node.
func decisionsDir(domainPath string) string {
	return filepath.Join(domainPath, "decisions")
}

func findDomainNode(path, domainCanonical string) (cosmosfs.DomainNode, error) {
	tree, err := load(path)
	if err != nil {
		return cosmosfs.DomainNode{}, err
	}
	for _, d := range tree.Domains {
		if strings.EqualFold(d.Metadata.CanonicalName, domainCanonical) ||
			strings.EqualFold(d.Metadata.DNSName, domainCanonical) ||
			strings.EqualFold(d.Name, domainCanonical) {
			return d, nil
		}
	}
	return cosmosfs.DomainNode{}, Error(CodeDomainNotFound, "Domain not found: "+domainCanonical, http.StatusNotFound, nil)
}

func ListDecisions(path, domainCanonical string) (DecisionsDTO, error) {
	d, err := findDomainNode(path, domainCanonical)
	if err != nil {
		return DecisionsDTO{}, err
	}
	nodes, err := cosmosfs.ScanDecisions(d.Path)
	if err != nil {
		return DecisionsDTO{}, err
	}
	items := make([]DecisionDTO, 0, len(nodes))
	for _, n := range nodes {
		items = append(items, decisionDTO(n))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return DecisionsDTO{Domain: domainCanonical, Items: items, Count: len(items)}, nil
}

func GetDecision(path, domainCanonical, id string) (DecisionDTO, error) {
	d, err := findDomainNode(path, domainCanonical)
	if err != nil {
		return DecisionDTO{}, err
	}
	nodes, err := cosmosfs.ScanDecisions(d.Path)
	if err != nil {
		return DecisionDTO{}, err
	}
	for _, n := range nodes {
		if n.Metadata.ID == id {
			return decisionDTO(n), nil
		}
	}
	return DecisionDTO{}, Error(CodeInvalidInput, "Decision not found: "+id, http.StatusNotFound, nil)
}

func CreateDecision(path, domainCanonical string, req CreateDecisionRequest) (DecisionDTO, error) {
	d, err := findDomainNode(path, domainCanonical)
	if err != nil {
		return DecisionDTO{}, err
	}
	id := strings.TrimSpace(req.ID)
	if id == "" {
		id = nextDecisionID(d.Path)
	}
	dir := filepath.Join(decisionsDir(d.Path), id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return DecisionDTO{}, err
	}
	dec := model.Decision{
		ID:      id,
		Type:    "decision",
		Name:    req.Name,
		Number:  req.Number,
		Version: fallback(req.Version, "0.1.0"),
		Status:  fallback(req.Status, "draft"),
		Owner:   req.Owner,
		Summary: req.Summary,
		Context: req.Context,
		Inputs:  decisionIOsFromDTO(req.Inputs),
		Outputs: decisionIOsFromDTO(req.Outputs),
	}
	yamlPath := filepath.Join(dir, "decision.yaml")
	if err := fsx.WriteYAML(yamlPath, dec); err != nil {
		return DecisionDTO{}, err
	}
	return decisionDTO(cosmosfs.DecisionNode{Path: dir, Metadata: dec}), nil
}

func UpdateDecision(path, domainCanonical, id string, req UpdateDecisionRequest) (DecisionDTO, error) {
	d, err := findDomainNode(path, domainCanonical)
	if err != nil {
		return DecisionDTO{}, err
	}
	nodes, err := cosmosfs.ScanDecisions(d.Path)
	if err != nil {
		return DecisionDTO{}, err
	}
	for _, n := range nodes {
		if n.Metadata.ID != id {
			continue
		}
		dec := n.Metadata
		if req.Name != "" {
			dec.Name = req.Name
		}
		if req.Number != "" {
			dec.Number = req.Number
		}
		if req.Version != "" {
			dec.Version = req.Version
		}
		if req.Status != "" {
			dec.Status = req.Status
		}
		if req.Owner != "" {
			dec.Owner = req.Owner
		}
		if req.Summary != "" {
			dec.Summary = req.Summary
		}
		if req.Context != "" {
			dec.Context = req.Context
		}
		if req.Inputs != nil {
			dec.Inputs = decisionIOsFromDTO(req.Inputs)
		}
		if req.Outputs != nil {
			dec.Outputs = decisionIOsFromDTO(req.Outputs)
		}
		if err := fsx.WriteYAML(filepath.Join(n.Path, "decision.yaml"), dec); err != nil {
			return DecisionDTO{}, err
		}
		n.Metadata = dec
		return decisionDTO(n), nil
	}
	return DecisionDTO{}, Error(CodeInvalidInput, "Decision not found: "+id, http.StatusNotFound, nil)
}

func DeleteDecision(path, domainCanonical, id string) error {
	d, err := findDomainNode(path, domainCanonical)
	if err != nil {
		return err
	}
	dir := filepath.Join(decisionsDir(d.Path), id)
	if _, err := os.Stat(dir); err != nil {
		return Error(CodeInvalidInput, "Decision not found: "+id, http.StatusNotFound, nil)
	}
	return os.RemoveAll(dir)
}

func GetDecisionDMN(path, domainCanonical, id string) (string, error) {
	d, err := findDomainNode(path, domainCanonical)
	if err != nil {
		return "", err
	}
	nodes, err := cosmosfs.ScanDecisions(d.Path)
	if err != nil {
		return "", err
	}
	for _, n := range nodes {
		if n.Metadata.ID != id {
			continue
		}
		if n.DMNPath == "" {
			return "", Error(CodeInvalidInput, "No DMN file for decision: "+id, http.StatusNotFound, nil)
		}
		data, err := os.ReadFile(n.DMNPath)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", Error(CodeInvalidInput, "Decision not found: "+id, http.StatusNotFound, nil)
}

func UpdateDecisionDMN(path, domainCanonical, id, dmnXML string) (DecisionDTO, error) {
	d, err := findDomainNode(path, domainCanonical)
	if err != nil {
		return DecisionDTO{}, err
	}
	nodes, err := cosmosfs.ScanDecisions(d.Path)
	if err != nil {
		return DecisionDTO{}, err
	}
	for _, n := range nodes {
		if n.Metadata.ID != id {
			continue
		}
		dmnPath := filepath.Join(n.Path, "decision.dmn")
		if err := os.WriteFile(dmnPath, []byte(dmnXML), 0o644); err != nil {
			return DecisionDTO{}, err
		}
		// Record the dmn_file reference in the YAML metadata if not already set.
		if n.Metadata.DMNFile == "" {
			n.Metadata.DMNFile = "decision.dmn"
			_ = fsx.WriteYAML(filepath.Join(n.Path, "decision.yaml"), n.Metadata)
		}
		n.DMNPath = dmnPath
		return decisionDTO(n), nil
	}
	return DecisionDTO{}, Error(CodeInvalidInput, "Decision not found: "+id, http.StatusNotFound, nil)
}

func decisionDTO(n cosmosfs.DecisionNode) DecisionDTO {
	return DecisionDTO{
		ID:      n.Metadata.ID,
		Name:    n.Metadata.Name,
		Number:  n.Metadata.Number,
		Version: n.Metadata.Version,
		Status:  n.Metadata.Status,
		Owner:   n.Metadata.Owner,
		Summary: n.Metadata.Summary,
		Context: n.Metadata.Context,
		DMNFile: n.Metadata.DMNFile,
		HasDMN:  n.DMNPath != "",
		Inputs:  decisionIODTOs(n.Metadata.Inputs),
		Outputs: decisionIODTOs(n.Metadata.Outputs),
		Path:    n.Path,
	}
}

func decisionIODTOs(ios []model.DecisionIO) []DecisionIODTO {
	out := make([]DecisionIODTO, 0, len(ios))
	for _, v := range ios {
		out = append(out, DecisionIODTO{Name: v.Name, Type: v.Type, Description: v.Description})
	}
	return out
}

func decisionIOsFromDTO(dtos []DecisionIODTO) []model.DecisionIO {
	out := make([]model.DecisionIO, 0, len(dtos))
	for _, v := range dtos {
		out = append(out, model.DecisionIO{Name: v.Name, Type: v.Type, Description: v.Description})
	}
	return out
}

func nextDecisionID(domainPath string) string {
	nodes, _ := cosmosfs.ScanDecisions(domainPath)
	return fmt.Sprintf("DEC-%03d", len(nodes)+1)
}
