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

// MethodParameter describes one typed input parameter of a service method (path, query, or header).
type MethodParameter struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"`
	In          string `yaml:"in,omitempty" json:"in,omitempty"` // path | query | header
	Required    bool   `yaml:"required,omitempty" json:"required,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// MethodHeader describes a static or variable-templated HTTP request header.
type MethodHeader struct {
	Name        string `yaml:"name" json:"name"`
	Value       string `yaml:"value,omitempty" json:"value,omitempty"` // static value or {{variable}}
	Required    bool   `yaml:"required,omitempty" json:"required,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// MethodSecurity describes the authentication scheme for a method endpoint.
type MethodSecurity struct {
	Scheme string `yaml:"scheme" json:"scheme"`                 // bearer | api-key | basic | none | oauth2
	In     string `yaml:"in,omitempty" json:"in,omitempty"`     // header | query (for api-key)
	Name   string `yaml:"name,omitempty" json:"name,omitempty"` // header or query-param name
}

// MethodPayloadField describes one field in a request body payload.
type MethodPayloadField struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"` // string | number | boolean | object | array
	Required    bool   `yaml:"required,omitempty" json:"required,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Example     string `yaml:"example,omitempty" json:"example,omitempty"`
}

// MethodPayload describes the request body of a method endpoint.
type MethodPayload struct {
	ContentType string               `yaml:"content_type,omitempty" json:"content_type,omitempty"` // application/json | multipart/form-data | …
	Fields      []MethodPayloadField `yaml:"fields,omitempty" json:"fields,omitempty"`
}

// MethodDefinition describes a named operation exposed by a service as a REST endpoint.
// It unmarshals from both the legacy plain-string format ("methodName") and the
// full object format so existing service.yaml files remain valid without migration.
type MethodDefinition struct {
	Name       string            `yaml:"name" json:"name"`
	Summary    string            `yaml:"summary,omitempty" json:"summary,omitempty"`
	HTTPMethod string            `yaml:"http_method,omitempty" json:"http_method,omitempty"` // GET | POST | PUT | DELETE | PATCH
	Path       string            `yaml:"path,omitempty" json:"path,omitempty"`               // /domains/{id}
	Parameters []MethodParameter `yaml:"parameters,omitempty" json:"parameters,omitempty"`   // path / query params
	Headers    []MethodHeader    `yaml:"headers,omitempty" json:"headers,omitempty"`
	Security   *MethodSecurity   `yaml:"security,omitempty" json:"security,omitempty"`
	Payload    *MethodPayload    `yaml:"payload,omitempty" json:"payload,omitempty"`
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

// Connector describes one access point (CLI, REST, MCP, or BPMN collaboration) for a capability.
type Connector struct {
	Type        string `yaml:"type" json:"type"` // cli | rest | mcp | collaboration
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
	// Collaboration (BPMN) — Capability = Collaboration (ArchiMate ↔ BPMN)
	ArtifactRef  string `yaml:"artifact_ref,omitempty" json:"artifact_ref,omitempty"`   // BPMN collaboration artifact ID
	ConsumerPool string `yaml:"consumer_pool,omitempty" json:"consumer_pool,omitempty"` // pool name for the consumer side
	ProviderPool string `yaml:"provider_pool,omitempty" json:"provider_pool,omitempty"` // pool name for the provider (service) side
}

// ServiceCapability describes one named capability of a service, with optional
// connector metadata. It unmarshals from both plain strings ("cap-name") and
// full objects so existing service.yaml files remain valid.
//
// MethodRefs and DataObjectRefs let a capability declare which of the
// surrounding service's methods and data objects it is composed of. The
// referenced values are method names and data-object IDs.
type ServiceCapability struct {
	ID             string      `yaml:"id" json:"id"`
	Name           string      `yaml:"name" json:"name"`
	Summary        string      `yaml:"summary,omitempty" json:"summary,omitempty"`
	Stability      string      `yaml:"stability,omitempty" json:"stability,omitempty"`
	SideEffect     string      `yaml:"side_effect,omitempty" json:"side_effect,omitempty"`
	Connectors     []Connector `yaml:"connectors,omitempty" json:"connectors,omitempty"`
	RelatedUCI     []string    `yaml:"related_uci,omitempty" json:"related_uci,omitempty"`
	MethodRefs     []string    `yaml:"method_refs,omitempty" json:"method_refs,omitempty"`
	DataObjectRefs []string    `yaml:"data_object_refs,omitempty" json:"data_object_refs,omitempty"`
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

// ServiceDataObject describes a named data object owned/exposed by a service
// (ArchiMate "data object" / application data). Unmarshals from plain string
// ("name") or full object.
type ServiceDataObject struct {
	ID        string `yaml:"id" json:"id"`
	Name      string `yaml:"name" json:"name"`
	Summary   string `yaml:"summary,omitempty" json:"summary,omitempty"`
	Schema    string `yaml:"schema,omitempty" json:"schema,omitempty"`
	Format    string `yaml:"format,omitempty" json:"format,omitempty"`
	Stability string `yaml:"stability,omitempty" json:"stability,omitempty"`
}

func (d *ServiceDataObject) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		d.ID = s
		d.Name = s
		return nil
	}
	type plain ServiceDataObject
	return unmarshal((*plain)(d))
}

// ServiceUserInterface describes a named user interface offered by a service
// (e.g. web UI, CLI surface, mobile app). Unmarshals from plain string
// ("name") or full object.
type ServiceUserInterface struct {
	ID        string `yaml:"id" json:"id"`
	Name      string `yaml:"name" json:"name"`
	Summary   string `yaml:"summary,omitempty" json:"summary,omitempty"`
	Channel   string `yaml:"channel,omitempty" json:"channel,omitempty"`
	URL       string `yaml:"url,omitempty" json:"url,omitempty"`
	Stability string `yaml:"stability,omitempty" json:"stability,omitempty"`
}

func (u *ServiceUserInterface) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		u.ID = s
		u.Name = s
		return nil
	}
	type plain ServiceUserInterface
	return unmarshal((*plain)(u))
}

// SelfModelRef records the embedded self-model bundle version imported into a cosmos.
type SelfModelRef struct {
	Version        string `yaml:"version" json:"version"`
	BundleChecksum string `yaml:"bundle_checksum,omitempty" json:"bundle_checksum,omitempty"`
	ImportedAt     string `yaml:"imported_at,omitempty" json:"imported_at,omitempty"`
}

type Service struct {
	ID                string                 `yaml:"id" json:"id"`
	Type              string                 `yaml:"type" json:"type"`
	Name              string                 `yaml:"name" json:"name"`
	Version           string                 `yaml:"version" json:"version"`
	Status            string                 `yaml:"status" json:"status"`
	Owner             string                 `yaml:"owner" json:"owner"`
	OwnedBy           string                 `yaml:"owned_by,omitempty" json:"owned_by,omitempty"`
	OperatedBy        []string               `yaml:"operated_by,omitempty" json:"operated_by,omitempty"`
	Capabilities      []ServiceCapability    `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	DataObjects       []ServiceDataObject    `yaml:"data_objects,omitempty" json:"data_objects,omitempty"`
	UserInterfaces    []ServiceUserInterface `yaml:"user_interfaces,omitempty" json:"user_interfaces,omitempty"`
	SupportedProducts []string               `yaml:"supported_products,omitempty" json:"supported_products,omitempty"`
	Methods           []MethodDefinition     `yaml:"methods,omitempty" json:"methods,omitempty"`
	Summary           string                 `yaml:"summary" json:"summary"`
	SLA               *ServiceLevelInfo      `yaml:"sla,omitempty" json:"sla,omitempty"`
	OLA               *ServiceLevelInfo      `yaml:"ola,omitempty" json:"ola,omitempty"`
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

// ProcessLane mirrors one bpmn:lane and binds it to the service whose
// methods (Service Tasks) and user interfaces (User Tasks) are valid
// inside that lane. The source of truth lives in the BPMN file as a
// nomos-service-ref documentation element; this mirror exists so the
// catalog API and validation can reason about lane bindings without
// re-parsing BPMN.
type ProcessLane struct {
	BPMNLaneID string `yaml:"bpmn_lane_id" json:"bpmn_lane_id"`
	Name       string `yaml:"name,omitempty" json:"name,omitempty"`
	ServiceRef string `yaml:"service_ref,omitempty" json:"service_ref,omitempty"`
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
	Lanes          []ProcessLane        `yaml:"lanes,omitempty" json:"lanes,omitempty"`
	BPMN           BPMNReference        `yaml:"bpmn" json:"bpmn"`
	TaskMappings   []ProcessTaskMapping `yaml:"task_mappings,omitempty" json:"task_mappings,omitempty"`
	Triggers       []ProcessTrigger     `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Notes          string               `yaml:"notes,omitempty" json:"notes,omitempty"`
}

// ProcessTrigger binds a BPMN start event to a runtime trigger (timer cron,
// message event, signal or conditional). The trigger type must match the
// underlying BPMN element's event definition; configuration sub-blocks
// carry the runtime details (cron expression, event topic, …).
//
// See ADR-0018 for the design rationale.
type ProcessTrigger struct {
	BPMNElementID string                    `yaml:"bpmn_element_id" json:"bpmn_element_id"`
	Type          string                    `yaml:"type" json:"type"`
	Name          string                    `yaml:"name,omitempty" json:"name,omitempty"`
	Description   string                    `yaml:"description,omitempty" json:"description,omitempty"`
	Timer         *TimerTriggerConfig       `yaml:"timer,omitempty" json:"timer,omitempty"`
	Message       *MessageTriggerConfig     `yaml:"message,omitempty" json:"message,omitempty"`
	Signal        *SignalTriggerConfig      `yaml:"signal,omitempty" json:"signal,omitempty"`
	Conditional   *ConditionalTriggerConfig `yaml:"conditional,omitempty" json:"conditional,omitempty"`
}

type TimerTriggerConfig struct {
	Cron        string `yaml:"cron,omitempty" json:"cron,omitempty"`
	ISODuration string `yaml:"iso_duration,omitempty" json:"iso_duration,omitempty"`
	ISODate     string `yaml:"iso_date,omitempty" json:"iso_date,omitempty"`
	Timezone    string `yaml:"timezone,omitempty" json:"timezone,omitempty"`
}

type MessageTriggerConfig struct {
	EventRef       string `yaml:"event_ref,omitempty" json:"event_ref,omitempty"`
	Topic          string `yaml:"topic,omitempty" json:"topic,omitempty"`
	Filter         string `yaml:"filter,omitempty" json:"filter,omitempty"`
	CorrelationKey string `yaml:"correlation_key,omitempty" json:"correlation_key,omitempty"`
}

type SignalTriggerConfig struct {
	SignalRef string `yaml:"signal_ref,omitempty" json:"signal_ref,omitempty"`
}

type ConditionalTriggerConfig struct {
	Expression string `yaml:"expression,omitempty" json:"expression,omitempty"`
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
// a service call (ArchiMate function/trigger). CapabilityRef identifies the
// service capability (= BPMN collaboration boundary) being invoked.
type ProcessStep struct {
	ID               string             `yaml:"id" json:"id"`
	Name             string             `yaml:"name" json:"name"`
	TaskType         string             `yaml:"task_type,omitempty" json:"task_type,omitempty"`
	ServiceRef       string             `yaml:"service_ref" json:"service_ref"`
	CapabilityRef    string             `yaml:"capability_ref,omitempty" json:"capability_ref,omitempty"`
	Method           string             `yaml:"method,omitempty" json:"method,omitempty"`
	UserInterfaceRef string             `yaml:"user_interface_ref,omitempty" json:"user_interface_ref,omitempty"` // for userTask: links to a ServiceUserInterface
	DecisionRef      string             `yaml:"decision_ref,omitempty" json:"decision_ref,omitempty"`
	Role             string             `yaml:"role,omitempty" json:"role,omitempty"`
	Required         bool               `yaml:"required" json:"required"`
	Notes            string             `yaml:"notes,omitempty" json:"notes,omitempty"`
	DependsOn        []string           `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Inputs           []StepInputBinding `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Outputs          []StepOutputSchema `yaml:"outputs,omitempty" json:"outputs,omitempty"`
	Decision         *DecisionTable     `yaml:"decision,omitempty" json:"decision,omitempty"`
	Gateway          *DecisionGateway   `yaml:"gateway,omitempty" json:"gateway,omitempty"`
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
	CapabilityRef   string `yaml:"capability_ref,omitempty" json:"capability_ref,omitempty"`
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

// Evaluator captures who/what triggered a decision evaluation.
type Evaluator struct {
	ID        string `yaml:"id,omitempty" json:"id,omitempty"`
	IP        string `yaml:"ip,omitempty" json:"ip,omitempty"`
	UserAgent string `yaml:"user_agent,omitempty" json:"user_agent,omitempty"`
}

// EngineInfo identifies the engine that produced a trace.
type EngineInfo struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
	Commit  string `yaml:"commit,omitempty" json:"commit,omitempty"`
}

// DecisionScenario is a named, persisted test case for a Decision.
// Scenarios are typically promoted from a previous DecisionTrace (Inputs +
// Outputs captured as ExpectedOutputs), and are stored in
// <decision>/scenarios/<VSC-id>.yaml so they're git-versioned and reviewable.
// Unlike traces (which are an audit log and may rotate), scenarios are
// curated regression cases keyed by a stable ID.
type DecisionScenario struct {
	SchemaVersion   int            `yaml:"schema_version" json:"schema_version"`
	ID              string         `yaml:"id" json:"id"`
	Name            string         `yaml:"name" json:"name"`
	Description     string         `yaml:"description,omitempty" json:"description,omitempty"`
	Domain          string         `yaml:"domain" json:"domain"`
	DecisionID      string         `yaml:"decision_id" json:"decision_id"`
	Inputs          map[string]any `yaml:"inputs" json:"inputs"`
	ExpectedOutputs map[string]any `yaml:"expected_outputs,omitempty" json:"expected_outputs,omitempty"`
	SourceTraceID   string         `yaml:"source_trace_id,omitempty" json:"source_trace_id,omitempty"`
	CreatedAt       string         `yaml:"created_at" json:"created_at"`
	UpdatedAt       string         `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// DecisionTrace is the persisted, content-addressed record of one decision evaluation.
// The TraceID is a sha256 over a canonical JSON encoding of all other fields (excluding
// TraceID itself), forming a hash chain via ParentTraceID for tamper-evidence.
type DecisionTrace struct {
	SchemaVersion   int            `yaml:"schema_version" json:"schema_version"`
	TraceID         string         `yaml:"trace_id" json:"trace_id"`
	ParentTraceID   string         `yaml:"parent_trace_id,omitempty" json:"parent_trace_id,omitempty"`
	Timestamp       string         `yaml:"timestamp" json:"timestamp"`
	Domain          string         `yaml:"domain" json:"domain"`
	DecisionID      string         `yaml:"decision_id" json:"decision_id"`
	DecisionName    string         `yaml:"decision_name,omitempty" json:"decision_name,omitempty"`
	DecisionVersion string         `yaml:"decision_version" json:"decision_version"`
	RuleHash        string         `yaml:"rule_hash" json:"rule_hash"`
	Inputs          map[string]any `yaml:"inputs" json:"inputs"`
	Outputs         map[string]any `yaml:"outputs" json:"outputs"`
	MatchedRules    []string       `yaml:"matched_rules,omitempty" json:"matched_rules,omitempty"`
	HitPolicy       string         `yaml:"hit_policy,omitempty" json:"hit_policy,omitempty"`
	Evaluator       Evaluator      `yaml:"evaluator,omitempty" json:"evaluator,omitempty"`
	Engine          EngineInfo     `yaml:"engine" json:"engine"`
}
