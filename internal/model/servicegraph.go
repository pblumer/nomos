package model

// Servicegraph represents a graph-based service composition model.
type Servicegraph struct {
	ID             string         `yaml:"id" json:"id"`
	Type           string         `yaml:"type" json:"type"`
	Name           string         `yaml:"name" json:"name"`
	Version        string         `yaml:"version" json:"version"`
	Status         string         `yaml:"status" json:"status"`
	Owner          string         `yaml:"owner" json:"owner"`
	Summary        string         `yaml:"summary" json:"summary"`
	RelatedProduct string         `yaml:"related_product" json:"related_product"`
	RelatedProcess string         `yaml:"related_process,omitempty" json:"related_process,omitempty"`
	Variants       []GraphVariant `yaml:"variants,omitempty" json:"variants,omitempty"`
	Nodes          []GraphNode    `yaml:"nodes" json:"nodes"`
	Edges          []GraphEdge    `yaml:"edges" json:"edges"`
	Rules          []GraphRule    `yaml:"rules,omitempty" json:"rules,omitempty"`
}

// GraphNode represents a node in the servicegraph.
type GraphNode struct {
	ID             string `yaml:"id" json:"id"`
	Type           string `yaml:"type" json:"type"`
	Name           string `yaml:"name" json:"name"`
	Description    string `yaml:"description,omitempty" json:"description,omitempty"`
	Mandatory      bool   `yaml:"mandatory" json:"mandatory"`
	Reusable       bool   `yaml:"reusable" json:"reusable"`
	Variant        string `yaml:"variant,omitempty" json:"variant,omitempty"`
	ActivationRule string `yaml:"activation_rule,omitempty" json:"activation_rule,omitempty"`
	Owner          string `yaml:"owner,omitempty" json:"owner,omitempty"`
	SkillRef       string `yaml:"skill_ref,omitempty" json:"skill_ref,omitempty"`
	DecisionRef    string `yaml:"decision_ref,omitempty" json:"decision_ref,omitempty"`
	RuleRef        string `yaml:"rule_ref,omitempty" json:"rule_ref,omitempty"`
}

// GraphEdge represents a relationship between two nodes.
type GraphEdge struct {
	ID          string `yaml:"id" json:"id"`
	Source      string `yaml:"source" json:"source"`
	Target      string `yaml:"target" json:"target"`
	Type        string `yaml:"type" json:"type"`
	Binding     string `yaml:"binding" json:"binding"`
	Condition   string `yaml:"condition,omitempty" json:"condition,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// GraphRule defines a rule that governs the servicegraph behavior.
type GraphRule struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"`
	Scope       string `yaml:"scope" json:"scope"`
	Binding     string `yaml:"binding" json:"binding"`
	Expression  string `yaml:"expression,omitempty" json:"expression,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// GraphVariant describes a variant of a service composition.
type GraphVariant struct {
	ID      string `yaml:"id" json:"id"`
	Name    string `yaml:"name" json:"name"`
	Context string `yaml:"context,omitempty" json:"context,omitempty"`
}
