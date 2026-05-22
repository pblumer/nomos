package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/graph"
	"github.com/nomos/nomos/internal/idgen"
	"github.com/nomos/nomos/internal/idmigrate"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
	"github.com/nomos/nomos/internal/validate"
)

func GetCosmos(path string) (CosmosDTO, error) {
	tree, err := load(path)
	if err != nil {
		return CosmosDTO{}, err
	}
	return CosmosDTO{Path: path, ID: tree.Cosmos.ID, Name: tree.Cosmos.Name, Version: tree.Cosmos.Version, Status: tree.Cosmos.Status, Owner: tree.Cosmos.Owner, ServiceCount: len(tree.Services), DecisionCount: len(tree.Decisions)}, nil
}

func ListServices(path string) (ServicesDTO, error) {
	tree, err := load(path)
	if err != nil {
		return ServicesDTO{}, err
	}
	out := ServicesDTO{Services: []ServiceDTO{}}
	for _, s := range tree.Services {
		out.Services = append(out.Services, serviceDTO(s))
	}
	sort.Slice(out.Services, func(i, j int) bool { return out.Services[i].Name < out.Services[j].Name })
	return out, nil
}

func GetService(path, serviceName string) (ServiceDTO, error) {
	tree, err := load(path)
	if err != nil {
		return ServiceDTO{}, err
	}
	resolved, _ := idmigrate.Resolve(path, serviceName)
	for _, s := range tree.Services {
		if s.Name == serviceName || s.Metadata.ID == serviceName || s.Metadata.ID == resolved {
			return serviceDTO(s), nil
		}
	}
	return ServiceDTO{}, Error(CodeServiceNotFound, "Service not found: "+serviceName, http.StatusNotFound, nil)
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

// BuildNamespaceTree returns a flat cosmos tree grouping the top-level
// Services, Decisions, and Products of the cosmos. Domains no longer exist;
// the explorer renders these flat groups directly.
func BuildNamespaceTree(path string) (NamespaceTreeDTO, error) {
	cosmos, err := GetCosmos(path)
	if err != nil {
		return NamespaceTreeDTO{}, err
	}
	tree, err := load(path)
	if err != nil {
		return NamespaceTreeDTO{}, err
	}
	root := NamespaceTreeNodeDTO{Label: fallback(cosmos.Name, "Local Cosmos"), Kind: "cosmos", CanOpenDetails: true}

	servicesParent := NamespaceTreeNodeDTO{Label: "Services", Kind: "service-parent", CanAddService: true}
	for _, sn := range tree.Services {
		s := serviceDTO(sn)
		sd := s
		servicesParent.Children = append(servicesParent.Children, NamespaceTreeNodeDTO{Label: s.Name, Kind: "service", Canonical: s.Name, Service: &sd, Persisted: true, CanOpenDetails: true})
	}
	root.Children = append(root.Children, servicesParent)

	decisionsParent := NamespaceTreeNodeDTO{Label: "Decisions", Kind: "decision-parent"}
	for _, dn := range tree.Decisions {
		dec := decisionDTO(dn)
		dd := dec
		decisionsParent.Children = append(decisionsParent.Children, NamespaceTreeNodeDTO{Label: firstNonEmpty(dec.Name, dec.ID), Kind: "decision", Canonical: dec.ID, Decision: &dd, Persisted: true, CanOpenDetails: true})
	}
	root.Children = append(root.Children, decisionsParent)

	products := allProductSummaries(tree)
	productsParent := NamespaceTreeNodeDTO{Label: "Products", Kind: "product-parent"}
	for _, product := range products {
		p := product
		productsParent.Children = append(productsParent.Children, NamespaceTreeNodeDTO{Label: product.Name, Kind: "product", Canonical: product.ID, Product: &p, Persisted: true, CanOpenDetails: true})
	}
	root.Children = append(root.Children, productsParent)

	sortTree(&root)
	return NamespaceTreeDTO{Root: root}, nil
}

func serviceDTO(s cosmosfs.ServiceNode) ServiceDTO {
	operations := make([]OperationDTO, 0, len(s.Metadata.Operations))
	for _, op := range s.Metadata.Operations {
		operations = append(operations, operationDTO(op))
	}
	capNames := make([]string, 0, len(s.Metadata.Capabilities))
	var capDefs []ServiceCapabilityDTO
	for _, c := range s.Metadata.Capabilities {
		capNames = append(capNames, c.Name)
		if len(c.Connectors) > 0 || c.Summary != "" || c.Stability != "" || len(c.OperationRefs) > 0 || len(c.DataObjectRefs) > 0 {
			connDTOs := make([]ConnectorDTO, 0, len(c.Connectors))
			for _, cn := range c.Connectors {
				connDTOs = append(connDTOs, ConnectorDTO{Type: cn.Type, Description: cn.Description, Invocation: cn.Invocation, Method: cn.Method, Path: cn.Path, Auth: cn.Auth, Tool: cn.Tool, Kind: cn.Kind, ArtifactRef: cn.ArtifactRef, ConsumerPool: cn.ConsumerPool, ProviderPool: cn.ProviderPool})
			}
			capDefs = append(capDefs, ServiceCapabilityDTO{ID: c.ID, Name: c.Name, Summary: c.Summary, Stability: c.Stability, SideEffect: c.SideEffect, Connectors: connDTOs, RelatedUCI: c.RelatedUCI, OperationRefs: c.OperationRefs, DataObjectRefs: c.DataObjectRefs, ConnectorTypes: connectorTypeLabel(c.Connectors)})
		}
	}
	doNames := make([]string, 0, len(s.Metadata.DataObjects))
	var doDefs []ServiceDataObjectDTO
	for _, d := range s.Metadata.DataObjects {
		doNames = append(doNames, d.Name)
		doDefs = append(doDefs, ServiceDataObjectDTO{ID: d.ID, Name: d.Name, Summary: d.Summary, Schema: d.Schema, Format: d.Format, Stability: d.Stability})
	}
	uiNames := make([]string, 0, len(s.Metadata.UserInterfaces))
	var uiDefs []ServiceUserInterfaceDTO
	for _, u := range s.Metadata.UserInterfaces {
		uiNames = append(uiNames, u.Name)
		ui := ServiceUserInterfaceDTO{ID: u.ID, Name: u.Name, Summary: u.Summary, Channel: u.Channel, URL: u.URL, Stability: u.Stability, Engine: u.Engine, EngineVersion: u.EngineVersion, Schema: u.Schema}
		if u.Binding != nil {
			ui.Binding = &ViewBindingDTO{DataObjectRef: u.Binding.DataObjectRef, Submit: u.Binding.Submit}
		}
		uiDefs = append(uiDefs, ui)
	}
	return ServiceDTO{ID: s.Metadata.ID, Name: s.Name, Owner: fallback(s.Metadata.Owner, "unknown"), Capabilities: capNames, CapabilityDefs: capDefs, DataObjects: doNames, DataObjectDefs: doDefs, UserInterfaces: uiNames, UserInterfaceDefs: uiDefs, SupportedProducts: s.Metadata.SupportedProducts, Operations: operations, Status: fallback(s.Metadata.Status, "unknown"), Path: s.Path}
}

func AddServiceOperation(path, serviceName, operation string) (ServiceDTO, error) {
	operation = strings.TrimSpace(operation)
	if operation == "" {
		return ServiceDTO{}, Error(CodeInvalidInput, "operation name is required", http.StatusBadRequest, nil)
	}
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	for _, op := range raw.Operations {
		if op.Name == operation {
			return ServiceDTO{}, Error(CodeInvalidInput, "operation already exists: "+operation, http.StatusConflict, nil)
		}
	}
	raw.Operations = append(raw.Operations, model.Operation{Name: operation, Protocol: "rest", REST: &model.RESTOperation{}})
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

func RemoveServiceOperation(path, serviceName, operation string) (ServiceDTO, error) {
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	filtered := raw.Operations[:0]
	for _, op := range raw.Operations {
		if op.Name != operation {
			filtered = append(filtered, op)
		}
	}
	raw.Operations = filtered
	for i := range raw.Capabilities {
		raw.Capabilities[i].OperationRefs = removeString(raw.Capabilities[i].OperationRefs, operation)
	}
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		case r == '-' || r == '_' || r == ' ':
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	return out
}

// AddServiceCapability appends a new capability to the service's service.yaml.
// The ID is taken from the request or derived from the name. Returns the
// updated ServiceDTO.
func AddServiceCapability(path, serviceName string, cap model.ServiceCapability) (ServiceDTO, error) {
	cap.Name = strings.TrimSpace(cap.Name)
	cap.ID = strings.TrimSpace(cap.ID)
	if cap.Name == "" {
		return ServiceDTO{}, Error(CodeInvalidInput, "capability name is required", http.StatusBadRequest, nil)
	}
	if cap.ID == "" {
		cap.ID = "cap-" + slugify(cap.Name)
	}
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	for _, c := range raw.Capabilities {
		if c.ID == cap.ID || (c.Name != "" && c.Name == cap.Name) {
			return ServiceDTO{}, Error(CodeInvalidInput, "capability already exists: "+cap.ID, http.StatusConflict, nil)
		}
	}
	raw.Capabilities = append(raw.Capabilities, cap)
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

// AddServiceDataObject appends a new data object to service.yaml.
func AddServiceDataObject(path, serviceName string, obj model.ServiceDataObject) (ServiceDTO, error) {
	obj.Name = strings.TrimSpace(obj.Name)
	obj.ID = strings.TrimSpace(obj.ID)
	if obj.Name == "" {
		return ServiceDTO{}, Error(CodeInvalidInput, "data object name is required", http.StatusBadRequest, nil)
	}
	if obj.ID == "" {
		obj.ID = "do-" + slugify(obj.Name)
	}
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	for _, d := range raw.DataObjects {
		if d.ID == obj.ID || (d.Name != "" && d.Name == obj.Name) {
			return ServiceDTO{}, Error(CodeInvalidInput, "data object already exists: "+obj.ID, http.StatusConflict, nil)
		}
	}
	raw.DataObjects = append(raw.DataObjects, obj)
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

// AddServiceUserInterface appends a new user interface to service.yaml.
func AddServiceUserInterface(path, serviceName string, ui model.ServiceUserInterface) (ServiceDTO, error) {
	ui.Name = strings.TrimSpace(ui.Name)
	ui.ID = strings.TrimSpace(ui.ID)
	if ui.Name == "" {
		return ServiceDTO{}, Error(CodeInvalidInput, "user interface name is required", http.StatusBadRequest, nil)
	}
	if ui.ID == "" {
		ui.ID = "ui-" + slugify(ui.Name)
	}
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	for _, u := range raw.UserInterfaces {
		if u.ID == ui.ID || (u.Name != "" && u.Name == ui.Name) {
			return ServiceDTO{}, Error(CodeInvalidInput, "user interface already exists: "+ui.ID, http.StatusConflict, nil)
		}
	}
	raw.UserInterfaces = append(raw.UserInterfaces, ui)
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

// UpdateServiceCapability performs a partial update of an existing capability
// identified by capID. Pointer / slice fields in patch overwrite their
// counterparts when non-nil; empty strings leave the existing value untouched.
// MethodRefs and DataObjectRefs are always replaced (use empty slice to clear).
func UpdateServiceCapability(path, serviceName, capID string, patch model.ServiceCapability) (ServiceDTO, error) {
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	found := false
	for i, c := range raw.Capabilities {
		if c.ID == capID {
			if patch.Name != "" {
				raw.Capabilities[i].Name = patch.Name
			}
			if patch.Summary != "" {
				raw.Capabilities[i].Summary = patch.Summary
			}
			if patch.Stability != "" {
				raw.Capabilities[i].Stability = patch.Stability
			}
			if patch.SideEffect != "" {
				raw.Capabilities[i].SideEffect = patch.SideEffect
			}
			if patch.Connectors != nil {
				raw.Capabilities[i].Connectors = patch.Connectors
			}
			if patch.RelatedUCI != nil {
				raw.Capabilities[i].RelatedUCI = patch.RelatedUCI
			}
			raw.Capabilities[i].OperationRefs = patch.OperationRefs
			raw.Capabilities[i].DataObjectRefs = patch.DataObjectRefs
			found = true
			break
		}
	}
	if !found {
		return ServiceDTO{}, Error(CodeServiceNotFound, "capability not found: "+capID, http.StatusNotFound, nil)
	}
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

// RemoveServiceCapability deletes the capability identified by capID.
func RemoveServiceCapability(path, serviceName, capID string) (ServiceDTO, error) {
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	filtered := raw.Capabilities[:0]
	found := false
	for _, c := range raw.Capabilities {
		if c.ID == capID {
			found = true
			continue
		}
		filtered = append(filtered, c)
	}
	if !found {
		return ServiceDTO{}, Error(CodeServiceNotFound, "capability not found: "+capID, http.StatusNotFound, nil)
	}
	raw.Capabilities = filtered
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

// UpdateServiceDataObject performs a partial update of a data object identified by doID.
func UpdateServiceDataObject(path, serviceName, doID string, patch model.ServiceDataObject) (ServiceDTO, error) {
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	found := false
	for i, d := range raw.DataObjects {
		if d.ID == doID {
			if patch.Name != "" {
				raw.DataObjects[i].Name = patch.Name
			}
			if patch.Summary != "" {
				raw.DataObjects[i].Summary = patch.Summary
			}
			if patch.Schema != "" {
				raw.DataObjects[i].Schema = patch.Schema
			}
			if patch.Format != "" {
				raw.DataObjects[i].Format = patch.Format
			}
			if patch.Stability != "" {
				raw.DataObjects[i].Stability = patch.Stability
			}
			found = true
			break
		}
	}
	if !found {
		return ServiceDTO{}, Error(CodeServiceNotFound, "data object not found: "+doID, http.StatusNotFound, nil)
	}
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

// RemoveServiceDataObject deletes the data object identified by doID and removes
// any references to it from capability DataObjectRefs.
func RemoveServiceDataObject(path, serviceName, doID string) (ServiceDTO, error) {
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	filtered := raw.DataObjects[:0]
	found := false
	for _, d := range raw.DataObjects {
		if d.ID == doID {
			found = true
			continue
		}
		filtered = append(filtered, d)
	}
	if !found {
		return ServiceDTO{}, Error(CodeServiceNotFound, "data object not found: "+doID, http.StatusNotFound, nil)
	}
	raw.DataObjects = filtered
	for i := range raw.Capabilities {
		raw.Capabilities[i].DataObjectRefs = removeString(raw.Capabilities[i].DataObjectRefs, doID)
	}
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

// UpdateServiceUserInterface performs a partial update of a UI identified by uiID.
func UpdateServiceUserInterface(path, serviceName, uiID string, patch model.ServiceUserInterface) (ServiceDTO, error) {
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	found := false
	for i, u := range raw.UserInterfaces {
		if u.ID == uiID {
			if patch.Name != "" {
				raw.UserInterfaces[i].Name = patch.Name
			}
			if patch.Summary != "" {
				raw.UserInterfaces[i].Summary = patch.Summary
			}
			if patch.Channel != "" {
				raw.UserInterfaces[i].Channel = patch.Channel
			}
			if patch.URL != "" {
				raw.UserInterfaces[i].URL = patch.URL
			}
			if patch.Stability != "" {
				raw.UserInterfaces[i].Stability = patch.Stability
			}
			if patch.Engine != "" {
				raw.UserInterfaces[i].Engine = patch.Engine
			}
			if patch.EngineVersion != "" {
				raw.UserInterfaces[i].EngineVersion = patch.EngineVersion
			}
			if patch.Schema != nil {
				raw.UserInterfaces[i].Schema = patch.Schema
			}
			if patch.Binding != nil {
				raw.UserInterfaces[i].Binding = patch.Binding
			}
			found = true
			break
		}
	}
	if !found {
		return ServiceDTO{}, Error(CodeServiceNotFound, "user interface not found: "+uiID, http.StatusNotFound, nil)
	}
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

// RemoveServiceUserInterface deletes the UI identified by uiID.
func RemoveServiceUserInterface(path, serviceName, uiID string) (ServiceDTO, error) {
	svc, err := GetService(path, serviceName)
	if err != nil {
		return ServiceDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	filtered := raw.UserInterfaces[:0]
	found := false
	for _, u := range raw.UserInterfaces {
		if u.ID == uiID {
			found = true
			continue
		}
		filtered = append(filtered, u)
	}
	if !found {
		return ServiceDTO{}, Error(CodeServiceNotFound, "user interface not found: "+uiID, http.StatusNotFound, nil)
	}
	raw.UserInterfaces = filtered
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, serviceName)
}

func removeString(s []string, v string) []string {
	out := s[:0]
	for _, x := range s {
		if x != v {
			out = append(out, x)
		}
	}
	return out
}

// UpdateMethodParameters replaces the parameter list of a named method on a service.
// Deprecated: prefer UpdateMethod which updates all endpoint fields.
func UpdateOperationParameters(path, serviceName, operationName string, params []model.MethodParameter) (OperationDTO, error) {
	return UpdateOperation(path, serviceName, operationName, model.Operation{Protocol: "rest", REST: &model.RESTOperation{Parameters: params}})
}

// UpdateOperation performs a partial update of a named operation's REST binding:
// only non-zero fields in patch are written.
func UpdateOperation(path, serviceName, operationName string, patch model.Operation) (OperationDTO, error) {
	svc, err := GetService(path, serviceName)
	if err != nil {
		return OperationDTO{}, err
	}
	yamlPath := filepath.Join(svc.Path, "service.yaml")
	var raw model.Service
	if err := fsx.ReadYAML(yamlPath, &raw); err != nil {
		return OperationDTO{}, Error(CodeInternalError, "Failed to read service: "+err.Error(), http.StatusInternalServerError, err)
	}
	pr := patch.REST
	if pr == nil {
		pr = &model.RESTOperation{}
	}
	found := false
	for i := range raw.Operations {
		if raw.Operations[i].Name != operationName {
			continue
		}
		if patch.Summary != "" {
			raw.Operations[i].Summary = patch.Summary
		}
		if patch.InputObject != "" {
			raw.Operations[i].InputObject = patch.InputObject
		}
		if patch.OutputObject != "" {
			raw.Operations[i].OutputObject = patch.OutputObject
		}
		if raw.Operations[i].REST == nil {
			raw.Operations[i].REST = &model.RESTOperation{}
		}
		r := raw.Operations[i].REST
		if pr.HTTPMethod != "" {
			r.HTTPMethod = pr.HTTPMethod
		}
		if pr.Path != "" {
			r.Path = pr.Path
		}
		if pr.BaseURL != "" {
			r.BaseURL = pr.BaseURL
		}
		if pr.Parameters != nil {
			r.Parameters = pr.Parameters
		}
		if pr.Headers != nil {
			r.Headers = pr.Headers
		}
		if pr.Security != nil {
			r.Security = pr.Security
		}
		if pr.Payload != nil {
			r.Payload = pr.Payload
		}
		found = true
		break
	}
	if !found {
		return OperationDTO{}, Error(CodeServiceNotFound, "operation not found: "+operationName, http.StatusNotFound, nil)
	}
	if err := fsx.WriteYAML(yamlPath, raw); err != nil {
		return OperationDTO{}, Error(CodeInternalError, "Failed to write service: "+err.Error(), http.StatusInternalServerError, err)
	}
	updated, err := GetService(path, serviceName)
	if err != nil {
		return OperationDTO{}, err
	}
	for _, op := range updated.Operations {
		if op.Name == operationName {
			return op, nil
		}
	}
	return OperationDTO{Name: operationName}, nil
}

// operationDTO converts a model.Operation to OperationDTO, flattening the REST
// binding for editor convenience.
func operationDTO(op model.Operation) OperationDTO {
	dto := OperationDTO{
		Name:         op.Name,
		Summary:      op.Summary,
		Protocol:     op.Protocol,
		InputObject:  op.InputObject,
		OutputObject: op.OutputObject,
	}
	if r := op.REST; r != nil {
		params := make([]MethodParameterDTO, 0, len(r.Parameters))
		for _, p := range r.Parameters {
			params = append(params, MethodParameterDTO{Name: p.Name, Type: p.Type, In: p.In, Required: p.Required, Description: p.Description})
		}
		headers := make([]MethodHeaderDTO, 0, len(r.Headers))
		for _, h := range r.Headers {
			headers = append(headers, MethodHeaderDTO{Name: h.Name, Value: h.Value, Required: h.Required, Description: h.Description})
		}
		dto.HTTPMethod = r.HTTPMethod
		dto.Path = r.Path
		dto.BaseURL = r.BaseURL
		dto.Parameters = params
		dto.Headers = headers
		if r.Security != nil {
			dto.Security = &MethodSecurityDTO{Scheme: r.Security.Scheme, In: r.Security.In, Name: r.Security.Name}
		}
		if r.Payload != nil {
			fields := make([]MethodPayloadFieldDTO, 0, len(r.Payload.Fields))
			for _, f := range r.Payload.Fields {
				fields = append(fields, MethodPayloadFieldDTO{Name: f.Name, Type: f.Type, Required: f.Required, Description: f.Description, Example: f.Example})
			}
			dto.Payload = &MethodPayloadDTO{ContentType: r.Payload.ContentType, Fields: fields}
		}
	}
	if op.MCP != nil {
		dto.MCP = &MCPOperationDTO{Transport: op.MCP.Transport, ServerURL: op.MCP.ServerURL, Tool: op.MCP.Tool}
	}
	if op.GRPC != nil {
		dto.GRPC = &GRPCOperationDTO{Target: op.GRPC.Target, Service: op.GRPC.Service, Method: op.GRPC.Method, ProtoRef: op.GRPC.ProtoRef}
	}
	return dto
}

// GetServiceOperation returns the definition of a single named operation.
func GetServiceOperation(path, serviceName, operationName string) (OperationDTO, error) {
	svc, err := GetService(path, serviceName)
	if err != nil {
		return OperationDTO{}, err
	}
	for _, op := range svc.Operations {
		if op.Name == operationName {
			return op, nil
		}
	}
	return OperationDTO{}, Error(CodeServiceNotFound, "operation not found: "+operationName, http.StatusNotFound, nil)
}

func allProductSummaries(tree cosmosfs.Tree) []ProductSummaryDTO {
	out := []ProductSummaryDTO{}
	for _, b := range tree.Blueprints {
		if b.Metadata.Type != "product_blueprint" && b.Metadata.Type != "product" {
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
					StepNum:     i + 1,
					ID:          step.ID,
					Name:        step.Name,
					TaskType:    normalizedStepTaskType(step.TaskType),
					ServiceRef:  step.ServiceRef,
					DecisionRef: step.DecisionRef,
					Method:      step.Method,
					Role:        step.Role,
					Required:    step.Required,
					DependsOn:   step.DependsOn,
					Inputs:      stepsInputsToDTO(step.Inputs),
					Outputs:     stepsOutputsToDTO(step.Outputs),
					Gateway:     gatewayToDTO(step.Gateway),
				})
			}
			processGroups = append(processGroups, group)
		}
	}
	return ProductSummaryDTO{ID: bp.ID, Name: bp.Name, Version: bp.Version, Status: bp.Status, SourcePath: sourcePath, CatalogPath: sourcePath, FulfillmentRequiredServicesCount: len(fulfillment.RequiredServices), FulfillmentUnresolvedCount: unresolved, ProcessCount: processCount, UnmappedTaskCount: unmappedTaskCount, Fulfillment: fulfillment, Processes: processGroups}
}

func AllServiceRefs(path string) (ServiceRefsDTO, error) {
	tree, err := load(path)
	if err != nil {
		return ServiceRefsDTO{}, err
	}
	out := ServiceRefsDTO{Services: []ServiceRefDTO{}}
	for _, s := range tree.Services {
		out.Services = append(out.Services, ServiceRefDTO{Service: s.Name, ServiceRef: s.Name})
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

func connectorTypeLabel(connectors []model.Connector) string {
	seen := map[string]bool{}
	order := []string{"cli", "rest", "mcp", "collaboration"}
	labels := map[string]string{"cli": "CLI", "rest": "REST", "mcp": "MCP", "collaboration": "BPMN"}
	for _, c := range connectors {
		seen[c.Type] = true
	}
	var parts []string
	for _, t := range order {
		if seen[t] {
			parts = append(parts, labels[t])
		}
	}
	return strings.Join(parts, " · ")
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
	// ADR-0020: Legacy-IDs werden transparent via id-history aufgelöst.
	resolved, _ := idmigrate.Resolve(path, id)
	for _, b := range items.Blueprints {
		if b.ID == id || b.ID == resolved {
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
	resolved, _ := idmigrate.Resolve(path, id)
	for _, i := range items.Instances {
		if i.ID == id || i.ID == resolved {
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
	return BlueprintDTO{ID: b.Metadata.ID, Type: b.Metadata.Type, Name: b.Metadata.Name, Version: b.Metadata.Version, Status: b.Metadata.Status, Owner: b.Metadata.Owner, Fulfillment: fulfillment, Summary: b.Metadata.Summary, Purpose: b.Metadata.Purpose, Description: b.Metadata.Description, Consumers: b.Metadata.Consumers, LifecycleStatus: b.Metadata.LifecycleStatus, Tags: b.Metadata.Tags, Processes: b.Metadata.Processes, ProcessSummary: processSummary, Path: b.Path, Variants: variants, Capabilities: b.Metadata.Capabilities, TargetSystems: b.Metadata.TargetSystems, RequiredInputs: b.Metadata.RequiredInputs, RequiredServiceBlueprints: b.Metadata.RequiredServiceBlueprints, RequiredServices: requiredServices, NamespaceServiceRef: b.Metadata.NamespaceServiceRef, Rules: b.Metadata.Rules, QualityCriteria: b.Metadata.QualityCriteria, EvidenceRequirements: b.Metadata.EvidenceRequirements, Requirements: requirements, RequirementsStatus: requirementsStatus(b.Metadata.Requirements), Attributes: attributes}
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
	fulfillmentType := "unresolved"
	if res.Status == "resolved" {
		fulfillmentType = "local"
	}
	slaDTO := serviceLevelDTO(firstServiceLevel(sla, res.SLA))
	olaDTO := serviceLevelDTO(firstServiceLevel(ola, res.OLA))
	treeTarget := ""
	if res.Status == "resolved" {
		treeTarget = "service:" + res.Service
	}
	return ProductRequiredServiceDTO{ServiceRef: serviceRef, Role: role, Required: required, Description: description, ResolutionStatus: res.Status, ResolvedService: res.Service, FulfillmentType: fulfillmentType, SLARef: slaRef, OLARef: olaRef, TreeTarget: treeTarget, SLA: slaDTO, OLA: olaDTO, ServiceLevelLabel: serviceLevelLabel(slaRef, olaRef, slaDTO, olaDTO)}
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

// resolveService resolves a flat service reference (service name or stable
// service ID) against the cosmos services.
func resolveService(tree cosmosfs.Tree, serviceRef string) ServiceResolution {
	ref := strings.TrimSpace(serviceRef)
	if ref == "" {
		return ServiceResolution{Status: "missing"}
	}
	for _, s := range tree.Services {
		if s.Name == ref || s.Metadata.ID == ref {
			return ServiceResolution{Status: "resolved", Service: s.Name, SLA: s.Metadata.SLA, OLA: s.Metadata.OLA}
		}
	}
	return ServiceResolution{Status: "unresolved_service", Service: ref}
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

func validServiceName(name string) bool {
	return name != "" && !strings.ContainsAny(name, `/\`) && name != "." && name != ".."
}

func DeleteService(path, serviceName string) error {
	svc, err := GetService(path, serviceName)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(svc.Path); err != nil {
		return Error(CodeInternalError, "Failed to delete service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

func RenameService(path, oldName, newName string) error {
	svc, err := GetService(path, oldName)
	if err != nil {
		return err
	}
	if !validServiceName(newName) {
		return Error(CodeInvalidNamespace, "Invalid service name: "+newName, http.StatusBadRequest, nil)
	}
	// Rename in place, preserving whatever folder the service currently lives in
	// (folders are a free organization layer; the ID is the stable reference).
	newDir := filepath.Join(filepath.Dir(svc.Path), newName)
	if _, err := os.Stat(newDir); err == nil {
		return Error(CodeInvalidNamespace, "Service already exists: "+newName, http.StatusConflict, nil)
	}
	if err := os.Rename(svc.Path, newDir); err != nil {
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

func AddService(path, name, owner string, force bool) (ServiceDTO, error) {
	return AddServiceIn(path, "", name, owner, force)
}

// AddServiceIn creates a service inside folderRel, a workspace-relative folder
// within the services root (empty → the root itself). Folders are a free
// organization layer; the service stays discoverable via the recursive scanner.
func AddServiceIn(path, folderRel, name, owner string, force bool) (ServiceDTO, error) {
	name = strings.TrimSpace(name)
	if !validServiceName(name) {
		return ServiceDTO{}, Error(CodeInvalidNamespace, "Invalid service name: "+name, http.StatusBadRequest, nil)
	}
	if strings.TrimSpace(owner) == "" {
		owner = "unknown"
	}
	if err := folderAllows(path, folderRel, "service"); err != nil {
		return ServiceDTO{}, err
	}
	base, err := resolveArtifactDir(path, storage.ServicesDir(path), folderRel)
	if err != nil {
		return ServiceDTO{}, err
	}
	sdir := filepath.Join(base, name)
	if _, err := os.Stat(sdir); err == nil && !force {
		return ServiceDTO{}, Error(CodeInvalidNamespace, "Service already exists: "+name, http.StatusConflict, nil)
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
	return GetService(path, name)
}

// MoveServiceElement relocates an embedded element (capability, data object,
// user interface or method) from one service to another (git-first).
func MoveServiceElement(path, fromService, kind, elementID, toService string) (ServiceDTO, error) {
	if fromService == toService {
		return ServiceDTO{}, Error(CodeInvalidInput, "source and target service are the same", http.StatusBadRequest, nil)
	}
	srcSvc, err := GetService(path, fromService)
	if err != nil {
		return ServiceDTO{}, err
	}
	dstSvc, err := GetService(path, toService)
	if err != nil {
		return ServiceDTO{}, err
	}
	srcPath := filepath.Join(srcSvc.Path, "service.yaml")
	dstPath := filepath.Join(dstSvc.Path, "service.yaml")
	var src model.Service
	if err := fsx.ReadYAML(srcPath, &src); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "read source service: "+err.Error(), http.StatusInternalServerError, err)
	}
	var dst model.Service
	if err := fsx.ReadYAML(dstPath, &dst); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "read target service: "+err.Error(), http.StatusInternalServerError, err)
	}
	moved, err := moveEmbeddedElement(&src, &dst, kind, elementID)
	if err != nil {
		return ServiceDTO{}, err
	}
	if !moved {
		return ServiceDTO{}, Error(CodeInvalidInput, kind+" not found: "+elementID, http.StatusNotFound, nil)
	}
	if err := fsx.WriteYAML(srcPath, src); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "write source service: "+err.Error(), http.StatusInternalServerError, err)
	}
	if err := fsx.WriteYAML(dstPath, dst); err != nil {
		return ServiceDTO{}, Error(CodeInternalError, "write target service: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetService(path, toService)
}

func moveEmbeddedElement(src, dst *model.Service, kind, id string) (bool, error) {
	switch kind {
	case "capability":
		for i, c := range src.Capabilities {
			if c.ID == id {
				src.Capabilities = append(src.Capabilities[:i], src.Capabilities[i+1:]...)
				dst.Capabilities = append(dst.Capabilities, c)
				return true, nil
			}
		}
	case "data-object":
		for i, o := range src.DataObjects {
			if o.ID == id {
				src.DataObjects = append(src.DataObjects[:i], src.DataObjects[i+1:]...)
				dst.DataObjects = append(dst.DataObjects, o)
				return true, nil
			}
		}
	case "user-interface":
		for i, u := range src.UserInterfaces {
			if u.ID == id {
				src.UserInterfaces = append(src.UserInterfaces[:i], src.UserInterfaces[i+1:]...)
				dst.UserInterfaces = append(dst.UserInterfaces, u)
				return true, nil
			}
		}
	case "operation":
		for i, op := range src.Operations {
			if op.Name == id {
				src.Operations = append(src.Operations[:i], src.Operations[i+1:]...)
				dst.Operations = append(dst.Operations, op)
				return true, nil
			}
		}
	default:
		return false, Error(CodeInvalidInput, "unsupported element kind: "+kind, http.StatusBadRequest, nil)
	}
	return false, nil
}

func CreateBlueprint(path string, bp model.Blueprint) error {
	return CreateBlueprintIn(path, "", bp)
}

// CreateBlueprintIn writes a blueprint into folderRel, a workspace-relative
// folder within the catalog root (empty → the conventional products/services
// subdirectory chosen by blueprint type).
func CreateBlueprintIn(path, folderRel string, bp model.Blueprint) error {
	if _, err := os.Stat(storage.CosmosFile(path)); err != nil {
		return Error(CodeCosmosMissing, ".nomos/cosmos.yaml not found", http.StatusNotFound, err)
	}
	if bp.Type != "product_blueprint" && bp.Type != "service_blueprint" {
		return Error(CodeInvalidInput, "Blueprint type must be product_blueprint or service_blueprint", http.StatusBadRequest, nil)
	}
	if strings.TrimSpace(bp.ID) == "" {
		generated, err := idgen.NewForType(bp.Type)
		if err != nil {
			return Error(CodeInternalError, "Failed to generate blueprint ID: "+err.Error(), http.StatusInternalServerError, err)
		}
		bp.ID = generated
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

	artifactType := "blueprint"
	if bp.Type == "product_blueprint" {
		artifactType = "product"
	}
	if err := folderAllows(path, folderRel, artifactType); err != nil {
		return err
	}
	var dir string
	if strings.TrimSpace(folderRel) != "" {
		resolved, err := resolveArtifactDir(path, filepath.Join(storage.CatalogDir(path), "blueprints"), folderRel)
		if err != nil {
			return err
		}
		dir = resolved
	} else {
		subdir := "services"
		if bp.Type == "product_blueprint" {
			subdir = "products"
		}
		dir = filepath.Join(storage.CatalogDir(path), "blueprints", subdir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Error(CodeInternalError, "Failed to create blueprint directory: "+err.Error(), http.StatusInternalServerError, err)
	}

	filePath := filepath.Join(dir, bp.ID+".yaml")
	if err := fsx.WriteYAML(filePath, bp); err != nil {
		return Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

func CreateProductOffering(path string, req CreateProductOfferingRequest) (BlueprintDTO, error) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		generated, err := idgen.NewForType("product_blueprint")
		if err != nil {
			return BlueprintDTO{}, Error(CodeInternalError, "Failed to generate product ID: "+err.Error(), http.StatusInternalServerError, err)
		}
		id = generated
	}
	version := firstNonEmpty(strings.TrimSpace(req.Version), "0.1.0")
	status := firstNonEmpty(strings.TrimSpace(req.Status), "draft")
	owner := firstNonEmpty(strings.TrimSpace(req.Owner), "unknown")
	bp := model.Blueprint{ID: id, Type: "product_blueprint", Name: strings.TrimSpace(req.Name), Version: version, Status: status, Owner: owner, Summary: strings.TrimSpace(req.Summary), Description: strings.TrimSpace(req.Description), Tags: req.Tags, Fulfillment: model.ProductFulfillment{RequiredServices: []model.ProductRequiredService{}}}
	if bp.Name == "" {
		bp.Name = id
	}
	if err := CreateBlueprint(path, bp); err != nil {
		return BlueprintDTO{}, err
	}
	return GetBlueprint(path, id)
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
	normalized := strings.TrimSpace(serviceRef)
	for _, existing := range raw.Fulfillment.RequiredServices {
		existingNormalized := strings.TrimSpace(existing.ServiceRef)
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
	normalized := strings.TrimSpace(serviceRef)
	for i, existing := range raw.Fulfillment.RequiredServices {
		if i == index {
			continue
		}
		existingNormalized := strings.TrimSpace(existing.ServiceRef)
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
	if inst.Type != "product_instance" && inst.Type != "service_instance" {
		return Error(CodeInvalidInput, "Instance type must be product_instance or service_instance", http.StatusBadRequest, nil)
	}
	if strings.TrimSpace(inst.BlueprintRef) == "" {
		return Error(CodeInvalidInput, "blueprint_ref is required", http.StatusBadRequest, nil)
	}
	if strings.TrimSpace(inst.ID) == "" {
		generated, err := idgen.NewForType(inst.Type)
		if err != nil {
			return Error(CodeInternalError, "Failed to generate instance ID: "+err.Error(), http.StatusInternalServerError, err)
		}
		inst.ID = generated
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
