package app

type CosmosDTO struct {
	Path         string `json:"path"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Status       string `json:"status"`
	Owner        string `json:"owner"`
	DomainCount  int    `json:"domainCount"`
	ServiceCount int    `json:"serviceCount"`
}

type NamespaceDTO struct {
	Canonical   string   `json:"canonical"`
	Parts       []string `json:"parts"`
	TreeParts   []string `json:"treeParts"`
	TreePath    string   `json:"treePath"`
	DisplayPath string   `json:"displayPath"`
	Leaf        string   `json:"leaf"`
}

type DomainDTO struct {
	Name         string       `json:"name"`
	Canonical    string       `json:"canonical"`
	Namespace    NamespaceDTO `json:"namespace"`
	DisplayName  string       `json:"displayName"`
	Owner        string       `json:"owner"`
	Status       string       `json:"status"`
	Path         string       `json:"path"`
	ServiceCount int          `json:"serviceCount"`
	Services     []ServiceDTO `json:"services,omitempty"`
}

type ServiceDTO struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Owner  string `json:"owner"`
	Status string `json:"status"`
	Path   string `json:"path"`
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
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
}
type GraphDTO struct {
	Format  string `json:"format"`
	Content string `json:"content"`
}

type NamespaceTreeDTO struct {
	Root NamespaceTreeNodeDTO `json:"root"`
}
type NamespaceTreeNodeDTO struct {
	Label       string                 `json:"label"`
	Kind        string                 `json:"kind"`
	Canonical   string                 `json:"canonical,omitempty"`
	DisplayPath string                 `json:"displayPath,omitempty"`
	TreePath    string                 `json:"treePath,omitempty"`
	Domain      *DomainDTO             `json:"domain,omitempty"`
	Service     *ServiceDTO            `json:"service,omitempty"`
	Children    []NamespaceTreeNodeDTO `json:"children,omitempty"`
}

type BlueprintDTO struct {
	ID                        string       `json:"id"`
	Type                      string       `json:"type"`
	Name                      string       `json:"name"`
	Version                   string       `json:"version"`
	Status                    string       `json:"status"`
	Owner                     string       `json:"owner"`
	Summary                   string       `json:"summary"`
	Path                      string       `json:"path"`
	Variants                  []VariantDTO `json:"variants,omitempty"`
	Capabilities              []string     `json:"capabilities,omitempty"`
	TargetSystems             []string     `json:"target_systems,omitempty"`
	RequiredInputs            []string     `json:"required_inputs"`
	RequiredServiceBlueprints []string     `json:"required_service_blueprints,omitempty"`
	Rules                     []string     `json:"rules,omitempty"`
	QualityCriteria           []string     `json:"quality_criteria"`
	EvidenceRequirements      []string     `json:"evidence_requirements"`
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
