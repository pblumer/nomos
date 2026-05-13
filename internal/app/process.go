package app

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

var bpmnTaskTypes = map[string]bool{
	"task": true, "userTask": true, "serviceTask": true, "businessRuleTask": true,
	"manualTask": true, "scriptTask": true, "callActivity": true,
}

type processNode struct {
	Path string
	Meta model.Process
}

func processesDir(path string) string {
	return filepath.Join(storage.CatalogDir(path), "blueprints", "processes")
}
func processesDirForRead(path string) string {
	return filepath.Join(storage.CatalogDirForRead(path), "blueprints", "processes")
}

func ListProcesses(path string) (ProcessesDTO, error) {
	nodes, err := scanProcesses(path)
	if err != nil {
		return ProcessesDTO{}, err
	}
	items := make([]ProcessDTO, 0, len(nodes))
	for _, n := range nodes {
		items = append(items, processDTO(path, n, true))
	}
	return ProcessesDTO{Items: items, Count: len(items)}, nil
}

func ListProductProcesses(path, productID string) (ProcessesDTO, error) {
	tree, err := load(path)
	if err != nil {
		return ProcessesDTO{}, err
	}
	want := map[string]bool{}
	found := false
	for _, bp := range tree.Blueprints {
		if bp.Metadata.ID == productID && bp.Metadata.Type == "product_blueprint" {
			found = true
			for _, id := range bp.Metadata.Processes {
				want[id] = true
			}
		}
	}
	if !found {
		return ProcessesDTO{}, Error(CodeBlueprintNotFound, "Blueprint not found: "+productID, http.StatusNotFound, nil)
	}
	all, err := scanProcesses(path)
	if err != nil {
		return ProcessesDTO{}, err
	}
	items := []ProcessDTO{}
	for _, n := range all {
		if n.Meta.RelatedProduct == productID || want[n.Meta.ID] {
			items = append(items, processDTO(path, n, false))
		}
	}
	return ProcessesDTO{Items: items, Count: len(items)}, nil
}

func GetProcess(path, id string) (ProcessDTO, error) {
	nodes, err := scanProcesses(path)
	if err != nil {
		return ProcessDTO{}, err
	}
	for _, n := range nodes {
		if n.Meta.ID == id {
			return processDTO(path, n, true), nil
		}
	}
	return ProcessDTO{}, Error(CodeInvalidInput, "Process not found: "+id, http.StatusNotFound, nil)
}

func CreateProductProcess(path, productID string, req CreateProcessRequest) (ProcessDTO, error) {
	product, err := GetBlueprint(path, productID)
	if err != nil {
		return ProcessDTO{}, err
	}
	id := strings.TrimSpace(req.ID)
	if id == "" {
		id = nextProcessID(path)
	}
	name := firstNonEmpty(strings.TrimSpace(req.Name), "Fulfillment process for "+product.Name)
	if _, err := GetProcess(path, id); err == nil {
		return ProcessDTO{}, Error(CodeInvalidInput, "Process already exists: "+id, http.StatusConflict, nil)
	}
	dir := processesDir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to create process directory: "+err.Error(), http.StatusInternalServerError, err)
	}
	fileBase := safeArtifactName(id)
	bpmnFile := fileBase + ".bpmn"
	meta := model.Process{ID: id, Type: "process", Name: name, Version: "0.1.0", Status: "draft", Owner: firstNonEmpty(product.OwningDomain, product.OfferedBy, product.Owner), Summary: strings.TrimSpace(req.Summary), RelatedProduct: productID, BPMN: model.BPMNReference{File: bpmnFile, ProcessID: "Process_" + sanitizeBPMNID(id), Primary: true}}
	if meta.Summary == "" {
		meta.Summary = "Product-level fulfillment process."
	}
	if err := fsx.WriteYAML(filepath.Join(dir, fileBase+".yaml"), meta); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write process: "+err.Error(), http.StatusInternalServerError, err)
	}
	if err := os.WriteFile(filepath.Join(dir, bpmnFile), []byte(DefaultBPMNTemplate(meta.BPMN.ProcessID, product.Name)), 0o644); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write BPMN: "+err.Error(), http.StatusInternalServerError, err)
	}
	_ = attachProcessToProduct(path, productID, id)
	return GetProcess(path, id)
}

func GetProcessBPMN(path, id string) (string, error) {
	dto, err := GetProcess(path, id)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(dto.BPMNPath)
	if err != nil {
		return "", Error(CodeInvalidInput, "BPMN file not found: "+dto.BPMN.File, http.StatusNotFound, err)
	}
	return string(data), nil
}

func UpdateProcessBPMN(path, id, xmlText string) (ProcessDTO, error) {
	dto, err := GetProcess(path, id)
	if err != nil {
		return ProcessDTO{}, err
	}
	if _, err := ExtractBPMNTasks(xmlText); err != nil {
		return ProcessDTO{}, Error(CodeInvalidInput, "Invalid BPMN XML: "+err.Error(), http.StatusBadRequest, err)
	}
	if err := os.WriteFile(dto.BPMNPath, []byte(xmlText), 0o644); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write BPMN: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetProcess(path, id)
}

func ProcessTasks(path, id string) ([]BPMNTaskDTO, error) {
	dto, err := GetProcess(path, id)
	if err != nil {
		return nil, err
	}
	return dto.Tasks, nil
}

func UpdateProcessTaskMappings(path, id string, req UpdateTaskMappingsRequest) (ProcessDTO, error) {
	node, err := findProcessNode(path, id)
	if err != nil {
		return ProcessDTO{}, err
	}
	tasks, _ := ProcessTasks(path, id)
	taskByID := map[string]BPMNTaskDTO{}
	for _, t := range tasks {
		taskByID[t.ID] = t
	}
	mappings := make([]model.ProcessTaskMapping, 0, len(req.TaskMappings))
	for _, m := range req.TaskMappings {
		if strings.TrimSpace(m.BPMNElementID) == "" || strings.TrimSpace(m.ServiceRef) == "" {
			continue
		}
		t := taskByID[m.BPMNElementID]
		mappings = append(mappings, model.ProcessTaskMapping{BPMNElementID: m.BPMNElementID, TaskName: firstNonEmpty(m.TaskName, t.Name), BPMNElementType: firstNonEmpty(m.BPMNElementType, t.ElementType), ServiceRef: strings.TrimSpace(m.ServiceRef), Role: firstNonEmpty(m.Role, "primary"), Required: m.Required, Notes: m.Notes})
	}
	node.Meta.TaskMappings = mappings
	if err := fsx.WriteYAML(node.Path, node.Meta); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write process: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetProcess(path, id)
}

func ValidateProcess(path, id string) (ProcessValidationDTO, error) {
	dto, err := GetProcess(path, id)
	if err != nil {
		return ProcessValidationDTO{}, err
	}
	return dto.Validation, nil
}

func scanProcesses(path string) ([]processNode, error) {
	root := processesDirForRead(path)
	nodes := []processNode{}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nodes, nil
		}
		return nil, err
	}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".yaml") {
			return nil
		}
		var pr model.Process
		if err := fsx.ReadYAML(p, &pr); err != nil {
			return err
		}
		if pr.Type == "process" {
			nodes = append(nodes, processNode{Path: p, Meta: pr})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Meta.ID < nodes[j].Meta.ID })
	return nodes, nil
}

func findProcessNode(path, id string) (processNode, error) {
	nodes, err := scanProcesses(path)
	if err != nil {
		return processNode{}, err
	}
	for _, n := range nodes {
		if n.Meta.ID == id {
			return n, nil
		}
	}
	return processNode{}, Error(CodeInvalidInput, "Process not found: "+id, http.StatusNotFound, nil)
}

func processDTO(path string, n processNode, includeValidation bool) ProcessDTO {
	bpmnPath := safeBPMNPath(filepath.Dir(n.Path), n.Meta.BPMN.File)
	dto := ProcessDTO{ID: n.Meta.ID, Type: n.Meta.Type, Name: n.Meta.Name, Version: n.Meta.Version, Status: n.Meta.Status, Owner: n.Meta.Owner, Summary: n.Meta.Summary, Tags: n.Meta.Tags, RelatedProduct: n.Meta.RelatedProduct, BPMN: BPMNReferenceDTO{File: n.Meta.BPMN.File, ProcessID: n.Meta.BPMN.ProcessID, Primary: n.Meta.BPMN.Primary}, Path: n.Path, BPMNPath: bpmnPath}
	for _, s := range n.Meta.Steps {
		dto.Steps = append(dto.Steps, ProcessStepDTO{ID: s.ID, Name: s.Name, ServiceRef: s.ServiceRef, Method: s.Method, Role: s.Role, Required: s.Required, Notes: s.Notes})
	}
	for _, m := range n.Meta.TaskMappings {
		dto.TaskMappings = append(dto.TaskMappings, ProcessTaskMappingDTO{BPMNElementID: m.BPMNElementID, TaskName: m.TaskName, BPMNElementType: m.BPMNElementType, ServiceRef: m.ServiceRef, Role: m.Role, Required: m.Required, Notes: m.Notes})
	}
	if data, err := os.ReadFile(bpmnPath); err == nil {
		if tasks, err := ExtractBPMNTasks(string(data)); err == nil {
			dto.Tasks = withMappingStatus(tasks, dto.TaskMappings)
		}
	}
	if includeValidation {
		dto.Validation = validateProcessDTO(path, dto)
	}
	return dto
}

func AddProcessStep(path, id string, req UpsertProcessStepRequest) (ProcessDTO, error) {
	if strings.TrimSpace(req.Name) == "" {
		return ProcessDTO{}, Error(CodeInvalidInput, "step name is required", http.StatusBadRequest, nil)
	}
	node, err := findProcessNode(path, id)
	if err != nil {
		return ProcessDTO{}, err
	}
	step := model.ProcessStep{
		ID:         fmt.Sprintf("step-%d", time.Now().UnixNano()),
		Name:       strings.TrimSpace(req.Name),
		ServiceRef: strings.TrimSpace(req.ServiceRef),
		Method:     strings.TrimSpace(req.Method),
		Role:       firstNonEmpty(strings.TrimSpace(req.Role), "supporting"),
		Required:   req.Required,
		Notes:      req.Notes,
		DependsOn:  req.DependsOn,
	}
	node.Meta.Steps = append(node.Meta.Steps, step)
	if err := fsx.WriteYAML(node.Path, node.Meta); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write process: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetProcess(path, id)
}

func UpdateProcessStep(path, id, stepID string, req UpsertProcessStepRequest) (ProcessDTO, error) {
	if strings.TrimSpace(req.Name) == "" {
		return ProcessDTO{}, Error(CodeInvalidInput, "step name is required", http.StatusBadRequest, nil)
	}
	node, err := findProcessNode(path, id)
	if err != nil {
		return ProcessDTO{}, err
	}
	found := false
	for i := range node.Meta.Steps {
		if node.Meta.Steps[i].ID == stepID {
			node.Meta.Steps[i].Name = strings.TrimSpace(req.Name)
			node.Meta.Steps[i].ServiceRef = strings.TrimSpace(req.ServiceRef)
			node.Meta.Steps[i].Method = strings.TrimSpace(req.Method)
			node.Meta.Steps[i].Role = firstNonEmpty(strings.TrimSpace(req.Role), "supporting")
			node.Meta.Steps[i].Required = req.Required
			node.Meta.Steps[i].Notes = req.Notes
			node.Meta.Steps[i].DependsOn = req.DependsOn
			found = true
			break
		}
	}
	if !found {
		return ProcessDTO{}, Error(CodeInvalidInput, "Step not found: "+stepID, http.StatusNotFound, nil)
	}
	if err := fsx.WriteYAML(node.Path, node.Meta); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write process: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetProcess(path, id)
}

func RemoveProcessStep(path, id, stepID string) (ProcessDTO, error) {
	node, err := findProcessNode(path, id)
	if err != nil {
		return ProcessDTO{}, err
	}
	n := len(node.Meta.Steps)
	filtered := node.Meta.Steps[:0]
	for _, s := range node.Meta.Steps {
		if s.ID != stepID {
			filtered = append(filtered, s)
		}
	}
	if len(filtered) == n {
		return ProcessDTO{}, Error(CodeInvalidInput, "Step not found: "+stepID, http.StatusNotFound, nil)
	}
	node.Meta.Steps = filtered
	if err := fsx.WriteYAML(node.Path, node.Meta); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write process: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetProcess(path, id)
}

func ExtractBPMNTasks(xmlText string) ([]BPMNTaskDTO, error) {
	dec := xml.NewDecoder(bytes.NewBufferString(xmlText))
	tasks := []BPMNTaskDTO{}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if !bpmnTaskTypes[se.Name.Local] {
			continue
		}
		var id, name string
		for _, a := range se.Attr {
			if a.Name.Local == "id" {
				id = a.Value
			}
			if a.Name.Local == "name" {
				name = a.Value
			}
		}
		if id != "" {
			tasks = append(tasks, BPMNTaskDTO{ID: id, Name: name, ElementType: "bpmn:" + se.Name.Local})
		}
	}
	return tasks, nil
}

func withMappingStatus(tasks []BPMNTaskDTO, mappings []ProcessTaskMappingDTO) []BPMNTaskDTO {
	by := map[string]ProcessTaskMappingDTO{}
	for _, m := range mappings {
		by[m.BPMNElementID] = m
	}
	for i := range tasks {
		if m, ok := by[tasks[i].ID]; ok {
			tasks[i].MappingStatus = "mapped"
			tasks[i].ServiceRef = m.ServiceRef
		} else {
			tasks[i].MappingStatus = "unmapped"
		}
	}
	return tasks
}

func validateProcessDTO(path string, p ProcessDTO) ProcessValidationDTO {
	findings := []FindingDTO{}
	add := func(code, severity, msg string) {
		findings = append(findings, FindingDTO{Code: code, Severity: severity, Message: msg, ArtifactType: "process", ArtifactID: p.ID})
	}
	if strings.TrimSpace(p.ID) == "" {
		add("PROCESS_ID_MISSING", "error", "Process ID is required")
	}
	if strings.TrimSpace(p.Name) == "" {
		add("PROCESS_NAME_MISSING", "error", "Process name is required")
	}
	if strings.TrimSpace(p.RelatedProduct) == "" {
		add("PROCESS_PRODUCT_MISSING", "error", "related_product is required")
	} else if _, err := GetBlueprint(path, p.RelatedProduct); err != nil {
		add("PROCESS_PRODUCT_UNRESOLVED", "error", "Related product does not exist")
	}
	hasSteps := len(p.Steps) > 0
	if strings.TrimSpace(p.BPMN.File) == "" {
		if !hasSteps {
			add("PROCESS_NO_DEFINITION", "warning", "Process has neither steps nor a BPMN file defined")
		}
	} else if _, err := os.Stat(p.BPMNPath); err != nil {
		if !hasSteps {
			add("BPMN_FILE_NOT_FOUND", "error", "BPMN XML file was not found")
		}
	} else if data, err := os.ReadFile(p.BPMNPath); err == nil {
		if tasks, err := ExtractBPMNTasks(string(data)); err != nil {
			add("BPMN_INVALID", "error", "BPMN XML cannot be parsed")
		} else if len(tasks) == 0 && !hasSteps {
			add("BPMN_NO_TASKS", "warning", "BPMN process has no mappable tasks")
		}
	}
	taskIDs := map[string]bool{}
	for _, t := range p.Tasks {
		taskIDs[t.ID] = true
	}
	mapped := map[string]bool{}
	productServices := map[string]bool{}
	if product, err := GetBlueprint(path, p.RelatedProduct); err == nil {
		for _, s := range product.Fulfillment.RequiredServices {
			productServices[s.ServiceRef] = true
			if s.Required && s.SLA == nil && s.SLARef == "" {
				add("SERVICE_SLA_MISSING", "warning", "Required service lacks SLA metadata: "+s.ServiceRef)
			}
		}
	}
	for _, m := range p.TaskMappings {
		mapped[m.BPMNElementID] = true
		if !taskIDs[m.BPMNElementID] {
			add("TASK_MAPPING_UNKNOWN_TASK", "error", "Task mapping references unknown BPMN task: "+m.BPMNElementID)
		}
		if !productServices[m.ServiceRef] {
			add("TASK_SERVICE_NOT_IN_COMPOSITION", "warning", "Mapped service is not in product fulfillment composition: "+m.ServiceRef)
		}
	}
	for _, t := range p.Tasks {
		if !mapped[t.ID] {
			add("BPMN_TASK_UNMAPPED", "warning", "BPMN task is not mapped to a service: "+firstNonEmpty(t.Name, t.ID))
		}
	}
	status := "ready"
	for _, f := range findings {
		if f.Severity == "error" {
			status = "invalid"
			break
		}
		status = "needs_attention"
	}
	return ProcessValidationDTO{Status: status, Findings: findings}
}

func safeBPMNPath(dir, file string) string {
	clean := filepath.Clean(file)
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return filepath.Join(dir, filepath.Base(clean))
	}
	return filepath.Join(dir, clean)
}
func safeArtifactName(id string) string {
	s := strings.ToLower(strings.NewReplacer(" ", "-", "_", "-").Replace(id))
	out := strings.Builder{}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			out.WriteRune(r)
		}
	}
	if out.Len() == 0 {
		return "process"
	}
	return out.String()
}
func sanitizeBPMNID(id string) string {
	s := safeArtifactName(id)
	return strings.ReplaceAll(s, "-", "_")
}
func nextProcessID(path string) string {
	nodes, _ := scanProcesses(path)
	return fmt.Sprintf("PRC-%07d", len(nodes)+1)
}

func attachProcessToProduct(path, productID, processID string) error {
	bp, err := GetBlueprint(path, productID)
	if err != nil {
		return err
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return err
	}
	for _, id := range raw.Processes {
		if id == processID {
			return nil
		}
	}
	raw.Processes = append(raw.Processes, processID)
	return fsx.WriteYAML(bp.Path, raw)
}

func DefaultBPMNTemplate(processID, productName string) string {
	if strings.TrimSpace(processID) == "" {
		processID = "Process_ProductFulfillment"
	}
	name := firstNonEmpty(productName, "Product fulfillment")
	return `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" xmlns:di="http://www.omg.org/spec/DD/20100524/DI" id="Definitions_ProductProcess" targetNamespace="https://nomos.local/bpmn">
  <bpmn:process id="` + processID + `" name="` + xmlEscape(name) + `" isExecutable="false">
    <bpmn:startEvent id="StartEvent_RequestReceived" name="Request received" />
    <bpmn:task id="Task_ValidateRequest" name="Validate request" />
    <bpmn:task id="Task_PerformFulfillment" name="Perform fulfillment" />
    <bpmn:task id="Task_QualityCheck" name="Quality check" />
    <bpmn:task id="Task_DocumentEvidence" name="Document evidence" />
    <bpmn:endEvent id="EndEvent_Fulfilled" name="Fulfilled" />
  </bpmn:process>
  <bpmndi:BPMNDiagram id="BPMNDiagram_ProductProcess"><bpmndi:BPMNPlane id="BPMNPlane_ProductProcess" bpmnElement="` + processID + `" /></bpmndi:BPMNDiagram>
</bpmn:definitions>
`
}
func xmlEscape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
