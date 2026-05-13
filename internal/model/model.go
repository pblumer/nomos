package model

type Cosmos struct {
	ID      string   `yaml:"id" json:"id"`
	Type    string   `yaml:"type" json:"type"`
	Name    string   `yaml:"name" json:"name"`
	Version string   `yaml:"version" json:"version"`
	Status  string   `yaml:"status" json:"status"`
	Owner   string   `yaml:"owner" json:"owner"`
	Summary string   `yaml:"summary" json:"summary"`
	Domains []string `yaml:"domains" json:"domains"`
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

type Service struct {
	ID                string            `yaml:"id" json:"id"`
	Type              string            `yaml:"type" json:"type"`
	Name              string            `yaml:"name" json:"name"`
	Version           string            `yaml:"version" json:"version"`
	Status            string            `yaml:"status" json:"status"`
	Owner             string            `yaml:"owner" json:"owner"`
	OwnedBy           string            `yaml:"owned_by,omitempty" json:"owned_by,omitempty"`
	OperatedBy        []string          `yaml:"operated_by,omitempty" json:"operated_by,omitempty"`
	Capabilities      []string          `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	SupportedProducts []string          `yaml:"supported_products,omitempty" json:"supported_products,omitempty"`
	Summary           string            `yaml:"summary" json:"summary"`
	SLA               *ServiceLevelInfo `yaml:"sla,omitempty" json:"sla,omitempty"`
	OLA               *ServiceLevelInfo `yaml:"ola,omitempty" json:"ola,omitempty"`
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
	Steps          []ProcessStep        `yaml:"steps,omitempty" json:"steps,omitempty"`
	BPMN           BPMNReference        `yaml:"bpmn" json:"bpmn"`
	TaskMappings   []ProcessTaskMapping `yaml:"task_mappings,omitempty" json:"task_mappings,omitempty"`
	Notes          string               `yaml:"notes,omitempty" json:"notes,omitempty"`
}

// ProcessStep is one ordered step in a fulfillment process. Each step maps to
// a service call (ArchiMate function/trigger).
type ProcessStep struct {
	ID         string `yaml:"id" json:"id"`
	Name       string `yaml:"name" json:"name"`
	ServiceRef string `yaml:"service_ref" json:"service_ref"`
	Method     string `yaml:"method,omitempty" json:"method,omitempty"`
	Role       string `yaml:"role,omitempty" json:"role,omitempty"`
	Required   bool   `yaml:"required" json:"required"`
	Notes      string `yaml:"notes,omitempty" json:"notes,omitempty"`
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
