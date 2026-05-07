package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"

	"github.com/nomos/nomos/internal/graph"
)

type Finding struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
}
type Result struct {
	Status   string    `json:"status"`
	Findings []Finding `json:"findings"`
}

var allowedComplianceStatuses = map[string]bool{
	"unknown":       true,
	"compliant":     true,
	"warning":       true,
	"non_compliant": true,
	"blocked":       true,
}

func Validate(path string) (Result, error) {
	res := Result{Status: "ok", Findings: []Finding{}}
	if _, err := os.Stat(storage.CosmosFile(path)); err != nil {
		res.add("COSMOS_MISSING", "error", ".nomos/cosmos.yaml fehlt", filepath.Join(".nomos", "cosmos.yaml"))
	}
	if err := validateCatalogArtifacts(path, &res); err != nil {
		return res, err
	}
	for _, f := range res.Findings {
		if f.Severity == "error" {
			res.Status = "failed"
			break
		}
	}
	return res, nil
}

func validateCatalogArtifacts(root string, res *Result) error {
	catalogDir := storage.CatalogDirForRead(root)
	if _, err := os.Stat(catalogDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return filepath.WalkDir(catalogDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isYAML(path) {
			return nil
		}
		var header struct {
			Type string `yaml:"type"`
		}
		if err := fsx.ReadYAML(path, &header); err != nil {
			return err
		}
		rel := relPath(root, path)
		switch header.Type {
		case "product_blueprint", "service_blueprint":
			var b model.Blueprint
			if err := fsx.ReadYAML(path, &b); err != nil {
				return err
			}
			validateBlueprint(b, rel, res)
		case "product_instance", "service_instance":
			var inst model.Instance
			if err := fsx.ReadYAML(path, &inst); err != nil {
				return err
			}
			validateInstance(inst, rel, res)
		case "servicegraph":
			var sg model.Servicegraph
			if err := fsx.ReadYAML(path, &sg); err != nil {
				return err
			}
			validateServicegraph(sg, rel, res)
		case "product":
			// Backward-compatible product artifacts keep their existing loose validation.
		}
		return nil
	})
}

func validateBlueprint(b model.Blueprint, path string, res *Result) {
	for _, field := range []struct{ name, value string }{{"id", b.ID}, {"type", b.Type}, {"name", b.Name}, {"version", b.Version}, {"status", b.Status}, {"owner", b.Owner}} {
		if strings.TrimSpace(field.value) == "" {
			res.add("BLUEPRINT_REQUIRED_FIELD", "error", "Blueprint Pflichtfeld fehlt: "+field.name, path)
		}
	}
	if b.Type == "product_blueprint" {
		if len(b.RequiredInputs) == 0 {
			res.add("PRODUCT_BLUEPRINT_INPUTS_EMPTY", "error", "Product Blueprint benoetigt required_inputs", path)
		}
		if len(b.RequiredServiceBlueprints) == 0 {
			res.add("PRODUCT_BLUEPRINT_SERVICES_EMPTY", "error", "Provisionierbarer Product Blueprint benoetigt required_service_blueprints", path)
		}
		if len(b.RequiredServices) == 0 {
			res.add("PRODUCT_BLUEPRINT_REQUIRED_SERVICES_RECOMMENDED", "warning", "Product Blueprint sollte required_services mit Namespace Services und Service Blueprints angeben", path)
		}
		for i, svc := range b.RequiredServices {
			entryPath := path + ":required_services[" + fmt.Sprint(i) + "]"
			if strings.TrimSpace(svc.ServiceRef) == "" {
				res.add("PRODUCT_BLUEPRINT_REQUIRED_SERVICE_REF_EMPTY", "error", "required_services Eintrag benoetigt service_ref", entryPath)
			}
			if strings.TrimSpace(svc.ServiceBlueprintRef) == "" {
				res.add("PRODUCT_BLUEPRINT_REQUIRED_SERVICE_BLUEPRINT_REF_EMPTY", "error", "required_services Eintrag benoetigt service_blueprint_ref", entryPath)
			}
		}
	}
	if b.Type == "service_blueprint" {
		if ref := strings.TrimSpace(b.NamespaceServiceRef); ref != "" && !strings.Contains(ref, "/") {
			res.add("SERVICE_BLUEPRINT_NAMESPACE_SERVICE_REF_SHAPE", "warning", "namespace_service_ref sollte die Form <domain>/<service> haben", path)
		}
		if len(b.TargetSystems) == 0 && len(b.Providers) == 0 && len(b.Capabilities) == 0 && len(b.Actions) == 0 && len(b.QualityCriteria) == 0 {
			res.add("SERVICE_BLUEPRINT_CAPABILITY_EMPTY", "error", "Service Blueprint benoetigt mindestens Provider/Zielsystem, Capability, Action oder Quality Criterion", path)
		}
	}
}

func validateInstance(inst model.Instance, path string, res *Result) {
	for _, field := range []struct{ name, value string }{{"id", inst.ID}, {"type", inst.Type}, {"blueprint_ref", inst.BlueprintRef}, {"blueprint_version", inst.BlueprintVersion}} {
		if strings.TrimSpace(field.value) == "" {
			res.add("INSTANCE_REQUIRED_FIELD", "error", "Instance Pflichtfeld fehlt: "+field.name, path)
		}
	}
	status := strings.TrimSpace(inst.ComplianceStatus)
	if !allowedComplianceStatuses[status] {
		res.add("INSTANCE_COMPLIANCE_STATUS_INVALID", "error", "compliance_status ist ungueltig: "+status, path)
	}
}

var allowedNodeTypes = map[string]bool{
	"service": true, "sub_service": true, "activity": true, "precondition": true,
	"decision": true, "validation": true, "manual": true, "compensation": true,
	"state": true, "reference": true,
}
var allowedEdgeTypes = map[string]bool{
	"composed_of": true, "depends_on": true, "enables": true, "blocks": true,
	"validates": true, "compensates": true, "alternative_to": true,
	"requires_condition": true, "produces_state": true, "consumes_state": true,
}
var allowedEdgeBindings = map[string]bool{"hard": true, "soft": true}

func validateServicegraph(sg model.Servicegraph, path string, res *Result) {
	for _, field := range []struct{ name, value string }{{"id", sg.ID}, {"type", sg.Type}, {"name", sg.Name}, {"version", sg.Version}, {"status", sg.Status}, {"owner", sg.Owner}} {
		if strings.TrimSpace(field.value) == "" {
			res.add("SG_REQUIRED_FIELD", "error", "Servicegraph Pflichtfeld fehlt: "+field.name, path)
		}
	}
	if sg.Type != "servicegraph" {
		res.add("SG_TYPE_INVALID", "error", "type muss servicegraph sein", path)
	}
	if len(sg.Nodes) == 0 {
		res.add("SG_NO_NODES", "error", "Servicegraph hat keine Knoten", path)
	}
	if len(sg.Edges) == 0 {
		res.add("SG_NO_EDGES", "error", "Servicegraph hat keine Kanten", path)
	}

	nodeIDs := make(map[string]bool)
	hasServiceNode := false
	for _, n := range sg.Nodes {
		if nodeIDs[n.ID] {
			res.add("SG_NODE_ID_DUPLICATE", "error", "Doppelte Knoten-ID: "+n.ID, path)
		}
		nodeIDs[n.ID] = true
		if !allowedNodeTypes[n.Type] {
			res.add("SG_NODE_TYPE_INVALID", "error", "Ungueltiger Knotentyp: "+n.Type+" bei "+n.ID, path)
		}
		if n.Type == "service" {
			hasServiceNode = true
		}
	}
	if !hasServiceNode && len(sg.Nodes) > 0 {
		res.add("SG_NO_SERVICE_NODE", "warning", "Kein Knoten mit type=service vorhanden", path)
	}

	edgeIDs := make(map[string]bool)
	hasComposedOf := false
	// Track produces_state targets and consumes_state sources for balance check
	producedStates := make(map[string]bool)
	consumedStateSources := make(map[string]bool)
	for _, e := range sg.Edges {
		if edgeIDs[e.ID] {
			res.add("SG_EDGE_ID_DUPLICATE", "error", "Doppelte Kanten-ID: "+e.ID, path)
		}
		edgeIDs[e.ID] = true
		if !allowedEdgeTypes[e.Type] {
			res.add("SG_EDGE_TYPE_INVALID", "error", "Ungueltiger Kantentyp: "+e.Type+" bei "+e.ID, path)
		}
		if !allowedEdgeBindings[e.Binding] {
			res.add("SG_EDGE_BINDING_INVALID", "error", "Ungueltige Verbindlichkeit: "+e.Binding+" bei "+e.ID, path)
		}
		if !nodeIDs[e.Source] {
			res.add("SG_EDGE_SOURCE_MISSING", "error", "Quell-Knoten nicht gefunden: "+e.Source+" bei "+e.ID, path)
		}
		if !nodeIDs[e.Target] {
			res.add("SG_EDGE_TARGET_MISSING", "error", "Ziel-Knoten nicht gefunden: "+e.Target+" bei "+e.ID, path)
		}
		if e.Type == "composed_of" {
			hasComposedOf = true
		}
		if e.Type == "produces_state" {
			producedStates[e.Target] = true
		}
		if e.Type == "consumes_state" {
			// Source of consumes_state is the state node being consumed
			consumedStateSources[e.Source] = true
		}
	}
	if !hasComposedOf && len(sg.Edges) > 0 {
		res.add("SG_NO_COMPOSED_OF", "warning", "Keine composed_of Beziehung vorhanden", path)
	}

	// State balance check: every state node consumed must also be produced
	for stateNode := range consumedStateSources {
		if !producedStates[stateNode] {
			res.add("SG_STATE_UNBALANCED", "warning", "consumes_state Ziel ohne produces_state: "+stateNode, path)
		}
	}

	// Cycle detection via topological sort
	if len(sg.Nodes) > 0 && len(sg.Edges) > 0 {
		if _, err := graph.TopologicalOrder(sg); err != nil {
			res.add("SG_CYCLE_DETECTED", "error", "Zyklus in harten depends_on Kanten erkannt", path)
		}
	}

	// Rule scope check
	for _, rule := range sg.Rules {
		if rule.Scope != "model" && rule.Scope != "" {
			scopes := strings.Split(rule.Scope, ",")
			for _, s := range scopes {
				s = strings.TrimSpace(s)
				if s != "" && !nodeIDs[s] {
					res.add("SG_RULE_SCOPE_MISSING", "warning", "Regel "+rule.ID+" referenziert unbekannten Knoten: "+s, path)
				}
			}
		}
	}
}

func (r *Result) add(code, severity, message, path string) {
	r.Findings = append(r.Findings, Finding{Code: code, Severity: severity, Message: message, Path: path})
}

func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
