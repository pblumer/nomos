package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
)

const scenariosDirName = "scenarios"

func scenariosDir(decisionDir string) string {
	return filepath.Join(decisionDir, scenariosDirName)
}

func scenarioPath(decisionDir, scenarioID string) string {
	return filepath.Join(scenariosDir(decisionDir), scenarioID+".yaml")
}

var scenarioIDRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

// slugifyScenarioID lower-cases and dash-joins the input so the ID is safe to use
// as a filename across platforms. Returns the original ID if it already matches
// the allowed pattern.
func slugifyScenarioID(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		case r == '_' || r == '-' || r == ' ' || r == '.' || r == '/':
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 63 {
		out = out[:63]
	}
	return out
}

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
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func loadScenario(decisionDir, id string) (model.DecisionScenario, error) {
	var s model.DecisionScenario
	p := scenarioPath(decisionDir, id)
	if err := fsx.ReadYAML(p, &s); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return model.DecisionScenario{}, Error(CodeInvalidInput, "Scenario not found: "+id, http.StatusNotFound, nil)
		}
		return model.DecisionScenario{}, err
	}
	return s, nil
}

func writeScenario(decisionDir string, s model.DecisionScenario) error {
	dir := scenariosDir(decisionDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return fsx.WriteYAML(scenarioPath(decisionDir, s.ID), s)
}

// validateScenario ensures basic invariants before persisting. Inputs map may
// be empty (decision without declared inputs); expected outputs/rules are
// optional and only checked when the scenario runs.
func validateScenario(s *model.DecisionScenario) error {
	if strings.TrimSpace(s.Name) == "" {
		return Error(CodeInvalidInput, "Scenario name is required", http.StatusBadRequest, nil)
	}
	if s.ID == "" {
		s.ID = slugifyScenarioID(s.Name)
	} else {
		s.ID = slugifyScenarioID(s.ID)
	}
	if !scenarioIDRe.MatchString(s.ID) {
		return Error(CodeInvalidInput, "Scenario id must be lowercase alphanumeric with - or _ (1-63 chars): got "+s.ID, http.StatusBadRequest, nil)
	}
	if s.Inputs == nil {
		s.Inputs = map[string]any{}
	}
	return nil
}

// ListDecisionScenarios returns all scenarios for a decision, sorted by name.
func ListDecisionScenarios(path, domainCanonical, id string) (DecisionScenariosDTO, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return DecisionScenariosDTO{}, err
	}
	items, err := loadScenarios(n.Path)
	if err != nil {
		return DecisionScenariosDTO{}, err
	}
	if items == nil {
		items = []model.DecisionScenario{}
	}
	return DecisionScenariosDTO{
		Domain:     domainCanonical,
		DecisionID: id,
		Count:      len(items),
		Items:      items,
	}, nil
}

// GetDecisionScenario fetches a single scenario by ID.
func GetDecisionScenario(path, domainCanonical, id, scenarioID string) (model.DecisionScenario, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	return loadScenario(n.Path, scenarioID)
}

// CreateDecisionScenario persists a new scenario and rejects duplicate IDs.
func CreateDecisionScenario(path, domainCanonical, id string, body model.DecisionScenario) (model.DecisionScenario, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	if err := validateScenario(&body); err != nil {
		return model.DecisionScenario{}, err
	}
	if _, err := os.Stat(scenarioPath(n.Path, body.ID)); err == nil {
		return model.DecisionScenario{}, Error(CodeInvalidInput, "Scenario already exists: "+body.ID, http.StatusConflict, nil)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	body.CreatedAt = now
	body.UpdatedAt = now
	if err := writeScenario(n.Path, body); err != nil {
		return model.DecisionScenario{}, err
	}
	return body, nil
}

// UpdateDecisionScenario overwrites an existing scenario. The ID in the URL
// wins over any ID in the body (body.ID is forced to match).
func UpdateDecisionScenario(path, domainCanonical, id, scenarioID string, body model.DecisionScenario) (model.DecisionScenario, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	existing, err := loadScenario(n.Path, scenarioID)
	if err != nil {
		return model.DecisionScenario{}, err
	}
	body.ID = scenarioID
	if err := validateScenario(&body); err != nil {
		return model.DecisionScenario{}, err
	}
	body.CreatedAt = existing.CreatedAt
	body.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := writeScenario(n.Path, body); err != nil {
		return model.DecisionScenario{}, err
	}
	return body, nil
}

// DeleteDecisionScenario removes the YAML file. Idempotent: returns 404 only if
// the scenario was never present.
func DeleteDecisionScenario(path, domainCanonical, id, scenarioID string) error {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return err
	}
	p := scenarioPath(n.Path, scenarioID)
	if err := os.Remove(p); err != nil {
		if os.IsNotExist(err) {
			return Error(CodeInvalidInput, "Scenario not found: "+scenarioID, http.StatusNotFound, nil)
		}
		return err
	}
	return nil
}

// equalAny compares two values JSON-style: numbers cross-compare across int/float64
// because YAML decoders may produce either; everything else uses reflect.DeepEqual.
func equalAny(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if reflect.DeepEqual(a, b) {
		return true
	}
	// Tolerate int/float cross-typing.
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return af == bf
	}
	return false
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	}
	return 0, false
}

// compareOutputs returns the list of human-readable diffs between expected and
// actual output maps. An empty slice means "match".
func compareOutputs(expected, actual map[string]any) []string {
	var diffs []string
	keys := map[string]struct{}{}
	for k := range expected {
		keys[k] = struct{}{}
	}
	for k := range actual {
		keys[k] = struct{}{}
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	for _, k := range sorted {
		ev, eok := expected[k]
		av, aok := actual[k]
		if !eok {
			diffs = append(diffs, fmt.Sprintf("%s: unexpected (actual=%v)", k, av))
			continue
		}
		if !aok {
			diffs = append(diffs, fmt.Sprintf("%s: missing (expected=%v)", k, ev))
			continue
		}
		if !equalAny(ev, av) {
			diffs = append(diffs, fmt.Sprintf("%s: expected=%v actual=%v", k, ev, av))
		}
	}
	return diffs
}

// compareRules returns differences as a set-based diff (order-independent).
func compareRules(expected, actual []string) []string {
	exp := map[string]struct{}{}
	for _, r := range expected {
		exp[r] = struct{}{}
	}
	act := map[string]struct{}{}
	for _, r := range actual {
		act[r] = struct{}{}
	}
	var diffs []string
	for r := range exp {
		if _, ok := act[r]; !ok {
			diffs = append(diffs, "missing: "+r)
		}
	}
	for r := range act {
		if _, ok := exp[r]; !ok {
			diffs = append(diffs, "unexpected: "+r)
		}
	}
	sort.Strings(diffs)
	return diffs
}

// RunDecisionScenario executes one scenario, persists a trace (tagged with the
// scenario_id), and returns a graded result. A run is not aborted by failed
// assertions — the result merely reports pass/fail.
func RunDecisionScenario(path, domainCanonical, id, scenarioID string, evaluator model.Evaluator) (ScenarioRunResultDTO, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return ScenarioRunResultDTO{}, err
	}
	s, err := loadScenario(n.Path, scenarioID)
	if err != nil {
		return ScenarioRunResultDTO{}, err
	}
	res := ScenarioRunResultDTO{ScenarioID: s.ID, ScenarioName: s.Name}
	start := time.Now()
	result, trace, err := EvaluateDecisionAsScenario(path, domainCanonical, id, s.ID, s.Inputs, evaluator)
	res.DurationMs = time.Since(start).Milliseconds()
	if err != nil && result == nil {
		res.Error = err.Error()
		return res, nil
	}
	if trace != nil {
		res.TraceID = trace.TraceID
	}
	if result != nil {
		res.ActualOutputs = result.Outputs
		res.ActualRules = result.MatchedRules
	}
	if s.ExpectedOutputs != nil {
		res.OutputDiff = compareOutputs(s.ExpectedOutputs, res.ActualOutputs)
		res.OutputsMatch = len(res.OutputDiff) == 0
	} else {
		res.OutputsMatch = true // not asserted
	}
	if s.ExpectedMatchedRules != nil {
		res.RuleDiff = compareRules(s.ExpectedMatchedRules, res.ActualRules)
		res.RulesMatch = len(res.RuleDiff) == 0
	} else {
		res.RulesMatch = true // not asserted
	}
	res.OK = res.OutputsMatch && res.RulesMatch && res.Error == ""
	return res, nil
}

// RunAllDecisionScenarios runs every enabled scenario sequentially. Disabled
// scenarios are counted as skipped without producing a trace.
func RunAllDecisionScenarios(path, domainCanonical, id string, evaluator model.Evaluator) (ScenarioRunBatchDTO, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return ScenarioRunBatchDTO{}, err
	}
	scenarios, err := loadScenarios(n.Path)
	if err != nil {
		return ScenarioRunBatchDTO{}, err
	}
	out := ScenarioRunBatchDTO{Domain: domainCanonical, DecisionID: id, Total: len(scenarios)}
	for _, s := range scenarios {
		if !s.Enabled {
			out.Items = append(out.Items, ScenarioRunResultDTO{ScenarioID: s.ID, ScenarioName: s.Name, OK: false, OutputsMatch: true, RulesMatch: true, Error: "disabled"})
			out.Skipped++
			continue
		}
		r, runErr := RunDecisionScenario(path, domainCanonical, id, s.ID, evaluator)
		if runErr != nil {
			r = ScenarioRunResultDTO{ScenarioID: s.ID, ScenarioName: s.Name, Error: runErr.Error()}
		}
		if r.OK {
			out.Passed++
		} else {
			out.Failed++
		}
		out.Items = append(out.Items, r)
	}
	return out, nil
}
