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
	"github.com/nomos/nomos/internal/idgen"
	"github.com/nomos/nomos/internal/idmigrate"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

var bpmnTaskTypes = map[string]bool{
	"task": true, "userTask": true, "serviceTask": true, "businessRuleTask": true,
	"manualTask": true, "scriptTask": true, "callActivity": true,
	"exclusiveGateway": true,
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
		if bp.Metadata.ID == productID && (bp.Metadata.Type == "product_blueprint" || bp.Metadata.Type == "product") {
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
	// ADR-0020: Legacy-IDs werden transparent via id-history aufgelöst.
	resolved, _ := idmigrate.Resolve(path, id)
	for _, n := range nodes {
		if n.Meta.ID == id || n.Meta.ID == resolved {
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
	if err := os.WriteFile(filepath.Join(dir, bpmnFile), []byte(DefaultBPMNTemplate(meta.BPMN.ProcessID, meta.Name)), 0o644); err != nil {
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
	// Auto-sync: if the BPMN has no tasks but the YAML has steps, the file
	// was created before step-sync was implemented — regenerate it now.
	tasks, _ := ExtractBPMNTasks(string(data))
	if len(tasks) == 0 && len(dto.Steps) > 0 {
		if node, nerr := findProcessNode(path, id); nerr == nil {
			if serr := syncBPMNFromSteps(node); serr == nil {
				data, _ = os.ReadFile(dto.BPMNPath)
			}
		}
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
	// Mirror lane → service bindings from BPMN into the YAML so the catalog
	// API can reason about lane membership without re-parsing the diagram.
	if lanes, err := ExtractBPMNLanes(xmlText); err == nil {
		node, nerr := findProcessNode(path, id)
		if nerr == nil {
			node.Meta.Lanes = lanes
			_ = fsx.WriteYAML(node.Path, node.Meta)
		}
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
		mappings = append(mappings, model.ProcessTaskMapping{BPMNElementID: m.BPMNElementID, TaskName: firstNonEmpty(m.TaskName, t.Name), BPMNElementType: firstNonEmpty(m.BPMNElementType, t.ElementType), ServiceRef: strings.TrimSpace(m.ServiceRef), Method: strings.TrimSpace(m.Method), Role: firstNonEmpty(m.Role, "primary"), Required: m.Required, Notes: m.Notes})
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
	if p := n.Meta.Participant; p != nil {
		dto.Participant = &ProcessParticipantDTO{Name: p.Name, Ref: p.Ref}
	}
	for _, s := range n.Meta.Steps {
		dto.Steps = append(dto.Steps, ProcessStepDTO{ID: s.ID, Name: s.Name, TaskType: normalizedStepTaskType(s.TaskType), ServiceRef: s.ServiceRef, CapabilityRef: s.CapabilityRef, Method: s.Method, DecisionRef: s.DecisionRef, Role: s.Role, Required: s.Required, Notes: s.Notes, DependsOn: s.DependsOn, Inputs: stepsInputsToDTO(s.Inputs), Outputs: stepsOutputsToDTO(s.Outputs), Decision: decisionToDTO(s.Decision), Gateway: gatewayToDTO(s.Gateway)})
	}
	for _, m := range n.Meta.TaskMappings {
		dto.TaskMappings = append(dto.TaskMappings, ProcessTaskMappingDTO{BPMNElementID: m.BPMNElementID, TaskName: m.TaskName, BPMNElementType: m.BPMNElementType, ServiceRef: m.ServiceRef, CapabilityRef: m.CapabilityRef, Method: m.Method, Role: m.Role, Required: m.Required, Notes: m.Notes})
	}
	for _, l := range n.Meta.Lanes {
		dto.Lanes = append(dto.Lanes, ProcessLaneDTO{BPMNLaneID: l.BPMNLaneID, Name: l.Name, ServiceRef: l.ServiceRef})
	}
	var starts []StartEventInfo
	if data, err := os.ReadFile(bpmnPath); err == nil {
		if tasks, err := ExtractBPMNTasks(string(data)); err == nil {
			dto.Tasks = withMappingStatus(tasks, dto.TaskMappings)
		}
		if evs, err := ExtractBPMNStartEvents(string(data)); err == nil {
			starts = evs
		}
	}
	dto.Triggers = mergeTriggerView(starts, n.Meta.Triggers)
	if includeValidation {
		dto.Validation = validateProcessDTO(path, dto, starts)
	}
	return dto
}

// syncBPMNFromSteps rewrites the process's BPMN file so the diagram reflects
// the structured step list. Called whenever steps are added, updated, or
// removed via the structured API so users see their tasks in the diagram.
func syncBPMNFromSteps(node processNode) error {
	bpmnPath := safeBPMNPath(filepath.Dir(node.Path), node.Meta.BPMN.File)
	if bpmnPath == "" {
		return nil
	}
	cosmosPath := workspaceFromPath(node.Path)
	xmlText := BuildBPMNFromSteps(cosmosPath, node.Meta.BPMN.ProcessID, node.Meta.Name, node.Meta.Steps)
	return os.WriteFile(bpmnPath, []byte(xmlText), 0o644)
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
		ID:            newStepID(),
		Name:          strings.TrimSpace(req.Name),
		TaskType:      normalizedStepTaskType(req.TaskType),
		ServiceRef:    strings.TrimSpace(req.ServiceRef),
		CapabilityRef: strings.TrimSpace(req.CapabilityRef),
		Method:        strings.TrimSpace(req.Method),
		DecisionRef:   strings.TrimSpace(req.DecisionRef),
		Role:          firstNonEmpty(strings.TrimSpace(req.Role), "supporting"),
		Required:      req.Required,
		Notes:         req.Notes,
		DependsOn:     req.DependsOn,
		Inputs:        dtoInputsToModel(req.Inputs),
		Outputs:       dtoOutputsToModel(req.Outputs),
		Decision:      decisionToModel(req.Decision, normalizedStepTaskType(req.TaskType)),
		Gateway:       gatewayToModel(req.Gateway, normalizedStepTaskType(req.TaskType)),
	}
	node.Meta.Steps = append(node.Meta.Steps, step)
	if err := fsx.WriteYAML(node.Path, node.Meta); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write process: "+err.Error(), http.StatusInternalServerError, err)
	}
	if err := syncBPMNFromSteps(node); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write BPMN: "+err.Error(), http.StatusInternalServerError, err)
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
			node.Meta.Steps[i].TaskType = normalizedStepTaskType(req.TaskType)
			node.Meta.Steps[i].ServiceRef = strings.TrimSpace(req.ServiceRef)
			node.Meta.Steps[i].CapabilityRef = strings.TrimSpace(req.CapabilityRef)
			node.Meta.Steps[i].Method = strings.TrimSpace(req.Method)
			node.Meta.Steps[i].DecisionRef = strings.TrimSpace(req.DecisionRef)
			node.Meta.Steps[i].Role = firstNonEmpty(strings.TrimSpace(req.Role), "supporting")
			node.Meta.Steps[i].Required = req.Required
			node.Meta.Steps[i].Notes = req.Notes
			node.Meta.Steps[i].DependsOn = req.DependsOn
			node.Meta.Steps[i].Inputs = dtoInputsToModel(req.Inputs)
			node.Meta.Steps[i].Outputs = dtoOutputsToModel(req.Outputs)
			node.Meta.Steps[i].Decision = decisionToModel(req.Decision, normalizedStepTaskType(req.TaskType))
			node.Meta.Steps[i].Gateway = gatewayToModel(req.Gateway, normalizedStepTaskType(req.TaskType))
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
	if err := syncBPMNFromSteps(node); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write BPMN: "+err.Error(), http.StatusInternalServerError, err)
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
	if err := syncBPMNFromSteps(node); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write BPMN: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetProcess(path, id)
}

// UpdateProcessParticipant sets the participant (pool) metadata on a process,
// identifying the surrounding system or actor responsible for executing it.
func UpdateProcessParticipant(path, id string, req UpdateParticipantRequest) (ProcessDTO, error) {
	node, err := findProcessNode(path, id)
	if err != nil {
		return ProcessDTO{}, err
	}
	if strings.TrimSpace(req.Name) == "" {
		node.Meta.Participant = nil
	} else {
		node.Meta.Participant = &model.ProcessParticipant{
			Name: strings.TrimSpace(req.Name),
			Ref:  strings.TrimSpace(req.Ref),
		}
	}
	if err := fsx.WriteYAML(node.Path, node.Meta); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write process: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetProcess(path, id)
}

// GetProductCollaboration returns the collaboration view for a product offering:
// all processes that belong to the product, each with their participant (surrounding
// system). This maps to a bpmn:collaboration where the product = collaboration and
// each process = one bpmn:participant pool.
func GetProductCollaboration(path, productID string) (CollaborationDTO, error) {
	bp, err := GetBlueprint(path, productID)
	if err != nil {
		return CollaborationDTO{}, err
	}
	processes, err := ListProductProcesses(path, productID)
	if err != nil {
		return CollaborationDTO{}, err
	}
	out := CollaborationDTO{
		ProductID:   productID,
		ProductName: bp.Name,
	}
	for _, p := range processes.Items {
		entry := CollaborationParticipantDTO{
			ProcessID:   p.ID,
			ProcessName: p.Name,
		}
		if p.Participant != nil {
			entry.Participant = ProcessParticipantDTO{Name: p.Participant.Name, Ref: p.Participant.Ref}
		} else {
			entry.Participant = ProcessParticipantDTO{Name: p.Name}
		}
		out.Participants = append(out.Participants, entry)
	}
	return out, nil
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
			tasks = append(tasks, BPMNTaskDTO{ID: id, Name: name, ElementType: "bpmn:" + se.Name.Local, Standard: isStandardBPMNTask(id)})
		}
	}
	return tasks, nil
}

// bpmnLaneServicePrefix marks a bpmn:documentation entry that binds the
// surrounding bpmn:lane to a Nomos service. The remainder of the text is
// treated as the service canonical ref (e.g. "identity.blumer.cloud/user-account").
const bpmnLaneServicePrefix = "nomos-service-ref:"

// ExtractBPMNLanes returns every bpmn:lane in the file together with its
// optional service binding (decoded from a bpmn:documentation child whose
// text starts with the nomos-service-ref: prefix).
func ExtractBPMNLanes(xmlText string) ([]model.ProcessLane, error) {
	dec := xml.NewDecoder(bytes.NewBufferString(xmlText))
	lanes := []model.ProcessLane{}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "lane" {
			continue
		}
		var lane model.ProcessLane
		for _, a := range se.Attr {
			switch a.Name.Local {
			case "id":
				lane.BPMNLaneID = a.Value
			case "name":
				lane.Name = a.Value
			}
		}
		// Walk children of <bpmn:lane> looking for <bpmn:documentation> entries.
		for {
			t, err := dec.Token()
			if err == io.EOF || err != nil {
				break
			}
			if end, ok := t.(xml.EndElement); ok && end.Name.Local == "lane" {
				break
			}
			child, ok := t.(xml.StartElement)
			if !ok || child.Name.Local != "documentation" {
				continue
			}
			var text string
			if err := dec.DecodeElement(&text, &child); err != nil {
				continue
			}
			text = strings.TrimSpace(text)
			if strings.HasPrefix(text, bpmnLaneServicePrefix) {
				lane.ServiceRef = strings.TrimSpace(strings.TrimPrefix(text, bpmnLaneServicePrefix))
			}
		}
		if lane.BPMNLaneID != "" {
			lanes = append(lanes, lane)
		}
	}
	return lanes, nil
}

func isStandardBPMNTask(id string) bool {
	switch id {
	case "Task_ValidateRequest", "Task_PerformFulfillment", "Task_QualityCheck", "Task_DocumentEvidence":
		return true
	default:
		return false
	}
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
			tasks[i].Method = m.Method
		} else if tasks[i].Standard {
			tasks[i].MappingStatus = "standard"
		} else {
			tasks[i].MappingStatus = "unmapped"
		}
	}
	return tasks
}

func validateProcessDTO(path string, p ProcessDTO, starts []StartEventInfo) ProcessValidationDTO {
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
	for _, step := range p.Steps {
		if normalizedStepTaskType(step.TaskType) != "businessRuleTask" {
			continue
		}
		if len(step.Outputs) == 0 {
			add("BUSINESS_RULE_OUTPUTS_MISSING", "warning", "Business Rule task should define output variables: "+firstNonEmpty(step.Name, step.ID))
		}
		if step.Decision == nil || len(step.Decision.Rules) == 0 {
			add("BUSINESS_RULE_DECISION_MISSING", "warning", "Business Rule task should define a DMN decision table: "+firstNonEmpty(step.Name, step.ID))
		}
		if step.Gateway == nil {
			add("BUSINESS_RULE_GATEWAY_MISSING", "error", "Business Rule task must be followed by a gateway: "+firstNonEmpty(step.Name, step.ID))
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
		if strings.TrimSpace(m.Method) == "" {
			add("TASK_METHOD_MISSING", "warning", "Mapped service task should point to a service method: "+firstNonEmpty(m.TaskName, m.BPMNElementID))
		}
	}
	for _, t := range p.Tasks {
		if t.Standard {
			continue
		}
		if !mapped[t.ID] {
			add("BPMN_TASK_UNMAPPED", "warning", "BPMN task is not mapped to a service: "+firstNonEmpty(t.Name, t.ID))
		}
	}
	validateTriggers(p, starts, add)
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

// newStepID erzeugt eine ProcessStep-ID gemäß ADR-0020.
// Bei einem Generator-Fehler fällt es auf eine Zeitstempel-ID zurück.
func newStepID() string {
	id, err := idgen.NewForType("process_step")
	if err == nil {
		return id
	}
	return fmt.Sprintf("step-%d", time.Now().UnixNano())
}

// nextProcessID erzeugt eine neue ID gemäß ADR-0020.
// Bei einem Generator-Fehler fällt auf das alte Schema zurück.
func nextProcessID(path string) string {
	id, err := idgen.NewForType("process")
	if err == nil {
		return id
	}
	nodes, _ := scanProcesses(path)
	return fmt.Sprintf("PRC-%07d", len(nodes)+1)
}

func DeleteProcess(path, id string) error {
	node, err := findProcessNode(path, id)
	if err != nil {
		return err
	}
	if err := os.Remove(node.Path); err != nil && !os.IsNotExist(err) {
		return Error(CodeInternalError, "Failed to delete process: "+err.Error(), http.StatusInternalServerError, err)
	}
	bpmnPath := safeBPMNPath(filepath.Dir(node.Path), node.Meta.BPMN.File)
	_ = os.Remove(bpmnPath)
	if node.Meta.RelatedProduct != "" {
		_ = detachProcessFromProduct(path, node.Meta.RelatedProduct, id)
	}
	return nil
}

func detachProcessFromProduct(path, productID, processID string) error {
	bp, err := GetBlueprint(path, productID)
	if err != nil {
		return err
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return err
	}
	filtered := raw.Processes[:0]
	for _, id := range raw.Processes {
		if id != processID {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == len(raw.Processes) {
		return nil
	}
	raw.Processes = filtered
	return fsx.WriteYAML(bp.Path, raw)
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

// DefaultBPMNTemplate generates a minimal blank BPMN collaboration with a
// single participant pool. The pool name is the process name so users can
// immediately see which process they are editing in the diagram.
func DefaultBPMNTemplate(processID, processName string) string {
	return BuildBPMNFromSteps("", processID, processName, nil)
}

// gwBranch is one outgoing branch of a BPMN exclusive gateway.
type gwBranch struct {
	flowID   string
	targetID string
	label    string
	cond     string // conditionExpression (empty = unconditional)
	errEndID string // non-empty → routes to an inline error end event
}

// workspaceFromPath walks up from a nested cosmos file path to the workspace root
// (the directory that contains ".nomos/").
func workspaceFromPath(p string) string {
	dir := filepath.Dir(p)
	for i := 0; i < 12; i++ {
		if filepath.Base(dir) == ".nomos" {
			return filepath.Dir(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
	return ""
}

// findDecisionByID scans all domain directories for a decision with the given ID.
func findDecisionByID(cosmosPath, id string) *model.Decision {
	if cosmosPath == "" || id == "" {
		return nil
	}
	root := storage.DomainsDirForRead(cosmosPath)
	var found *model.Decision
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || found != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		for _, name := range []string{id + ".yaml", "decision.yaml"} {
			var dec model.Decision
			if fsx.ReadYAML(filepath.Join(p, name), &dec) == nil && dec.ID == id {
				found = &dec
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

// gatewayBranches returns the outgoing branches for an exclusive gateway step.
// When the step's Gateway.Conditions is empty, branches are auto-resolved from
// the preceding businessRuleTask's decision outputs (boolean → Ja/Nein).
func gatewayBranches(cosmosPath string, steps []model.ProcessStep, gwIdx int, gwID, defaultNextID string) []gwBranch {
	s := steps[gwIdx]
	sid := sanitizeBPMNID(gwID)

	if s.Gateway != nil && len(s.Gateway.Conditions) > 0 {
		out := make([]gwBranch, 0, len(s.Gateway.Conditions))
		for ci, c := range s.Gateway.Conditions {
			fid := fmt.Sprintf("Flow_%s_B%d", sid, ci+1)
			target := defaultNextID
			errEndID := ""
			if c.TargetStep != "" {
				for si, ss := range steps {
					if ss.ID == c.TargetStep {
						target = bpmnTaskIDForStep(ss, si)
						break
					}
				}
			} else if !isHappyGatewayValue(c.Value) {
				errEndID = fmt.Sprintf("ErrEnd_%s_%d", sid, ci+1)
				target = errEndID
			}
			cond := ""
			if c.Output != "" {
				op := firstNonEmpty(c.Operator, "==")
				cond = fmt.Sprintf("${%s %s %s}", c.Output, op, c.Value)
			}
			out = append(out, gwBranch{flowID: fid, targetID: target, label: c.Label, cond: cond, errEndID: errEndID})
		}
		return out
	}

	// Auto-resolve: find the primary boolean output of the preceding businessRuleTask's decision.
	gateVar := "result"
	for j := gwIdx - 1; j >= 0; j-- {
		if normalizedStepTaskType(steps[j].TaskType) == "businessRuleTask" && steps[j].DecisionRef != "" {
			if dec := findDecisionByID(cosmosPath, steps[j].DecisionRef); dec != nil {
				for _, o := range dec.Outputs {
					if strings.EqualFold(o.Type, "boolean") {
						gateVar = o.Name
						break
					}
				}
			}
			break
		}
	}

	errEndID := "ErrEnd_" + sid
	return []gwBranch{
		{flowID: "Flow_" + sid + "_Yes", targetID: defaultNextID, label: "Ja", cond: "${" + gateVar + " == true}"},
		{flowID: "Flow_" + sid + "_No", targetID: errEndID, label: "Nein", cond: "${" + gateVar + " == false}", errEndID: errEndID},
	}
}

func isHappyGatewayValue(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "true" || v == "yes" || v == "1"
}

// BuildBPMNFromSteps generates a BPMN 2.0 collaboration XML where the pool is
// labelled with the process name and each ProcessStep becomes a task (or
// exclusive gateway) laid out between Start and End. Exclusive gateways are
// rendered with a diamond shape and an error end event for the rejection branch,
// with conditions auto-resolved from the preceding businessRuleTask's decision.
func BuildBPMNFromSteps(cosmosPath, processID, processName string, steps []model.ProcessStep) string {
	if strings.TrimSpace(processID) == "" {
		processID = "Process_ProductFulfillment"
	}
	poolName := firstNonEmpty(processName, "Process")
	collabID := "Collab_" + sanitizeBPMNID(processID)
	participantID := "Participant_" + sanitizeBPMNID(processID)

	const (
		laneX          = 100
		laneY          = 80
		laneHeightBase = 160
		laneHeightGW   = 280
		eventSize      = 36
		taskWidth      = 120
		taskHeight     = 80
		gwSize         = 50
		gap            = 50
		marginLeft     = 80
		centerY        = 160
		taskTop        = 120
		labelOffsetY   = 43
		errorEndCY     = 260 // center-Y of error end events below main flow
		errorEndSize   = 28
	)

	startEventID := "StartEvent_Begin"
	endEventID := "EndEvent_Done"

	hasGateways := false
	for _, s := range steps {
		if normalizedStepTaskType(s.TaskType) == "exclusiveGateway" {
			hasGateways = true
			break
		}
	}
	laneHeight := laneHeightBase
	if hasGateways {
		laneHeight = laneHeightGW
	}

	var (
		processBody bytes.Buffer
		planeBody   bytes.Buffer
		flows       bytes.Buffer
		shapes      bytes.Buffer
	)

	xCursor := laneX + marginLeft

	// ── Start event ───────────────────────────────────────────────────────────
	startOutFlow := "Flow_Start_End"
	if len(steps) > 0 {
		startOutFlow = "Flow_Start_" + sanitizeBPMNID(bpmnTaskIDForStep(steps[0], 0))
	}
	processBody.WriteString("    <bpmn:startEvent id=\"" + startEventID + "\" name=\"Start\">\n")
	processBody.WriteString("      <bpmn:outgoing>" + startOutFlow + "</bpmn:outgoing>\n")
	processBody.WriteString("    </bpmn:startEvent>\n")
	shapes.WriteString("      <bpmndi:BPMNShape id=\"Shape_Start\" bpmnElement=\"" + startEventID + "\">\n")
	shapes.WriteString(fmt.Sprintf("        <dc:Bounds x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" />\n", xCursor, centerY-eventSize/2, eventSize, eventSize))
	shapes.WriteString(fmt.Sprintf("        <bpmndi:BPMNLabel><dc:Bounds x=\"%d\" y=\"%d\" width=\"24\" height=\"14\" /></bpmndi:BPMNLabel>\n", xCursor+6, centerY-eventSize/2+labelOffsetY))
	shapes.WriteString("      </bpmndi:BPMNShape>\n")
	startRightX := xCursor + eventSize
	xCursor += eventSize + gap

	// ── Steps ─────────────────────────────────────────────────────────────────
	type stepLayout struct {
		taskID    string
		outFlow   string // primary (happy) outgoing flow
		leftX     int
		rightX    int
		isGateway bool
		branches  []gwBranch
	}
	layouts := make([]stepLayout, len(steps))
	prevOutFlow := startOutFlow

	for i, s := range steps {
		taskID := bpmnTaskIDForStep(s, i)

		if normalizedStepTaskType(s.TaskType) == "exclusiveGateway" {
			nextID := endEventID
			if i+1 < len(steps) {
				nextID = bpmnTaskIDForStep(steps[i+1], i+1)
			}
			branches := gatewayBranches(cosmosPath, steps, i, taskID, nextID)

			// Primary (happy) flow = first branch without an error end.
			primaryFlow := ""
			for _, b := range branches {
				if b.errEndID == "" {
					primaryFlow = b.flowID
					break
				}
			}
			if primaryFlow == "" && len(branches) > 0 {
				primaryFlow = branches[0].flowID
			}

			// Gateway element.
			processBody.WriteString(fmt.Sprintf("    <bpmn:exclusiveGateway id=\"%s\" name=\"%s\">\n", taskID, xmlEscape(firstNonEmpty(s.Name, taskID))))
			processBody.WriteString("      <bpmn:incoming>" + prevOutFlow + "</bpmn:incoming>\n")
			for _, b := range branches {
				processBody.WriteString("      <bpmn:outgoing>" + b.flowID + "</bpmn:outgoing>\n")
			}
			processBody.WriteString("    </bpmn:exclusiveGateway>\n")

			// Inline error end events.
			for _, b := range branches {
				if b.errEndID != "" {
					processBody.WriteString(fmt.Sprintf("    <bpmn:endEvent id=\"%s\" name=\"Abgelehnt\">\n", b.errEndID))
					processBody.WriteString(fmt.Sprintf("      <bpmn:incoming>%s</bpmn:incoming>\n", b.flowID))
					processBody.WriteString("      <bpmn:errorEventDefinition />\n")
					processBody.WriteString("    </bpmn:endEvent>\n")
				}
			}

			// Gateway diamond shape.
			gwCenterX := xCursor + gwSize/2
			shapes.WriteString(fmt.Sprintf("      <bpmndi:BPMNShape id=\"Shape_%s\" bpmnElement=\"%s\" isMarkerVisible=\"true\">\n", taskID, taskID))
			shapes.WriteString(fmt.Sprintf("        <dc:Bounds x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" />\n", xCursor, centerY-gwSize/2, gwSize, gwSize))
			shapes.WriteString("      </bpmndi:BPMNShape>\n")

			// Error end event shapes.
			for _, b := range branches {
				if b.errEndID != "" {
					errX := gwCenterX - errorEndSize/2
					shapes.WriteString(fmt.Sprintf("      <bpmndi:BPMNShape id=\"Shape_%s\" bpmnElement=\"%s\">\n", b.errEndID, b.errEndID))
					shapes.WriteString(fmt.Sprintf("        <dc:Bounds x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" />\n", errX, errorEndCY-errorEndSize/2, errorEndSize, errorEndSize))
					shapes.WriteString("      </bpmndi:BPMNShape>\n")
				}
			}

			layouts[i] = stepLayout{taskID: taskID, outFlow: primaryFlow, leftX: xCursor, rightX: xCursor + gwSize, isGateway: true, branches: branches}
			prevOutFlow = primaryFlow
			xCursor += gwSize + gap
			continue
		}

		// Regular task.
		element := bpmnElementForStep(s.TaskType)
		var primaryFlow string
		if i == len(steps)-1 {
			primaryFlow = "Flow_" + sanitizeBPMNID(taskID) + "_End"
		} else {
			primaryFlow = "Flow_" + sanitizeBPMNID(taskID) + "_" + sanitizeBPMNID(bpmnTaskIDForStep(steps[i+1], i+1))
		}
		processBody.WriteString(fmt.Sprintf("    <bpmn:%s id=\"%s\" name=\"%s\">\n", element, taskID, xmlEscape(firstNonEmpty(s.Name, taskID))))
		processBody.WriteString("      <bpmn:incoming>" + prevOutFlow + "</bpmn:incoming>\n")
		processBody.WriteString("      <bpmn:outgoing>" + primaryFlow + "</bpmn:outgoing>\n")
		processBody.WriteString(fmt.Sprintf("    </bpmn:%s>\n", element))
		shapes.WriteString(fmt.Sprintf("      <bpmndi:BPMNShape id=\"Shape_%s\" bpmnElement=\"%s\">\n", taskID, taskID))
		shapes.WriteString(fmt.Sprintf("        <dc:Bounds x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" />\n", xCursor, taskTop, taskWidth, taskHeight))
		shapes.WriteString("      </bpmndi:BPMNShape>\n")
		layouts[i] = stepLayout{taskID: taskID, outFlow: primaryFlow, leftX: xCursor, rightX: xCursor + taskWidth}
		prevOutFlow = primaryFlow
		xCursor += taskWidth + gap
	}

	// ── End event ─────────────────────────────────────────────────────────────
	processBody.WriteString("    <bpmn:endEvent id=\"" + endEventID + "\" name=\"End\">\n")
	processBody.WriteString("      <bpmn:incoming>" + prevOutFlow + "</bpmn:incoming>\n")
	processBody.WriteString("    </bpmn:endEvent>\n")
	shapes.WriteString("      <bpmndi:BPMNShape id=\"Shape_End\" bpmnElement=\"" + endEventID + "\">\n")
	shapes.WriteString(fmt.Sprintf("        <dc:Bounds x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" />\n", xCursor, centerY-eventSize/2, eventSize, eventSize))
	shapes.WriteString(fmt.Sprintf("        <bpmndi:BPMNLabel><dc:Bounds x=\"%d\" y=\"%d\" width=\"24\" height=\"14\" /></bpmndi:BPMNLabel>\n", xCursor+6, centerY-eventSize/2+labelOffsetY))
	shapes.WriteString("      </bpmndi:BPMNShape>\n")
	endLeftX := xCursor
	endRightX := xCursor + eventSize

	// ── Sequence flows + DI edges ──────────────────────────────────────────────
	firstTargetID := endEventID
	firstTargetLeft := endLeftX
	if len(layouts) > 0 {
		firstTargetID = layouts[0].taskID
		firstTargetLeft = layouts[0].leftX
	}
	processBody.WriteString(fmt.Sprintf("    <bpmn:sequenceFlow id=\"%s\" sourceRef=\"%s\" targetRef=\"%s\" />\n", startOutFlow, startEventID, firstTargetID))
	flows.WriteString("      <bpmndi:BPMNEdge id=\"Edge_" + startOutFlow + "\" bpmnElement=\"" + startOutFlow + "\">\n")
	flows.WriteString(fmt.Sprintf("        <di:waypoint x=\"%d\" y=\"%d\" /><di:waypoint x=\"%d\" y=\"%d\" />\n", startRightX, centerY, firstTargetLeft, centerY))
	flows.WriteString("      </bpmndi:BPMNEdge>\n")

	for i, lay := range layouts {
		nextRef := endEventID
		nextLeftX := endLeftX
		if i < len(layouts)-1 {
			nextRef = layouts[i+1].taskID
			nextLeftX = layouts[i+1].leftX
		}

		if lay.isGateway {
			gwCenterX := lay.leftX + gwSize/2
			gwRightX := lay.rightX
			gwBottomY := centerY + gwSize/2

			for _, b := range lay.branches {
				// Sequence flow with condition.
				if b.cond != "" {
					processBody.WriteString(fmt.Sprintf("    <bpmn:sequenceFlow id=\"%s\" name=\"%s\" sourceRef=\"%s\" targetRef=\"%s\">\n", b.flowID, xmlEscape(b.label), lay.taskID, b.targetID))
					processBody.WriteString(fmt.Sprintf("      <bpmn:conditionExpression>%s</bpmn:conditionExpression>\n", xmlEscape(b.cond)))
					processBody.WriteString("    </bpmn:sequenceFlow>\n")
				} else {
					processBody.WriteString(fmt.Sprintf("    <bpmn:sequenceFlow id=\"%s\" name=\"%s\" sourceRef=\"%s\" targetRef=\"%s\" />\n", b.flowID, xmlEscape(b.label), lay.taskID, b.targetID))
				}

				// DI edge.
				if b.errEndID != "" {
					// Vertical: from gateway bottom → error end.
					midY := (gwBottomY + errorEndCY) / 2
					flows.WriteString(fmt.Sprintf("      <bpmndi:BPMNEdge id=\"Edge_%s\" bpmnElement=\"%s\">\n", b.flowID, b.flowID))
					flows.WriteString(fmt.Sprintf("        <di:waypoint x=\"%d\" y=\"%d\" />\n", gwCenterX, gwBottomY))
					flows.WriteString(fmt.Sprintf("        <di:waypoint x=\"%d\" y=\"%d\" />\n", gwCenterX, errorEndCY))
					flows.WriteString(fmt.Sprintf("        <bpmndi:BPMNLabel><dc:Bounds x=\"%d\" y=\"%d\" width=\"40\" height=\"14\" /></bpmndi:BPMNLabel>\n", gwCenterX+5, midY))
					flows.WriteString("      </bpmndi:BPMNEdge>\n")
				} else {
					// Horizontal: from gateway right → next element.
					labelMidX := (gwRightX + nextLeftX) / 2
					flows.WriteString(fmt.Sprintf("      <bpmndi:BPMNEdge id=\"Edge_%s\" bpmnElement=\"%s\">\n", b.flowID, b.flowID))
					flows.WriteString(fmt.Sprintf("        <di:waypoint x=\"%d\" y=\"%d\" /><di:waypoint x=\"%d\" y=\"%d\" />\n", gwRightX, centerY, nextLeftX, centerY))
					flows.WriteString(fmt.Sprintf("        <bpmndi:BPMNLabel><dc:Bounds x=\"%d\" y=\"%d\" width=\"24\" height=\"14\" /></bpmndi:BPMNLabel>\n", labelMidX-12, centerY-20))
					flows.WriteString("      </bpmndi:BPMNEdge>\n")
				}
			}
			continue
		}

		// Normal task → next.
		processBody.WriteString(fmt.Sprintf("    <bpmn:sequenceFlow id=\"%s\" sourceRef=\"%s\" targetRef=\"%s\" />\n", lay.outFlow, lay.taskID, nextRef))
		flows.WriteString("      <bpmndi:BPMNEdge id=\"Edge_" + lay.outFlow + "\" bpmnElement=\"" + lay.outFlow + "\">\n")
		flows.WriteString(fmt.Sprintf("        <di:waypoint x=\"%d\" y=\"%d\" /><di:waypoint x=\"%d\" y=\"%d\" />\n", lay.rightX, centerY, nextLeftX, centerY))
		flows.WriteString("      </bpmndi:BPMNEdge>\n")
	}

	// ── Assemble ──────────────────────────────────────────────────────────────
	laneWidth := endRightX - laneX + marginLeft
	if laneWidth < 400 {
		laneWidth = 400
	}

	planeBody.WriteString(fmt.Sprintf("      <bpmndi:BPMNShape id=\"Shape_Participant\" bpmnElement=\"%s\" isHorizontal=\"true\">\n", participantID))
	planeBody.WriteString(fmt.Sprintf("        <dc:Bounds x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" />\n", laneX, laneY, laneWidth, laneHeight))
	planeBody.WriteString("      </bpmndi:BPMNShape>\n")
	planeBody.Write(shapes.Bytes())
	planeBody.Write(flows.Bytes())

	var out bytes.Buffer
	out.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	out.WriteString("<bpmn:definitions xmlns:bpmn=\"http://www.omg.org/spec/BPMN/20100524/MODEL\" xmlns:bpmndi=\"http://www.omg.org/spec/BPMN/20100524/DI\" xmlns:dc=\"http://www.omg.org/spec/DD/20100524/DC\" xmlns:di=\"http://www.omg.org/spec/DD/20100524/DI\" id=\"Definitions_ProductProcess\" targetNamespace=\"https://nomos.local/bpmn\">\n")
	out.WriteString("  <bpmn:collaboration id=\"" + collabID + "\">\n")
	out.WriteString("    <bpmn:participant id=\"" + participantID + "\" name=\"" + xmlEscape(poolName) + "\" processRef=\"" + processID + "\" />\n")
	out.WriteString("  </bpmn:collaboration>\n")
	out.WriteString("  <bpmn:process id=\"" + processID + "\" name=\"" + xmlEscape(poolName) + "\" isExecutable=\"false\">\n")
	out.Write(processBody.Bytes())
	out.WriteString("  </bpmn:process>\n")
	out.WriteString("  <bpmndi:BPMNDiagram id=\"BPMNDiagram_ProductProcess\">\n")
	out.WriteString("    <bpmndi:BPMNPlane id=\"BPMNPlane_ProductProcess\" bpmnElement=\"" + collabID + "\">\n")
	out.Write(planeBody.Bytes())
	out.WriteString("    </bpmndi:BPMNPlane>\n")
	out.WriteString("  </bpmndi:BPMNDiagram>\n")
	out.WriteString("</bpmn:definitions>\n")
	return out.String()
}

func bpmnTaskIDForStep(s model.ProcessStep, idx int) string {
	base := strings.TrimSpace(s.ID)
	if base == "" {
		base = fmt.Sprintf("step-%d", idx+1)
	}
	return "Task_" + sanitizeBPMNID(base)
}

func bpmnElementForStep(taskType string) string {
	switch normalizedStepTaskType(taskType) {
	case "businessRuleTask":
		return "businessRuleTask"
	case "exclusiveGateway":
		return "exclusiveGateway"
	default:
		return "serviceTask"
	}
}

type bpmnStepLayout struct {
	taskID  string
	outFlow string
	leftX   int
	rightX  int
}

func bpmnFirstTargetRef(layouts []bpmnStepLayout, fallback string) string {
	if len(layouts) == 0 {
		return fallback
	}
	return layouts[0].taskID
}
func xmlEscape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func stepsInputsToDTO(ins []model.StepInputBinding) []StepInputBindingDTO {
	out := make([]StepInputBindingDTO, len(ins))
	for i, b := range ins {
		out[i] = StepInputBindingDTO{Name: b.Name, Source: b.Source, Required: b.Required}
	}
	return out
}

func stepsOutputsToDTO(outs []model.StepOutputSchema) []StepOutputSchemaDTO {
	out := make([]StepOutputSchemaDTO, len(outs))
	for i, s := range outs {
		out[i] = StepOutputSchemaDTO{Name: s.Name, Type: s.Type, Description: s.Description}
	}
	return out
}

func dtoInputsToModel(ins []StepInputBindingDTO) []model.StepInputBinding {
	out := make([]model.StepInputBinding, len(ins))
	for i, b := range ins {
		out[i] = model.StepInputBinding{Name: b.Name, Source: b.Source, Required: b.Required}
	}
	return out
}

func dtoOutputsToModel(outs []StepOutputSchemaDTO) []model.StepOutputSchema {
	out := make([]model.StepOutputSchema, len(outs))
	for i, s := range outs {
		out[i] = model.StepOutputSchema{Name: s.Name, Type: s.Type, Description: s.Description}
	}
	return out
}

func normalizedStepTaskType(taskType string) string {
	switch strings.TrimSpace(taskType) {
	case "businessRuleTask", "bpmn:businessRuleTask", "business_rule_task":
		return "businessRuleTask"
	case "exclusiveGateway", "exclusive_gateway", "bpmn:exclusiveGateway":
		return "exclusiveGateway"
	default:
		return "serviceTask"
	}
}

func decisionToDTO(d *model.DecisionTable) *DecisionTableDTO {
	if d == nil {
		return nil
	}
	out := &DecisionTableDTO{Name: d.Name, HitPolicy: d.HitPolicy, Description: d.Description}
	for _, r := range d.Rules {
		out.Rules = append(out.Rules, DecisionRuleDTO{Input: r.Input, Operator: r.Operator, Value: r.Value, Output: r.Output, OutputValue: r.OutputValue, Label: r.Label})
	}
	return out
}

func decisionToModel(d *DecisionTableDTO, taskType string) *model.DecisionTable {
	if normalizedStepTaskType(taskType) != "businessRuleTask" || d == nil {
		return nil
	}
	out := &model.DecisionTable{Name: strings.TrimSpace(firstNonEmpty(d.Name, "DMN Entscheidung")), HitPolicy: strings.TrimSpace(firstNonEmpty(d.HitPolicy, "UNIQUE")), Description: strings.TrimSpace(d.Description)}
	for _, r := range d.Rules {
		if strings.TrimSpace(r.Input) == "" || strings.TrimSpace(r.Output) == "" {
			continue
		}
		op := strings.TrimSpace(r.Operator)
		if op == "" {
			op = "<="
		}
		out.Rules = append(out.Rules, model.DecisionRule{Input: strings.TrimSpace(r.Input), Operator: op, Value: strings.TrimSpace(r.Value), Output: strings.TrimSpace(r.Output), OutputValue: strings.TrimSpace(firstNonEmpty(r.OutputValue, "true")), Label: strings.TrimSpace(r.Label)})
	}
	if len(out.Rules) == 0 {
		return nil
	}
	return out
}

func gatewayToDTO(g *model.DecisionGateway) *DecisionGatewayDTO {
	if g == nil {
		return nil
	}
	out := &DecisionGatewayDTO{Name: g.Name, DefaultTo: g.DefaultTo}
	for _, c := range g.Conditions {
		out.Conditions = append(out.Conditions, GatewayConditionDTO{Output: c.Output, Operator: c.Operator, Value: c.Value, TargetStep: c.TargetStep, Label: c.Label})
	}
	return out
}

func gatewayToModel(g *DecisionGatewayDTO, taskType string) *model.DecisionGateway {
	if normalizedStepTaskType(taskType) != "businessRuleTask" {
		return nil
	}
	out := &model.DecisionGateway{Name: strings.TrimSpace("Entscheidung")}
	if g != nil {
		out.Name = strings.TrimSpace(firstNonEmpty(g.Name, "Entscheidung"))
		out.DefaultTo = strings.TrimSpace(g.DefaultTo)
		for _, c := range g.Conditions {
			if strings.TrimSpace(c.Output) == "" {
				continue
			}
			operator := strings.TrimSpace(c.Operator)
			if operator == "" {
				operator = "=="
			}
			out.Conditions = append(out.Conditions, model.GatewayCondition{Output: strings.TrimSpace(c.Output), Operator: operator, Value: strings.TrimSpace(c.Value), TargetStep: strings.TrimSpace(c.TargetStep), Label: strings.TrimSpace(c.Label)})
		}
	}
	return out
}
