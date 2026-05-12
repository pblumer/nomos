package app

type CosmosDTO struct {
	Path               string `json:"path"`
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Version            string `json:"version"`
	Status             string `json:"status"`
	Owner              string `json:"owner"`
	DomainCount        int    `json:"domainCount"`
	VirtualDomainCount int    `json:"virtualDomainCount"`
	ServiceCount       int    `json:"serviceCount"`
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
	Products           []ProductSummaryDTO `json:"products,omitempty"`
	Services           []ServiceDTO        `json:"services,omitempty"`
}

type ServiceDTO struct {
	Name              string   `json:"name"`
	Domain            string   `json:"domain"`
	Owner             string   `json:"owner"`
	OwnedBy           string   `json:"owned_by,omitempty"`
	OperatedBy        []string `json:"operated_by,omitempty"`
	Capabilities      []string `json:"capabilities,omitempty"`
	SupportedProducts []string `json:"supported_products,omitempty"`
	Status            string   `json:"status"`
	Path              string   `json:"path"`
}

type DomainsDTO struct {
	Domains []DomainDTO `json:"domains"`
}
type ServicesDTO struct {
	Domain   string       `json:"domain"`
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
	Persisted            bool                       `json:"persisted"`
	Virtual              bool                       `json:"virtual"`
	CanCreateChildDomain bool                       `json:"canCreateChildDomain"`
	CanAddService        bool                       `json:"canAddService"`
	CanOpenDetails       bool                       `json:"canOpenDetails"`
	CanVerifyDomain      bool                       `json:"canVerifyDomain"`
	CanMaterializeDomain bool                       `json:"canMaterializeDomain"`
	FulfillmentCount     int                        `json:"fulfillmentCount,omitempty"`
	TreeTarget           string                     `json:"treeTarget,omitempty"`
	Children             []NamespaceTreeNodeDTO     `json:"children,omitempty"`
}

type RequiredServiceRefDTO struct {
	ServiceRef          string `json:"service_ref"`
	ServiceBlueprintRef string `json:"service_blueprint_ref"`
	Purpose             string `json:"purpose,omitempty"`
	Required            bool   `json:"required"`
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
	Fulfillment                      ProductFulfillmentDTO `json:"fulfillment,omitempty"`
}

type ServiceRefDTO struct {
	Domain     string `json:"domain"`
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
	Name         string `json:"name,omitempty"`
	Target       string `json:"target,omitempty"`
	Availability string `json:"availability,omitempty"`
	Description  string `json:"description,omitempty"`
}

type ProductRequiredServiceDTO struct {
	Index                int              `json:"index"`
	ServiceRef           string           `json:"service_ref"`
	Label                string           `json:"label"`
	Role                 string           `json:"role,omitempty"`
	Required             bool             `json:"required"`
	Description          string           `json:"description,omitempty"`
	ResolutionStatus     string           `json:"resolution_status"`
	ResolvedDomain       string           `json:"resolved_domain,omitempty"`
	ResolvedService      string           `json:"resolved_service,omitempty"`
	ResolvedLabel        string           `json:"resolved_label,omitempty"`
	FulfillmentType      string           `json:"fulfillment_type"`
	CrossDomain          bool             `json:"cross_domain"`
	SLARef               string           `json:"sla_ref,omitempty"`
	OLARef               string           `json:"ola_ref,omitempty"`
	TreeTarget           string           `json:"tree_target,omitempty"`
	ServiceSelection     string           `json:"service_selection,omitempty"`
	FulfillmentSelection string           `json:"fulfillment_selection"`
	SLA                  *ServiceLevelDTO `json:"sla,omitempty"`
	OLA                  *ServiceLevelDTO `json:"ola,omitempty"`
	ServiceLevelLabel    string           `json:"service_level_label,omitempty"`
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
	Domain    string `json:"domain"`
	Record    string `json:"record"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
}

type VerificationDTO struct {
	Evidence []VerificationEvidenceDTO `json:"evidence"`
	Count    int                       `json:"count"`
}
