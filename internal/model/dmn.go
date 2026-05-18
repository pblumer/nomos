package model

// DMN 1.5 metamodel as Nomos first-class types.
//
// Names mirror the OMG DMN 1.5 specification (omg.org/spec/DMN/1.5) as closely
// as the Go type system allows. The wrapper artifact stored on disk is
// model.Decision (decision.yaml). The DMN XML file referenced by Decision.DMNFile
// parses into DMNDefinitions and represents one Decision Requirements Graph
// (DRG). Several DRDs may visualise the same DRG.
//
// Element mapping summary (DMN 1.5 → Nomos):
//
//   definitions                → DMNDefinitions          (DRG root)
//   itemDefinition             → DMNItemDefinition       (typed data)
//   inputData                  → DMNInputData            (DRG node)
//   decision                   → DMNDecision             (DRG node)
//   businessKnowledgeModel     → DMNBusinessKnowledgeModel (DRG node, "BKM")
//   decisionService            → DMNDecisionService      (DRG node)
//   knowledgeSource            → DMNKnowledgeSource      (DRG node)
//   informationRequirement     → DMNInformationRequirement  (DRG edge)
//   knowledgeRequirement       → DMNKnowledgeRequirement    (DRG edge)
//   authorityRequirement       → DMNAuthorityRequirement    (DRG edge)
//   informationItem / variable → DMNInformationItem
//   decisionTable              → DMNDecisionTable        (logic)
//   literalExpression          → DMNLiteralExpression    (logic)
//   context / contextEntry     → DMNContext / DMNContextEntry (logic)
//   invocation / binding       → DMNInvocation / DMNBinding   (logic)
//   functionDefinition         → DMNFunctionDefinition   (logic)
//   list                       → DMNList                 (logic)
//   relation                   → DMNRelation             (logic)
//   conditional                → DMNConditional          (logic, DMN 1.4+)
//   for / every / some         → DMNFor / DMNEvery / DMNSome (logic, DMN 1.4+)
//   filter                     → DMNFilter               (logic, DMN 1.4+)
//   dmndi:DMNDI / DMNDiagram   → DMNDiagram              (visual layout)
//
// The DMN logic body of a DMNDecision (or DMNBusinessKnowledgeModel's
// encapsulatedLogic) is a sum type: exactly one *DMNDecisionTable,
// *DMNLiteralExpression, *DMNContext, *DMNInvocation, *DMNFunctionDefinition,
// *DMNList, *DMNRelation, *DMNConditional, *DMNFor, *DMNEvery, *DMNSome, or
// *DMNFilter is populated; the others stay nil. DMNLogic.Kind names the active
// variant so YAML readers do not need to guess.

// DMNDefinitions is the root <definitions> element of a DMN file
// (one Decision Requirements Graph + optional DRD layout).
type DMNDefinitions struct {
	ID              string                      `yaml:"id" json:"id"`
	Name            string                      `yaml:"name" json:"name"`
	Namespace       string                      `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	ExpressionLang  string                      `yaml:"expression_language,omitempty" json:"expression_language,omitempty"`
	TypeLang        string                      `yaml:"type_language,omitempty" json:"type_language,omitempty"`
	ExporterName    string                      `yaml:"exporter,omitempty" json:"exporter,omitempty"`
	ExporterVersion string                      `yaml:"exporter_version,omitempty" json:"exporter_version,omitempty"`
	Description     string                      `yaml:"description,omitempty" json:"description,omitempty"`
	ItemDefinitions []DMNItemDefinition         `yaml:"item_definitions,omitempty" json:"item_definitions,omitempty"`
	InputData       []DMNInputData              `yaml:"input_data,omitempty" json:"input_data,omitempty"`
	Decisions       []DMNDecision               `yaml:"decisions,omitempty" json:"decisions,omitempty"`
	BKMs            []DMNBusinessKnowledgeModel `yaml:"business_knowledge_models,omitempty" json:"business_knowledge_models,omitempty"`
	DecisionService []DMNDecisionService        `yaml:"decision_services,omitempty" json:"decision_services,omitempty"`
	KnowledgeSource []DMNKnowledgeSource        `yaml:"knowledge_sources,omitempty" json:"knowledge_sources,omitempty"`
	Imports         []DMNImport                 `yaml:"imports,omitempty" json:"imports,omitempty"`
	Diagrams        []DMNDiagram                `yaml:"diagrams,omitempty" json:"diagrams,omitempty"`
}

// DMNImport references another DMN model (<import>).
type DMNImport struct {
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Namespace   string `yaml:"namespace" json:"namespace"`
	ImportType  string `yaml:"import_type,omitempty" json:"import_type,omitempty"`
	LocationURI string `yaml:"location_uri,omitempty" json:"location_uri,omitempty"`
}

// DMNItemDefinition describes a reusable type (<itemDefinition>).
// Nested ItemComponents form a composite/struct type; AllowedValues constrain
// scalar types via FEEL unary tests.
type DMNItemDefinition struct {
	ID             string              `yaml:"id,omitempty" json:"id,omitempty"`
	Name           string              `yaml:"name" json:"name"`
	Description    string              `yaml:"description,omitempty" json:"description,omitempty"`
	Label          string              `yaml:"label,omitempty" json:"label,omitempty"`
	TypeRef        string              `yaml:"type_ref,omitempty" json:"type_ref,omitempty"`
	TypeLanguage   string              `yaml:"type_language,omitempty" json:"type_language,omitempty"`
	IsCollection   bool                `yaml:"is_collection,omitempty" json:"is_collection,omitempty"`
	FunctionItem   *DMNFunctionItem    `yaml:"function_item,omitempty" json:"function_item,omitempty"`
	AllowedValues  []string            `yaml:"allowed_values,omitempty" json:"allowed_values,omitempty"`
	TypeConstraint []string            `yaml:"type_constraint,omitempty" json:"type_constraint,omitempty"`
	ItemComponents []DMNItemDefinition `yaml:"item_components,omitempty" json:"item_components,omitempty"`
}

// DMNFunctionItem describes the signature of a function-typed
// itemDefinition (<functionItem>).
type DMNFunctionItem struct {
	OutputTypeRef string               `yaml:"output_type_ref,omitempty" json:"output_type_ref,omitempty"`
	Parameters    []DMNInformationItem `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// DMNInformationItem is a typed named value (<variable> / <informationItem>).
type DMNInformationItem struct {
	ID      string `yaml:"id,omitempty" json:"id,omitempty"`
	Name    string `yaml:"name" json:"name"`
	TypeRef string `yaml:"type_ref,omitempty" json:"type_ref,omitempty"`
	Label   string `yaml:"label,omitempty" json:"label,omitempty"`
}

// --- DRG nodes ------------------------------------------------------------

// DMNInputData is the source of values fed into a DRG (<inputData>).
type DMNInputData struct {
	ID          string             `yaml:"id" json:"id"`
	Name        string             `yaml:"name" json:"name"`
	Label       string             `yaml:"label,omitempty" json:"label,omitempty"`
	Description string             `yaml:"description,omitempty" json:"description,omitempty"`
	Variable    DMNInformationItem `yaml:"variable" json:"variable"`
}

// DMNDecision is a DMN decision node (<decision>).
// Logic carries the decision's evaluation body and Requirements link to
// upstream InputData / Decisions / BKMs / KnowledgeSources.
type DMNDecision struct {
	ID                      string                      `yaml:"id" json:"id"`
	Name                    string                      `yaml:"name" json:"name"`
	Label                   string                      `yaml:"label,omitempty" json:"label,omitempty"`
	Description             string                      `yaml:"description,omitempty" json:"description,omitempty"`
	Question                string                      `yaml:"question,omitempty" json:"question,omitempty"`
	AllowedAnswers          string                      `yaml:"allowed_answers,omitempty" json:"allowed_answers,omitempty"`
	Variable                DMNInformationItem          `yaml:"variable" json:"variable"`
	InformationRequirements []DMNInformationRequirement `yaml:"information_requirements,omitempty" json:"information_requirements,omitempty"`
	KnowledgeRequirements   []DMNKnowledgeRequirement   `yaml:"knowledge_requirements,omitempty" json:"knowledge_requirements,omitempty"`
	AuthorityRequirements   []DMNAuthorityRequirement   `yaml:"authority_requirements,omitempty" json:"authority_requirements,omitempty"`
	Logic                   *DMNLogic                   `yaml:"logic,omitempty" json:"logic,omitempty"`
}

// DMNBusinessKnowledgeModel (BKM) packages reusable decision logic
// (<businessKnowledgeModel>).
type DMNBusinessKnowledgeModel struct {
	ID                    string                    `yaml:"id" json:"id"`
	Name                  string                    `yaml:"name" json:"name"`
	Label                 string                    `yaml:"label,omitempty" json:"label,omitempty"`
	Description           string                    `yaml:"description,omitempty" json:"description,omitempty"`
	Variable              DMNInformationItem        `yaml:"variable" json:"variable"`
	EncapsulatedLogic     *DMNFunctionDefinition    `yaml:"encapsulated_logic,omitempty" json:"encapsulated_logic,omitempty"`
	KnowledgeRequirements []DMNKnowledgeRequirement `yaml:"knowledge_requirements,omitempty" json:"knowledge_requirements,omitempty"`
	AuthorityRequirements []DMNAuthorityRequirement `yaml:"authority_requirements,omitempty" json:"authority_requirements,omitempty"`
}

// DMNDecisionService publishes a subset of decisions as a callable service
// (<decisionService>).
type DMNDecisionService struct {
	ID                    string             `yaml:"id" json:"id"`
	Name                  string             `yaml:"name" json:"name"`
	Label                 string             `yaml:"label,omitempty" json:"label,omitempty"`
	Description           string             `yaml:"description,omitempty" json:"description,omitempty"`
	Variable              DMNInformationItem `yaml:"variable" json:"variable"`
	OutputDecisions       []string           `yaml:"output_decisions,omitempty" json:"output_decisions,omitempty"`
	EncapsulatedDecisions []string           `yaml:"encapsulated_decisions,omitempty" json:"encapsulated_decisions,omitempty"`
	InputDecisions        []string           `yaml:"input_decisions,omitempty" json:"input_decisions,omitempty"`
	InputData             []string           `yaml:"input_data,omitempty" json:"input_data,omitempty"`
}

// DMNKnowledgeSource is a non-executable authority such as a policy or expert
// (<knowledgeSource>).
type DMNKnowledgeSource struct {
	ID                    string                    `yaml:"id" json:"id"`
	Name                  string                    `yaml:"name" json:"name"`
	Label                 string                    `yaml:"label,omitempty" json:"label,omitempty"`
	Description           string                    `yaml:"description,omitempty" json:"description,omitempty"`
	Type                  string                    `yaml:"source_type,omitempty" json:"source_type,omitempty"`
	LocationURI           string                    `yaml:"location_uri,omitempty" json:"location_uri,omitempty"`
	Owner                 string                    `yaml:"owner,omitempty" json:"owner,omitempty"`
	AuthorityRequirements []DMNAuthorityRequirement `yaml:"authority_requirements,omitempty" json:"authority_requirements,omitempty"`
}

// --- DRG edges ------------------------------------------------------------

// DMNInformationRequirement links a Decision to an upstream Decision or
// InputData (<informationRequirement>). Exactly one of RequiredDecision /
// RequiredInput is set.
type DMNInformationRequirement struct {
	ID               string `yaml:"id,omitempty" json:"id,omitempty"`
	RequiredDecision string `yaml:"required_decision,omitempty" json:"required_decision,omitempty"`
	RequiredInput    string `yaml:"required_input,omitempty" json:"required_input,omitempty"`
}

// DMNKnowledgeRequirement links a Decision/BKM to an invoked BKM
// (<knowledgeRequirement>).
type DMNKnowledgeRequirement struct {
	ID                string `yaml:"id,omitempty" json:"id,omitempty"`
	RequiredKnowledge string `yaml:"required_knowledge" json:"required_knowledge"`
}

// DMNAuthorityRequirement links a Decision/BKM/KnowledgeSource to a
// KnowledgeSource (<authorityRequirement>).
type DMNAuthorityRequirement struct {
	ID                string `yaml:"id,omitempty" json:"id,omitempty"`
	RequiredAuthority string `yaml:"required_authority,omitempty" json:"required_authority,omitempty"`
	RequiredDecision  string `yaml:"required_decision,omitempty" json:"required_decision,omitempty"`
	RequiredInput     string `yaml:"required_input,omitempty" json:"required_input,omitempty"`
}

// --- Decision logic (boxed expressions) ----------------------------------

// DMNLogic is the union/discriminator of all DMN boxed expression types.
// Exactly one of the typed pointer fields is non-nil; Kind names the
// active variant.
type DMNLogic struct {
	Kind               string                 `yaml:"kind" json:"kind"`
	ID                 string                 `yaml:"id,omitempty" json:"id,omitempty"`
	TypeRef            string                 `yaml:"type_ref,omitempty" json:"type_ref,omitempty"`
	Label              string                 `yaml:"label,omitempty" json:"label,omitempty"`
	Description        string                 `yaml:"description,omitempty" json:"description,omitempty"`
	DecisionTable      *DMNDecisionTable      `yaml:"decision_table,omitempty" json:"decision_table,omitempty"`
	LiteralExpression  *DMNLiteralExpression  `yaml:"literal_expression,omitempty" json:"literal_expression,omitempty"`
	Context            *DMNContext            `yaml:"context,omitempty" json:"context,omitempty"`
	Invocation         *DMNInvocation         `yaml:"invocation,omitempty" json:"invocation,omitempty"`
	FunctionDefinition *DMNFunctionDefinition `yaml:"function_definition,omitempty" json:"function_definition,omitempty"`
	List               *DMNList               `yaml:"list,omitempty" json:"list,omitempty"`
	Relation           *DMNRelation           `yaml:"relation,omitempty" json:"relation,omitempty"`
	Conditional        *DMNConditional        `yaml:"conditional,omitempty" json:"conditional,omitempty"`
	For                *DMNFor                `yaml:"for,omitempty" json:"for,omitempty"`
	Every              *DMNEvery              `yaml:"every,omitempty" json:"every,omitempty"`
	Some               *DMNSome               `yaml:"some,omitempty" json:"some,omitempty"`
	Filter             *DMNFilter             `yaml:"filter,omitempty" json:"filter,omitempty"`
}

// DMNDecisionTable is a tabular boxed expression (<decisionTable>).
type DMNDecisionTable struct {
	HitPolicy       string                    `yaml:"hit_policy" json:"hit_policy"`                       // UNIQUE | FIRST | PRIORITY | ANY | COLLECT | RULE ORDER | OUTPUT ORDER
	Aggregation     string                    `yaml:"aggregation,omitempty" json:"aggregation,omitempty"` // SUM | MIN | MAX | COUNT (COLLECT only)
	PreferredOrient string                    `yaml:"preferred_orientation,omitempty" json:"preferred_orientation,omitempty"`
	OutputLabel     string                    `yaml:"output_label,omitempty" json:"output_label,omitempty"`
	Inputs          []DMNDecisionTableInput   `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Outputs         []DMNDecisionTableOutput  `yaml:"outputs,omitempty" json:"outputs,omitempty"`
	Annotations     []DMNRuleAnnotationClause `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	Rules           []DMNDecisionRule         `yaml:"rules,omitempty" json:"rules,omitempty"`
}

// DMNDecisionTableInput models <input> of a decision table.
type DMNDecisionTableInput struct {
	ID          string   `yaml:"id,omitempty" json:"id,omitempty"`
	Label       string   `yaml:"label,omitempty" json:"label,omitempty"`
	Expression  string   `yaml:"expression" json:"expression"`
	TypeRef     string   `yaml:"type_ref,omitempty" json:"type_ref,omitempty"`
	InputValues []string `yaml:"input_values,omitempty" json:"input_values,omitempty"`
}

// DMNDecisionTableOutput models <output> of a decision table.
type DMNDecisionTableOutput struct {
	ID           string   `yaml:"id,omitempty" json:"id,omitempty"`
	Name         string   `yaml:"name,omitempty" json:"name,omitempty"`
	Label        string   `yaml:"label,omitempty" json:"label,omitempty"`
	TypeRef      string   `yaml:"type_ref,omitempty" json:"type_ref,omitempty"`
	OutputValues []string `yaml:"output_values,omitempty" json:"output_values,omitempty"`
	DefaultValue string   `yaml:"default_value,omitempty" json:"default_value,omitempty"`
}

// DMNRuleAnnotationClause names an annotation column (<ruleAnnotationClause>).
type DMNRuleAnnotationClause struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
}

// DMNDecisionRule is one row in a decision table (<rule>).
type DMNDecisionRule struct {
	ID            string   `yaml:"id,omitempty" json:"id,omitempty"`
	Description   string   `yaml:"description,omitempty" json:"description,omitempty"`
	InputEntries  []string `yaml:"input_entries" json:"input_entries"`
	OutputEntries []string `yaml:"output_entries" json:"output_entries"`
	Annotations   []string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

// DMNLiteralExpression is a FEEL expression body (<literalExpression>).
type DMNLiteralExpression struct {
	Text               string `yaml:"text" json:"text"`
	ExpressionLanguage string `yaml:"expression_language,omitempty" json:"expression_language,omitempty"`
	ImportedValuesURI  string `yaml:"imported_values_uri,omitempty" json:"imported_values_uri,omitempty"`
}

// DMNContext is a boxed list of named entries (<context>).
type DMNContext struct {
	Entries []DMNContextEntry `yaml:"entries" json:"entries"`
}

// DMNContextEntry is one row of a boxed context (<contextEntry>).
type DMNContextEntry struct {
	ID       string              `yaml:"id,omitempty" json:"id,omitempty"`
	Variable *DMNInformationItem `yaml:"variable,omitempty" json:"variable,omitempty"`
	Value    *DMNLogic           `yaml:"value" json:"value"`
}

// DMNInvocation calls a BKM with positional or named bindings (<invocation>).
type DMNInvocation struct {
	CalledFunction *DMNLogic    `yaml:"called_function" json:"called_function"`
	Bindings       []DMNBinding `yaml:"bindings,omitempty" json:"bindings,omitempty"`
}

// DMNBinding is one parameter binding inside an invocation (<binding>).
type DMNBinding struct {
	Parameter DMNInformationItem `yaml:"parameter" json:"parameter"`
	Value     *DMNLogic          `yaml:"value" json:"value"`
}

// DMNFunctionDefinition wraps a body in a parameterised callable
// (<functionDefinition>).
type DMNFunctionDefinition struct {
	Kind       string               `yaml:"kind,omitempty" json:"kind,omitempty"` // FEEL | Java | PMML
	Parameters []DMNInformationItem `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Body       *DMNLogic            `yaml:"body" json:"body"`
}

// DMNList is a boxed list literal (<list>).
type DMNList struct {
	Items []DMNLogic `yaml:"items,omitempty" json:"items,omitempty"`
}

// DMNRelation is a boxed table (<relation>).
type DMNRelation struct {
	Columns []DMNInformationItem `yaml:"columns,omitempty" json:"columns,omitempty"`
	Rows    [][]DMNLogic         `yaml:"rows,omitempty" json:"rows,omitempty"`
}

// DMNConditional is a boxed if/then/else (DMN 1.4+, <conditional>).
type DMNConditional struct {
	If   *DMNLogic `yaml:"if" json:"if"`
	Then *DMNLogic `yaml:"then" json:"then"`
	Else *DMNLogic `yaml:"else" json:"else"`
}

// DMNFor is a boxed for-loop (DMN 1.4+, <for>).
type DMNFor struct {
	Iterator string    `yaml:"iterator" json:"iterator"`
	In       *DMNLogic `yaml:"in" json:"in"`
	Return   *DMNLogic `yaml:"return" json:"return"`
}

// DMNEvery is a universal quantifier (DMN 1.4+, <every>).
type DMNEvery struct {
	Iterator  string    `yaml:"iterator" json:"iterator"`
	In        *DMNLogic `yaml:"in" json:"in"`
	Satisfies *DMNLogic `yaml:"satisfies" json:"satisfies"`
}

// DMNSome is an existential quantifier (DMN 1.4+, <some>).
type DMNSome struct {
	Iterator  string    `yaml:"iterator" json:"iterator"`
	In        *DMNLogic `yaml:"in" json:"in"`
	Satisfies *DMNLogic `yaml:"satisfies" json:"satisfies"`
}

// DMNFilter is a boxed list filter (DMN 1.4+, <filter>).
type DMNFilter struct {
	In    *DMNLogic `yaml:"in" json:"in"`
	Match *DMNLogic `yaml:"match" json:"match"`
}

// --- DRD (diagram) layout -------------------------------------------------

// DMNDiagram captures one DRD: a visual layout over the DRG (<DMNDI:DMNDiagram>).
type DMNDiagram struct {
	ID     string     `yaml:"id,omitempty" json:"id,omitempty"`
	Name   string     `yaml:"name,omitempty" json:"name,omitempty"`
	Shapes []DMNShape `yaml:"shapes,omitempty" json:"shapes,omitempty"`
	Edges  []DMNEdge  `yaml:"edges,omitempty" json:"edges,omitempty"`
}

// DMNShape places a DRG element on the diagram (<DMNDI:DMNShape>).
type DMNShape struct {
	ID            string  `yaml:"id,omitempty" json:"id,omitempty"`
	DMNElementRef string  `yaml:"dmn_element_ref" json:"dmn_element_ref"`
	X             float64 `yaml:"x" json:"x"`
	Y             float64 `yaml:"y" json:"y"`
	Width         float64 `yaml:"width" json:"width"`
	Height        float64 `yaml:"height" json:"height"`
}

// DMNEdge routes a DRG requirement on the diagram (<DMNDI:DMNEdge>).
type DMNEdge struct {
	ID            string        `yaml:"id,omitempty" json:"id,omitempty"`
	DMNElementRef string        `yaml:"dmn_element_ref" json:"dmn_element_ref"`
	Waypoints     []DMNWaypoint `yaml:"waypoints,omitempty" json:"waypoints,omitempty"`
}

// DMNWaypoint is one point of an edge polyline.
type DMNWaypoint struct {
	X float64 `yaml:"x" json:"x"`
	Y float64 `yaml:"y" json:"y"`
}
