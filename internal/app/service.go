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

func ListBlueprints(path string) (BlueprintsDTO, error) {
	tree, err := load(path)
	if err != nil {
		return BlueprintsDTO{}, err
	}
	out := BlueprintsDTO{Blueprints: []BlueprintDTO{}}
	for _, b := range tree.Blueprints {
		out.Blueprints = append(out.Blueprints, blueprintDTO(b))
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

func blueprintDTO(b cosmosfs.BlueprintNode) BlueprintDTO {
	variants := make([]VariantDTO, 0, len(b.Metadata.Variants))
	for _, v := range b.Metadata.Variants {
		variants = append(variants, VariantDTO{ID: v.ID, Name: v.Name})
	}
	requiredServices := make([]RequiredServiceRefDTO, 0, len(b.Metadata.RequiredServices))
	for _, svc := range b.Metadata.RequiredServices {
		requiredServices = append(requiredServices, RequiredServiceRefDTO{ServiceRef: svc.ServiceRef, ServiceBlueprintRef: svc.ServiceBlueprintRef, Purpose: svc.Purpose, Required: svc.Required})
	}
	return BlueprintDTO{ID: b.Metadata.ID, Type: b.Metadata.Type, Name: b.Metadata.Name, Version: b.Metadata.Version, Status: b.Metadata.Status, Owner: b.Metadata.Owner, Summary: b.Metadata.Summary, Path: b.Path, Variants: variants, Capabilities: b.Metadata.Capabilities, TargetSystems: b.Metadata.TargetSystems, RequiredInputs: b.Metadata.RequiredInputs, RequiredServiceBlueprints: b.Metadata.RequiredServiceBlueprints, RequiredServices: requiredServices, NamespaceServiceRef: b.Metadata.NamespaceServiceRef, Rules: b.Metadata.Rules, QualityCriteria: b.Metadata.QualityCriteria, EvidenceRequirements: b.Metadata.EvidenceRequirements}
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
	return InstanceDTO{ID: i.Metadata.ID, Type: i.Metadata.Type, Name: i.Metadata.Name, BlueprintRef: i.Metadata.BlueprintRef, BlueprintVersion: i.Metadata.BlueprintVersion, Status: i.Metadata.Status, Owner: i.Metadata.Owner, Path: i.Path, Inputs: i.Metadata.Inputs, ObservedState: i.Metadata.ObservedState, ProvisionedServiceInstances: i.Metadata.ProvisionedServiceInstances, OwningProductInstance: i.Metadata.OwningProductInstance, ProviderRef: i.Metadata.ProviderRef, ComplianceStatus: i.Metadata.ComplianceStatus, Evidence: evidence, Findings: findings}
}

func DoctorCosmos(path string) (DoctorDTO, error) {
	checks := []DoctorCheckDTO{}
	add := func(name, status, message, p string) {
		checks = append(checks, DoctorCheckDTO{Name: name, Status: status, Message: message, Path: p})
	}
	cosmosYAML := filepath.Join(path, "cosmos.yaml")
	if _, err := os.Stat(cosmosYAML); err != nil {
		add("cosmos.yaml", "error", "cosmos.yaml is missing", cosmosYAML)
	} else {
		add("cosmos.yaml", "ok", "cosmos.yaml exists", cosmosYAML)
	}
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		add("git repository", "warning", "Git repository is not initialized", gitDir)
	} else {
		add("git repository", "ok", "Git repository exists", gitDir)
	}
	nomosDir := filepath.Join(path, ".nomos")
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

func AddDomain(path, dns, owner string, force bool) (DomainDTO, error) {
	dns = namespace.Canonical(strings.TrimSpace(dns))
	if dns == "" || !strings.Contains(dns, ".") || strings.ContainsAny(dns, `/\\`) {
		return DomainDTO{}, Error(CodeInvalidNamespace, "Invalid domain name: "+dns, http.StatusBadRequest, nil)
	}
	if strings.TrimSpace(owner) == "" {
		owner = "unknown"
	}
	ddir := filepath.Join(path, "domains", dns)
	if _, err := os.Stat(ddir); err == nil && !force {
		return DomainDTO{}, Error(CodeInvalidNamespace, "Domain already exists: "+dns, http.StatusConflict, nil)
	}
	if err := os.MkdirAll(filepath.Join(ddir, "services"), 0o755); err != nil {
		return DomainDTO{}, err
	}
	d := model.Domain{ID: "domain-" + strings.ReplaceAll(dns, ".", "-"), Type: "domain", Name: dns, Version: "0.1.0", Status: "draft", Owner: owner, DNSName: dns, Summary: "Nomos Domain " + dns + "."}
	if err := fsx.WriteYAML(filepath.Join(ddir, "domain.yaml"), d); err != nil {
		return DomainDTO{}, err
	}
	if err := os.WriteFile(filepath.Join(ddir, "README.md"), []byte("# Domain\n"), 0o644); err != nil {
		return DomainDTO{}, err
	}
	return GetDomain(path, dns)
}

func DeleteDomain(path, domainName string) error {
	canonical := namespace.Canonical(strings.TrimSpace(domainName))
	if canonical == "" || !strings.Contains(canonical, ".") || strings.ContainsAny(canonical, `/\\`) {
		return Error(CodeInvalidNamespace, "Invalid domain name: "+domainName, http.StatusBadRequest, nil)
	}
	_, err := GetDomain(path, canonical)
	if err != nil {
		return err
	}
	dir := filepath.Join(path, "domains", canonical)
	if err := os.RemoveAll(dir); err != nil {
		return Error(CodeInternalError, "Failed to delete domain: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

func DeleteService(path, domainName, serviceName string) error {
	canonical := namespace.Canonical(strings.TrimSpace(domainName))
	_, err := GetService(path, canonical, serviceName)
	if err != nil {
		return err
	}
	dir := filepath.Join(path, "domains", canonical, "services", serviceName)
	if err := os.RemoveAll(dir); err != nil {
		return Error(CodeInternalError, "Failed to delete service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
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
	sdir := filepath.Join(path, "domains", d.Canonical, "services", name)
	if _, err := os.Stat(sdir); err == nil && !force {
		return ServiceDTO{}, Error(CodeInvalidNamespace, "Service already exists: "+d.Canonical+"/"+name, http.StatusConflict, nil)
	}
	for _, dir := range []string{"capabilities", "requirements", "rules", "processes", "skills", "findings", "evidence"} {
		if err := os.MkdirAll(filepath.Join(sdir, dir), 0o755); err != nil {
			return ServiceDTO{}, err
		}
	}
	s := model.Service{ID: "service-" + name, Type: "service", Name: name, Version: "0.1.0", Status: "draft", Owner: owner, Summary: "Nomos Service " + name + "."}
	if err := fsx.WriteYAML(filepath.Join(sdir, "service.yaml"), s); err != nil {
		return ServiceDTO{}, err
	}
	if err := os.WriteFile(filepath.Join(sdir, "README.md"), []byte("# Service\n"), 0o644); err != nil {
		return ServiceDTO{}, err
	}
	return GetService(path, d.Canonical, name)
}

func ListVerificationEvidence(path string) (VerificationDTO, error) {
	root := filepath.Join(path, ".nomos", "evidence")
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
	if err := os.MkdirAll(filepath.Join(path, ".nomos", "evidence"), 0o755); err != nil {
		return VerificationEvidenceDTO{}, err
	}
	now := time.Now().UTC()
	ev := VerificationEvidenceDTO{ID: "evidence-" + now.Format("20060102-150405"), Type: "evidence", Domain: dns, Record: rec, Status: status, Timestamp: now.Format(time.RFC3339), Path: filepath.Join(path, ".nomos", "evidence", strings.ReplaceAll(dns, ".", "-")+"-dns.yaml")}
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
	if _, err := os.Stat(filepath.Join(path, "cosmos.yaml")); err != nil {
		return Error(CodeCosmosMissing, "cosmos.yaml not found", http.StatusNotFound, err)
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
	dir := filepath.Join(path, "catalog", "blueprints", subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Error(CodeInternalError, "Failed to create blueprint directory: "+err.Error(), http.StatusInternalServerError, err)
	}

	filePath := filepath.Join(dir, bp.ID+".yaml")
	if err := fsx.WriteYAML(filePath, bp); err != nil {
		return Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
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
