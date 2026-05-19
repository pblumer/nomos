package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/dmn"
	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/idgen"
	"github.com/nomos/nomos/internal/idmigrate"
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
	// ADR-0020: Legacy-IDs werden transparent via id-history aufgelöst.
	resolved, _ := idmigrate.Resolve(path, id)
	for _, n := range nodes {
		if n.Metadata.ID == id || n.Metadata.ID == resolved {
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
	// Initial version snapshot. A fresh decision has no DMN yet, so the
	// snapshot directory only carries decision.yaml — it gets the DMN as
	// soon as UpdateDecisionDMN is called for the same version.
	if err := writeVersionSnapshot(dir, dec, nil); err != nil {
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
		// Version is auto-managed: every material change bumps patch and
		// creates a new snapshot. The Version field in the request is
		// ignored — clients can't accidentally tear the trace audit.
		if !metadataChanged(n.Metadata, dec) {
			return decisionDTO(n), nil
		}
		dec.Version = bumpPatch(n.Metadata.Version)
		if err := fsx.WriteYAML(filepath.Join(n.Path, "decision.yaml"), dec); err != nil {
			return DecisionDTO{}, err
		}
		// Snapshot the new state. We copy the existing DMN bytes (if any)
		// into the version directory so each snapshot is self-contained,
		// even though only metadata changed.
		var dmnBytes []byte
		if n.DMNPath != "" {
			if data, err := os.ReadFile(n.DMNPath); err == nil {
				dmnBytes = data
			}
		}
		if err := writeVersionSnapshot(n.Path, dec, dmnBytes); err != nil {
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
		newBytes := []byte(dmnXML)
		// Detect whether anything that the audit chain cares about actually
		// changed: the DMN bytes themselves (rule_hash) or the I/O contract
		// re-derived from the DMN.
		newDec := n.Metadata
		if newDec.DMNFile == "" {
			newDec.DMNFile = "decision.dmn"
		}
		if defs, perr := dmn.ParseDefinitions(newBytes); perr == nil {
			// Conservative: only overwrite when the DMN actually declares
			// the corresponding elements — empty seeds shouldn't wipe
			// manually-entered metadata before the user drags <inputData>
			// pills into the DRD.
			if len(defs.InputData) > 0 {
				newDec.Inputs = inputsFromDMN(defs)
			}
			if len(defs.Decisions) > 0 {
				newDec.Outputs = outputsFromDMN(defs)
			}
		}
		var existingBytes []byte
		if n.DMNPath != "" {
			existingBytes, _ = os.ReadFile(n.DMNPath)
		}
		bytesUnchanged := existingBytes != nil && string(existingBytes) == dmnXML
		metaUnchanged := !metadataChanged(n.Metadata, newDec)
		if bytesUnchanged && metaUnchanged {
			// Nothing to do — same DMN, same derived I/O. No new version.
			n.DMNPath = dmnPath
			return decisionDTO(n), nil
		}
		// Auto-bump patch. The new state is written to HEAD and to its own
		// version snapshot directory so traces created against this version
		// can resolve back to exactly these bytes via rule_hash.
		newDec.Version = bumpPatch(n.Metadata.Version)
		if err := os.WriteFile(dmnPath, newBytes, 0o644); err != nil {
			return DecisionDTO{}, err
		}
		if err := fsx.WriteYAML(filepath.Join(n.Path, "decision.yaml"), newDec); err != nil {
			return DecisionDTO{}, err
		}
		if err := writeVersionSnapshot(n.Path, newDec, newBytes); err != nil {
			return DecisionDTO{}, err
		}
		n.Metadata = newDec
		n.DMNPath = dmnPath
		return decisionDTO(n), nil
	}
	return DecisionDTO{}, Error(CodeInvalidInput, "Decision not found: "+id, http.StatusNotFound, nil)
}

// inputsFromDMN maps DMN <inputData> nodes to the decision artifact's inputs.
func inputsFromDMN(defs *model.DMNDefinitions) []model.DecisionIO {
	if defs == nil {
		return nil
	}
	out := make([]model.DecisionIO, 0, len(defs.InputData))
	for _, in := range defs.InputData {
		name := strings.TrimSpace(in.Variable.Name)
		if name == "" {
			name = strings.TrimSpace(in.Name)
		}
		if name == "" {
			continue
		}
		out = append(out, model.DecisionIO{
			Name:        name,
			Type:        normalizeDMNType(in.Variable.TypeRef),
			Description: strings.TrimSpace(in.Description),
		})
	}
	return out
}

// outputsFromDMN maps each top-level <decision>'s variable to an output.
// "Top-level" = decisions not required by any other decision in the DRG; if
// the requirement graph is empty (single-decision DMN) every decision counts.
func outputsFromDMN(defs *model.DMNDefinitions) []model.DecisionIO {
	if defs == nil {
		return nil
	}
	required := map[string]bool{}
	for _, d := range defs.Decisions {
		for _, ir := range d.InformationRequirements {
			if ir.RequiredDecision != "" {
				required[strings.TrimPrefix(ir.RequiredDecision, "#")] = true
			}
		}
	}
	out := make([]model.DecisionIO, 0, len(defs.Decisions))
	for _, d := range defs.Decisions {
		if required[d.ID] {
			continue
		}
		name := strings.TrimSpace(d.Variable.Name)
		if name == "" {
			name = strings.TrimSpace(d.Name)
		}
		if name == "" {
			continue
		}
		out = append(out, model.DecisionIO{
			Name: name,
			Type: normalizeDMNType(d.Variable.TypeRef),
		})
	}
	return out
}

// normalizeDMNType lower-cases standard DMN/FEEL type refs and leaves custom
// itemDefinition references untouched.
func normalizeDMNType(t string) string {
	s := strings.TrimSpace(t)
	if s == "" {
		return ""
	}
	switch strings.ToLower(s) {
	case "string", "number", "boolean", "date", "time", "date and time", "days and time duration", "years and months duration", "any":
		return strings.ToLower(s)
	}
	return s
}

func decisionIOsEqual(a, b []model.DecisionIO) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].Type != b[i].Type || a[i].Description != b[i].Description {
			return false
		}
	}
	return true
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

// GetDecisionDefinitions parses the DMN file backing the decision and returns
// the full DMN 1.5 Decision Requirements Graph as Nomos types. Used by the
// Cosmos Explorer to render decision metadata around the dmn-js editor.
func GetDecisionDefinitions(path, domainCanonical, id string) (*model.DMNDefinitions, error) {
	d, err := findDomainNode(path, domainCanonical)
	if err != nil {
		return nil, err
	}
	nodes, err := cosmosfs.ScanDecisions(d.Path)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.Metadata.ID != id {
			continue
		}
		if n.DMNPath == "" {
			return nil, Error(CodeInvalidInput, "No DMN file for decision: "+id, http.StatusNotFound, nil)
		}
		data, err := os.ReadFile(n.DMNPath)
		if err != nil {
			return nil, err
		}
		defs, err := dmn.ParseDefinitions(data)
		if err != nil {
			return nil, Error(CodeInvalidInput, "DMN parse error: "+err.Error(), http.StatusUnprocessableEntity, err)
		}
		return defs, nil
	}
	return nil, Error(CodeInvalidInput, "Decision not found: "+id, http.StatusNotFound, nil)
}

// EvaluateDecisionRequest holds the input map for a decision evaluation.
type EvaluateDecisionRequest struct {
	Inputs map[string]any `json:"inputs"`
}

// EvaluateDecision loads the DMN file for a decision and evaluates it against the provided inputs.
// Note: callers that need a persisted audit trail should use EvaluateDecisionWithTrace.
func EvaluateDecision(path, domainCanonical, id string, req EvaluateDecisionRequest) (*dmn.Result, error) {
	d, err := findDomainNode(path, domainCanonical)
	if err != nil {
		return nil, err
	}
	nodes, err := cosmosfs.ScanDecisions(d.Path)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.Metadata.ID != id {
			continue
		}
		if n.DMNPath == "" {
			return nil, Error(CodeInvalidInput, "No DMN file for decision: "+id, http.StatusNotFound, nil)
		}
		data, err := os.ReadFile(n.DMNPath)
		if err != nil {
			return nil, err
		}
		table, err := dmn.ParseDMN(data)
		if err != nil {
			return nil, Error(CodeInvalidInput, "DMN parse error: "+err.Error(), http.StatusUnprocessableEntity, err)
		}
		result, err := dmn.Evaluate(table, req.Inputs)
		if err != nil {
			return nil, Error(CodeInvalidInput, "DMN evaluation error: "+err.Error(), http.StatusUnprocessableEntity, err)
		}
		return result, nil
	}
	return nil, Error(CodeInvalidInput, "Decision not found: "+id, http.StatusNotFound, nil)
}

// nextDecisionID erzeugt eine neue ID gemäß ADR-0020.
// Bei einem Generator-Fehler (extrem unwahrscheinlich) fällt auf
// das alte Schema zurück, damit Create-Aufrufe nie hart fehlschlagen.
func nextDecisionID(domainPath string) string {
	id, err := idgen.NewForType("decision")
	if err == nil {
		return id
	}
	nodes, _ := cosmosfs.ScanDecisions(domainPath)
	return fmt.Sprintf("DEC-%03d", len(nodes)+1)
}
