package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/graph"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/namespace"
	"github.com/nomos/nomos/internal/storage"
	"github.com/nomos/nomos/internal/validate"
)

func GetCosmos(path string) (CosmosDTO, error) {
	tree, err := load(path)
	if err != nil {
		return CosmosDTO{}, err
	}
	serviceCount := 0
	persisted := map[string]bool{}
	virtual := map[string]bool{}
	for _, d := range tree.Domains {
		canonical := namespace.Canonical(d.Name)
		persisted[canonical] = true
		serviceCount += len(d.Services)
		parts := namespace.Parts(canonical)
		for len(parts) > 2 {
			parts = parts[1:]
			parent := strings.Join(parts, ".")
			if !persisted[parent] {
				virtual[parent] = true
			}
		}
	}
	for canonical := range persisted {
		delete(virtual, canonical)
	}
	return CosmosDTO{Path: path, ID: tree.Cosmos.ID, Name: tree.Cosmos.Name, Version: tree.Cosmos.Version, Status: tree.Cosmos.Status, Owner: tree.Cosmos.Owner, DomainCount: len(tree.Domains), VirtualDomainCount: len(virtual), ServiceCount: serviceCount}, nil
}

func ListDomains(path string) (DomainsDTO, error) {
	tree, err := load(path)
	if err != nil {
		return DomainsDTO{}, err
	}
	out := DomainsDTO{Domains: []DomainDTO{}}
	for _, d := range tree.Domains {
		out.Domains = append(out.Domains, domainDTO(tree, d, false))
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
	if strings.HasPrefix(canonical, "/") {
		if converted, err := namespace.TreePathToCanonical(canonical); err == nil {
			canonical = converted
		}
	}
	for _, d := range tree.Domains {
		if strings.EqualFold(d.Name, canonical) {
			return domainDTO(tree, d, true), nil
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
		out.Findings = append(out.Findings, FindingDTO{Code: f.Code, Severity: f.Severity, Message: f.Message, Path: f.Path, ArtifactType: f.ArtifactType, ArtifactID: f.ArtifactID, Suggestion: f.Suggestion})
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
	root := NamespaceTreeNodeDTO{Label: fallback(cosmos.Name, "Local Cosmos"), Kind: "cosmos", CanOpenDetails: true}
	namespaces := NamespaceTreeNodeDTO{Label: "Namespaces", Kind: "namespace-parent"}
	for _, d := range domains.Domains {
		full, err := GetDomain(path, d.Canonical)
		if err != nil {
			return NamespaceTreeDTO{}, err
		}
		insertDomain(&namespaces, full)
	}
	root.Children = append(root.Children, namespaces)
	sortTree(&root)
	return NamespaceTreeDTO{Root: root}, nil
}

func domainDTO(tree cosmosfs.Tree, d cosmosfs.DomainNode, includeServices bool) DomainDTO {
	canonical := namespace.Canonical(d.Name)
	v := namespace.View(canonical)
	products := productSummariesOfferedBy(tree, canonical)
	dto := DomainDTO{Name: canonical, Canonical: canonical, CanonicalName: canonical, Namespace: namespaceDTO(v), Label: v.Label, NamespaceName: v.Namespace, ParentCanonical: v.ParentCanonical, ParentTreePath: v.ParentTreePath, TreePath: v.TreePath, GitPath: v.GitPath, VerificationStatus: verificationStatus(fallback(d.Metadata.Status, "unknown")), DisplayName: v.Label, Owner: fallback(d.Metadata.Owner, "unknown"), Status: fallback(d.Metadata.Status, "unknown"), Path: d.Path, ServiceCount: len(d.Services), ProductCount: len(products), Persisted: true, Virtual: false}
	if includeServices {
		dto.Products = products
		dto.Services = make([]ServiceDTO, 0, len(d.Services))
		for _, s := range d.Services {
			dto.Services = append(dto.Services, serviceDTO(canonical, s))
		}
	}
	return dto
}

func serviceDTO(domain string, s cosmosfs.ServiceNode) ServiceDTO {
	ownedBy := firstNonEmpty(s.Metadata.OwnedBy, domain, s.Metadata.Owner)
	methods := make([]MethodDefinitionDTO, 0, len(s.Metadata.Methods))
	for _, m := range s.Metadata.Methods {
		params := make([]MethodParameterDTO, 0, len(m.Parameters))
		for _, p := range m.Parameters {
			params = append(params, MethodParameterDTO{Name: p.Name, Type: p.Type, Required: p.Required, Description: p.Description})
		}
		methods = append(methods, MethodDefinitionDTO{Name: m.Name, Summary: m.Summary, Parameters: params})
	}
	return ServiceDTO{Name: s.Name, Domain: domain, Owner: fallback(s.Metadata.Owner, "unknown"), OwnedBy: ownedBy, OperatedBy: s.Metadata.OperatedBy, Capabilities: s.Metadata.Capabilities, SupportedProducts: s.Metadata.SupportedProducts, Methods: methods, Status: fallback(s.Metadata.Status, "unknown"), Path: s.Path}
}

func AddServiceMethod(path, domainName, serviceName, method string) (ServiceDTO, error) {
	method = strings.TrimSpace(method)
	if method == "" {
		return ServiceDTO{}, Error(CodeInvalidInput, "method name is required", http.StatusBadRequest, nil)
	}
	svc, err := GetService(path, domainName, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	for _, m := range raw.Methods {
		if m.Name == method {
			return ServiceDTO{}, Error(CodeInvalidInput, "method already exists: "+method, http.StatusConflict, nil)
		}
	}
	raw.Methods = append(raw.Methods, model.MethodDefinition{Name: method})
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, domainName, serviceName)
}

func RemoveServiceMethod(path, domainName, serviceName, method string) (ServiceDTO, error) {
	svc, err := GetService(path, domainName, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	filtered := raw.Methods[:0]
	for _, m := range raw.Methods {
		if m.Name != method {
			filtered = append(filtered, m)
		}
	}
	raw.Methods = filtered
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, domainName, serviceName)
}

// UpdateMethodParameters replaces the parameter list of a named method on a service.
func UpdateMethodParameters(path, domainName, serviceName, methodName string, params []model.MethodParameter) (MethodDefinitionDTO, error) {
	svc, err := GetService(path, domainName, serviceName)
	if err != nil {
		return MethodDefinitionDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return MethodDefinitionDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	found := false
	for i, m := range raw.Methods {
		if m.Name == methodName {
			raw.Methods[i].Parameters = params
			found = true
			break
		}
	}
	if !found {
		return MethodDefinitionDTO{}, Error(CodeServiceNotFound, "method not found: "+methodName, http.StatusNotFound, nil)
	}
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return MethodDefinitionDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	updated, err := GetService(path, domainName, serviceName)
	if err != nil {
		return MethodDefinitionDTO{}, err
	}
	for _, m := range updated.Methods {
		if m.Name == methodName {
			return m, nil
		}
	}
	return MethodDefinitionDTO{Name: methodName}, nil
}

// GetServiceMethod returns the definition of a single named method on a service.
func GetServiceMethod(path, domainName, serviceName, methodName string) (MethodDefinitionDTO, error) {
	svc, err := GetService(path, domainName, serviceName)
	if err != nil {
		return MethodDefinitionDTO{}, err
	}
	for _, m := range svc.Methods {
		if m.Name == methodName {
			return m, nil
		}
	}
	return MethodDefinitionDTO{}, Error(CodeServiceNotFound, "method not found: "+methodName, http.StatusNotFound, nil)
}

func productSummariesOfferedBy(tree cosmosfs.Tree, domainCanonical string) []ProductSummaryDTO {
	canonical := namespace.Canonical(domainCanonical)
	out := []ProductSummaryDTO{}
	for _, b := range tree.Blueprints {
		if b.Metadata.Type != "product_blueprint" || namespace.Canonical(b.Metadata.OfferedBy) != canonical {
			continue
		}
		out = append(out, productSummaryDTO(tree, b.Metadata, b.Path))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func productSummaryDTO(tree cosmosfs.Tree, bp model.Blueprint, sourcePath string) ProductSummaryDTO {
	fulfillment := fulfillmentDTO(tree, bp)
	unresolved := 0
	for _, svc := range fulfillment.RequiredServices {
		if svc.ResolutionStatus != "resolved" {
			unresolved++
		}
	}
	processCount := 0
	unmappedTaskCount := 0
	var processGroups []ProcessGroupDTO
	if processes, err := scanProcesses(tree.Path); err == nil {
		processRefs := map[string]bool{}
		for _, id := range bp.Processes {
			processRefs[id] = true
		}
		for _, process := range processes {
			if process.Meta.RelatedProduct != bp.ID && !processRefs[process.Meta.ID] {
				continue
			}
			processCount++
			dto := processDTO(tree.Path, process, false)
			for _, task := range dto.Tasks {
				if task.MappingStatus == "unmapped" {
					unmappedTaskCount++
				}
			}
			group := ProcessGroupDTO{ProcessID: process.Meta.ID, ProcessName: firstNonEmpty(process.Meta.Name, process.Meta.ID)}
			for i, step := range process.Meta.Steps {
				group.Steps = append(group.Steps, ProcessStepSummaryDTO{
					StepNum:    i + 1,
					ID:         step.ID,
					Name:       step.Name,
					TaskType:   normalizedStepTaskType(step.TaskType),
					ServiceRef: step.ServiceRef,
					Method:     step.Method,
					Role:       step.Role,
					Required:   step.Required,
					DependsOn:  step.DependsOn,
					Inputs:     stepsInputsToDTO(step.Inputs),
					Outputs:    stepsOutputsToDTO(step.Outputs),
					Gateway:    gatewayToDTO(step.Gateway),
				})
			}
			processGroups = append(processGroups, group)
		}
	}
	return ProductSummaryDTO{ID: bp.ID, Name: bp.Name, Version: bp.Version, Status: bp.Status, OfferedBy: bp.OfferedBy, OwningDomain: firstNonEmpty(bp.OwningDomain, bp.OfferedBy), SourcePath: sourcePath, CatalogPath: sourcePath, FulfillmentRequiredServicesCount: len(fulfillment.RequiredServices), FulfillmentUnresolvedCount: unresolved, ProcessCount: processCount, UnmappedTaskCount: unmappedTaskCount, Fulfillment: fulfillment, Processes: processGroups}
}

func ProductsOfferedBy(path, domainCanonical string) ([]ProductSummaryDTO, error) {
	tree, err := load(path)
	if err != nil {
		return nil, err
	}
	if !resolveDomain(tree, domainCanonical) {
		return nil, Error(CodeDomainNotFound, "Domain not found: "+namespace.Canonical(domainCanonical), http.StatusNotFound, nil)
	}
	return productSummariesOfferedBy(tree, domainCanonical), nil
}

func ServicesOwnedBy(path, domainCanonical string) ([]ServiceDTO, error) {
	d, err := GetDomain(path, domainCanonical)
	if err != nil {
		return nil, err
	}
	return d.Services, nil
}

func AllServiceRefs(path string) (ServiceRefsDTO, error) {
	tree, err := load(path)
	if err != nil {
		return ServiceRefsDTO{}, err
	}
	out := ServiceRefsDTO{Services: []ServiceRefDTO{}}
	for _, d := range tree.Domains {
		domain := namespace.Canonical(d.Name)
		for _, s := range d.Services {
			out.Services = append(out.Services, ServiceRefDTO{Domain: domain, Service: s.Name, ServiceRef: domain + "/" + s.Name})
		}
	}
	sort.Slice(out.Services, func(i, j int) bool { return out.Services[i].ServiceRef < out.Services[j].ServiceRef })
	return out, nil
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

func namespaceDTO(v namespace.NamespaceView) NamespaceDTO {
	return NamespaceDTO{
		Canonical:       v.Canonical,
		CanonicalName:   v.CanonicalName,
		Namespace:       v.Namespace,
		Labels:          v.Labels,
		Label:           v.Label,
		ParentCanonical: v.ParentCanonical,
		Parts:           v.Parts,
		TreeParts:       v.TreeParts,
		TreePath:        v.TreePath,
		GitPath:         v.GitPath,
		ParentTreePath:  v.ParentTreePath,
		DisplayPath:     v.DisplayPath,
		Leaf:            v.Leaf,
	}
}

func verificationStatus(status string) string {
	switch strings.ToLower(status) {
	case "verified", "active", "ok", "compliant":
		return "verified"
	case "failed", "error", "blocking":
		return "failed"
	case "pending", "warning":
		return "pending"
	default:
		return status
	}
}

func insertDomain(root *NamespaceTreeNodeDTO, d DomainDTO) {
	node := root
	parts := d.Namespace.TreeParts
	for i, label := range parts {
		kind := "domain"
		if i == 0 {
			kind = "namespace"
		}
		idx := findTreeChild(node, label, kind)
		if idx == -1 {
			child := NamespaceTreeNodeDTO{Label: label, Kind: kind}
			if kind == "namespace" {
				child.DisplayPath = label
				child.TreePath = "/" + label
			} else {
				child.Virtual = true
				child.CanCreateChildDomain = true
				child.CanMaterializeDomain = true
				canonical, err := namespace.ComposeCanonical(parts[0], parts[1:i+1]...)
				if err == nil {
					child.Canonical = canonical
					child.CanonicalName = canonical
					v := namespace.View(canonical)
					child.DisplayPath = v.DisplayPath
					child.TreePath = v.TreePath
					child.GitPath = v.GitPath
				}
			}
			node.Children = append(node.Children, child)
			idx = len(node.Children) - 1
		}
		if kind == "domain" && i == len(parts)-1 {
			node.Children[idx].Persisted = true
			node.Children[idx].Virtual = false
			node.Children[idx].CanCreateChildDomain = true
			node.Children[idx].CanAddService = true
			node.Children[idx].CanOpenDetails = true
			node.Children[idx].CanVerifyDomain = true
			node.Children[idx].CanMaterializeDomain = false
			node.Children[idx].Canonical = d.Canonical
			node.Children[idx].CanonicalName = d.Canonical
			node.Children[idx].DisplayPath = d.Namespace.DisplayPath
			node.Children[idx].TreePath = d.Namespace.TreePath
			node.Children[idx].GitPath = d.GitPath
			dd := d
			node.Children[idx].Domain = &dd
		}
		node = &node.Children[idx]
	}
	if len(d.Products) > 0 {
		idx := findTreeChild(node, "Products", "product-parent")
		if idx == -1 {
			node.Children = append(node.Children, NamespaceTreeNodeDTO{Label: "Products", Kind: "product-parent", Canonical: d.Canonical, CanonicalName: d.Canonical, DisplayPath: d.Namespace.DisplayPath + " / Products", TreePath: d.Namespace.TreePath + "/products", GitPath: d.GitPath})
			idx = len(node.Children) - 1
		}
		productParent := &node.Children[idx]
		for _, product := range d.Products {
			p := product
			productNode := NamespaceTreeNodeDTO{Label: product.Name, Kind: "product", Canonical: product.ID, CanonicalName: d.Canonical, Product: &p, Persisted: true, CanOpenDetails: true}
			for _, group := range product.Processes {
				g := group
				label := fmt.Sprintf("%s (%d)", g.ProcessName, len(g.Steps))
				stepsParent := NamespaceTreeNodeDTO{Label: label, Kind: "process-steps-parent", Canonical: product.ID + "/" + g.ProcessID, CanonicalName: d.Canonical, Product: &p, Persisted: true, CanOpenDetails: true, FulfillmentCount: len(g.Steps), TreeTarget: g.ProcessID}
				for _, step := range g.Steps {
					s := step
					stepLabel := fmt.Sprintf("%d · %s", step.StepNum, step.Name)
					stepsParent.Children = append(stepsParent.Children, NamespaceTreeNodeDTO{Label: stepLabel, Kind: "process-step", Canonical: product.ID, CanonicalName: d.Canonical, Product: &p, ProcessStep: &s, Persisted: true, CanOpenDetails: true, TreeTarget: "service:" + step.ServiceRef})
				}
				productNode.Children = append(productNode.Children, stepsParent)
			}
			if len(product.Processes) == 0 && len(product.Fulfillment.RequiredServices) > 0 {
				fulfillmentParent := NamespaceTreeNodeDTO{Label: "Fulfillment Services", Kind: "product-fulfillment-parent", Canonical: product.ID, CanonicalName: d.Canonical, Product: &p, Persisted: true, CanOpenDetails: true, FulfillmentCount: len(product.Fulfillment.RequiredServices)}
				for _, svc := range product.Fulfillment.RequiredServices {
					f := svc
					label := firstNonEmpty(svc.ServiceRef, svc.ResolvedService)
					fulfillmentParent.Children = append(fulfillmentParent.Children, NamespaceTreeNodeDTO{Label: label, Kind: "product-fulfillment-service", Canonical: svc.ServiceRef, CanonicalName: d.Canonical, Product: &p, Fulfillment: &f, Persisted: true, CanOpenDetails: true, TreeTarget: svc.TreeTarget})
				}
				productNode.Children = append(productNode.Children, fulfillmentParent)
			}
			productParent.Children = append(productParent.Children, productNode)
		}
	}
	if len(d.Services) > 0 {
		idx := findTreeChild(node, "Services", "service-parent")
		if idx == -1 {
			node.Children = append(node.Children, NamespaceTreeNodeDTO{Label: "Services", Kind: "service-parent", Canonical: d.Canonical, CanonicalName: d.Canonical, DisplayPath: d.Namespace.DisplayPath + " / Services", TreePath: d.Namespace.TreePath + "/services", GitPath: d.GitPath})
			idx = len(node.Children) - 1
		}
		serviceParent := &node.Children[idx]
		for _, svc := range d.Services {
			s := svc
			svcNode := NamespaceTreeNodeDTO{Label: svc.Name, Kind: "service", Canonical: d.Canonical + "/" + svc.Name, CanonicalName: d.Canonical, Service: &s, Persisted: true, CanOpenDetails: true}
			for _, method := range svc.Methods {
				m := method
				svcNode.Children = append(svcNode.Children, NamespaceTreeNodeDTO{Label: m.Name, Kind: "service-method", Canonical: d.Canonical + "/" + svc.Name, CanonicalName: d.Canonical, MethodName: m.Name, Service: &s, Persisted: true, CanOpenDetails: true})
			}
			serviceParent.Children = append(serviceParent.Children, svcNode)
		}
	}
}

func findTreeChild(node *NamespaceTreeNodeDTO, label, kind string) int {
	for j := range node.Children {
		if node.Children[j].Label == label && node.Children[j].Kind == kind {
			return j
		}
	}
	return -1
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func ListBlueprints(path string) (BlueprintsDTO, error) {
	tree, err := load(path)
	if err != nil {
		return BlueprintsDTO{}, err
	}
	out := BlueprintsDTO{Blueprints: []BlueprintDTO{}}
	for _, b := range tree.Blueprints {
		out.Blueprints = append(out.Blueprints, blueprintDTO(tree, b))
	}
	out.Count = len(out.Blueprints)
	return out, nil
}

func GetBlueprint(path, id string) (BlueprintDTO, error) {
	items, err := ListBlueprints(path)
	if err != nil {
		return BlueprintDTO{}, err
	}
	for _, b := range items.Blueprints {
		if b.ID == id {
			return b, nil
		}
	}
	return BlueprintDTO{}, Error(CodeBlueprintNotFound, "Blueprint not found: "+id, http.StatusNotFound, nil)
}

func ListInstances(path string) (InstancesDTO, error) {
	tree, err := load(path)
	if err != nil {
		return InstancesDTO{}, err
	}
	out := InstancesDTO{Instances: []InstanceDTO{}}
	for _, i := range tree.Instances {
		out.Instances = append(out.Instances, instanceDTO(i))
	}
	out.Count = len(out.Instances)
	return out, nil
}

func GetInstance(path, id string) (InstanceDTO, error) {
	items, err := ListInstances(path)
	if err != nil {
		return InstanceDTO{}, err
	}
	for _, i := range items.Instances {
		if i.ID == id {
			return i, nil
		}
	}
	return InstanceDTO{}, Error(CodeInstanceNotFound, "Instance not found: "+id, http.StatusNotFound, nil)
}

func GetInstanceCompliance(path, id string) (ComplianceDTO, error) {
	inst, err := GetInstance(path, id)
	if err != nil {
		return ComplianceDTO{}, err
	}
	return ComplianceDTO{InstanceID: inst.ID, Status: inst.ComplianceStatus, Evidence: inst.Evidence, Findings: inst.Findings}, nil
}

func blueprintDTO(tree cosmosfs.Tree, b cosmosfs.BlueprintNode) BlueprintDTO {
	variants := make([]VariantDTO, 0, len(b.Metadata.Variants))
	for _, v := range b.Metadata.Variants {
		variants = append(variants, VariantDTO{ID: v.ID, Name: v.Name})
	}
	requiredServices := make([]RequiredServiceRefDTO, 0, len(b.Metadata.RequiredServices))
	for _, svc := range b.Metadata.RequiredServices {
		requiredServices = append(requiredServices, RequiredServiceRefDTO{ServiceRef: svc.ServiceRef, ServiceBlueprintRef: svc.ServiceBlueprintRef, Purpose: svc.Purpose, Required: svc.Required})
	}
	fulfillment := fulfillmentDTO(tree, b.Metadata)
	requirements := make([]BlueprintRequirementDTO, 0, len(b.Metadata.Requirements))
	for _, r := range b.Metadata.Requirements {
		requirements = append(requirements, BlueprintRequirementDTO{ID: r.ID, Label: r.Label, Status: r.Status, AttributeRefs: r.AttributeRefs})
	}
	attributes := make([]BlueprintAttributeDTO, 0, len(b.Metadata.Attributes))
	for _, a := range b.Metadata.Attributes {
		rules := make([]AttributeRuleDTO, 0, len(a.Rules))
		for _, rule := range a.Rules {
			rules = append(rules, AttributeRuleDTO{ID: rule.ID, Label: rule.Label, Type: rule.Type, Value: rule.Value})
		}
		attributes = append(attributes, BlueprintAttributeDTO{ID: a.ID, Label: a.Label, Type: a.Type, Required: a.Required, ServiceRef: a.ServiceRef, Rules: rules})
	}
	processSummary := ProcessesDTO{}
	if b.Metadata.Type == "product_blueprint" {
		if ps, err := ListProductProcesses(tree.Path, b.Metadata.ID); err == nil {
			processSummary = ps
		}
	}
	return BlueprintDTO{ID: b.Metadata.ID, Type: b.Metadata.Type, Name: b.Metadata.Name, Version: b.Metadata.Version, Status: b.Metadata.Status, Owner: b.Metadata.Owner, OfferedBy: b.Metadata.OfferedBy, OwningDomain: firstNonEmpty(b.Metadata.OwningDomain, b.Metadata.OfferedBy), Fulfillment: fulfillment, Summary: b.Metadata.Summary, Purpose: b.Metadata.Purpose, Description: b.Metadata.Description, Consumers: b.Metadata.Consumers, LifecycleStatus: b.Metadata.LifecycleStatus, Tags: b.Metadata.Tags, Processes: b.Metadata.Processes, ProcessSummary: processSummary, Path: b.Path, PrimaryHome: primaryProductHome(b.Metadata), PrimaryHomeDomain: namespace.Canonical(b.Metadata.OfferedBy), Variants: variants, Capabilities: b.Metadata.Capabilities, TargetSystems: b.Metadata.TargetSystems, RequiredInputs: b.Metadata.RequiredInputs, RequiredServiceBlueprints: b.Metadata.RequiredServiceBlueprints, RequiredServices: requiredServices, NamespaceServiceRef: b.Metadata.NamespaceServiceRef, Rules: b.Metadata.Rules, QualityCriteria: b.Metadata.QualityCriteria, EvidenceRequirements: b.Metadata.EvidenceRequirements, Requirements: requirements, RequirementsStatus: requirementsStatus(b.Metadata.Requirements), Attributes: attributes}
}

func fulfillmentDTO(tree cosmosfs.Tree, bp model.Blueprint) ProductFulfillmentDTO {
	required := make([]ProductRequiredServiceDTO, 0, len(bp.Fulfillment.RequiredServices)+len(bp.RequiredServices))
	for _, svc := range bp.Fulfillment.RequiredServices {
		res := resolveService(tree, svc.ServiceRef)
		required = append(required, productRequiredServiceDTO(bp, svc.ServiceRef, svc.Role, svc.Required, svc.Description, svc.SLARef, svc.OLARef, svc.SLA, svc.OLA, res))
	}
	if len(required) == 0 {
		for _, svc := range bp.RequiredServices {
			res := resolveService(tree, svc.ServiceRef)
			required = append(required, productRequiredServiceDTO(bp, svc.ServiceRef, "", svc.Required, svc.Purpose, "", "", nil, nil, res))
		}
	}
	return ProductFulfillmentDTO{RequiredServices: required}
}

type ServiceResolution struct {
	Status  string
	Domain  string
	Service string
	SLA     *model.ServiceLevelInfo
	OLA     *model.ServiceLevelInfo
}

func productRequiredServiceDTO(bp model.Blueprint, serviceRef, role string, required bool, description, slaRef, olaRef string, sla, ola *model.ServiceLevelInfo, res ServiceResolution) ProductRequiredServiceDTO {
	offeredBy := namespace.Canonical(bp.OfferedBy)
	fulfillmentType := "unresolved"
	crossDomain := false
	if res.Status == "resolved" {
		if offeredBy != "" && res.Domain != "" && res.Domain != offeredBy {
			fulfillmentType = "cross-domain"
			crossDomain = true
		} else {
			fulfillmentType = "local"
		}
	} else if res.Status == "unresolved_domain" {
		fulfillmentType = "unresolved"
	} else if res.Status == "unresolved_service" || res.Status == "missing" {
		fulfillmentType = "unresolved"
	}
	slaDTO := serviceLevelDTO(firstServiceLevel(sla, res.SLA))
	olaDTO := serviceLevelDTO(firstServiceLevel(ola, res.OLA))
	treeTarget := ""
	if res.Status == "resolved" {
		treeTarget = "service:" + res.Domain + "/" + res.Service
	}
	return ProductRequiredServiceDTO{ServiceRef: serviceRef, Role: role, Required: required, Description: description, ResolutionStatus: res.Status, ResolvedDomain: res.Domain, ResolvedService: res.Service, FulfillmentType: fulfillmentType, CrossDomain: crossDomain, SLARef: slaRef, OLARef: olaRef, TreeTarget: treeTarget, SLA: slaDTO, OLA: olaDTO, ServiceLevelLabel: serviceLevelLabel(slaRef, olaRef, slaDTO, olaDTO)}
}

func firstServiceLevel(values ...*model.ServiceLevelInfo) *model.ServiceLevelInfo {
	for _, v := range values {
		if v != nil && (strings.TrimSpace(v.Name) != "" || strings.TrimSpace(v.Owner) != "" || strings.TrimSpace(v.Target) != "" || strings.TrimSpace(v.Availability) != "" || strings.TrimSpace(v.SupportWindow) != "" || strings.TrimSpace(v.Description) != "") {
			return v
		}
	}
	return nil
}

func serviceLevelDTO(info *model.ServiceLevelInfo) *ServiceLevelDTO {
	if info == nil {
		return nil
	}
	return &ServiceLevelDTO{Name: info.Name, Owner: info.Owner, Target: info.Target, Availability: info.Availability, SupportWindow: info.SupportWindow, Description: info.Description}
}

func serviceLevelLabel(slaRef, olaRef string, sla, ola *ServiceLevelDTO) string {
	parts := []string{}
	if strings.TrimSpace(slaRef) != "" {
		parts = append(parts, "SLA")
	} else if sla != nil {
		parts = append(parts, compactServiceLevelLabel("SLA", sla))
	}
	if strings.TrimSpace(olaRef) != "" {
		parts = append(parts, "OLA")
	} else if ola != nil {
		parts = append(parts, compactServiceLevelLabel("OLA", ola))
	}
	return strings.Join(parts, " · ")
}

func compactServiceLevelLabel(kind string, info *ServiceLevelDTO) string {
	if strings.TrimSpace(info.Target) != "" {
		return kind + " " + strings.TrimSpace(info.Target)
	}
	if strings.TrimSpace(info.Name) != "" {
		return kind + " " + strings.TrimSpace(info.Name)
	}
	return kind
}

func primaryProductHome(bp model.Blueprint) string {
	if bp.Type != "product_blueprint" || strings.TrimSpace(bp.OfferedBy) == "" {
		return ""
	}
	name := firstNonEmpty(strings.TrimSpace(bp.Name), strings.TrimSpace(bp.ID))
	return namespace.Canonical(bp.OfferedBy) + " / Products / " + name
}

func normalizeServiceRef(serviceRef string) (string, string, string) {
	serviceRef = strings.TrimSpace(serviceRef)
	parts := strings.SplitN(serviceRef, "/", 2)
	if len(parts) != 2 {
		return serviceRef, "", ""
	}
	domain := namespace.Canonical(strings.TrimSpace(parts[0]))
	service := strings.TrimSpace(parts[1])
	return domain + "/" + service, domain, service
}

func resolveDomain(tree cosmosfs.Tree, canonical string) bool {
	canonical = namespace.Canonical(strings.TrimSpace(canonical))
	for _, d := range tree.Domains {
		if namespace.Canonical(d.Name) == canonical {
			return true
		}
	}
	return false
}

func resolveService(tree cosmosfs.Tree, serviceRef string) ServiceResolution {
	_, domain, service := normalizeServiceRef(serviceRef)
	if strings.TrimSpace(serviceRef) == "" {
		return ServiceResolution{Status: "missing"}
	}
	if domain == "" || service == "" {
		return ServiceResolution{Status: "unresolved_domain"}
	}
	for _, d := range tree.Domains {
		if namespace.Canonical(d.Name) != domain {
			continue
		}
		for _, s := range d.Services {
			if s.Name == service {
				return ServiceResolution{Status: "resolved", Domain: domain, Service: service, SLA: s.Metadata.SLA, OLA: s.Metadata.OLA}
			}
		}
		return ServiceResolution{Status: "unresolved_service", Domain: domain, Service: service}
	}
	return ServiceResolution{Status: "unresolved_domain", Domain: domain, Service: service}
}

// requirementsStatus derives the overall status from a set of requirements:
// no requirements → "" (null), any open → "open", all fulfilled → "fulfilled"
func requirementsStatus(reqs []model.BlueprintRequirement) string {
	if len(reqs) == 0 {
		return ""
	}
	for _, r := range reqs {
		if r.Status != "fulfilled" {
			return "open"
		}
	}
	return "fulfilled"
}

func instanceDTO(i cosmosfs.InstanceNode) InstanceDTO {
	evidence := make([]EvidenceDTO, 0, len(i.Metadata.Evidence))
	for _, e := range i.Metadata.Evidence {
		evidence = append(evidence, EvidenceDTO{ID: e.ID, Type: e.Type, Summary: e.Summary})
	}
	findings := make([]CatalogFindingDTO, 0, len(i.Metadata.Findings))
	for _, f := range i.Metadata.Findings {
		findings = append(findings, CatalogFindingDTO{ID: f.ID, Severity: f.Severity, Category: f.Category, Summary: f.Summary, Code: f.Code, Message: f.Message, Path: f.Path})
	}
	return InstanceDTO{ID: i.Metadata.ID, Type: i.Metadata.Type, Name: i.Metadata.Name, BlueprintRef: i.Metadata.BlueprintRef, BlueprintVersion: i.Metadata.BlueprintVersion, Status: i.Metadata.Status, Owner: i.Metadata.Owner, Path: i.Path, Inputs: i.Metadata.Inputs, ObservedState: i.Metadata.ObservedState, ProvisionedServiceInstances: i.Metadata.ProvisionedServiceInstances, OwningProductInstance: i.Metadata.OwningProductInstance, ProviderRef: i.Metadata.ProviderRef, ComplianceStatus: i.Metadata.ComplianceStatus, Evidence: evidence, Findings: findings, AttributeValues: i.Metadata.AttributeValues}
}

func DoctorCosmos(path string) (DoctorDTO, error) {
	checks := []DoctorCheckDTO{}
	add := func(name, status, message, p string) {
		checks = append(checks, DoctorCheckDTO{Name: name, Status: status, Message: message, Path: p})
	}
	cosmosYAML := storage.CosmosFile(path)
	if _, err := os.Stat(cosmosYAML); err != nil {
		add(".nomos/cosmos.yaml", "error", ".nomos/cosmos.yaml is missing", cosmosYAML)
	} else {
		add(".nomos/cosmos.yaml", "ok", ".nomos/cosmos.yaml exists", cosmosYAML)
	}
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		add("git repository", "warning", "Git repository is not initialized", gitDir)
	} else {
		add("git repository", "ok", "Git repository exists", gitDir)
	}
	nomosDir := storage.NomosDir(path)
	if _, err := os.Stat(nomosDir); err != nil {
		add(".nomos directory", "warning", ".nomos directory is not present yet", nomosDir)
	} else {
		add(".nomos directory", "ok", ".nomos directory exists", nomosDir)
	}
	status := "ok"
	for _, c := range checks {
		if c.Status == "error" {
			status = "error"
			break
		}
		if c.Status == "warning" && status == "ok" {
			status = "warning"
		}
	}
	return DoctorDTO{Status: status, Checks: checks}, nil
}

func domainDirFromTreePath(workspace, treePath string) string {
	return filepath.Join(storage.DomainsDir(workspace), filepath.FromSlash(strings.TrimPrefix(treePath, "/")))
}

func AddDomain(path, dns, owner string, force bool) (DomainDTO, error) {
	return addDomain(path, dns, owner, force, false)
}

func addDomain(path, dns, owner string, force bool, materializedFromTree bool) (DomainDTO, error) {
	dns = namespace.Canonical(strings.TrimSpace(dns))
	identity, err := namespace.Identity(dns)
	if err != nil || strings.ContainsAny(dns, `/\`) {
		return DomainDTO{}, Error(CodeInvalidNamespace, "Invalid domain name: "+dns, http.StatusBadRequest, err)
	}
	dns = identity.CanonicalName
	if strings.TrimSpace(owner) == "" {
		owner = "unknown"
	}
	ddir := domainDirFromTreePath(path, identity.TreePath)
	domainFile := filepath.Join(ddir, "domain.yaml")
	if _, err := os.Stat(domainFile); err == nil && !force {
		return DomainDTO{}, Error(CodeInvalidNamespace, "Domain already exists: "+dns, http.StatusConflict, nil)
	}
	if err := os.MkdirAll(filepath.Join(ddir, "services"), 0o755); err != nil {
		return DomainDTO{}, err
	}
	d := model.Domain{ID: dns, Type: "domain", Name: dns, Version: "0.1.0", Status: "draft", Owner: owner, DNSName: dns, Namespace: identity.Namespace, Label: identity.Label, Labels: identity.Labels, CanonicalName: dns, TreePath: identity.TreePath, ParentCanonical: identity.ParentCanonical, ParentTreePath: identity.ParentTreePath, MaterializedFromTree: materializedFromTree, Summary: "Nomos Domain " + dns + "."}
	if err := fsx.WriteYAML(domainFile, d); err != nil {
		return DomainDTO{}, err
	}
	if err := os.WriteFile(filepath.Join(ddir, "README.md"), []byte("# Domain\n"), 0o644); err != nil {
		return DomainDTO{}, err
	}
	return GetDomain(path, dns)
}

func DeleteDomain(path, domainName string) error {
	canonical := namespace.Canonical(strings.TrimSpace(domainName))
	if canonical == "" || !strings.Contains(canonical, ".") || strings.ContainsAny(canonical, `/\`) {
		return Error(CodeInvalidNamespace, "Invalid domain name: "+domainName, http.StatusBadRequest, nil)
	}
	d, err := GetDomain(path, canonical)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(d.Path); err != nil {
		return Error(CodeInternalError, "Failed to delete domain: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

func DeleteService(path, domainName, serviceName string) error {
	canonical := namespace.Canonical(strings.TrimSpace(domainName))
	d, err := GetDomain(path, canonical)
	if err != nil {
		return err
	}
	if _, err := GetService(path, canonical, serviceName); err != nil {
		return err
	}
	dir := filepath.Join(d.Path, "services", serviceName)
	if err := os.RemoveAll(dir); err != nil {
		return Error(CodeInternalError, "Failed to delete service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

func RenameDomain(path, oldName, newName string) error {
	oldCanon := namespace.Canonical(strings.TrimSpace(oldName))
	newIdentity, err := namespace.Identity(newName)
	if err != nil || strings.ContainsAny(newName, `/\`) {
		return Error(CodeInvalidNamespace, "Invalid domain name: "+newName, http.StatusBadRequest, err)
	}
	oldDomain, err := GetDomain(path, oldCanon)
	if err != nil {
		return err
	}
	newCanon := newIdentity.CanonicalName
	newDir := domainDirFromTreePath(path, newIdentity.TreePath)
	if _, err := os.Stat(newDir); err == nil {
		return Error(CodeInvalidNamespace, "Domain already exists: "+newCanon, http.StatusConflict, nil)
	}
	if err := os.Rename(oldDomain.Path, newDir); err != nil {
		return Error(CodeInternalError, "Failed to rename domain: "+err.Error(), http.StatusInternalServerError, err)
	}
	var d model.Domain
	f := filepath.Join(newDir, "domain.yaml")
	if err := fsx.ReadYAML(f, &d); err != nil {
		return Error(CodeInternalError, "read domain.yaml: "+err.Error(), http.StatusInternalServerError, err)
	}
	d.Name = newCanon
	d.DNSName = newCanon
	d.CanonicalName = newCanon
	d.Namespace = newIdentity.Namespace
	d.Label = newIdentity.Label
	d.Labels = newIdentity.Labels
	d.TreePath = newIdentity.TreePath
	d.ParentCanonical = newIdentity.ParentCanonical
	d.ParentTreePath = newIdentity.ParentTreePath
	return fsx.WriteYAML(f, d)
}

func RenameService(path, domainName, oldName, newName string) error {
	canonical := namespace.Canonical(strings.TrimSpace(domainName))
	d, err := GetDomain(path, canonical)
	if err != nil {
		return err
	}
	if _, err := GetService(path, canonical, oldName); err != nil {
		return err
	}
	if strings.TrimSpace(newName) == "" || strings.ContainsAny(newName, `/\\`) || newName == "." || newName == ".." {
		return Error(CodeInvalidNamespace, "Invalid service name: "+newName, http.StatusBadRequest, nil)
	}
	oldDir := filepath.Join(d.Path, "services", oldName)
	newDir := filepath.Join(d.Path, "services", newName)
	if _, err := os.Stat(newDir); err == nil {
		return Error(CodeInvalidNamespace, "Service already exists: "+newName, http.StatusConflict, nil)
	}
	if err := os.Rename(oldDir, newDir); err != nil {
		return Error(CodeInternalError, "Failed to rename service: "+err.Error(), http.StatusInternalServerError, err)
	}
	var s model.Service
	f := filepath.Join(newDir, "service.yaml")
	if err := fsx.ReadYAML(f, &s); err != nil {
		return Error(CodeInternalError, "read service.yaml: "+err.Error(), http.StatusInternalServerError, err)
	}
	s.Name = newName
	return fsx.WriteYAML(f, s)
}

func AddService(path, domainName, name, owner string, force bool) (ServiceDTO, error) {
	d, err := GetDomain(path, domainName)
	if err != nil {
		return ServiceDTO{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" || strings.ContainsAny(name, `/\\`) || name == "." || name == ".." {
		return ServiceDTO{}, Error(CodeInvalidNamespace, "Invalid service name: "+name, http.StatusBadRequest, nil)
	}
	if strings.TrimSpace(owner) == "" {
		owner = "unknown"
	}
	sdir := filepath.Join(d.Path, "services", name)
	if _, err := os.Stat(sdir); err == nil && !force {
		return ServiceDTO{}, Error(CodeInvalidNamespace, "Service already exists: "+d.Canonical+"/"+name, http.StatusConflict, nil)
	}
	for _, dir := range []string{"capabilities", "requirements", "rules", "processes", "skills", "findings", "evidence"} {
		if err := os.MkdirAll(filepath.Join(sdir, dir), 0o755); err != nil {
			return ServiceDTO{}, err
		}
	}
	s := model.Service{ID: "service-" + name, Type: "service", Name: name, Version: "0.1.0", Status: "draft", Owner: owner, OwnedBy: d.Canonical, OperatedBy: []string{d.Canonical}, Summary: "Nomos Service " + name + "."}
	if err := fsx.WriteYAML(filepath.Join(sdir, "service.yaml"), s); err != nil {
		return ServiceDTO{}, err
	}
	if err := os.WriteFile(filepath.Join(sdir, "README.md"), []byte("# Service\n"), 0o644); err != nil {
		return ServiceDTO{}, err
	}
	return GetService(path, d.Canonical, name)
}

func ListVerificationEvidence(path string) (VerificationDTO, error) {
	root := storage.EvidenceDir(path)
	out := VerificationDTO{Evidence: []VerificationEvidenceDTO{}}
	ents, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return out, err
	}
	for _, e := range ents {
		if e.IsDir() || !(strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml")) {
			continue
		}
		var ev struct {
			ID        string `yaml:"id"`
			Type      string `yaml:"type"`
			Domain    string `yaml:"domain"`
			Record    string `yaml:"record"`
			Status    string `yaml:"status"`
			Timestamp string `yaml:"timestamp"`
		}
		full := filepath.Join(root, e.Name())
		if err := fsx.ReadYAML(full, &ev); err != nil {
			return out, err
		}
		out.Evidence = append(out.Evidence, VerificationEvidenceDTO{ID: ev.ID, Type: ev.Type, Domain: ev.Domain, Record: ev.Record, Status: ev.Status, Timestamp: ev.Timestamp, Path: full})
	}
	out.Count = len(out.Evidence)
	return out, nil
}

func VerifyDomain(ctx context.Context, path, dns string) (VerificationEvidenceDTO, error) {
	dns = namespace.Canonical(strings.TrimSpace(dns))
	if dns == "" || !strings.Contains(dns, ".") {
		return VerificationEvidenceDTO{}, Error(CodeInvalidNamespace, "Invalid domain name: "+dns, http.StatusBadRequest, nil)
	}
	rec := "_nomos." + dns
	txt, lookupErr := net.DefaultResolver.LookupTXT(ctx, rec)
	status := "failed"
	exp := "nomos-domain=" + dns
	for _, t := range txt {
		if strings.Contains(t, exp) {
			status = "verified"
		}
	}
	if err := os.MkdirAll(storage.EvidenceDir(path), 0o755); err != nil {
		return VerificationEvidenceDTO{}, err
	}
	now := time.Now().UTC()
	ev := VerificationEvidenceDTO{ID: "evidence-" + now.Format("20060102-150405"), Type: "evidence", Domain: dns, Record: rec, Status: status, Timestamp: now.Format(time.RFC3339), Path: filepath.Join(storage.EvidenceDir(path), strings.ReplaceAll(dns, ".", "-")+"-dns.yaml")}
	content := fmt.Sprintf("id: %s\ntype: evidence\nevidence_type: dns_verification\ndomain: %s\nrecord: %s\nstatus: %s\ntimestamp: %q\n", ev.ID, ev.Domain, ev.Record, ev.Status, ev.Timestamp)
	if err := os.WriteFile(ev.Path, []byte(content), 0o644); err != nil {
		return ev, err
	}
	if lookupErr != nil || status != "verified" {
		return ev, Error(CodeValidationFailed, "DNS verification failed", http.StatusBadRequest, lookupErr)
	}
	return ev, nil
}

func CreateBlueprint(path string, bp model.Blueprint) error {
	if _, err := os.Stat(storage.CosmosFile(path)); err != nil {
		return Error(CodeCosmosMissing, ".nomos/cosmos.yaml not found", http.StatusNotFound, err)
	}
	if strings.TrimSpace(bp.ID) == "" {
		return Error(CodeInvalidInput, "Blueprint ID is required", http.StatusBadRequest, nil)
	}
	if bp.Type != "product_blueprint" && bp.Type != "service_blueprint" {
		return Error(CodeInvalidInput, "Blueprint type must be product_blueprint or service_blueprint", http.StatusBadRequest, nil)
	}

	existing, err := ListBlueprints(path)
	if err != nil {
		return err
	}
	for _, b := range existing.Blueprints {
		if b.ID == bp.ID {
			return Error(CodeBlueprintAlreadyExists, "Blueprint already exists: "+bp.ID, http.StatusConflict, nil)
		}
	}

	var subdir string
	if bp.Type == "product_blueprint" {
		subdir = "products"
	} else {
		subdir = "services"
	}
	dir := filepath.Join(storage.CatalogDir(path), "blueprints", subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Error(CodeInternalError, "Failed to create blueprint directory: "+err.Error(), http.StatusInternalServerError, err)
	}

	filePath := filepath.Join(dir, bp.ID+".yaml")
	if err := fsx.WriteYAML(filePath, bp); err != nil {
		return Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

func CreateProductOffering(path, domainCanonical string, req CreateProductOfferingRequest) (BlueprintDTO, error) {
	canonical := namespace.Canonical(strings.TrimSpace(domainCanonical))
	tree, err := load(path)
	if err != nil {
		return BlueprintDTO{}, err
	}
	if !resolveDomain(tree, canonical) {
		return BlueprintDTO{}, Error(CodeDomainNotFound, "Domain not found: "+canonical, http.StatusNotFound, nil)
	}
	id := strings.TrimSpace(req.ID)
	if id == "" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "Product ID is required", http.StatusBadRequest, nil)
	}
	version := firstNonEmpty(strings.TrimSpace(req.Version), "0.1.0")
	status := firstNonEmpty(strings.TrimSpace(req.Status), "draft")
	owner := firstNonEmpty(strings.TrimSpace(req.Owner), canonical)
	owning := firstNonEmpty(strings.TrimSpace(req.OwningDomain), canonical)
	bp := model.Blueprint{ID: id, Type: "product_blueprint", Name: strings.TrimSpace(req.Name), Version: version, Status: status, Owner: owner, OfferedBy: canonical, OwningDomain: namespace.Canonical(owning), Summary: strings.TrimSpace(req.Summary), Description: strings.TrimSpace(req.Description), Tags: req.Tags, Fulfillment: model.ProductFulfillment{RequiredServices: []model.ProductRequiredService{}}}
	if bp.Name == "" {
		bp.Name = id
	}
	if err := CreateBlueprint(path, bp); err != nil {
		return BlueprintDTO{}, err
	}
	return GetBlueprint(path, id)
}

func normalizeDomainInput(value string) string {
	v := strings.TrimSpace(value)
	v = strings.Trim(v, `"'`)
	return namespace.Canonical(strings.TrimSpace(v))
}

func validCanonicalDomain(value string) bool {
	if value == "" || strings.ContainsAny(value, `/\\`) || strings.ContainsAny(value, ` "'`) {
		return false
	}
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" || strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return false
		}
		for _, r := range part {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' {
				return false
			}
		}
	}
	return true
}

func MoveProductOffering(path string, req MoveProductOfferingRequest) (BlueprintDTO, error) {
	productID := strings.TrimSpace(firstNonEmpty(req.ProductID, ""))
	if productID == "" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "Product ID is required", http.StatusBadRequest, nil)
	}
	rawTarget := strings.TrimSpace(req.TargetDomain)
	if rawTarget == "" {
		return BlueprintDTO{}, Error(CodeProductMoveInvalidTarget, "Bitte eine Zieldomäne wählen.", http.StatusBadRequest, nil)
	}
	targetDomain := normalizeDomainInput(rawTarget)
	if !validCanonicalDomain(targetDomain) {
		return BlueprintDTO{}, Error(CodeProductMoveInvalidTarget, "Die gewählte Zieldomäne ist keine gültige kanonische Domäne.", http.StatusBadRequest, nil)
	}
	tree, err := load(path)
	if err != nil {
		return BlueprintDTO{}, err
	}
	if !resolveDomain(tree, targetDomain) {
		return BlueprintDTO{}, Error(CodeTargetDomainNotFound, "Die gewählte Zieldomäne existiert nicht.", http.StatusNotFound, nil)
	}
	var product *cosmosfs.BlueprintNode
	for i := range tree.Blueprints {
		b := &tree.Blueprints[i]
		if b.Metadata.ID == productID && b.Metadata.Type == "product_blueprint" {
			product = b
			break
		}
	}
	if product == nil {
		return BlueprintDTO{}, Error(CodeProductNotFound, "Product not found: "+productID, http.StatusNotFound, nil)
	}
	oldOfferedBy := namespace.Canonical(strings.TrimSpace(product.Metadata.OfferedBy))
	if oldOfferedBy == "" {
		return BlueprintDTO{}, Error(CodeProductMoveInvalidTarget, "Product has no current offered_by domain", http.StatusBadRequest, nil)
	}
	if oldOfferedBy == targetDomain {
		return BlueprintDTO{}, Error(CodeProductMoveNoop, "Bitte eine andere Zieldomäne wählen.", http.StatusConflict, nil)
	}

	var raw model.Blueprint
	if err := fsx.ReadYAML(product.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeProductMoveWriteFailed, "Failed to read product: "+err.Error(), http.StatusInternalServerError, err)
	}
	if raw.Type != "product_blueprint" {
		return BlueprintDTO{}, Error(CodeProductNotFound, "Product not found: "+productID, http.StatusNotFound, nil)
	}
	raw.OfferedBy = targetDomain
	oldOwning := namespace.Canonical(strings.TrimSpace(raw.OwningDomain))
	if oldOwning == "" || oldOwning == oldOfferedBy || req.UpdateOwningDomain {
		raw.OwningDomain = targetDomain
	}
	if err := fsx.WriteYAML(product.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeProductMoveWriteFailed, "Failed to write product: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, productID)
}

func AddProductFulfillmentService(path, productID string, req AddFulfillmentServiceRequest) (BlueprintDTO, error) {
	serviceRef := strings.TrimSpace(req.ServiceRef)
	if serviceRef == "" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "service_ref is required", http.StatusBadRequest, nil)
	}
	product, err := GetBlueprint(path, productID)
	if err != nil {
		return BlueprintDTO{}, err
	}
	if product.Type != "product_blueprint" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "target blueprint is not a product_blueprint", http.StatusBadRequest, nil)
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(product.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read product: "+err.Error(), http.StatusInternalServerError, err)
	}
	normalized, _, _ := normalizeServiceRef(serviceRef)
	for _, existing := range raw.Fulfillment.RequiredServices {
		existingNormalized, _, _ := normalizeServiceRef(existing.ServiceRef)
		if strings.EqualFold(existingNormalized, normalized) || strings.EqualFold(strings.TrimSpace(existing.ServiceRef), serviceRef) {
			return BlueprintDTO{}, Error(CodeInvalidInput, "Fulfillment service already exists for product: "+normalized, http.StatusConflict, nil)
		}
	}
	role := firstNonEmpty(strings.TrimSpace(req.Role), "supporting")
	raw.Fulfillment.RequiredServices = append(raw.Fulfillment.RequiredServices, model.ProductRequiredService{ServiceRef: normalized, Role: role, Required: req.Required, Description: strings.TrimSpace(req.Description)})
	if err := fsx.WriteYAML(product.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write product: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, productID)
}

func UpdateProductFulfillmentService(path, productID string, index int, req UpdateFulfillmentServiceRequest) (BlueprintDTO, error) {
	if index < 0 {
		return BlueprintDTO{}, Error(CodeInvalidInput, "fulfillment index is invalid", http.StatusBadRequest, nil)
	}
	serviceRef := strings.TrimSpace(req.ServiceRef)
	if serviceRef == "" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "service_ref is required", http.StatusBadRequest, nil)
	}
	product, err := GetBlueprint(path, productID)
	if err != nil {
		return BlueprintDTO{}, err
	}
	if product.Type != "product_blueprint" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "target blueprint is not a product_blueprint", http.StatusBadRequest, nil)
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(product.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read product: "+err.Error(), http.StatusInternalServerError, err)
	}
	materializeLegacyFulfillment(&raw)
	if index >= len(raw.Fulfillment.RequiredServices) {
		return BlueprintDTO{}, Error(CodeInvalidInput, "fulfillment index not found", http.StatusNotFound, nil)
	}
	normalized, _, _ := normalizeServiceRef(serviceRef)
	for i, existing := range raw.Fulfillment.RequiredServices {
		if i == index {
			continue
		}
		existingNormalized, _, _ := normalizeServiceRef(existing.ServiceRef)
		if strings.EqualFold(existingNormalized, normalized) || strings.EqualFold(strings.TrimSpace(existing.ServiceRef), serviceRef) {
			return BlueprintDTO{}, Error(CodeInvalidInput, "Fulfillment service already exists for product: "+normalized, http.StatusConflict, nil)
		}
	}
	role := firstNonEmpty(strings.TrimSpace(req.Role), "supporting")
	old := raw.Fulfillment.RequiredServices[index]
	raw.Fulfillment.RequiredServices[index] = model.ProductRequiredService{ServiceRef: normalized, Role: role, Required: req.Required, Description: strings.TrimSpace(req.Description), SLARef: old.SLARef, OLARef: old.OLARef, SLA: old.SLA, OLA: old.OLA}
	if err := fsx.WriteYAML(product.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write product: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, productID)
}

func RemoveProductFulfillmentService(path, productID string, index int) (BlueprintDTO, error) {
	if index < 0 {
		return BlueprintDTO{}, Error(CodeInvalidInput, "fulfillment index is invalid", http.StatusBadRequest, nil)
	}
	product, err := GetBlueprint(path, productID)
	if err != nil {
		return BlueprintDTO{}, err
	}
	if product.Type != "product_blueprint" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "target blueprint is not a product_blueprint", http.StatusBadRequest, nil)
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(product.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read product: "+err.Error(), http.StatusInternalServerError, err)
	}
	materializeLegacyFulfillment(&raw)
	if index >= len(raw.Fulfillment.RequiredServices) {
		return BlueprintDTO{}, Error(CodeInvalidInput, "fulfillment index not found", http.StatusNotFound, nil)
	}
	raw.Fulfillment.RequiredServices = append(raw.Fulfillment.RequiredServices[:index], raw.Fulfillment.RequiredServices[index+1:]...)
	if err := fsx.WriteYAML(product.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write product: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, productID)
}

func materializeLegacyFulfillment(bp *model.Blueprint) {
	if len(bp.Fulfillment.RequiredServices) > 0 || len(bp.RequiredServices) == 0 {
		return
	}
	bp.Fulfillment.RequiredServices = make([]model.ProductRequiredService, 0, len(bp.RequiredServices))
	for _, svc := range bp.RequiredServices {
		bp.Fulfillment.RequiredServices = append(bp.Fulfillment.RequiredServices, model.ProductRequiredService{ServiceRef: svc.ServiceRef, Required: svc.Required, Description: svc.Purpose})
	}
}

func DeleteBlueprint(path, id string) error {
	bp, err := GetBlueprint(path, id)
	if err != nil {
		return err
	}
	if err := os.Remove(bp.Path); err != nil {
		return Error(CodeInternalError, "Failed to delete blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

func AddServiceBlueprintToProduct(path, productID, serviceID string) (BlueprintDTO, error) {
	product, err := GetBlueprint(path, productID)
	if err != nil {
		return BlueprintDTO{}, err
	}
	if product.Type != "product_blueprint" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "target blueprint is not a product_blueprint", http.StatusBadRequest, nil)
	}
	svc, err := GetBlueprint(path, serviceID)
	if err != nil {
		return BlueprintDTO{}, Error(CodeInvalidInput, "service blueprint not found: "+serviceID, http.StatusBadRequest, nil)
	}
	if svc.Type != "service_blueprint" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "referenced blueprint is not a service_blueprint", http.StatusBadRequest, nil)
	}
	for _, existing := range product.RequiredServiceBlueprints {
		if existing == serviceID {
			return product, nil
		}
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(product.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	raw.RequiredServiceBlueprints = append(raw.RequiredServiceBlueprints, serviceID)
	if err := fsx.WriteYAML(product.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, productID)
}

func RemoveServiceBlueprintFromProduct(path, productID, serviceID string) (BlueprintDTO, error) {
	product, err := GetBlueprint(path, productID)
	if err != nil {
		return BlueprintDTO{}, err
	}
	if product.Type != "product_blueprint" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "target blueprint is not a product_blueprint", http.StatusBadRequest, nil)
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(product.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	filtered := raw.RequiredServiceBlueprints[:0]
	for _, s := range raw.RequiredServiceBlueprints {
		if s != serviceID {
			filtered = append(filtered, s)
		}
	}
	raw.RequiredServiceBlueprints = filtered
	if err := fsx.WriteYAML(product.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, productID)
}

func AddBlueprintRequirement(path, id, label string) (BlueprintDTO, error) {
	return AddBlueprintRequirementWithAttributeRefs(path, id, label, nil)
}

func AddBlueprintRequirementWithAttributeRefs(path, id, label string, attributeRefs []string) (BlueprintDTO, error) {
	bp, err := GetBlueprint(path, id)
	if err != nil {
		return BlueprintDTO{}, err
	}
	label = strings.TrimSpace(label)
	if label == "" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "requirement label is required", http.StatusBadRequest, nil)
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	cleanRefs, err := normalizeRequirementAttributeRefs(attributeRefs, raw.Attributes)
	if err != nil {
		return BlueprintDTO{}, err
	}
	reqID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	raw.Requirements = append(raw.Requirements, model.BlueprintRequirement{ID: reqID, Label: label, Status: "open", AttributeRefs: cleanRefs})
	if err := fsx.WriteYAML(bp.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, id)
}

func normalizeRequirementAttributeRefs(attributeRefs []string, attributes []model.BlueprintAttribute) ([]string, error) {
	known := make(map[string]bool, len(attributes))
	for _, attr := range attributes {
		known[attr.ID] = true
	}
	seen := make(map[string]bool, len(attributeRefs))
	cleanRefs := make([]string, 0, len(attributeRefs))
	for _, ref := range attributeRefs {
		ref = strings.TrimSpace(ref)
		if ref == "" || seen[ref] {
			continue
		}
		if !known[ref] {
			return nil, Error(CodeInvalidInput, "requirement references unknown attribute: "+ref, http.StatusBadRequest, nil)
		}
		seen[ref] = true
		cleanRefs = append(cleanRefs, ref)
	}
	return cleanRefs, nil
}

func SetBlueprintRequirementStatus(path, id, reqID, status string) (BlueprintDTO, error) {
	if status != "open" && status != "fulfilled" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "status must be 'open' or 'fulfilled'", http.StatusBadRequest, nil)
	}
	bp, err := GetBlueprint(path, id)
	if err != nil {
		return BlueprintDTO{}, err
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	found := false
	for i := range raw.Requirements {
		if raw.Requirements[i].ID == reqID {
			raw.Requirements[i].Status = status
			found = true
			break
		}
	}
	if !found {
		return BlueprintDTO{}, Error(CodeInvalidInput, "requirement not found: "+reqID, http.StatusNotFound, nil)
	}
	if err := fsx.WriteYAML(bp.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, id)
}

func DeleteBlueprintRequirement(path, id, reqID string) (BlueprintDTO, error) {
	bp, err := GetBlueprint(path, id)
	if err != nil {
		return BlueprintDTO{}, err
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	filtered := raw.Requirements[:0]
	for _, r := range raw.Requirements {
		if r.ID != reqID {
			filtered = append(filtered, r)
		}
	}
	raw.Requirements = filtered
	if err := fsx.WriteYAML(bp.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, id)
}

func PublishBlueprint(path, id string) (BlueprintDTO, error) {
	bp, err := GetBlueprint(path, id)
	if err != nil {
		return BlueprintDTO{}, err
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	raw.Status = "published"
	if err := fsx.WriteYAML(bp.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, id)
}

func ValidateBlueprint(path, id string) (ValidationResultDTO, error) {
	bp, err := GetBlueprint(path, id)
	if err != nil {
		return ValidationResultDTO{}, err
	}
	full, err := ValidateCosmos(path)
	if err != nil {
		return full, err
	}
	filtered := ValidationResultDTO{Status: "ok", Findings: []FindingDTO{}}
	for _, f := range full.Findings {
		if strings.Contains(f.Path, bp.ID) || f.ArtifactID == bp.ID {
			filtered.Findings = append(filtered.Findings, f)
		}
	}
	for _, f := range filtered.Findings {
		if f.Severity == "error" {
			filtered.Status = "failed"
			break
		}
	}
	return filtered, nil
}

func CreateInstance(path string, inst model.Instance) error {
	if _, err := os.Stat(storage.CosmosFile(path)); err != nil {
		return Error(CodeCosmosMissing, ".nomos/cosmos.yaml not found", http.StatusNotFound, err)
	}
	if strings.TrimSpace(inst.ID) == "" {
		return Error(CodeInvalidInput, "Instance ID is required", http.StatusBadRequest, nil)
	}
	if inst.Type != "product_instance" && inst.Type != "service_instance" {
		return Error(CodeInvalidInput, "Instance type must be product_instance or service_instance", http.StatusBadRequest, nil)
	}
	if strings.TrimSpace(inst.BlueprintRef) == "" {
		return Error(CodeInvalidInput, "blueprint_ref is required", http.StatusBadRequest, nil)
	}

	existing, err := ListInstances(path)
	if err != nil {
		return err
	}
	for _, i := range existing.Instances {
		if i.ID == inst.ID {
			return Error(CodeInstanceAlreadyExists, "Instance already exists: "+inst.ID, http.StatusConflict, nil)
		}
	}

	var subdir string
	if inst.Type == "product_instance" {
		subdir = "products"
	} else {
		subdir = "services"
	}
	dir := filepath.Join(storage.CatalogDir(path), "instances", subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Error(CodeInternalError, "Failed to create instance directory: "+err.Error(), http.StatusInternalServerError, err)
	}
	filePath := filepath.Join(dir, inst.ID+".yaml")
	return fsx.WriteYAML(filePath, inst)
}

func DeleteInstance(path, id string) error {
	inst, err := GetInstance(path, id)
	if err != nil {
		return err
	}
	if err := os.Remove(inst.Path); err != nil {
		return Error(CodeInternalError, "Failed to delete instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

func PatchBlueprint(path, id string, fields map[string]string) (BlueprintDTO, error) {
	bp, err := GetBlueprint(path, id)
	if err != nil {
		return BlueprintDTO{}, err
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	for k, v := range fields {
		switch k {
		case "name":
			raw.Name = v
		case "status":
			raw.Status = v
		case "version":
			raw.Version = v
		case "owner":
			raw.Owner = v
		case "summary":
			raw.Summary = v
		}
	}
	if err := fsx.WriteYAML(bp.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, id)
}

func PatchInstance(path, id string, fields map[string]string) (InstanceDTO, error) {
	inst, err := GetInstance(path, id)
	if err != nil {
		return InstanceDTO{}, err
	}
	var raw model.Instance
	if err := fsx.ReadYAML(inst.Path, &raw); err != nil {
		return InstanceDTO{}, Error(CodeInternalError, "Failed to read instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	for k, v := range fields {
		switch k {
		case "name":
			raw.Name = v
		case "status":
			raw.Status = v
		case "owner":
			raw.Owner = v
		case "compliance_status":
			raw.ComplianceStatus = v
		case "provider_ref":
			raw.ProviderRef = v
		}
	}
	if err := fsx.WriteYAML(inst.Path, raw); err != nil {
		return InstanceDTO{}, Error(CodeInternalError, "Failed to write instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetInstance(path, id)
}

func VerifyInstance(path, id string) (InstanceDTO, error) {
	inst, err := GetInstance(path, id)
	if err != nil {
		return InstanceDTO{}, err
	}
	var raw model.Instance
	if err := fsx.ReadYAML(inst.Path, &raw); err != nil {
		return InstanceDTO{}, Error(CodeInternalError, "Failed to read instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	now := time.Now().UTC()
	ev := model.Evidence{
		ID:      "evidence-verify-" + now.Format("20060102-150405"),
		Type:    "manual_verification",
		Summary: "Manual verification completed at " + now.Format(time.RFC3339),
	}
	raw.Evidence = append(raw.Evidence, ev)
	raw.ComplianceStatus = "compliant"
	if err := fsx.WriteYAML(inst.Path, raw); err != nil {
		return InstanceDTO{}, Error(CodeInternalError, "Failed to write instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetInstance(path, id)
}

// ProvisionServiceInstance creates a service_instance linked to an existing product_instance.
// It sets OwningProductInstance on the new instance and appends the new ID to the
// product's ProvisionedServiceInstances list atomically (best-effort).
func ProvisionServiceInstance(cosmosPath, productID string, inst model.Instance) (InstanceDTO, error) {
	product, err := GetInstance(cosmosPath, productID)
	if err != nil {
		return InstanceDTO{}, err
	}
	if product.Type != "product_instance" {
		return InstanceDTO{}, Error(CodeInvalidInput, "target instance is not a product_instance", http.StatusBadRequest, nil)
	}
	inst.Type = "service_instance"
	inst.OwningProductInstance = productID
	if err := CreateInstance(cosmosPath, inst); err != nil {
		return InstanceDTO{}, err
	}
	var productRaw model.Instance
	if err := fsx.ReadYAML(product.Path, &productRaw); err == nil {
		productRaw.ProvisionedServiceInstances = append(productRaw.ProvisionedServiceInstances, inst.ID)
		_ = fsx.WriteYAML(product.Path, productRaw)
	}
	return GetInstance(cosmosPath, inst.ID)
}
