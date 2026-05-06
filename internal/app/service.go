package app

import (
	"errors"
	"net/http"
	"os"
	"sort"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/graph"
	"github.com/nomos/nomos/internal/namespace"
	"github.com/nomos/nomos/internal/validate"
)

func GetCosmos(path string) (CosmosDTO, error) {
	tree, err := load(path)
	if err != nil {
		return CosmosDTO{}, err
	}
	serviceCount := 0
	for _, d := range tree.Domains {
		serviceCount += len(d.Services)
	}
	return CosmosDTO{Path: path, ID: tree.Cosmos.ID, Name: tree.Cosmos.Name, Version: tree.Cosmos.Version, Status: tree.Cosmos.Status, Owner: tree.Cosmos.Owner, DomainCount: len(tree.Domains), ServiceCount: serviceCount}, nil
}

func ListDomains(path string) (DomainsDTO, error) {
	tree, err := load(path)
	if err != nil {
		return DomainsDTO{}, err
	}
	out := DomainsDTO{Domains: []DomainDTO{}}
	for _, d := range tree.Domains {
		out.Domains = append(out.Domains, domainDTO(d, false))
	}
	sort.Slice(out.Domains, func(i, j int) bool { return out.Domains[i].Canonical < out.Domains[j].Canonical })
	return out, nil
}

func GetDomain(path, domainName string) (DomainDTO, error) {
	tree, err := load(path)
	if err != nil {
		return DomainDTO{}, err
	}
	canonical := namespace.Canonical(domainName)
	for _, d := range tree.Domains {
		if d.Name == canonical {
			return domainDTO(d, true), nil
		}
	}
	return DomainDTO{}, Error(CodeDomainNotFound, "Domain not found: "+canonical, http.StatusNotFound, nil)
}

func ListServices(path, domainName string) (ServicesDTO, error) {
	d, err := GetDomain(path, domainName)
	if err != nil {
		return ServicesDTO{}, err
	}
	return ServicesDTO{Domain: d.Canonical, Services: d.Services}, nil
}

func GetService(path, domainName, serviceName string) (ServiceDTO, error) {
	d, err := GetDomain(path, domainName)
	if err != nil {
		return ServiceDTO{}, err
	}
	for _, s := range d.Services {
		if s.Name == serviceName {
			return s, nil
		}
	}
	return ServiceDTO{}, Error(CodeServiceNotFound, "Service not found: "+d.Canonical+"/"+serviceName, http.StatusNotFound, nil)
}

func ValidateCosmos(path string) (ValidationResultDTO, error) {
	res, err := validate.Validate(path)
	out := ValidationResultDTO{Status: res.Status, Findings: []FindingDTO{}}
	for _, f := range res.Findings {
		out.Findings = append(out.Findings, FindingDTO{Code: f.Code, Severity: f.Severity, Message: f.Message, Path: f.Path})
	}
	if err != nil {
		return out, Error(CodeValidationFailed, "Validation failed", http.StatusInternalServerError, err)
	}
	return out, nil
}

func BuildGraph(path string) (GraphDTO, error) {
	tree, err := load(path)
	if err != nil {
		return GraphDTO{}, err
	}
	return GraphDTO{Format: "mermaid", Content: graph.Mermaid(tree)}, nil
}

func BuildNamespaceTree(path string) (NamespaceTreeDTO, error) {
	cosmos, err := GetCosmos(path)
	if err != nil {
		return NamespaceTreeDTO{}, err
	}
	domains, err := ListDomains(path)
	if err != nil {
		return NamespaceTreeDTO{}, err
	}
	root := NamespaceTreeNodeDTO{Label: fallback(cosmos.Name, "Local Cosmos"), Kind: "cosmos"}
	for _, d := range domains.Domains {
		full, err := GetDomain(path, d.Canonical)
		if err != nil {
			return NamespaceTreeDTO{}, err
		}
		insertDomain(&root, full)
	}
	sortTree(&root)
	return NamespaceTreeDTO{Root: root}, nil
}

func domainDTO(d cosmosfs.DomainNode, includeServices bool) DomainDTO {
	canonical := namespace.Canonical(d.Name)
	v := namespace.View(canonical)
	dto := DomainDTO{Name: canonical, Canonical: canonical, Namespace: NamespaceDTO(v), DisplayName: v.Leaf, Owner: fallback(d.Metadata.Owner, "unknown"), Status: fallback(d.Metadata.Status, "unknown"), Path: d.Path, ServiceCount: len(d.Services)}
	if includeServices {
		dto.Services = make([]ServiceDTO, 0, len(d.Services))
		for _, s := range d.Services {
			dto.Services = append(dto.Services, serviceDTO(canonical, s))
		}
	}
	return dto
}

func serviceDTO(domain string, s cosmosfs.ServiceNode) ServiceDTO {
	return ServiceDTO{Name: s.Name, Domain: domain, Owner: fallback(s.Metadata.Owner, "unknown"), Status: fallback(s.Metadata.Status, "unknown"), Path: s.Path}
}

func load(path string) (cosmosfs.Tree, error) {
	tree, err := cosmosfs.LoadTree(path)
	if err == nil {
		return tree, nil
	}
	code := CodeCosmosLoadFailed
	status := http.StatusInternalServerError
	if errors.Is(err, os.ErrNotExist) {
		code = CodeCosmosMissing
		status = http.StatusNotFound
	}
	return cosmosfs.Tree{}, Error(code, err.Error(), status, err)
}

func insertDomain(root *NamespaceTreeNodeDTO, d DomainDTO) {
	node := root
	for i, label := range d.Namespace.TreeParts {
		kind := "namespace"
		if i == len(d.Namespace.TreeParts)-1 {
			kind = "domain"
		}
		idx := -1
		for j := range node.Children {
			if node.Children[j].Label == label && node.Children[j].Kind == kind {
				idx = j
				break
			}
		}
		if idx == -1 {
			child := NamespaceTreeNodeDTO{Label: label, Kind: kind}
			if kind == "domain" {
				child.Canonical = d.Canonical
				child.DisplayPath = d.Namespace.DisplayPath
				child.TreePath = d.Namespace.TreePath
				dd := d
				child.Domain = &dd
			}
			node.Children = append(node.Children, child)
			idx = len(node.Children) - 1
		}
		node = &node.Children[idx]
	}
	for _, svc := range d.Services {
		s := svc
		node.Children = append(node.Children, NamespaceTreeNodeDTO{Label: svc.Name, Kind: "service", Canonical: d.Canonical + "/" + svc.Name, Service: &s})
	}
}

func sortTree(n *NamespaceTreeNodeDTO) {
	sort.Slice(n.Children, func(i, j int) bool { return n.Children[i].Label < n.Children[j].Label })
	for i := range n.Children {
		sortTree(&n.Children[i])
	}
}

func fallback(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
