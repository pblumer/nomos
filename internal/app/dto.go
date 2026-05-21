package app

import "github.com/nomos/nomos/internal/model"

type CosmosDTO struct {
	Path          string `json:"path"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Status        string `json:"status"`
	Owner         string `json:"owner"`
	ServiceCount  int    `json:"serviceCount"`
	DecisionCount int    `json:"decisionCount"`
}

type RepositoryDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Kind          string `json:"kind"`
	Location      string `json:"location"`
	DefaultBranch string `json:"default_branch,omitempty"`
	Status        string `json:"status,omitempty"`
	Head          string `json:"head,omitempty"`
}

type RepositoriesDTO struct {
	Repositories []RepositoryDTO `json:"repositories"`
}

// ServerDTO describes a Nomos server mounted in the Cosmos Explorer tree
// (ADR-0022 §1). The server answers on its endpoint (port 7373) and manages the
// repositories beneath it.
type ServerDTO struct {
	MountID         string `json:"mountId,omitempty"`
	Endpoint        string `json:"endpoint"`
	Label           string `json:"label,omitempty"`
	Local           bool   `json:"local"`
	Authenticated   bool   `json:"authenticated"`
	Status          string `json:"status,omitempty"`
	RepositoryCount int    `json:"repositoryCount"`
}

type NamespaceDTO struct {
	Canonical       string   `json:"canonical"`
	CanonicalName   string   `json:"canonicalName"`
	Namespace       string   `json:"namespace"`
	Labels          []string `json:"labels"`
	Label           string   `json:"label"`
	ParentCanonical string   `json:"parentCanonical"`
	Parts           []string `json:"parts"`
	TreeParts       []string `json:"treeParts"`
	TreePath        string   `json:"treePath"`
	GitPath         string   `json:"gitPath"`
	ParentTreePath  string   `json:"parentTreePath"`
	DisplayPath     string   `json:"displayPath"`
	Leaf            string   `json:"leaf"`
}

type DomainDTO struct {
	Name               string              `json:"name"`
	Canonical          string              `json:"canonical"`
	CanonicalName      string              `json:"canonicalName"`
	Namespace          NamespaceDTO        `json:"namespace"`
	Label              string              `json:"label"`
	NamespaceName      string              `json:"namespaceName"`
	ParentCanonical    string              `json:"parentCanonical"`
	ParentTreePath     string              `json:"parentTreePath"`
	TreePath           string              `json:"treePath"`
	GitPath            string              `json:"gitPath"`
	VerificationStatus string              `json:"verificationStatus"`
	DisplayName        string              `json:"displayName"`
	Owner              string              `json:"owner"`
	Status             string              `json:"status"`
	Path               string              `json:"path"`
	ServiceCount       int                 `json:"serviceCount"`
	ProductCount       int                 `json:"productCount"`
	Persisted          bool                `json:"persisted"`
	Virtual            bool                `json:"virtual"`
	IsFolder           bool                `json:"isFolder"`
	Products           []ProductSummaryDTO `json:"products,omitempty"`
	Services           []ServiceDTO        `json:"services,omitempty"`
	Decisions          []DecisionDTO       `json:"decisions,omitempty"`
}

type MethodParameterDTO struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	In          string `json:"in,omitempty"` // path | query | header
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
}

type MethodHeaderDTO struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
}

type MethodSecurityDTO struct {
	Scheme string `json:"scheme"`
	In     string `json:"in,omitempty"`
	Name   string `json:"name,omitempty"`
}

type MethodPayloadFieldDTO struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
	Example     string `json:"example,omitempty"`
}

type MethodPayloadDTO struct {
	ContentType string                  `json:"content_type,omitempty"`
	Fields      []MethodPayloadFieldDTO `json:"fields,omitempty"`
}

type MethodDefinitionDTO struct {
	Name       string               `json:"name"`
	Summary    string               `json:"summary,omitempty"`
	HTTPMethod string               `json:"http_method,omitempty"`
	Path       string               `json:"path,omitempty"`
	Parameters []MethodParameterDTO `json:"parameters,omitempty"`
	Headers    []MethodHeaderDTO    `json:"headers,omitempty"`
	Security   *MethodSecurityDTO   `json:"security,omitempty"`
	Payload    *MethodPayloadDTO    `json:"payload,omitempty"`
}

type ConnectorDTO struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Invocation  string `json:"invocation,omitempty"`
	Method      string `json:"method,omitempty"`
	Path        string `json:"path,omitempty"`
	Auth        string `json:"auth,omitempty"`
	Tool        string `json:"tool,omitempty"`
	Kind        string `json:"kind,omitempty"`
	// Collaboration (BPMN)
	ArtifactRef  string `json:"artifact_ref,omitempty"`
	ConsumerPool string `json:"consumer_pool,omitempty"`
	ProviderPool string `json:"provider_pool,omitempty"`
}

type ServiceCapabilityDTO struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Summary        string         `json:"summary,omitempty"`
	Stability      string         `json:"stability,omitempty"`
	SideEffect     string         `json:"side_effect,omitempty"`
	Connectors     []ConnectorDTO `json:"connectors,omitempty"`
	RelatedUCI     []string       `json:"related_uci,omitempty"`
	MethodRefs     []string       `json:"method_refs,omitempty"`
	DataObjectRefs []string       `json:"data_object_refs,omitempty"`
	ConnectorTypes string         `json:"connector_types,omitempty"` // e.g. "CLI · REST · MCP"
}

type ServiceDataObjectDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Summary   string `json:"summary,omitempty"`
	Schema    string `json:"schema,omitempty"`
	Format    string `json:"format,omitempty"`
	Stability string `json:"stability,omitempty"`
}

type ServiceUserInterfaceDTO struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Summary       string          `json:"summary,omitempty"`
	Channel       string          `json:"channel,omitempty"`
	URL           string          `json:"url,omitempty"`
	Stability     string          `json:"stability,omitempty"`
	Engine        string          `json:"engine,omitempty"`
	EngineVersion string          `json:"engine_version,omitempty"`
	Schema        map[string]any  `json:"schema,omitempty"`
	Binding       *ViewBindingDTO `json:"binding,omitempty"`
}

type ViewBindingDTO struct {
	DataObjectRef string `json:"data_object_ref,omitempty"`
	Submit        string `json:"submit,omitempty"`
}

type ServiceDTO struct {
	ID                string                    `json:"id,omitempty"`
	Name              string                    `json:"name"`
	Owner             string                    `json:"owner"`
	Capabilities      []string                  `json:"capabilities,omitempty"`
	CapabilityDefs    []ServiceCapabilityDTO    `json:"capability_defs,omitempty"`
	DataObjects       []string                  `json:"data_objects,omitempty"`
	DataObjectDefs    []ServiceDataObjectDTO    `json:"data_object_defs,omitempty"`
	UserInterfaces    []string                  `json:"user_interfaces,omitempty"`
	UserInterfaceDefs []ServiceUserInterfaceDTO `json:"user_interface_defs,omitempty"`
	SupportedProducts []string                  `json:"supported_products,omitempty"`
	Methods           []MethodDefinitionDTO     `json:"methods,omitempty"`
	Status            string                    `json:"status"`
	Path              string                    `json:"path"`
}

type DomainsDTO struct {
	Domains []DomainDTO `json:"domains"`
}
type ServicesDTO struct {
	Services []ServiceDTO `json:"services"`
}

type ValidationResultDTO struct {
	Status   string       `json:"status"`
	Findings []FindingDTO `json:"findings"`
}
type FindingDTO struct {
	Code         string `json:"code"`
	Severity     string `json:"severity"`
	Message      string `json:"message"`
	Path         string `json:"path,omitempty"`
	ArtifactType string `json:"artifact_type,omitempty"`
	ArtifactID   string `json:"artifact_id,omitempty"`
	Suggestion   string `json:"suggestion,omitempty"`
}
type GraphDTO struct {
	Format  string `json:"format"`
	Content string `json:"content"`
}

type NamespaceTreeDTO struct {
	Root NamespaceTreeNodeDTO `json:"root"`
}
type NamespaceTreeNodeDTO struct {
	Label                string                     `json:"label"`
	Kind                 string                     `json:"kind"`
	Canonical            string                     `json:"canonical,omitempty"`
	CanonicalName        string                     `json:"canonicalName,omitempty"`
	GitPath              string                     `json:"gitPath,omitempty"`
	DisplayPath          string                     `json:"displayPath,omitempty"`
	TreePath             string                     `json:"treePath,omitempty"`
	Domain               *DomainDTO                 `json:"domain,omitempty"`
	Service              *ServiceDTO                `json:"service,omitempty"`
	Product              *ProductSummaryDTO         `json:"product,omitempty"`
	Fulfillment          *ProductRequiredServiceDTO `json:"fulfillment,omitempty"`
	ProcessStep          *ProcessStepSummaryDTO     `json:"processStep,omitempty"`
	Persisted            bool                       `json:"persisted"`
	Virtual              bool                       `json:"virtual"`
	IsFolder             bool                       `json:"isFolder,omitempty"`
	CanCreateChildDomain bool                       `json:"canCreateChildDomain"`
	CanAddService        bool                       `json:"canAddService"`
	CanOpenDetails       bool                       `json:"canOpenDetails"`
	CanVerifyDomain      bool                       `json:"canVerifyDomain"`
	CanMaterializeDomain bool                       `json:"canMaterializeDomain"`
	FulfillmentCount     int                        `json:"fulfillmentCount,omitempty"`
	TreeTarget           string                     `json:"treeTarget,omitempty"`
	MethodName           string                     `json:"methodName,omitempty"`
	Decision             *DecisionDTO               `json:"decision,omitempty"`
	Server               *ServerDTO                 `json:"server,omitempty"`
	Repository           *RepositoryDTO             `json:"repository,omitempty"`
	Capability           *ServiceCapabilityDTO      `json:"capability,omitempty"`
	DataObject           *ServiceDataObjectDTO      `json:"dataObject,omitempty"`
	UserInterface        *ServiceUserInterfaceDTO   `json:"userInterface,omitempty"`
	Children             []NamespaceTreeNodeDTO     `json:"children,omitempty"`
}

type RequiredServiceRefDTO struct {
	ServiceRef          string `json:"service_ref"`
	ServiceBlueprintRef string `json:"service_blueprint_ref"`
	Purpose             string `json:"purpose,omitempty"`
	Required            bool   `json:"required"`
}

type StepInputBindingDTO struct {
	Name     string `json:"name"`
	Source   string `json:"source"`
	Required bool   `json:"required,omitempty"`
}

type StepOutputSchemaDTO struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type GatewayConditionDTO struct {
	Output     string `json:"output"`
	Operator   string `json:"operator"`
	Value      string `json:"value"`
	TargetStep string `json:"target_step,omitempty"`
	Label      string `json:"label,omitempty"`
}

type DecisionGatewayDTO struct {
	Name       string                `json:"name,omitempty"`
	DefaultTo  string                `json:"default_to,omitempty"`
	Conditions []GatewayConditionDTO `json:"conditions,omitempty"`
}

type DecisionRuleDTO struct {
	Input       string `json:"input"`
	Operator    string `json:"operator"`
	Value       string `json:"value"`
	Output      string `json:"output"`
	OutputValue string `json:"output_value"`
	Label       string `json:"label,omitempty"`
}

type DecisionTableDTO struct {
	Name        string            `json:"name,omitempty"`
	HitPolicy   string            `json:"hit_policy,omitempty"`
	Description string            `json:"description,omitempty"`
	Rules       []DecisionRuleDTO `json:"rules,omitempty"`
}

type ProcessStepSummaryDTO struct {
	StepNum     int                   `json:"step_num"`
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	TaskType    string                `json:"task_type,omitempty"`
	ServiceRef  string                `json:"service_ref"`
	DecisionRef string                `json:"decision_ref,omitempty"`
	Method      string                `json:"method,omitempty"`
	Role        string                `json:"role,omitempty"`
	Required    bool                  `json:"required"`
	DependsOn   []string              `json:"depends_on,omitempty"`
	Inputs      []StepInputBindingDTO `json:"inputs,omitempty"`
	Outputs     []StepOutputSchemaDTO `json:"outputs,omitempty"`
	Decision    *DecisionTableDTO     `json:"decision,omitempty"`
	Gateway     *DecisionGatewayDTO   `json:"gateway,omitempty"`
}

type ProcessGroupDTO struct {
	ProcessID   string                  `json:"process_id"`
	ProcessName string                  `json:"process_name"`
	Steps       []ProcessStepSummaryDTO `json:"steps"`
}

type ProductSummaryDTO struct {
	ID                               string                `json:"id"`
	Name                             string                `json:"name"`
	Version                          string                `json:"version"`
	Status                           string                `json:"status"`
	OfferedBy                        string                `json:"offered_by"`
	OwningDomain                     string                `json:"owning_domain"`
	SourcePath                       string                `json:"source_path,omitempty"`
	CatalogPath                      string                `json:"catalog_path,omitempty"`
	FulfillmentRequiredServicesCount int                   `json:"fulfillment_required_services_count"`
	FulfillmentUnresolvedCount       int                   `json:"fulfillment_unresolved_count"`
	ProcessCount                     int                   `json:"process_count"`
	UnmappedTaskCount                int                   `json:"unmapped_task_count"`
	Fulfillment                      ProductFulfillmentDTO `json:"fulfillment,omitempty"`
	Processes                        []ProcessGroupDTO     `json:"processes,omitempty"`
}

type ServiceRefDTO struct {
	Service    string `json:"service"`
	ServiceRef string `json:"service_ref"`
}

type ServiceRefsDTO struct {
	Services []ServiceRefDTO `json:"services"`
}

type CreateProductOfferingRequest struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Status       string   `json:"status"`
	Summary      string   `json:"summary"`
	Description  string   `json:"description"`
	Owner        string   `json:"owner"`
	OwningDomain string   `json:"owning_domain"`
	Tags         []string `json:"tags"`
}

type MoveProductOfferingRequest struct {
	ProductID          string `json:"product_id"`
	TargetDomain       string `json:"target_domain"`
	UpdateOwningDomain bool   `json:"update_owning_domain"`
}

type AddFulfillmentServiceRequest struct {
	ServiceRef  string `json:"service_ref"`
	Role        string `json:"role"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// UpdateFulfillmentServiceRequest updates an existing product fulfillment entry by stable index.
type UpdateFulfillmentServiceRequest = AddFulfillmentServiceRequest

type ProductFulfillmentDTO struct {
	RequiredServices []ProductRequiredServiceDTO `json:"required_services,omitempty"`
}

type ServiceLevelDTO struct {
	Name          string `json:"name,omitempty"`
	Owner         string `json:"owner,omitempty"`
	Target        string `json:"target,omitempty"`
	Availability  string `json:"availability,omitempty"`
	SupportWindow string `json:"support_window,omitempty"`
	Description   string `json:"description,omitempty"`
}

type ProductRequiredServiceDTO struct {
	ServiceRef        string           `json:"service_ref"`
	Role              string           `json:"role,omitempty"`
	Required          bool             `json:"required"`
	Description       string           `json:"description,omitempty"`
	ResolutionStatus  string           `json:"resolution_status"`
	ResolvedDomain    string           `json:"resolved_domain,omitempty"`
	ResolvedService   string           `json:"resolved_service,omitempty"`
	FulfillmentType   string           `json:"fulfillment_type"`
	CrossDomain       bool             `json:"cross_domain"`
	SLARef            string           `json:"sla_ref,omitempty"`
	OLARef            string           `json:"ola_ref,omitempty"`
	TreeTarget        string           `json:"tree_target,omitempty"`
	SLA               *ServiceLevelDTO `json:"sla,omitempty"`
	OLA               *ServiceLevelDTO `json:"ola,omitempty"`
	ServiceLevelLabel string           `json:"service_level_label,omitempty"`
}

type BlueprintDTO struct {
	ID                        string                    `json:"id"`
	Type                      string                    `json:"type"`
	Name                      string                    `json:"name"`
	Version                   string                    `json:"version"`
	Status                    string                    `json:"status"`
	Owner                     string                    `json:"owner"`
	OfferedBy                 string                    `json:"offered_by,omitempty"`
	OwningDomain              string                    `json:"owning_domain,omitempty"`
	Fulfillment               ProductFulfillmentDTO     `json:"fulfillment,omitempty"`
	Summary                   string                    `json:"summary"`
	Purpose                   string                    `json:"purpose,omitempty"`
	Description               string                    `json:"description,omitempty"`
	Consumers                 []string                  `json:"consumers,omitempty"`
	LifecycleStatus           string                    `json:"lifecycle_status,omitempty"`
	Tags                      []string                  `json:"tags,omitempty"`
	Processes                 []string                  `json:"processes,omitempty"`
	ProcessSummary            ProcessesDTO              `json:"process_summary,omitempty"`
	Path                      string                    `json:"path"`
	PrimaryHome               string                    `json:"primary_home,omitempty"`
	PrimaryHomeDomain         string                    `json:"primary_home_domain,omitempty"`
	Variants                  []VariantDTO              `json:"variants,omitempty"`
	Capabilities              []string                  `json:"capabilities,omitempty"`
	TargetSystems             []string                  `json:"target_systems,omitempty"`
	RequiredInputs            []string                  `json:"required_inputs"`
	RequiredServiceBlueprints []string                  `json:"required_service_blueprints,omitempty"`
	RequiredServices          []RequiredServiceRefDTO   `json:"required_services,omitempty"`
	NamespaceServiceRef       string                    `json:"namespace_service_ref,omitempty"`
	Rules                     []string                  `json:"rules,omitempty"`
	QualityCriteria           []string                  `json:"quality_criteria"`
	EvidenceRequirements      []string                  `json:"evidence_requirements"`
	Requirements              []BlueprintRequirementDTO `json:"requirements,omitempty"`
	RequirementsStatus        string                    `json:"requirements_status,omitempty"`
	Attributes                []BlueprintAttributeDTO   `json:"attributes,omitempty"`
}

type BlueprintAttributeDTO struct {
	ID         string             `json:"id"`
	Label      string             `json:"label"`
	Type       string             `json:"type"`
	Required   bool               `json:"required"`
	ServiceRef string             `json:"service_ref,omitempty"`
	Rules      []AttributeRuleDTO `json:"rules,omitempty"`
}

type AttributeRuleDTO struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// AttributeValidationDTO is the result of validating an instance's attribute values.
type AttributeValidationDTO struct {
	Status     string                 `json:"status"` // valid, invalid, missing_values
	Attributes []AttrValidationResult `json:"attributes"`
}

type AttrValidationResult struct {
	AttributeID string          `json:"attribute_id"`
	Label       string          `json:"label"`
	Value       string          `json:"value,omitempty"`
	Status      string          `json:"status"` // valid, invalid, missing
	Rules       []RuleResultDTO `json:"rules,omitempty"`
}

type RuleResultDTO struct {
	RuleID  string `json:"rule_id"`
	Label   string `json:"label"`
	Type    string `json:"type"`
	Status  string `json:"status"` // pass, fail, manual
	Message string `json:"message,omitempty"`
}

type BlueprintRequirementDTO struct {
	ID            string   `json:"id"`
	Label         string   `json:"label"`
	Status        string   `json:"status"`
	AttributeRefs []string `json:"attribute_refs,omitempty"`
}

type VariantDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type BlueprintsDTO struct {
	Blueprints []BlueprintDTO `json:"blueprints"`
	Count      int            `json:"count"`
}

type InstanceDTO struct {
	ID                          string              `json:"id"`
	Type                        string              `json:"type"`
	Name                        string              `json:"name"`
	BlueprintRef                string              `json:"blueprint_ref"`
	BlueprintVersion            string              `json:"blueprint_version"`
	Status                      string              `json:"status"`
	Owner                       string              `json:"owner"`
	Path                        string              `json:"path"`
	Inputs                      map[string]string   `json:"inputs,omitempty"`
	ObservedState               map[string]string   `json:"observed_state,omitempty"`
	ProvisionedServiceInstances []string            `json:"provisioned_service_instances,omitempty"`
	OwningProductInstance       string              `json:"owning_product_instance,omitempty"`
	ProviderRef                 string              `json:"provider_ref,omitempty"`
	ComplianceStatus            string              `json:"compliance_status"`
	Evidence                    []EvidenceDTO       `json:"evidence,omitempty"`
	Findings                    []CatalogFindingDTO `json:"findings"`
	AttributeValues             map[string]string   `json:"attribute_values,omitempty"`
}

type EvidenceDTO struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

type CatalogFindingDTO struct {
	ID       string `json:"id,omitempty"`
	Severity string `json:"severity"`
	Category string `json:"category,omitempty"`
	Summary  string `json:"summary,omitempty"`
	Code     string `json:"code,omitempty"`
	Message  string `json:"message,omitempty"`
	Path     string `json:"path,omitempty"`
}

type InstancesDTO struct {
	Instances []InstanceDTO `json:"instances"`
	Count     int           `json:"count"`
}

type ComplianceDTO struct {
	InstanceID string              `json:"instance_id"`
	Status     string              `json:"status"`
	Evidence   []EvidenceDTO       `json:"evidence,omitempty"`
	Findings   []CatalogFindingDTO `json:"findings"`
}

type DoctorCheckDTO struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

type DoctorDTO struct {
	Status string           `json:"status"`
	Checks []DoctorCheckDTO `json:"checks"`
}

type VerificationEvidenceDTO struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Record    string `json:"record"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
}

type VerificationDTO struct {
	Evidence []VerificationEvidenceDTO `json:"evidence"`
	Count    int                       `json:"count"`
}

// ProcessParticipantDTO identifies the actor or system that executes a process,
// corresponding to a bpmn:participant (pool) in the collaboration diagram.
type ProcessParticipantDTO struct {
	Name string `json:"name"`
	Ref  string `json:"ref,omitempty"`
}

// ProcessLaneDTO mirrors a bpmn:lane and its service binding so clients
// can render lane→service relationships without re-parsing BPMN.
type ProcessLaneDTO struct {
	BPMNLaneID string `json:"bpmn_lane_id"`
	Name       string `json:"name,omitempty"`
	ServiceRef string `json:"service_ref,omitempty"`
}

// CollaborationParticipantDTO is one entry in a product's collaboration view,
// pairing a process with the participant (surrounding system) that executes it.
type CollaborationParticipantDTO struct {
	ProcessID   string                `json:"process_id"`
	ProcessName string                `json:"process_name"`
	Participant ProcessParticipantDTO `json:"participant"`
}

// CollaborationDTO represents the full BPMN collaboration for a product offering:
// one participant pool per process, showing all surrounding systems involved.
type CollaborationDTO struct {
	ProductID    string                        `json:"product_id"`
	ProductName  string                        `json:"product_name"`
	Participants []CollaborationParticipantDTO `json:"participants"`
}

type ProcessDTO struct {
	ID             string                  `json:"id"`
	Type           string                  `json:"type"`
	Name           string                  `json:"name"`
	Version        string                  `json:"version"`
	Status         string                  `json:"status"`
	Owner          string                  `json:"owner"`
	Summary        string                  `json:"summary,omitempty"`
	Tags           []string                `json:"tags,omitempty"`
	RelatedProduct string                  `json:"related_product"`
	Participant    *ProcessParticipantDTO  `json:"participant,omitempty"`
	Steps          []ProcessStepDTO        `json:"steps,omitempty"`
	Lanes          []ProcessLaneDTO        `json:"lanes,omitempty"`
	BPMN           BPMNReferenceDTO        `json:"bpmn"`
	TaskMappings   []ProcessTaskMappingDTO `json:"task_mappings,omitempty"`
	Triggers       []ProcessTriggerDTO     `json:"triggers,omitempty"`
	Path           string                  `json:"path,omitempty"`
	BPMNPath       string                  `json:"bpmn_path,omitempty"`
	Tasks          []BPMNTaskDTO           `json:"tasks,omitempty"`
	Validation     ProcessValidationDTO    `json:"validation"`
}

// ProcessTriggerDTO is the merged view of a start event's BPMN definition
// and its YAML-configured runtime binding. `DetectedType` is taken from the
// BPMN XML; `Type` and config sub-blocks come from the YAML.
type ProcessTriggerDTO struct {
	BPMNElementID string                       `json:"bpmn_element_id"`
	Name          string                       `json:"name,omitempty"`
	Description   string                       `json:"description,omitempty"`
	Type          string                       `json:"type"`
	DetectedType  string                       `json:"detected_type,omitempty"`
	Timer         *TimerTriggerConfigDTO       `json:"timer,omitempty"`
	Message       *MessageTriggerConfigDTO     `json:"message,omitempty"`
	Signal        *SignalTriggerConfigDTO      `json:"signal,omitempty"`
	Conditional   *ConditionalTriggerConfigDTO `json:"conditional,omitempty"`
	Configured    bool                         `json:"configured"`
}

type TimerTriggerConfigDTO struct {
	Cron        string `json:"cron,omitempty"`
	ISODuration string `json:"iso_duration,omitempty"`
	ISODate     string `json:"iso_date,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
}

type MessageTriggerConfigDTO struct {
	EventRef       string `json:"event_ref,omitempty"`
	Topic          string `json:"topic,omitempty"`
	Filter         string `json:"filter,omitempty"`
	CorrelationKey string `json:"correlation_key,omitempty"`
}

type SignalTriggerConfigDTO struct {
	SignalRef string `json:"signal_ref,omitempty"`
}

type ConditionalTriggerConfigDTO struct {
	Expression string `json:"expression,omitempty"`
}

type UpdateProcessTriggersRequest struct {
	Triggers []ProcessTriggerDTO `json:"triggers"`
}

type ProcessStepDTO struct {
	ID               string                `json:"id"`
	Name             string                `json:"name"`
	TaskType         string                `json:"task_type,omitempty"`
	ServiceRef       string                `json:"service_ref"`
	CapabilityRef    string                `json:"capability_ref,omitempty"`
	Method           string                `json:"method,omitempty"`
	UserInterfaceRef string                `json:"user_interface_ref,omitempty"`
	DecisionRef      string                `json:"decision_ref,omitempty"`
	Role             string                `json:"role,omitempty"`
	Required         bool                  `json:"required"`
	Notes            string                `json:"notes,omitempty"`
	DependsOn        []string              `json:"depends_on,omitempty"`
	Inputs           []StepInputBindingDTO `json:"inputs,omitempty"`
	Outputs          []StepOutputSchemaDTO `json:"outputs,omitempty"`
	Decision         *DecisionTableDTO     `json:"decision,omitempty"`
	Gateway          *DecisionGatewayDTO   `json:"gateway,omitempty"`
}

type UpsertProcessStepRequest struct {
	ID               string                `json:"id,omitempty"`
	Name             string                `json:"name"`
	TaskType         string                `json:"task_type"`
	ServiceRef       string                `json:"service_ref"`
	CapabilityRef    string                `json:"capability_ref"`
	Method           string                `json:"method"`
	UserInterfaceRef string                `json:"user_interface_ref"`
	DecisionRef      string                `json:"decision_ref"`
	Role             string                `json:"role"`
	Required         bool                  `json:"required"`
	Notes            string                `json:"notes"`
	DependsOn        []string              `json:"depends_on,omitempty"`
	Inputs           []StepInputBindingDTO `json:"inputs,omitempty"`
	Outputs          []StepOutputSchemaDTO `json:"outputs,omitempty"`
	Decision         *DecisionTableDTO     `json:"decision,omitempty"`
	Gateway          *DecisionGatewayDTO   `json:"gateway,omitempty"`
}

type BPMNReferenceDTO struct {
	File      string `json:"file"`
	ProcessID string `json:"process_id,omitempty"`
	Primary   bool   `json:"primary"`
}

type ProcessTaskMappingDTO struct {
	BPMNElementID   string `json:"bpmn_element_id"`
	TaskName        string `json:"task_name,omitempty"`
	BPMNElementType string `json:"bpmn_element_type,omitempty"`
	ServiceRef      string `json:"service_ref"`
	CapabilityRef   string `json:"capability_ref,omitempty"`
	Method          string `json:"method,omitempty"`
	Role            string `json:"role,omitempty"`
	Required        bool   `json:"required"`
	Notes           string `json:"notes,omitempty"`
}

type BPMNTaskDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ElementType   string `json:"element_type"`
	MappingStatus string `json:"mapping_status,omitempty"`
	ServiceRef    string `json:"service_ref,omitempty"`
	Method        string `json:"method,omitempty"`
	Standard      bool   `json:"standard,omitempty"`
}

type ProcessValidationDTO struct {
	Status   string       `json:"status"`
	Findings []FindingDTO `json:"findings"`
}

type ProcessesDTO struct {
	Items []ProcessDTO `json:"items"`
	Count int          `json:"count"`
}

type CreateProcessRequest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

type UpdateTaskMappingsRequest struct {
	TaskMappings []ProcessTaskMappingDTO `json:"task_mappings"`
}

type UpdateParticipantRequest struct {
	Name string `json:"name"`
	Ref  string `json:"ref,omitempty"`
}

type DecisionIODTO struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type DecisionDTO struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Number  string          `json:"number,omitempty"`
	Version string          `json:"version"`
	Status  string          `json:"status"`
	Owner   string          `json:"owner"`
	Summary string          `json:"summary,omitempty"`
	Context string          `json:"context,omitempty"`
	DMNFile string          `json:"dmn_file,omitempty"`
	HasDMN  bool            `json:"has_dmn"`
	Inputs  []DecisionIODTO `json:"inputs,omitempty"`
	Outputs []DecisionIODTO `json:"outputs,omitempty"`
	Path    string          `json:"path,omitempty"`
}

type DecisionsDTO struct {
	Items []DecisionDTO `json:"items"`
	Count int           `json:"count"`
}

type CreateDecisionRequest struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Number  string          `json:"number,omitempty"`
	Version string          `json:"version"`
	Status  string          `json:"status"`
	Owner   string          `json:"owner"`
	Summary string          `json:"summary,omitempty"`
	Context string          `json:"context,omitempty"`
	Inputs  []DecisionIODTO `json:"inputs,omitempty"`
	Outputs []DecisionIODTO `json:"outputs,omitempty"`
	// Folder is an optional workspace-relative folder within the decisions root
	// in which to create the decision (empty → the root itself).
	Folder string `json:"folder,omitempty"`
}

// UpdateDecisionRequest is the partial update payload for a decision.
// Version is intentionally absent: it is auto-managed (patch-bumped) on every
// material change so that each persisted trace points at exactly one immutable
// snapshot. Clients should not try to control it.
type UpdateDecisionRequest struct {
	Name    string          `json:"name,omitempty"`
	Number  string          `json:"number,omitempty"`
	Status  string          `json:"status,omitempty"`
	Owner   string          `json:"owner,omitempty"`
	Summary string          `json:"summary,omitempty"`
	Context string          `json:"context,omitempty"`
	Inputs  []DecisionIODTO `json:"inputs,omitempty"`
	Outputs []DecisionIODTO `json:"outputs,omitempty"`
}

// DecisionVersionDTO is one entry in the version snapshot list of a decision.
// RuleHash is the sha256 of the snapshot's DMN file (matching trace.rule_hash);
// it's omitted when the snapshot has no DMN attached.
type DecisionVersionDTO struct {
	Version  string `json:"version"`
	Status   string `json:"status,omitempty"`
	HasDMN   bool   `json:"has_dmn"`
	RuleHash string `json:"rule_hash,omitempty"`
}

// DecisionVersionsDTO is the response for listing a decision's version snapshots.
type DecisionVersionsDTO struct {
	DecisionID string               `json:"decision_id"`
	Current    string               `json:"current"`
	Items      []DecisionVersionDTO `json:"items"`
	Count      int                  `json:"count"`
}

// DecisionScenariosDTO is the response for listing decision scenarios.
type DecisionScenariosDTO struct {
	DecisionID string                   `json:"decision_id"`
	Count      int                      `json:"count"`
	Items      []model.DecisionScenario `json:"items"`
}

// DecisionTracesDTO is the response for listing decision traces.
type DecisionTracesDTO struct {
	DecisionID string                `json:"decision_id"`
	Count      int                   `json:"count"`
	Items      []model.DecisionTrace `json:"items"`
}

// TraceVerifyEntryDTO is the per-trace verification status.
type TraceVerifyEntryDTO struct {
	TraceID   string `json:"trace_id"`
	Timestamp string `json:"timestamp,omitempty"`
	OK        bool   `json:"ok"`
	Issue     string `json:"issue,omitempty"`
}

// DecisionTraceVerifyDTO is the response for the verify endpoint.
type DecisionTraceVerifyDTO struct {
	DecisionID string                `json:"decision_id"`
	Count      int                   `json:"count"`
	OK         bool                  `json:"ok"`
	Items      []TraceVerifyEntryDTO `json:"items"`
}
