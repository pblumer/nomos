package app

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/nomos/nomos/internal/dmn"
	"github.com/nomos/nomos/internal/model"
)

// --- 1. Decision table enumeration ---------------------------------------

// DecisionTableSpecDTO is a flat, test-friendly projection of one decision table in
// the DRG: its columns and rules laid out as rows. It complements the full
// /definitions structure for callers that only want to enumerate rules (e.g. to
// generate or audit test cases).
type DecisionTableSpecDTO struct {
	DecisionID   string                     `json:"decision_id"`
	DecisionName string                     `json:"decision_name"`
	HitPolicy    string                     `json:"hit_policy"`
	Aggregation  string                     `json:"aggregation,omitempty"`
	Inputs       []DecisionTableColumnDTO   `json:"inputs"`
	Outputs      []DecisionTableColumnDTO   `json:"outputs"`
	Rules        []DecisionTableRuleSpecDTO `json:"rules"`
}

// DecisionTableColumnDTO is one input or output column.
type DecisionTableColumnDTO struct {
	Name    string `json:"name"`
	Label   string `json:"label,omitempty"`
	TypeRef string `json:"type_ref,omitempty"`
}

// DecisionTableRuleSpecDTO is one rule row, with When (input entries) aligned to the
// input columns and Then (output entries) aligned to the output columns.
type DecisionTableRuleSpecDTO struct {
	Index       int      `json:"index"`
	ID          string   `json:"id,omitempty"`
	Description string   `json:"description,omitempty"`
	When        []string `json:"when"`
	Then        []string `json:"then"`
}

// DecisionTables returns every decision table in the decision's DMN, projected
// into a flat rules-and-columns shape.
func DecisionTables(path, id string) ([]DecisionTableSpecDTO, error) {
	defs, err := GetDecisionDefinitions(path, id)
	if err != nil {
		return nil, err
	}
	out := []DecisionTableSpecDTO{}
	for _, d := range defs.Decisions {
		if d.Logic == nil || d.Logic.DecisionTable == nil {
			continue
		}
		out = append(out, projectDecisionTable(d))
	}
	return out, nil
}

func projectDecisionTable(d model.DMNDecision) DecisionTableSpecDTO {
	t := d.Logic.DecisionTable
	dto := DecisionTableSpecDTO{
		DecisionID:   d.ID,
		DecisionName: d.Name,
		HitPolicy:    t.HitPolicy,
		Aggregation:  t.Aggregation,
		Inputs:       make([]DecisionTableColumnDTO, 0, len(t.Inputs)),
		Outputs:      make([]DecisionTableColumnDTO, 0, len(t.Outputs)),
		Rules:        make([]DecisionTableRuleSpecDTO, 0, len(t.Rules)),
	}
	for _, in := range t.Inputs {
		dto.Inputs = append(dto.Inputs, DecisionTableColumnDTO{Name: in.Expression, Label: in.Label, TypeRef: in.TypeRef})
	}
	for _, o := range t.Outputs {
		dto.Outputs = append(dto.Outputs, DecisionTableColumnDTO{Name: o.Name, Label: o.Label, TypeRef: o.TypeRef})
	}
	for i, r := range t.Rules {
		dto.Rules = append(dto.Rules, DecisionTableRuleSpecDTO{
			Index:       i,
			ID:          r.ID,
			Description: r.Description,
			When:        r.InputEntries,
			Then:        r.OutputEntries,
		})
	}
	return dto
}

// --- 2. Coverage ----------------------------------------------------------

// DecisionCoverageDTO reports how the decision's stored scenarios exercise its
// rules: per-rule hit counts plus the rules no scenario reaches. The evaluator
// works on the first decision table (the one /evaluate uses), so coverage is
// reported against that table.
type DecisionCoverageDTO struct {
	DecisionID     string                `json:"decision_id"`
	HitPolicy      string                `json:"hit_policy"`
	TotalRules     int                   `json:"total_rules"`
	CoveredRules   int                   `json:"covered_rules"`
	UncoveredRules []string              `json:"uncovered_rules"`
	RulePercent    int                   `json:"rule_percent"`
	ScenarioCount  int                   `json:"scenario_count"`
	Rules          []RuleCoverageDTO     `json:"rules"`
	Scenarios      []ScenarioCoverageDTO `json:"scenarios"`
}

// RuleCoverageDTO is one rule with the scenarios that hit it.
type RuleCoverageDTO struct {
	Index int      `json:"index"`
	ID    string   `json:"id,omitempty"`
	Hits  int      `json:"hits"`
	HitBy []string `json:"hit_by,omitempty"` // scenario ids
}

// ScenarioCoverageDTO records which rules a single scenario matched, or an error.
type ScenarioCoverageDTO struct {
	ScenarioID   string   `json:"scenario_id"`
	Name         string   `json:"name,omitempty"`
	MatchedRules []string `json:"matched_rules"`
	Error        string   `json:"error,omitempty"`
}

// DecisionCoverage evaluates every stored scenario against the decision table
// and aggregates per-rule coverage.
func DecisionCoverage(path, id string) (DecisionCoverageDTO, error) {
	n, err := findDecisionNode(path, id)
	if err != nil {
		return DecisionCoverageDTO{}, err
	}
	if n.DMNPath == "" {
		return DecisionCoverageDTO{}, Error(CodeInvalidInput, "No DMN file for decision: "+id, http.StatusNotFound, nil)
	}
	data, err := os.ReadFile(n.DMNPath)
	if err != nil {
		return DecisionCoverageDTO{}, Error(CodeInternalError, "failed to read DMN: "+err.Error(), http.StatusInternalServerError, err)
	}
	table, err := dmn.ParseDMN(data)
	if err != nil {
		return DecisionCoverageDTO{}, Error(CodeInvalidInput, "DMN parse error: "+err.Error(), http.StatusUnprocessableEntity, err)
	}
	scenarios, err := loadScenarios(n.Path)
	if err != nil {
		return DecisionCoverageDTO{}, err
	}

	// Index rules by id (falling back to a positional key when ids are absent).
	type ruleAgg struct {
		idx   int
		id    string
		hitBy []string
	}
	rules := make([]ruleAgg, len(table.Rules))
	byID := map[string]int{}
	for i, r := range table.Rules {
		rules[i] = ruleAgg{idx: i, id: r.ID}
		if r.ID != "" {
			byID[r.ID] = i
		}
	}

	dto := DecisionCoverageDTO{
		DecisionID:    id,
		HitPolicy:     table.HitPolicy,
		TotalRules:    len(table.Rules),
		ScenarioCount: len(scenarios),
		Scenarios:     make([]ScenarioCoverageDTO, 0, len(scenarios)),
	}
	for _, s := range scenarios {
		sc := ScenarioCoverageDTO{ScenarioID: s.ID, Name: s.Name, MatchedRules: []string{}}
		res, evErr := dmn.Evaluate(table, s.Inputs)
		if evErr != nil {
			sc.Error = evErr.Error()
			dto.Scenarios = append(dto.Scenarios, sc)
			continue
		}
		sc.MatchedRules = res.MatchedRules
		for _, rid := range res.MatchedRules {
			if idx, ok := byID[rid]; ok {
				rules[idx].hitBy = append(rules[idx].hitBy, s.ID)
			}
		}
		dto.Scenarios = append(dto.Scenarios, sc)
	}

	for _, r := range rules {
		rc := RuleCoverageDTO{Index: r.idx, ID: r.id, Hits: len(r.hitBy), HitBy: r.hitBy}
		dto.Rules = append(dto.Rules, rc)
		if rc.Hits > 0 {
			dto.CoveredRules++
		} else {
			key := r.id
			if key == "" {
				key = "rule#" + strconv.Itoa(r.idx)
			}
			dto.UncoveredRules = append(dto.UncoveredRules, key)
		}
	}
	if dto.TotalRules > 0 {
		dto.RulePercent = dto.CoveredRules * 100 / dto.TotalRules
	}
	return dto, nil
}

// --- 3. Business Knowledge Models ----------------------------------------

// DecisionBKMs returns all business knowledge models in the decision's DMN.
func DecisionBKMs(path, id string) ([]model.DMNBusinessKnowledgeModel, error) {
	defs, err := GetDecisionDefinitions(path, id)
	if err != nil {
		return nil, err
	}
	out := make([]model.DMNBusinessKnowledgeModel, 0, len(defs.BKMs))
	out = append(out, defs.BKMs...)
	return out, nil
}

// DecisionBKM returns one business knowledge model by name or id.
func DecisionBKM(path, id, nameOrID string) (model.DMNBusinessKnowledgeModel, error) {
	bkms, err := DecisionBKMs(path, id)
	if err != nil {
		return model.DMNBusinessKnowledgeModel{}, err
	}
	want := strings.TrimSpace(nameOrID)
	for _, b := range bkms {
		if b.ID == want || strings.EqualFold(b.Name, want) {
			return b, nil
		}
	}
	return model.DMNBusinessKnowledgeModel{}, Error(CodeInvalidInput, "BKM not found: "+nameOrID, http.StatusNotFound, nil)
}
