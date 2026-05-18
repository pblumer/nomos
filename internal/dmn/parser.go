package dmn

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// Table is the parsed, engine-ready representation of a DMN 1.x decision table.
type Table struct {
	DecisionID   string
	DecisionName string
	HitPolicy    string // UNIQUE | FIRST | ANY | COLLECT | RULE ORDER
	Aggregation  string // for COLLECT: SUM | MIN | MAX | COUNT
	Inputs       []InputColumn
	Outputs      []OutputColumn
	Rules        []RuleRow
}

// InputColumn describes one input column (variable to match against).
type InputColumn struct {
	ID         string
	Label      string
	Expression string // FEEL path, e.g. "orderAmount"
	TypeRef    string // number | string | boolean | date | date and time
}

// OutputColumn describes one output column.
type OutputColumn struct {
	ID      string
	Name    string
	Label   string
	TypeRef string
}

// RuleRow is one rule (row) in the decision table.
type RuleRow struct {
	ID            string
	Description   string
	InputEntries  []string // raw FEEL unary-test text per input column
	OutputEntries []string // raw FEEL expression text per output column
	Annotation    string
}

// ParseDMN parses a DMN 1.x XML document and returns the first decision table found.
func ParseDMN(xmlData []byte) (*Table, error) {
	// DMN namespaces (1.1 through 1.5)
	var raw xmlDefinitions
	if err := xml.Unmarshal(xmlData, &raw); err != nil {
		return nil, fmt.Errorf("dmn xml parse: %w", err)
	}
	if len(raw.Decisions) == 0 {
		return nil, fmt.Errorf("no <decision> element found in DMN document")
	}
	dec := raw.Decisions[0]
	if dec.DecisionTable == nil {
		return nil, fmt.Errorf("decision %q has no <decisionTable>", dec.ID)
	}
	dt := dec.DecisionTable

	t := &Table{
		DecisionID:   dec.ID,
		DecisionName: dec.Name,
		HitPolicy:    normaliseHitPolicy(dt.HitPolicy),
		Aggregation:  normaliseAggregation(dt.Aggregation),
	}

	for _, in := range dt.Inputs {
		col := InputColumn{
			ID:      in.ID,
			Label:   in.Label,
			TypeRef: in.InputExpression.TypeRef,
		}
		col.Expression = strings.TrimSpace(in.InputExpression.Text)
		if col.Expression == "" {
			col.Expression = in.Label
		}
		t.Inputs = append(t.Inputs, col)
	}

	for _, out := range dt.Outputs {
		t.Outputs = append(t.Outputs, OutputColumn{
			ID:      out.ID,
			Name:    out.Name,
			Label:   out.Label,
			TypeRef: out.TypeRef,
		})
	}

	for _, r := range dt.Rules {
		row := RuleRow{
			ID:          r.ID,
			Description: strings.TrimSpace(r.Description),
		}
		for _, ie := range r.InputEntries {
			row.InputEntries = append(row.InputEntries, strings.TrimSpace(ie.Text))
		}
		for _, oe := range r.OutputEntries {
			row.OutputEntries = append(row.OutputEntries, strings.TrimSpace(oe.Text))
		}
		if len(r.AnnotationEntries) > 0 {
			row.Annotation = strings.TrimSpace(r.AnnotationEntries[0].Text)
		}
		t.Rules = append(t.Rules, row)
	}

	return t, nil
}

// normaliseHitPolicy maps shorthand codes to canonical names.
func normaliseHitPolicy(hp string) string {
	switch strings.ToUpper(strings.TrimSpace(hp)) {
	case "U", "UNIQUE", "":
		return "UNIQUE"
	case "F", "FIRST":
		return "FIRST"
	case "A", "ANY":
		return "ANY"
	case "C", "COLLECT":
		return "COLLECT"
	case "R", "RULE ORDER":
		return "RULE ORDER"
	case "P", "PRIORITY":
		return "PRIORITY"
	case "O", "OUTPUT ORDER":
		return "OUTPUT ORDER"
	default:
		return strings.ToUpper(hp)
	}
}

// normaliseAggregation maps shorthand aggregation codes to canonical names.
func normaliseAggregation(a string) string {
	switch strings.TrimSpace(a) {
	case "+", "SUM":
		return "SUM"
	case "<", "MIN":
		return "MIN"
	case ">", "MAX":
		return "MAX"
	case "#", "COUNT":
		return "COUNT"
	default:
		return strings.ToUpper(strings.TrimSpace(a))
	}
}

// --- raw XML structs (namespace-agnostic via local names) ---

type xmlDefinitions struct {
	XMLName   xml.Name      `xml:"definitions"`
	Decisions []xmlDecision `xml:"decision"`
}

type xmlDecision struct {
	ID            string            `xml:"id,attr"`
	Name          string            `xml:"name,attr"`
	DecisionTable *xmlDecisionTable `xml:"decisionTable"`
}

type xmlDecisionTable struct {
	ID          string      `xml:"id,attr"`
	HitPolicy   string      `xml:"hitPolicy,attr"`
	Aggregation string      `xml:"aggregation,attr"`
	Inputs      []xmlInput  `xml:"input"`
	Outputs     []xmlOutput `xml:"output"`
	Rules       []xmlRule   `xml:"rule"`
}

type xmlInput struct {
	ID              string             `xml:"id,attr"`
	Label           string             `xml:"label,attr"`
	InputExpression xmlInputExpression `xml:"inputExpression"`
}

type xmlInputExpression struct {
	ID      string `xml:"id,attr"`
	TypeRef string `xml:"typeRef,attr"`
	Text    string `xml:"text"`
}

type xmlOutput struct {
	ID      string `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	Label   string `xml:"label,attr"`
	TypeRef string `xml:"typeRef,attr"`
}

type xmlRule struct {
	ID                string     `xml:"id,attr"`
	Description       string     `xml:"description"`
	InputEntries      []xmlEntry `xml:"inputEntry"`
	OutputEntries     []xmlEntry `xml:"outputEntry"`
	AnnotationEntries []xmlEntry `xml:"annotationEntry"`
}

type xmlEntry struct {
	ID   string `xml:"id,attr"`
	Text string `xml:"text"`
}
