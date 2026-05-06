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
