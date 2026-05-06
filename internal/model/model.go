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
	ID       string   `yaml:"id" json:"id"`
	Type     string   `yaml:"type" json:"type"`
	Name     string   `yaml:"name" json:"name"`
	Version  string   `yaml:"version" json:"version"`
	Status   string   `yaml:"status" json:"status"`
	Owner    string   `yaml:"owner" json:"owner"`
	DNSName  string   `yaml:"dns_name" json:"dns_name"`
	Summary  string   `yaml:"summary" json:"summary"`
	Services []string `yaml:"services" json:"services"`
}

type Service struct {
	ID      string `yaml:"id" json:"id"`
	Type    string `yaml:"type" json:"type"`
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
	Status  string `yaml:"status" json:"status"`
	Owner   string `yaml:"owner" json:"owner"`
	Summary string `yaml:"summary" json:"summary"`
}

type Variant struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
}

type Blueprint struct {
	ID                        string    `yaml:"id" json:"id"`
	Type                      string    `yaml:"type" json:"type"`
	Name                      string    `yaml:"name" json:"name"`
	Version                   string    `yaml:"version" json:"version"`
	Status                    string    `yaml:"status" json:"status"`
	Owner                     string    `yaml:"owner" json:"owner"`
	Summary                   string    `yaml:"summary" json:"summary"`
	Variants                  []Variant `yaml:"variants" json:"variants,omitempty"`
	Capabilities              []string  `yaml:"capabilities" json:"capabilities,omitempty"`
	TargetSystems             []string  `yaml:"target_systems" json:"target_systems,omitempty"`
	Providers                 []string  `yaml:"providers" json:"providers,omitempty"`
	Actions                   []string  `yaml:"actions" json:"actions,omitempty"`
	RequiredInputs            []string  `yaml:"required_inputs" json:"required_inputs"`
	RequiredServiceBlueprints []string  `yaml:"required_service_blueprints" json:"required_service_blueprints,omitempty"`
	Rules                     []string  `yaml:"rules" json:"rules,omitempty"`
	QualityCriteria           []string  `yaml:"quality_criteria" json:"quality_criteria"`
	EvidenceRequirements      []string  `yaml:"evidence_requirements" json:"evidence_requirements"`
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
