package app

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
)

// ── Attribute CRUD ────────────────────────────────────────────────────────────

func AddBlueprintAttribute(path, bpID, label, attrType string, required bool) (BlueprintDTO, error) {
	if strings.TrimSpace(label) == "" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "attribute label is required", http.StatusBadRequest, nil)
	}
	validTypes := map[string]bool{"text": true, "number": true, "boolean": true, "date": true, "enum": true}
	if attrType == "" {
		attrType = "text"
	}
	if !validTypes[attrType] {
		return BlueprintDTO{}, Error(CodeInvalidInput, "invalid attribute type: "+attrType, http.StatusBadRequest, nil)
	}
	bp, err := GetBlueprint(path, bpID)
	if err != nil {
		return BlueprintDTO{}, err
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	attrID := fmt.Sprintf("attr-%d", time.Now().UnixNano())
	raw.Attributes = append(raw.Attributes, model.BlueprintAttribute{
		ID:       attrID,
		Label:    label,
		Type:     attrType,
		Required: required,
	})
	if err := fsx.WriteYAML(bp.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, bpID)
}

func DeleteBlueprintAttribute(path, bpID, attrID string) (BlueprintDTO, error) {
	bp, err := GetBlueprint(path, bpID)
	if err != nil {
		return BlueprintDTO{}, err
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	n := len(raw.Attributes)
	filtered := raw.Attributes[:0]
	for _, a := range raw.Attributes {
		if a.ID != attrID {
			filtered = append(filtered, a)
		}
	}
	if len(filtered) == n {
		return BlueprintDTO{}, Error(CodeInvalidInput, "attribute not found: "+attrID, http.StatusNotFound, nil)
	}
	raw.Attributes = filtered
	if err := fsx.WriteYAML(bp.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, bpID)
}

// ── Rule CRUD ─────────────────────────────────────────────────────────────────

var validRuleTypes = map[string]bool{
	"regex":       true,
	"max_length":  true,
	"min_length":  true,
	"starts_with": true,
	"ends_with":   true,
	"one_of":      true,
	"manual":      true,
}

func AddAttributeRule(path, bpID, attrID, label, ruleType, value string) (BlueprintDTO, error) {
	if strings.TrimSpace(label) == "" {
		return BlueprintDTO{}, Error(CodeInvalidInput, "rule label is required", http.StatusBadRequest, nil)
	}
	if !validRuleTypes[ruleType] {
		return BlueprintDTO{}, Error(CodeInvalidInput, "invalid rule type: "+ruleType, http.StatusBadRequest, nil)
	}
	if ruleType == "regex" {
		if _, err := regexp.Compile(value); err != nil {
			return BlueprintDTO{}, Error(CodeInvalidInput, "invalid regex: "+err.Error(), http.StatusBadRequest, err)
		}
	}
	bp, err := GetBlueprint(path, bpID)
	if err != nil {
		return BlueprintDTO{}, err
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	found := false
	ruleID := fmt.Sprintf("rule-%d", time.Now().UnixNano())
	for i := range raw.Attributes {
		if raw.Attributes[i].ID == attrID {
			raw.Attributes[i].Rules = append(raw.Attributes[i].Rules, model.AttributeRule{
				ID:    ruleID,
				Label: label,
				Type:  ruleType,
				Value: value,
			})
			found = true
			break
		}
	}
	if !found {
		return BlueprintDTO{}, Error(CodeInvalidInput, "attribute not found: "+attrID, http.StatusNotFound, nil)
	}
	if err := fsx.WriteYAML(bp.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, bpID)
}

func DeleteAttributeRule(path, bpID, attrID, ruleID string) (BlueprintDTO, error) {
	bp, err := GetBlueprint(path, bpID)
	if err != nil {
		return BlueprintDTO{}, err
	}
	var raw model.Blueprint
	if err := fsx.ReadYAML(bp.Path, &raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to read blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	attrFound := false
	for i := range raw.Attributes {
		if raw.Attributes[i].ID == attrID {
			attrFound = true
			filtered := raw.Attributes[i].Rules[:0]
			for _, r := range raw.Attributes[i].Rules {
				if r.ID != ruleID {
					filtered = append(filtered, r)
				}
			}
			raw.Attributes[i].Rules = filtered
			break
		}
	}
	if !attrFound {
		return BlueprintDTO{}, Error(CodeInvalidInput, "attribute not found: "+attrID, http.StatusNotFound, nil)
	}
	if err := fsx.WriteYAML(bp.Path, raw); err != nil {
		return BlueprintDTO{}, Error(CodeInternalError, "Failed to write blueprint: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetBlueprint(path, bpID)
}

// ── Instance attribute values ─────────────────────────────────────────────────

func SetInstanceAttributeValues(path, instanceID string, values map[string]string) (InstanceDTO, error) {
	inst, err := GetInstance(path, instanceID)
	if err != nil {
		return InstanceDTO{}, err
	}
	var raw model.Instance
	if err := fsx.ReadYAML(inst.Path, &raw); err != nil {
		return InstanceDTO{}, Error(CodeInternalError, "Failed to read instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	if raw.AttributeValues == nil {
		raw.AttributeValues = make(map[string]string)
	}
	for k, v := range values {
		raw.AttributeValues[k] = v
	}
	if err := fsx.WriteYAML(inst.Path, raw); err != nil {
		return InstanceDTO{}, Error(CodeInternalError, "Failed to write instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetInstance(path, instanceID)
}

// ── Validation ────────────────────────────────────────────────────────────────

func ValidateInstanceAttributes(path, instanceID string) (AttributeValidationDTO, error) {
	inst, err := GetInstance(path, instanceID)
	if err != nil {
		return AttributeValidationDTO{}, err
	}
	bp, err := GetBlueprint(path, inst.BlueprintRef)
	if err != nil {
		return AttributeValidationDTO{}, Error(CodeInternalError, "blueprint not found: "+inst.BlueprintRef, http.StatusNotFound, err)
	}
	if len(bp.Attributes) == 0 {
		return AttributeValidationDTO{Status: "valid", Attributes: []AttrValidationResult{}}, nil
	}
	results := make([]AttrValidationResult, 0, len(bp.Attributes))
	overallOK := true
	for _, attr := range bp.Attributes {
		value, present := inst.AttributeValues[attr.ID]
		result := validateAttributeValue(attr, value, present)
		results = append(results, result)
		if result.Status != "valid" {
			overallOK = false
		}
	}
	status := "valid"
	if !overallOK {
		status = "invalid"
	}
	return AttributeValidationDTO{Status: status, Attributes: results}, nil
}

func validateAttributeValue(attr BlueprintAttributeDTO, value string, present bool) AttrValidationResult {
	result := AttrValidationResult{
		AttributeID: attr.ID,
		Label:       attr.Label,
		Value:       value,
	}
	if !present || value == "" {
		if attr.Required {
			result.Status = "missing"
			return result
		}
		result.Status = "valid"
		return result
	}
	ruleResults := make([]RuleResultDTO, 0, len(attr.Rules))
	allPass := true
	for _, rule := range attr.Rules {
		rr := evalRule(rule, value)
		ruleResults = append(ruleResults, rr)
		if rr.Status == "fail" {
			allPass = false
		}
	}
	result.Rules = ruleResults
	if allPass {
		result.Status = "valid"
	} else {
		result.Status = "invalid"
	}
	return result
}

func evalRule(rule AttributeRuleDTO, value string) RuleResultDTO {
	rr := RuleResultDTO{RuleID: rule.ID, Label: rule.Label, Type: rule.Type}
	switch rule.Type {
	case "regex":
		re, err := regexp.Compile(rule.Value)
		if err != nil {
			rr.Status = "fail"
			rr.Message = "invalid regex: " + err.Error()
		} else if re.MatchString(value) {
			rr.Status = "pass"
		} else {
			rr.Status = "fail"
			rr.Message = fmt.Sprintf("value %q does not match pattern %q", value, rule.Value)
		}
	case "max_length":
		max, _ := strconv.Atoi(rule.Value)
		if len(value) <= max {
			rr.Status = "pass"
		} else {
			rr.Status = "fail"
			rr.Message = fmt.Sprintf("length %d exceeds maximum %d", len(value), max)
		}
	case "min_length":
		min, _ := strconv.Atoi(rule.Value)
		if len(value) >= min {
			rr.Status = "pass"
		} else {
			rr.Status = "fail"
			rr.Message = fmt.Sprintf("length %d is below minimum %d", len(value), min)
		}
	case "starts_with":
		if strings.HasPrefix(value, rule.Value) {
			rr.Status = "pass"
		} else {
			rr.Status = "fail"
			rr.Message = fmt.Sprintf("value must start with %q", rule.Value)
		}
	case "ends_with":
		if strings.HasSuffix(value, rule.Value) {
			rr.Status = "pass"
		} else {
			rr.Status = "fail"
			rr.Message = fmt.Sprintf("value must end with %q", rule.Value)
		}
	case "one_of":
		allowed := strings.Split(rule.Value, ",")
		for _, a := range allowed {
			if strings.TrimSpace(a) == value {
				rr.Status = "pass"
				return rr
			}
		}
		rr.Status = "fail"
		rr.Message = "value must be one of: " + rule.Value
	case "manual":
		rr.Status = "manual"
		rr.Message = "requires manual verification"
	default:
		rr.Status = "manual"
		rr.Message = "unknown rule type: " + rule.Type
	}
	return rr
}
