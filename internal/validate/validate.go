package validate

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
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
	if _, err := os.Stat(filepath.Join(path, "cosmos.yaml")); err != nil {
		res.add("COSMOS_MISSING", "error", "cosmos.yaml fehlt", "cosmos.yaml")
	}
	if err := validateCatalogArtifacts(path, &res); err != nil {
		return res, err
	}
	if len(res.Findings) > 0 {
		res.Status = "failed"
	}
	return res, nil
}

func validateCatalogArtifacts(root string, res *Result) error {
	catalogDir := filepath.Join(root, "catalog")
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
	}
	if b.Type == "service_blueprint" && len(b.TargetSystems) == 0 && len(b.Providers) == 0 && len(b.Capabilities) == 0 && len(b.Actions) == 0 && len(b.QualityCriteria) == 0 {
		res.add("SERVICE_BLUEPRINT_CAPABILITY_EMPTY", "error", "Service Blueprint benoetigt mindestens Provider/Zielsystem, Capability, Action oder Quality Criterion", path)
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
