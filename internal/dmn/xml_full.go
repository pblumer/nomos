package dmn

// XML structs used by ParseDefinitions. Field names are local-name based so
// the parser accepts every DMN 1.x namespace (1.1 — 1.5) without manual
// namespace bookkeeping. Suffix "Full" distinguishes these from the
// decision-table-only structs in parser.go that drive the FEEL evaluator.
//
// typeRef in DMN may appear either as an attribute (DMN 1.2+, dmn-js writes
// this form) or as a child element with text (older DMN files). The
// xmlTypedXXX helpers expose both via typeRef() / pickTypeRef().

import "strings"

func pickTypeRef(attr, elem string) string {
	if a := strings.TrimSpace(attr); a != "" {
		return a
	}
	return strings.TrimSpace(elem)
}

// xmlLogicChildren contains all possible boxed expression children of a
// container element. Embedded anonymously into every parent that hosts a
// boxed-expression body, so encoding/xml decodes the matching child into
// the right pointer field without a custom UnmarshalXML.
type xmlLogicChildren struct {
	DecisionTable      *xmlDecisionTableFull  `xml:"decisionTable"`
	LiteralExpression  *xmlLiteralExpression  `xml:"literalExpression"`
	Context            *xmlContext            `xml:"context"`
	Invocation         *xmlInvocation         `xml:"invocation"`
	FunctionDefinition *xmlFunctionDefinition `xml:"functionDefinition"`
	List               *xmlList               `xml:"list"`
	Relation           *xmlRelation           `xml:"relation"`
	Conditional        *xmlConditional        `xml:"conditional"`
	For                *xmlIteratorExpr       `xml:"for"`
	Every              *xmlIteratorExpr       `xml:"every"`
	Some               *xmlIteratorExpr       `xml:"some"`
	Filter             *xmlFilter             `xml:"filter"`
}

type xmlDefinitionsFull struct {
	XMLName          struct{}             `xml:"definitions"`
	ID               string               `xml:"id,attr"`
	Name             string               `xml:"name,attr"`
	Namespace        string               `xml:"namespace,attr"`
	ExpressionLang   string               `xml:"expressionLanguage,attr"`
	TypeLang         string               `xml:"typeLanguage,attr"`
	Exporter         string               `xml:"exporter,attr"`
	ExporterVersion  string               `xml:"exporterVersion,attr"`
	Description      string               `xml:"description"`
	Imports          []xmlImport          `xml:"import"`
	ItemDefinitions  []xmlItemDefinition  `xml:"itemDefinition"`
	InputData        []xmlInputData       `xml:"inputData"`
	Decisions        []xmlDecisionFull    `xml:"decision"`
	BKMs             []xmlBKM             `xml:"businessKnowledgeModel"`
	DecisionServices []xmlDecisionService `xml:"decisionService"`
	KnowledgeSources []xmlKnowledgeSource `xml:"knowledgeSource"`
	DMNDI            xmlDMNDI             `xml:"DMNDI"`
}

type xmlImport struct {
	Name        string `xml:"name,attr"`
	Namespace   string `xml:"namespace,attr"`
	ImportType  string `xml:"importType,attr"`
	LocationURI string `xml:"locationURI,attr"`
}

type xmlItemDefinition struct {
	ID             string              `xml:"id,attr"`
	Name           string              `xml:"name,attr"`
	Label          string              `xml:"label,attr"`
	TypeRefAttr    string              `xml:"typeRef,attr"`
	TypeRefElem    string              `xml:"typeRef"`
	TypeLanguage   string              `xml:"typeLanguage,attr"`
	IsCollection   bool                `xml:"isCollection,attr"`
	Description    string              `xml:"description"`
	AllowedValues  *xmlTextList        `xml:"allowedValues"`
	TypeConstraint *xmlTextList        `xml:"typeConstraint"`
	ItemComponents []xmlItemDefinition `xml:"itemComponent"`
	FunctionItem   *xmlFunctionItem    `xml:"functionItem"`
}

func (x xmlItemDefinition) typeRef() string { return pickTypeRef(x.TypeRefAttr, x.TypeRefElem) }

type xmlFunctionItem struct {
	OutputTypeRef string               `xml:"outputTypeRef,attr"`
	Parameters    []xmlInformationItem `xml:"formalParameter"`
}

type xmlTextList struct {
	Texts []xmlText `xml:"text"`
}

type xmlText struct {
	Text string `xml:",chardata"`
}

type xmlHRef struct {
	HRef string `xml:"href,attr"`
}

type xmlInformationItem struct {
	ID          string `xml:"id,attr"`
	Name        string `xml:"name,attr"`
	Label       string `xml:"label,attr"`
	TypeRefAttr string `xml:"typeRef,attr"`
	TypeRefElem string `xml:"typeRef"`
}

func (x xmlInformationItem) typeRef() string { return pickTypeRef(x.TypeRefAttr, x.TypeRefElem) }

type xmlInputData struct {
	ID          string             `xml:"id,attr"`
	Name        string             `xml:"name,attr"`
	Label       string             `xml:"label,attr"`
	Description string             `xml:"description"`
	Variable    xmlInformationItem `xml:"variable"`
}

type xmlKnowledgeSource struct {
	ID                    string                    `xml:"id,attr"`
	Name                  string                    `xml:"name,attr"`
	Label                 string                    `xml:"label,attr"`
	Type                  string                    `xml:"type"`
	LocationURI           string                    `xml:"locationURI"`
	Description           string                    `xml:"description"`
	Owner                 xmlHRef                   `xml:"owner"`
	AuthorityRequirements []xmlAuthorityRequirement `xml:"authorityRequirement"`
}

type xmlBKM struct {
	ID                    string                    `xml:"id,attr"`
	Name                  string                    `xml:"name,attr"`
	Label                 string                    `xml:"label,attr"`
	Description           string                    `xml:"description"`
	Variable              xmlInformationItem        `xml:"variable"`
	EncapsulatedLogic     *xmlFunctionDefinition    `xml:"encapsulatedLogic"`
	KnowledgeRequirements []xmlKnowledgeRequirement `xml:"knowledgeRequirement"`
	AuthorityRequirements []xmlAuthorityRequirement `xml:"authorityRequirement"`
}

type xmlDecisionService struct {
	ID                    string             `xml:"id,attr"`
	Name                  string             `xml:"name,attr"`
	Label                 string             `xml:"label,attr"`
	Description           string             `xml:"description"`
	Variable              xmlInformationItem `xml:"variable"`
	OutputDecisions       []xmlHRef          `xml:"outputDecision"`
	EncapsulatedDecisions []xmlHRef          `xml:"encapsulatedDecision"`
	InputDecisions        []xmlHRef          `xml:"inputDecision"`
	InputData             []xmlHRef          `xml:"inputData"`
}

type xmlKnowledgeRequirement struct {
	ID                string  `xml:"id,attr"`
	RequiredKnowledge xmlHRef `xml:"requiredKnowledge"`
}

type xmlAuthorityRequirement struct {
	ID                string  `xml:"id,attr"`
	RequiredAuthority xmlHRef `xml:"requiredAuthority"`
	RequiredDecision  xmlHRef `xml:"requiredDecision"`
	RequiredInput     xmlHRef `xml:"requiredInput"`
}

type xmlInformationRequirement struct {
	ID               string  `xml:"id,attr"`
	RequiredDecision xmlHRef `xml:"requiredDecision"`
	RequiredInput    xmlHRef `xml:"requiredInput"`
}

// xmlDecisionFull mirrors <decision> with full DMN 1.5 children.
type xmlDecisionFull struct {
	ID                      string                      `xml:"id,attr"`
	Name                    string                      `xml:"name,attr"`
	Label                   string                      `xml:"label,attr"`
	Description             string                      `xml:"description"`
	Question                string                      `xml:"question"`
	AllowedAnswers          string                      `xml:"allowedAnswers"`
	Variable                xmlInformationItem          `xml:"variable"`
	InformationRequirements []xmlInformationRequirement `xml:"informationRequirement"`
	KnowledgeRequirements   []xmlKnowledgeRequirement   `xml:"knowledgeRequirement"`
	AuthorityRequirements   []xmlAuthorityRequirement   `xml:"authorityRequirement"`
	xmlLogicChildren
}

// xmlLogic captures any nested boxed expression. Used as the value of
// <contextEntry>, <binding>, <list> items, <relation> rows / cells, etc.
type xmlLogic struct {
	ID      string `xml:"id,attr"`
	TypeRef string `xml:"typeRef,attr"`
	Label   string `xml:"label,attr"`
	xmlLogicChildren
}

type xmlDecisionTableFull struct {
	ID                   string                `xml:"id,attr"`
	TypeRefAttr          string                `xml:"typeRef,attr"`
	TypeRefElem          string                `xml:"typeRef"`
	Label                string                `xml:"label,attr"`
	HitPolicy            string                `xml:"hitPolicy,attr"`
	Aggregation          string                `xml:"aggregation,attr"`
	PreferredOrientation string                `xml:"preferredOrientation,attr"`
	OutputLabel          string                `xml:"outputLabel,attr"`
	Inputs               []xmlInputFull        `xml:"input"`
	Outputs              []xmlOutputFull       `xml:"output"`
	AnnotationClauses    []xmlAnnotationClause `xml:"annotation"`
	Rules                []xmlRuleFull         `xml:"rule"`
}

func (x xmlDecisionTableFull) typeRef() string { return pickTypeRef(x.TypeRefAttr, x.TypeRefElem) }

type xmlInputFull struct {
	ID              string           `xml:"id,attr"`
	Label           string           `xml:"label,attr"`
	InputExpression xmlInputExprFull `xml:"inputExpression"`
	InputValues     *xmlTextList     `xml:"inputValues"`
}

type xmlInputExprFull struct {
	ID          string `xml:"id,attr"`
	TypeRefAttr string `xml:"typeRef,attr"`
	TypeRefElem string `xml:"typeRef"`
	Text        string `xml:"text"`
}

func (x xmlInputExprFull) typeRef() string { return pickTypeRef(x.TypeRefAttr, x.TypeRefElem) }

type xmlOutputFull struct {
	ID                 string       `xml:"id,attr"`
	Name               string       `xml:"name,attr"`
	Label              string       `xml:"label,attr"`
	TypeRefAttr        string       `xml:"typeRef,attr"`
	TypeRefElem        string       `xml:"typeRef"`
	OutputValues       *xmlTextList `xml:"outputValues"`
	DefaultOutputEntry xmlText      `xml:"defaultOutputEntry>text"`
}

func (x xmlOutputFull) typeRef() string { return pickTypeRef(x.TypeRefAttr, x.TypeRefElem) }

type xmlAnnotationClause struct {
	Name string `xml:"name,attr"`
}

type xmlRuleFull struct {
	ID                string         `xml:"id,attr"`
	Description       string         `xml:"description"`
	InputEntries      []xmlEntryText `xml:"inputEntry"`
	OutputEntries     []xmlEntryText `xml:"outputEntry"`
	AnnotationEntries []xmlEntryText `xml:"annotationEntry"`
}

type xmlEntryText struct {
	ID   string `xml:"id,attr"`
	Text string `xml:"text"`
}

type xmlLiteralExpression struct {
	ID                 string  `xml:"id,attr"`
	TypeRefAttr        string  `xml:"typeRef,attr"`
	TypeRefElem        string  `xml:"typeRef"`
	Label              string  `xml:"label,attr"`
	ExpressionLanguage string  `xml:"expressionLanguage,attr"`
	Text               xmlText `xml:"text"`
	ImportedValues     xmlHRef `xml:"importedValues"`
}

func (x xmlLiteralExpression) typeRef() string { return pickTypeRef(x.TypeRefAttr, x.TypeRefElem) }

type xmlContext struct {
	ID      string            `xml:"id,attr"`
	TypeRef string            `xml:"typeRef,attr"`
	Label   string            `xml:"label,attr"`
	Entries []xmlContextEntry `xml:"contextEntry"`
}

type xmlContextEntry struct {
	ID       string              `xml:"id,attr"`
	Variable *xmlInformationItem `xml:"variable"`
	xmlLogicChildren
}

type xmlInvocation struct {
	ID      string `xml:"id,attr"`
	TypeRef string `xml:"typeRef,attr"`
	Label   string `xml:"label,attr"`
	xmlLogicChildren
	Bindings []xmlBinding `xml:"binding"`
}

type xmlBinding struct {
	Parameter xmlInformationItem `xml:"parameter"`
	xmlLogicChildren
}

type xmlFunctionDefinition struct {
	ID               string               `xml:"id,attr"`
	TypeRef          string               `xml:"typeRef,attr"`
	Label            string               `xml:"label,attr"`
	Kind             string               `xml:"kind,attr"`
	FormalParameters []xmlInformationItem `xml:"formalParameter"`
	xmlLogicChildren
}

type xmlList struct {
	ID      string `xml:"id,attr"`
	TypeRef string `xml:"typeRef,attr"`
	Label   string `xml:"label,attr"`
	// Items must be heterogeneous, so we collect each variant into its own
	// slice and merge them in document order via parser convention.
	DecisionTableItems      []xmlDecisionTableFull  `xml:"decisionTable"`
	LiteralExpressionItems  []xmlLiteralExpression  `xml:"literalExpression"`
	ContextItems            []xmlContext            `xml:"context"`
	InvocationItems         []xmlInvocation         `xml:"invocation"`
	FunctionDefinitionItems []xmlFunctionDefinition `xml:"functionDefinition"`
	ListItems               []xmlList               `xml:"list"`
	RelationItems           []xmlRelation           `xml:"relation"`
	ConditionalItems        []xmlConditional        `xml:"conditional"`
	ForItems                []xmlIteratorExpr       `xml:"for"`
	EveryItems              []xmlIteratorExpr       `xml:"every"`
	SomeItems               []xmlIteratorExpr       `xml:"some"`
	FilterItems             []xmlFilter             `xml:"filter"`
}

type xmlRelation struct {
	ID      string               `xml:"id,attr"`
	TypeRef string               `xml:"typeRef,attr"`
	Label   string               `xml:"label,attr"`
	Columns []xmlInformationItem `xml:"column"`
	Rows    []xmlRelationRow     `xml:"row"`
}

type xmlRelationRow struct {
	xmlLogicChildren
}

type xmlConditional struct {
	ID      string         `xml:"id,attr"`
	TypeRef string         `xml:"typeRef,attr"`
	If      xmlLogicHolder `xml:"if"`
	Then    xmlLogicHolder `xml:"then"`
	Else    xmlLogicHolder `xml:"else"`
}

type xmlIteratorExpr struct {
	ID               string         `xml:"id,attr"`
	TypeRef          string         `xml:"typeRef,attr"`
	IteratorVariable string         `xml:"iteratorVariable,attr"`
	In               xmlLogicHolder `xml:"in"`
	Return           xmlLogicHolder `xml:"return"`
	Satisfies        xmlLogicHolder `xml:"satisfies"`
}

type xmlFilter struct {
	ID      string         `xml:"id,attr"`
	TypeRef string         `xml:"typeRef,attr"`
	In      xmlLogicHolder `xml:"in"`
	Match   xmlLogicHolder `xml:"match"`
}

// xmlLogicHolder is the named child wrapper (<if>, <then>, <in>, <return>,
// <satisfies>, <match>) that contains an arbitrary boxed expression.
type xmlLogicHolder struct {
	xmlLogicChildren
}

// --- DMN DI (diagram interchange) ----------------------------------------

type xmlDMNDI struct {
	Diagrams []xmlDMNDiagram `xml:"DMNDiagram"`
}

type xmlDMNDiagram struct {
	ID     string        `xml:"id,attr"`
	Name   string        `xml:"name,attr"`
	Shapes []xmlDMNShape `xml:"DMNShape"`
	Edges  []xmlDMNEdge  `xml:"DMNEdge"`
}

type xmlDMNShape struct {
	ID            string      `xml:"id,attr"`
	DMNElementRef string      `xml:"dmnElementRef,attr"`
	Bounds        xmlDIBounds `xml:"Bounds"`
}

type xmlDIBounds struct {
	X      float64 `xml:"x,attr"`
	Y      float64 `xml:"y,attr"`
	Width  float64 `xml:"width,attr"`
	Height float64 `xml:"height,attr"`
}

type xmlDMNEdge struct {
	ID            string       `xml:"id,attr"`
	DMNElementRef string       `xml:"dmnElementRef,attr"`
	Waypoints     []xmlDIPoint `xml:"waypoint"`
}

type xmlDIPoint struct {
	X float64 `xml:"x,attr"`
	Y float64 `xml:"y,attr"`
}
