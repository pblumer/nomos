package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/idgen"
	"github.com/nomos/nomos/internal/model"
)

// scenarioSchemaVersion is the on-disk schema for DecisionScenario YAMLs.
const scenarioSchemaVersion = 1

// scenariosDir returns the scenarios directory for a decision.
func scenariosDir(decisionDir string) string {
	return filepath.Join(decisionDir, "scenarios")
}

// scenarioFileName builds a deterministic file name from the scenario ID.
func scenarioFileName(s model.DecisionScenario) string {
	return s.ID + ".yaml"
}

// loadScenarios reads all scenario YAMLs for a decision, sorted by created_at
// ascending (then by ID for tiebreaks).
func loadScenarios(decisionDir string) ([]model.DecisionScenario, error) {
	root := scenariosDir(decisionDir)
	ents, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []model.DecisionScenario
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".yaml") && !strings.HasSuffix(e.Name(), ".yml") {
			continue
		}
		var s model.DecisionScenario
		if err := fsx.ReadYAML(filepath.Join(root, e.Name()), &s); err != nil {
			return nil, fmt.Errorf("read scenario %s: %w", e.Name(), err)
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt < out[j].CreatedAt
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func writeScenario(decisionDir string, s model.DecisionScenario) (string, error) {
	dir := scenariosDir(decisionDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, scenarioFileName(s))
	if err := fsx.WriteYAML(p, s); err != nil {
		return "", err
	}
	return p, nil
}

// CreateDecisionScenarioRequest captures the payload for a free-form scenario
// create. Either Inputs is supplied (and optionally ExpectedOutputs), or
// FromTraceID points at an existing trace to seed both fields.
type CreateDecisionScenarioRequest struct {
	Name            string         `json:"name"`
	Description     string         `json:"description,omitempty"`
	Inputs          map[string]any `json:"inputs,omitempty"`
	ExpectedOutputs map[string]any `json:"expected_outputs,omitempty"`
	FromTraceID     string         `json:"from_trace_id,omitempty"`
}

// UpdateDecisionScenarioRequest mirrors Create but every field is optional.
// Nil maps mean "leave unchanged"; an empty (non-nil) map clears the field.
type UpdateDecisionScenarioRequest struct {
	Name            *string        `json:"name,omitempty"`
	Description     *string        `json:"description,omitempty"`
	Inputs          map[string]any `json:"inputs,omitempty"`
	ExpectedOutputs map[string]any `json:"expected_outputs,omitempty"`
	// ClearExpected lets callers explicitly drop the expected_outputs baseline.
	ClearExpected bool `json:"clear_expected,omitempty"`
}

// ListDecisionScenarios returns all scenarios for a decision, oldest first.
func ListDecisionScenarios(path, domainCanonical, id string) (DecisionScenariosDTO, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return DecisionScenariosDTO{}, err
	}
	scenarios, err := loadScenarios(n.Path)
	if err != nil {
		return DecisionScenariosDTO{}, err
	}
	items := make([]model.DecisionScenario, 0, len(scenarios))
	items = append(items, scenarios...)
	return DecisionScenariosDTO{
		Domain:     domainCanonical,
		DecisionID: id,
		Count:      len(items),
		Items:      items,
	}, nil
}

// GetDecisionScenario returns a single scenario by its ID.
func GetDecisionScenario(path, domainCanonical, id, scenarioID string) (model.DecisionScenario, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	scenarios, err := loadScenarios(n.Path)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	for _, s := range scenarios {
		if s.ID == scenarioID {
			return s, nil
		}
	}
	return model.DecisionScenario{}, Error(CodeInvalidInput, "Scenario not found: "+scenarioID, http.StatusNotFound, nil)
}

// CreateDecisionScenario persists a new scenario for a decision. The free-form
// path expects Name + Inputs; if FromTraceID is set it instead delegates to
// CreateDecisionScenarioFromTrace so callers can use one endpoint for both.
func CreateDecisionScenario(path, domainCanonical, id string, req CreateDecisionScenarioRequest) (model.DecisionScenario, error) {
	if req.FromTraceID != "" {
		return CreateDecisionScenarioFromTrace(path, domainCanonical, id, req.FromTraceID, req.Name, req.Description)
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return model.DecisionScenario{}, Error(CodeInvalidInput, "Scenario name required", http.StatusBadRequest, nil)
	}
	if req.Inputs == nil {
		return model.DecisionScenario{}, Error(CodeInvalidInput, "Scenario inputs required", http.StatusBadRequest, nil)
	}
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	sid, err := idgen.NewForType("validation_scenario")
	if err != nil {
		return model.DecisionScenario{}, fmt.Errorf("alloc scenario id: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	s := model.DecisionScenario{
		SchemaVersion:   scenarioSchemaVersion,
		ID:              sid,
		Name:            name,
		Description:     strings.TrimSpace(req.Description),
		Domain:          domainCanonical,
		DecisionID:      n.Metadata.ID,
		Inputs:          req.Inputs,
		ExpectedOutputs: req.ExpectedOutputs,
		CreatedAt:       now,
	}
	if _, err := writeScenario(n.Path, s); err != nil {
		return model.DecisionScenario{}, fmt.Errorf("write scenario: %w", err)
	}
	return s, nil
}

// CreateDecisionScenarioFromTrace promotes an existing trace into a saved
// scenario: inputs are copied verbatim and the trace's outputs become the
// scenario's expected_outputs baseline.
func CreateDecisionScenarioFromTrace(path, domainCanonical, id, traceID, name, description string) (model.DecisionScenario, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.DecisionScenario{}, Error(CodeInvalidInput, "Scenario name required", http.StatusBadRequest, nil)
	}
	tr, err := GetDecisionTrace(path, domainCanonical, id, traceID)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	sid, err := idgen.NewForType("validation_scenario")
	if err != nil {
		return model.DecisionScenario{}, fmt.Errorf("alloc scenario id: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	s := model.DecisionScenario{
		SchemaVersion:   scenarioSchemaVersion,
		ID:              sid,
		Name:            name,
		Description:     strings.TrimSpace(description),
		Domain:          domainCanonical,
		DecisionID:      n.Metadata.ID,
		Inputs:          copyAnyMap(tr.Inputs),
		ExpectedOutputs: copyAnyMap(tr.Outputs),
		SourceTraceID:   tr.TraceID,
		CreatedAt:       now,
	}
	if _, err := writeScenario(n.Path, s); err != nil {
		return model.DecisionScenario{}, fmt.Errorf("write scenario: %w", err)
	}
	return s, nil
}

// UpdateDecisionScenario applies a partial update. Pointer fields are only
// written when non-nil; ClearExpected wipes the expected_outputs baseline.
func UpdateDecisionScenario(path, domainCanonical, id, scenarioID string, req UpdateDecisionScenarioRequest) (model.DecisionScenario, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	existing, err := GetDecisionScenario(path, domainCanonical, id, scenarioID)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return model.DecisionScenario{}, Error(CodeInvalidInput, "Scenario name required", http.StatusBadRequest, nil)
		}
		existing.Name = name
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}
	if req.Inputs != nil {
		existing.Inputs = req.Inputs
	}
	if req.ClearExpected {
		existing.ExpectedOutputs = nil
	} else if req.ExpectedOutputs != nil {
		existing.ExpectedOutputs = req.ExpectedOutputs
	}
	existing.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := writeScenario(n.Path, existing); err != nil {
		return model.DecisionScenario{}, fmt.Errorf("write scenario: %w", err)
	}
	return existing, nil
}

// DeleteDecisionScenario removes the scenario YAML.
func DeleteDecisionScenario(path, domainCanonical, id, scenarioID string) error {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return err
	}
	// Resolve the file name via list so unknown IDs return a 404 rather than
	// silently no-oping.
	scenarios, err := loadScenarios(n.Path)
	if err != nil {
		return err
	}
	for _, s := range scenarios {
		if s.ID == scenarioID {
			return os.Remove(filepath.Join(scenariosDir(n.Path), scenarioFileName(s)))
		}
	}
	return Error(CodeInvalidInput, "Scenario not found: "+scenarioID, http.StatusNotFound, nil)
}

// copyAnyMap returns a shallow copy of m so mutations on the trace's maps
// don't bleed into the scenario (and vice versa).
func copyAnyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
