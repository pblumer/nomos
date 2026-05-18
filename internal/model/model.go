package model

type Cosmos struct {
	ID        string        `yaml:"id" json:"id"`
	Type      string        `yaml:"type" json:"type"`
	Name      string        `yaml:"name" json:"name"`
	Version   string        `yaml:"version" json:"version"`
	Status    string        `yaml:"status" json:"status"`
	Owner     string        `yaml:"owner" json:"owner"`
	Summary   string        `yaml:"summary" json:"summary"`
	Domains   []string      `yaml:"domains" json:"domains"`
	SelfModel *SelfModelRef `yaml:"self_model,omitempty" json:"self_model,omitempty"`
}

type Domain struct {
	ID                   string   `yaml:"id" json:"id"`
	Type                 string   `yaml:"type" json:"type"`
	Name                 string   `yaml:"name" json:"name"`
	Version              string   `yaml:"version" json:"version"`
	Status               string   `yaml:"status" json:"status"`
	Owner                string   `yaml:"owner" json:"owner"`
	DNSName              string   `yaml:"dns_name" json:"dns_name"`
	Namespace            string   `yaml:"namespace" json:"namespace"`
	Label                string   `yaml:"label" json:"label"`
	Labels               []string `yaml:"labels" json:"labels"`
	CanonicalName        string   `yaml:"canonicalName" json:"canonicalName"`
	TreePath             string   `yaml:"treePath" json:"treePath"`
	ParentCanonical      string   `yaml:"parentCanonical" json:"parentCanonical"`
	ParentTreePath       string   `yaml:"parentTreePath" json:"parentTreePath"`
	MaterializedFromTree bool     `yaml:"materializedFromTree,omitempty" json:"materializedFromTree,omitempty"`
	Summary              string   `yaml:"summary" json:"summary"`
	Services             []string `yaml:"services" json:"services"`
}

type ServiceLevelInfo struct {
	Name          string `yaml:"name,omitempty" json:"name,omitempty"`
	Owner         string `yaml:"owner,omitempty" json:"owner,omitempty"`
	Target        string `yaml:"target,omitempty" json:"target,omitempty"`
	Availability  string `yaml:"availability,omitempty" json:"availability,omitempty"`
	SupportWindow string `yaml:"support_window,omitempty" json:"support_window,omitempty"`
	Description   string `yaml:"description,omitempty" json:"description,omitempty"`
}

// MethodParameter describes one typed input parameter of a service method.
type MethodParameter struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"`
	Required    bool   `yaml:"required,omitempty" json:"required,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// MethodDefinition describes a named operation exposed by a service, with
// optional typed parameters. It unmarshals from both the legacy plain-string
// format ("methodName") and the new object format ({name:…, parameters:[…]}).
type MethodDefinition struct {
	Name       string            `yaml:"name" json:"name"`
	Summary    string            `yaml:"summary,omitempty" json:"summary,omitempty"`
	Parameters []MethodParameter `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// UnmarshalYAML lets MethodDefinition parse both "methodName" strings and full
// objects so existing service.yaml files remain valid without migration.
func (m *MethodDefinition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		m.Name = s
		return nil
	}
	type plain MethodDefinition
	return unmarshal((*plain)(m))
}

// ConnectorArg describes one argument of a CLI connector.
type ConnectorArg struct {
	Name        string   `yaml:"name" json:"name"`
	Flag        string   `yaml:"flag,omitempty" json:"flag,omitempty"`
	Positional  bool     `yaml:"positional,omitempty" json:"positional,omitempty"`
	Required    bool     `yaml:"required,omitempty" json:"required,omitempty"`
	Default     string   `yaml:"default,omitempty" json:"default,omitempty"`
	Enum        []string `yaml:"enum,omitempty" json:"enum,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
}

// ConnectorExitCode documents a CLI exit code.
type ConnectorExitCode struct {
	Code    int    `yaml:"code" json:"code"`
	Meaning string `yaml:"meaning" json:"meaning"`
}

// Connector describes one access point (CLI, REST, or MCP) for a capability.
type Connector struct {
	Type        string `yaml:"type" json:"type"` // cli | rest | mcp
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// CLI
	Invocation string              `yaml:"invocation,omitempty" json:"invocation,omitempty"`
	Args       []ConnectorArg      `yaml:"args,omitempty" json:"args,omitempty"`
	ExitCodes  []ConnectorExitCode `yaml:"exit_codes,omitempty" json:"exit_codes,omitempty"`
	// REST
	Method              string `yaml:"method,omitempty" json:"method,omitempty"`
	Path                string `yaml:"path,omitempty" json:"path,omitempty"`
	RequestContentType  string `yaml:"request_content_type,omitempty" json:"request_content_type,omitempty"`
	ResponseContentType string `yaml:"response_content_type,omitempty" json:"response_content_type,omitempty"`
	Auth                string `yaml:"auth,omitempty" json:"auth,omitempty"`
	// MCP
	Tool       string `yaml:"tool,omitempty" json:"tool,omitempty"`
	Kind       string `yaml:"kind,omitempty" json:"kind,omitempty"`
	Idempotent *bool  `yaml:"idempotent,omitempty" json:"idempotent,omitempty"`
}

// ServiceCapability describes one named capability of a service, with optional
// connector metadata. It unmarshals from both plain strings ("cap-name") and
// full objects so existing service.yaml files remain valid.
type ServiceCapability struct {
	ID         string      `yaml:"id" json:"id"`
	Name       string      `yaml:"name" json:"name"`
	Summary    string      `yaml:"summary,omitempty" json:"summary,omitempty"`
	Stability  string      `yaml:"stability,omitempty" json:"stability,omitempty"`
	SideEffect string      `yaml:"side_effect,omitempty" json:"side_effect,omitempty"`
	Connectors []Connector `yaml:"connectors,omitempty" json:"connectors,omitempty"`
	RelatedUCI []string    `yaml:"related_uci,omitempty" json:"related_uci,omitempty"`
}

func (c *ServiceCapability) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		c.ID = s
		c.Name = s
		return nil
	}
	type plain ServiceCapability
	return unmarshal((*plain)(c))
}

// SelfModelRef records the embedded self-model bundle version imported into a cosmos.
type SelfModelRef struct {
	Version        string `yaml:"version" json:"version"`
	BundleChecksum string `yaml:"bundle_checksum,omitempty" json:"bundle_checksum,omitempty"`
	ImportedAt     string `yaml:"imported_at,omitempty" json:"imported_at,omitempty"`
}

type Service struct {
	ID                string              `yaml:"id" json:"id"`
	Type              string              `yaml:"type" json:"type"`
	Name              string              `yaml:"name" json:"name"`
	Version           string              `yaml:"version" json:"version"`
	Status            string              `yaml:"status" json:"status"`
	Owner             string              `yaml:"owner" json:"owner"`
	OwnedBy           string              `yaml:"owned_by,omitempty" json:"owned_by,omitempty"`
	OperatedBy        []string            `yaml:"operated_by,omitempty" json:"operated_by,omitempty"`
	Capabilities      []ServiceCapability `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	SupportedProducts []string            `yaml:"supported_products,omitempty" json:"supported_products,omitempty"`
	Methods           []MethodDefinition  `yaml:"methods,omitempty" json:"methods,omitempty"`
	Summary           string              `yaml:"summary" json:"summary"`
	SLA               *ServiceLevelInfo   `yaml:"sla,omitempty" json:"sla,omitempty"`
	OLA               *ServiceLevelInfo   `yaml:"ola,omitempty" json:"ola,omitempty"`
}

type Variant struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
}

type RequiredServiceRef struct {
	ServiceRef          string `yaml:"service_ref" json:"service_ref"`
	ServiceBlueprintRef string `yaml:"service_blueprint_ref" json:"service_blueprint_ref"`
	Purpose             string `yaml:"purpose,omitempty" json:"purpose,omitempty"`
	Required            bool   `yaml:"required" json:"required"`
}

type ProductFulfillment struct {
	RequiredServices []ProductRequiredService `yaml:"required_services,omitempty" json:"required_services,omitempty"`
}

type ProductRequiredService struct {
	ServiceRef  string            `yaml:"service_ref" json:"service_ref"`
	Role        string            `yaml:"role,omitempty" json:"role,omitempty"`
	Required    bool              `yaml:"required" json:"required"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	SLARef      string            `yaml:"sla_ref,omitempty" json:"sla_ref,omitempty"`
	OLARef      string            `yaml:"ola_ref,omitempty" json:"ola_ref,omitempty"`
	SLA         *ServiceLevelInfo `yaml:"sla,omitempty" json:"sla,omitempty"`
	OLA         *ServiceLevelInfo `yaml:"ola,omitempty" json:"ola,omitempty"`
}

type Blueprint struct {
	ID                        string                 `yaml:"id" json:"id"`
	Type                      string                 `yaml:"type" json:"type"`
	Name                      string                 `yaml:"name" json:"name"`
	Version                   string                 `yaml:"version" json:"version"`
	Status                    string                 `yaml:"status" json:"status"`
	Owner                     string                 `yaml:"owner" json:"owner"`
	OfferedBy                 string                 `yaml:"offered_by,omitempty" json:"offered_by,omitempty"`
	OwningDomain              string                 `yaml:"owning_domain,omitempty" json:"owning_domain,omitempty"`
	Fulfillment               ProductFulfillment     `yaml:"fulfillment,omitempty" json:"fulfillment,omitempty"`
	Summary                   string                 `yaml:"summary" json:"summary"`
	Purpose                   string                 `yaml:"purpose,omitempty" json:"purpose,omitempty"`
	Description               string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Consumers                 []string               `yaml:"consumers,omitempty" json:"consumers,omitempty"`
	LifecycleStatus           string                 `yaml:"lifecycle_status,omitempty" json:"lifecycle_status,omitempty"`
	Tags                      []string               `yaml:"tags,omitempty" json:"tags,omitempty"`
	Processes                 []string               `yaml:"processes,omitempty" json:"processes,omitempty"`
	Variants                  []Variant              `yaml:"variants" json:"variants,omitempty"`
	Capabilities              []string               `yaml:"capabilities" json:"capabilities,omitempty"`
	TargetSystems             []string               `yaml:"target_systems" json:"target_systems,omitempty"`
	Providers                 []string               `yaml:"providers" json:"providers,omitempty"`
	Actions                   []string               `yaml:"actions" json:"actions,omitempty"`
	RequiredInputs            []string               `yaml:"required_inputs" json:"required_inputs"`
	RequiredServiceBlueprints []string               `yaml:"required_service_blueprints" json:"required_service_blueprints,omitempty"`
	RequiredServices          []RequiredServiceRef   `yaml:"required_services" json:"required_services,omitempty"`
	NamespaceServiceRef       string                 `yaml:"namespace_service_ref" json:"namespace_service_ref,omitempty"`
	Rules                     []string               `yaml:"rules" json:"rules,omitempty"`
	QualityCriteria           []string               `yaml:"quality_criteria" json:"quality_criteria"`
	EvidenceRequirements      []string               `yaml:"evidence_requirements" json:"evidence_requirements"`
	Requirements              []BlueprintRequirement `yaml:"requirements,omitempty" json:"requirements,omitempty"`
	Attributes                []BlueprintAttribute   `yaml:"attributes,omitempty" json:"attributes,omitempty"`
}

// BlueprintRequirement represents a design-time check on a blueprint.
// Status: "" (not yet assessed) → "open" → "fulfilled"
type BlueprintRequirement struct {
	ID            string   `yaml:"id" json:"id"`
	Label         string   `yaml:"label" json:"label"`
	Status        string   `yaml:"status" json:"status"`
	AttributeRefs []string `yaml:"attribute_refs,omitempty" json:"attribute_refs,omitempty"`
}

// BlueprintAttribute is a typed property of a blueprint with attached validation rules.
type BlueprintAttribute struct {
	ID         string          `yaml:"id" json:"id"`
	Label      string          `yaml:"label" json:"label"`
	Type       string          `yaml:"type" json:"type"` // text, number, boolean, date, enum, service_ref
	Required   bool            `yaml:"required" json:"required"`
	ServiceRef string          `yaml:"service_ref,omitempty" json:"service_ref,omitempty"`
	Rules      []AttributeRule `yaml:"rules,omitempty" json:"rules,omitempty"`
}

// AttributeRule defines one validation constraint on a BlueprintAttribute.
// Automatic types: regex, max_length, min_length, starts_with, ends_with, one_of
// Manual type: manual (user marks pass/fail themselves)
type AttributeRule struct {
	ID    string `yaml:"id" json:"id"`
	Label string `yaml:"label" json:"label"`
	Type  string `yaml:"type" json:"type"`
	Value string `yaml:"value" json:"value"`
}

// ProcessParticipant identifies the actor or system that executes a process.
// It maps to a bpmn:participant / pool in the BPMN collaboration diagram and
// names the surrounding system responsible for this process.
type ProcessParticipant struct {
	Name string `yaml:"name" json:"name"`
	Ref  string `yaml:"ref,omitempty" json:"ref,omitempty"` // domain or service canonical ref
}

// Process describes a product-level process artifact. Steps are the primary
// definition; BPMN XML (stored in a separate .bpmn file) is kept for future
// visual modelling and remains fully compatible.
type Process struct {
	ID             string               `yaml:"id" json:"id"`
	Type           string               `yaml:"type" json:"type"`
	Name           string               `yaml:"name" json:"name"`
	Version        string               `yaml:"version" json:"version"`
	Status         string               `yaml:"status" json:"status"`
	Owner          string               `yaml:"owner" json:"owner"`
	Summary        string               `yaml:"summary,omitempty" json:"summary,omitempty"`
	Tags           []string             `yaml:"tags,omitempty" json:"tags,omitempty"`
	RelatedProduct string               `yaml:"related_product" json:"related_product"`
	Participant    *ProcessParticipant  `yaml:"participant,omitempty" json:"participant,omitempty"`
	Steps          []ProcessStep        `yaml:"steps,omitempty" json:"steps,omitempty"`
	BPMN           BPMNReference        `yaml:"bpmn" json:"bpmn"`
	TaskMappings   []ProcessTaskMapping `yaml:"task_mappings,omitempty" json:"task_mappings,omitempty"`
	Notes          string               `yaml:"notes,omitempty" json:"notes,omitempty"`
}

type StepInputBinding struct {
	Name     string `yaml:"name" json:"name"`
	Source   string `yaml:"source" json:"source"`
	Required bool   `yaml:"required,omitempty" json:"required,omitempty"`
}

type StepOutputSchema struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// GatewayCondition describes one outgoing branch of the exclusive gateway that
// follows a business rule task. The expression is evaluated against the
// business rule task's output variables by the orchestration runtime.
type GatewayCondition struct {
	Output     string `yaml:"output" json:"output"`
	Operator   string `yaml:"operator" json:"operator"`
	Value      string `yaml:"value" json:"value"`
	TargetStep string `yaml:"target_step,omitempty" json:"target_step,omitempty"`
	Label      string `yaml:"label,omitempty" json:"label,omitempty"`
}

// DecisionGateway configures the exclusive gateway that is generated directly
// after a business rule task.
type DecisionGateway struct {
	Name       string             `yaml:"name,omitempty" json:"name,omitempty"`
	DefaultTo  string             `yaml:"default_to,omitempty" json:"default_to,omitempty"`
	Conditions []GatewayCondition `yaml:"conditions,omitempty" json:"conditions,omitempty"`
}

// DecisionRule is one DMN-like decision table row evaluated by a business rule task.
type DecisionRule struct {
	Input       string `yaml:"input" json:"input"`
	Operator    string `yaml:"operator" json:"operator"`
	Value       string `yaml:"value" json:"value"`
	Output      string `yaml:"output" json:"output"`
	OutputValue string `yaml:"output_value" json:"output_value"`
	Label       string `yaml:"label,omitempty" json:"label,omitempty"`
}

// DecisionTable captures DMN-style decision logic for business rule tasks.
type DecisionTable struct {
	Name        string         `yaml:"name,omitempty" json:"name,omitempty"`
	HitPolicy   string         `yaml:"hit_policy,omitempty" json:"hit_policy,omitempty"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Rules       []DecisionRule `yaml:"rules,omitempty" json:"rules,omitempty"`
}

// ProcessStep is one ordered step in a fulfillment process. Each step maps to
// a service call (ArchiMate function/trigger).
type ProcessStep struct {
	ID          string             `yaml:"id" json:"id"`
	Name        string             `yaml:"name" json:"name"`
	TaskType    string             `yaml:"task_type,omitempty" json:"task_type,omitempty"`
	ServiceRef  string             `yaml:"service_ref" json:"service_ref"`
	Method      string             `yaml:"method,omitempty" json:"method,omitempty"`
	DecisionRef string             `yaml:"decision_ref,omitempty" json:"decision_ref,omitempty"`
	Role        string             `yaml:"role,omitempty" json:"role,omitempty"`
	Required    bool               `yaml:"required" json:"required"`
	Notes       string             `yaml:"notes,omitempty" json:"notes,omitempty"`
	DependsOn   []string           `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Inputs      []StepInputBinding `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Outputs     []StepOutputSchema `yaml:"outputs,omitempty" json:"outputs,omitempty"`
	Decision    *DecisionTable     `yaml:"decision,omitempty" json:"decision,omitempty"`
	Gateway     *DecisionGateway   `yaml:"gateway,omitempty" json:"gateway,omitempty"`
}

type BPMNReference struct {
	File      string `yaml:"file" json:"file"`
	ProcessID string `yaml:"process_id,omitempty" json:"process_id,omitempty"`
	Primary   bool   `yaml:"primary" json:"primary"`
}

type ProcessTaskMapping struct {
	BPMNElementID   string `yaml:"bpmn_element_id" json:"bpmn_element_id"`
	TaskName        string `yaml:"task_name,omitempty" json:"task_name,omitempty"`
	BPMNElementType string `yaml:"bpmn_element_type,omitempty" json:"bpmn_element_type,omitempty"`
	ServiceRef      string `yaml:"service_ref" json:"service_ref"`
	Method          string `yaml:"method,omitempty" json:"method,omitempty"`
	Role            string `yaml:"role,omitempty" json:"role,omitempty"`
	Required        bool   `yaml:"required" json:"required"`
	Notes           string `yaml:"notes,omitempty" json:"notes,omitempty"`
}

type Instance struct {
	ID                          string            `yaml:"id" json:"id"`
	Type                        string            `yaml:"type" json:"type"`
	Name                        string            `yaml:"name" json:"name"`
	BlueprintRef                string            `yaml:"blueprint_ref" json:"blueprint_ref"`
	BlueprintVersion            string            `yaml:"blueprint_version" json:"blueprint_version"`
	Status                      string            `yaml:"status" json:"status"`
	Owner                       string            `yaml:"owner" json:"owner"`
	Inputs                      map[string]string `yaml:"inputs" json:"inputs,omitempty"`
	ObservedState               map[string]string `yaml:"observed_state" json:"observed_state,omitempty"`
	ProvisionedServiceInstances []string          `yaml:"provisioned_service_instances" json:"provisioned_service_instances,omitempty"`
	OwningProductInstance       string            `yaml:"owning_product_instance" json:"owning_product_instance,omitempty"`
	ProviderRef                 string            `yaml:"provider_ref" json:"provider_ref,omitempty"`
	ComplianceStatus            string            `yaml:"compliance_status" json:"compliance_status"`
	Evidence                    []Evidence        `yaml:"evidence" json:"evidence,omitempty"`
	Findings                    []Finding         `yaml:"findings" json:"findings"`
	AttributeValues             map[string]string `yaml:"attribute_values,omitempty" json:"attribute_values,omitempty"`
}

type Evidence struct {
	ID      string `yaml:"id" json:"id"`
	Type    string `yaml:"type" json:"type"`
	Summary string `yaml:"summary" json:"summary"`
}

type ComplianceResult struct {
	Status   string     `yaml:"status" json:"status"`
	Evidence []Evidence `yaml:"evidence" json:"evidence,omitempty"`
	Findings []Finding  `yaml:"findings" json:"findings"`
}

type Finding struct {
	ID             string `yaml:"id,omitempty" json:"id,omitempty"`
	Severity       string `yaml:"severity" json:"severity"`
	Category       string `yaml:"category,omitempty" json:"category,omitempty"`
	Summary        string `yaml:"summary,omitempty" json:"summary,omitempty"`
	Code           string `yaml:"code,omitempty" json:"code,omitempty"`
	Message        string `yaml:"message,omitempty" json:"message,omitempty"`
	Path           string `yaml:"path,omitempty" json:"path,omitempty"`
	Recommendation string `yaml:"recommendation,omitempty" json:"recommendation,omitempty"`
}

// DecisionIO describes one input or output variable of a Decision artifact.
type DecisionIO struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"` // string, number, boolean, date
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// Decision is a standalone decision artifact scoped to a domain.
// It captures decision metadata and optionally references an external DMN file.
// A businessRuleTask in a BPMN process can call a Decision via its ID and
// receives the output variables for downstream gateway routing.
type Decision struct {
	ID      string       `yaml:"id" json:"id"`
	Type    string       `yaml:"type" json:"type"` // "decision"
	Name    string       `yaml:"name" json:"name"`
	Number  string       `yaml:"number,omitempty" json:"number,omitempty"` // e.g. "DEC-001"
	Version string       `yaml:"version" json:"version"`
	Status  string       `yaml:"status" json:"status"` // draft, active, deprecated
	Owner   string       `yaml:"owner" json:"owner"`
	Summary string       `yaml:"summary,omitempty" json:"summary,omitempty"`
	Context string       `yaml:"context,omitempty" json:"context,omitempty"`
	DMNFile string       `yaml:"dmn_file,omitempty" json:"dmn_file,omitempty"` // rel. path to .dmn file
	Inputs  []DecisionIO `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Outputs []DecisionIO `yaml:"outputs,omitempty" json:"outputs,omitempty"`
}
