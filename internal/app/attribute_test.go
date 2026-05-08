package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

// setupAttrCosmos builds a cosmos with one product blueprint that has one
// attribute ("Konto-ID", text, required) and two rules (regex ^[UX], max_length 9).
func setupAttrCosmos(t *testing.T) (cosmosPath, blueprintID, attrID string) {
	t.Helper()
	p, productID, _ := createProductAndService(t)

	added, err := AddBlueprintAttribute(p, productID, "Konto-ID", "text", true)
	if err != nil {
		t.Fatal(err)
	}
	aID := added.Attributes[0].ID

	if _, err := AddAttributeRule(p, productID, aID, "Beginnt mit U oder X", "regex", "^[UX]"); err != nil {
		t.Fatal(err)
	}
	if _, err := AddAttributeRule(p, productID, aID, "Max. 9 Zeichen", "max_length", "9"); err != nil {
		t.Fatal(err)
	}
	return p, productID, aID
}

// makeInstance writes a product_instance YAML file for the given blueprint.
func makeInstance(t *testing.T, p, bpID, instID string) {
	t.Helper()
	dir := filepath.Join(storage.CatalogDir(p), "instances", "products")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	inst := model.Instance{
		ID:               instID,
		Type:             "product_instance",
		Name:             instID,
		BlueprintRef:     bpID,
		BlueprintVersion: "0.1.0",
		Status:           "draft",
		Owner:            "Team",
		ComplianceStatus: "unknown",
	}
	if err := CreateInstance(p, inst); err != nil {
		t.Fatalf("CreateInstance: %v", err)
	}
}

// ── AddBlueprintAttribute ─────────────────────────────────────────────────────

func TestAddBlueprintAttribute_Success(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	got, err := AddBlueprintAttribute(p, productID, "Konto-ID", "text", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Attributes) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(got.Attributes))
	}
	attr := got.Attributes[0]
	if attr.Label != "Konto-ID" {
		t.Fatalf("unexpected label: %s", attr.Label)
	}
	if attr.Type != "text" {
		t.Fatalf("expected type text, got %s", attr.Type)
	}
	if !attr.Required {
		t.Fatal("expected required=true")
	}
	if attr.ID == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestAddBlueprintAttribute_DefaultsToText(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	got, err := AddBlueprintAttribute(p, productID, "Bezeichnung", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Attributes[0].Type != "text" {
		t.Fatalf("expected default type text, got %s", got.Attributes[0].Type)
	}
}

func TestAddBlueprintAttribute_ServiceRef(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	got, err := AddBlueprintAttributeWithServiceRef(p, productID, "Provisionierungsservice", "service_ref", true, "identity.blumer.cloud/user-account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	attr := got.Attributes[0]
	if attr.Type != "service_ref" {
		t.Fatalf("expected type service_ref, got %s", attr.Type)
	}
	if attr.ServiceRef != "identity.blumer.cloud/user-account" {
		t.Fatalf("unexpected service_ref: %s", attr.ServiceRef)
	}
}

func TestAddBlueprintAttribute_ServiceRefOnlyForServiceRefType(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	_, err := AddBlueprintAttributeWithServiceRef(p, productID, "Feld", "text", false, "identity.blumer.cloud/user-account")
	if err == nil {
		t.Fatal("expected error for service_ref on text attribute")
	}
}

func TestAddBlueprintAttribute_EmptyLabel(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	_, err := AddBlueprintAttribute(p, productID, "", "text", false)
	if err == nil {
		t.Fatal("expected error for empty label")
	}
}

func TestAddBlueprintAttribute_InvalidType(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	_, err := AddBlueprintAttribute(p, productID, "Feld", "invalid_type", false)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestAddBlueprintAttribute_BlueprintNotFound(t *testing.T) {
	p := createTestCosmosWithCatalog(t)
	_, err := AddBlueprintAttribute(p, "PB-NONEXISTENT", "Label", "text", false)
	if err == nil {
		t.Fatal("expected error for non-existent blueprint")
	}
}

// ── DeleteBlueprintAttribute ──────────────────────────────────────────────────

func TestDeleteBlueprintAttribute_Success(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	added, _ := AddBlueprintAttribute(p, productID, "Konto-ID", "text", true)
	attrID := added.Attributes[0].ID

	got, err := DeleteBlueprintAttribute(p, productID, attrID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Attributes) != 0 {
		t.Fatalf("expected 0 attributes after deletion, got %d", len(got.Attributes))
	}
}

func TestDeleteBlueprintAttribute_NotFound(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	_, err := DeleteBlueprintAttribute(p, productID, "attr-nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent attribute")
	}
}

// ── AddAttributeRule ──────────────────────────────────────────────────────────

func TestAddAttributeRule_Regex(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	added, _ := AddBlueprintAttribute(p, productID, "Konto-ID", "text", true)
	attrID := added.Attributes[0].ID

	got, err := AddAttributeRule(p, productID, attrID, "Beginnt mit U oder X", "regex", "^[UX]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Attributes[0].Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(got.Attributes[0].Rules))
	}
	rule := got.Attributes[0].Rules[0]
	if rule.Type != "regex" || rule.Value != "^[UX]" {
		t.Fatalf("unexpected rule: %+v", rule)
	}
}

func TestAddAttributeRule_AllAutomaticTypes(t *testing.T) {
	cases := []struct {
		ruleType string
		value    string
	}{
		{"max_length", "9"},
		{"min_length", "1"},
		{"starts_with", "U"},
		{"ends_with", "X"},
		{"one_of", "A,B,C"},
		{"manual", ""},
	}
	for _, tc := range cases {
		p, productID, _ := createProductAndService(t)
		added, _ := AddBlueprintAttribute(p, productID, "Feld", "text", false)
		attrID := added.Attributes[0].ID
		_, err := AddAttributeRule(p, productID, attrID, "label", tc.ruleType, tc.value)
		if err != nil {
			t.Errorf("rule type %s: unexpected error: %v", tc.ruleType, err)
		}
	}
}

func TestAddAttributeRule_InvalidRegex(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	added, _ := AddBlueprintAttribute(p, productID, "Konto-ID", "text", true)
	attrID := added.Attributes[0].ID

	_, err := AddAttributeRule(p, productID, attrID, "bad regex", "regex", "[invalid")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestAddAttributeRule_InvalidType(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	added, _ := AddBlueprintAttribute(p, productID, "Konto-ID", "text", true)
	attrID := added.Attributes[0].ID

	_, err := AddAttributeRule(p, productID, attrID, "bad", "not_a_type", "x")
	if err == nil {
		t.Fatal("expected error for invalid rule type")
	}
}

func TestAddAttributeRule_EmptyLabel(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	added, _ := AddBlueprintAttribute(p, productID, "Konto-ID", "text", true)
	attrID := added.Attributes[0].ID

	_, err := AddAttributeRule(p, productID, attrID, "", "regex", "^[UX]")
	if err == nil {
		t.Fatal("expected error for empty label")
	}
}

func TestAddAttributeRule_AttributeNotFound(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	_, err := AddAttributeRule(p, productID, "attr-nonexistent", "Label", "regex", "^[UX]")
	if err == nil {
		t.Fatal("expected error for non-existent attribute")
	}
}

// ── DeleteAttributeRule ───────────────────────────────────────────────────────

func TestDeleteAttributeRule_Success(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	added, _ := AddBlueprintAttribute(p, productID, "Konto-ID", "text", true)
	attrID := added.Attributes[0].ID
	withRule, _ := AddAttributeRule(p, productID, attrID, "Pattern", "regex", "^[UX]")
	ruleID := withRule.Attributes[0].Rules[0].ID

	got, err := DeleteAttributeRule(p, productID, attrID, ruleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Attributes[0].Rules) != 0 {
		t.Fatalf("expected 0 rules after deletion, got %d", len(got.Attributes[0].Rules))
	}
}

func TestDeleteAttributeRule_AttrNotFound(t *testing.T) {
	p, productID, _ := createProductAndService(t)
	_, err := DeleteAttributeRule(p, productID, "attr-nonexistent", "rule-x")
	if err == nil {
		t.Fatal("expected error for non-existent attribute")
	}
}

// ── evalRule ──────────────────────────────────────────────────────────────────

func TestEvalRule_AllTypes(t *testing.T) {
	cases := []struct {
		ruleType string
		value    string
		ruleVal  string
		want     string
	}{
		{"regex", "U123", "^[UX]", "pass"},
		{"regex", "A123", "^[UX]", "fail"},
		{"max_length", "hello", "10", "pass"},
		{"max_length", "hello world!", "10", "fail"},
		{"min_length", "hello", "3", "pass"},
		{"min_length", "hi", "3", "fail"},
		{"starts_with", "Ufoo", "U", "pass"},
		{"starts_with", "Xfoo", "U", "fail"},
		{"ends_with", "fooX", "X", "pass"},
		{"ends_with", "fooY", "X", "fail"},
		{"one_of", "A", "A,B,C", "pass"},
		{"one_of", "D", "A,B,C", "fail"},
		{"manual", "anything", "", "manual"},
	}
	for _, tc := range cases {
		rule := AttributeRuleDTO{ID: "r1", Label: "test", Type: tc.ruleType, Value: tc.ruleVal}
		got := evalRule(rule, tc.value)
		if got.Status != tc.want {
			t.Errorf("type=%s value=%q ruleVal=%q: want %s got %s (msg:%s)",
				tc.ruleType, tc.value, tc.ruleVal, tc.want, got.Status, got.Message)
		}
	}
}

// ── ValidateInstanceAttributes ────────────────────────────────────────────────

func TestValidateInstanceAttributes_AllPass(t *testing.T) {
	p, bpID, attrID := setupAttrCosmos(t)
	makeInstance(t, p, bpID, "PI-VALID-001")
	if _, err := SetInstanceAttributeValues(p, "PI-VALID-001", map[string]string{attrID: "U12345678"}); err != nil {
		t.Fatalf("set values: %v", err)
	}

	result, err := ValidateInstanceAttributes(p, "PI-VALID-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "valid" {
		t.Fatalf("expected valid, got %s: %+v", result.Status, result)
	}
	for _, ar := range result.Attributes {
		for _, rr := range ar.Rules {
			if rr.Status == "fail" {
				t.Fatalf("rule %q failed: %s", rr.Label, rr.Message)
			}
		}
	}
}

func TestValidateInstanceAttributes_Fail_WrongPrefix(t *testing.T) {
	p, bpID, attrID := setupAttrCosmos(t)
	makeInstance(t, p, bpID, "PI-FAIL-PREFIX")
	if _, err := SetInstanceAttributeValues(p, "PI-FAIL-PREFIX", map[string]string{attrID: "A12345"}); err != nil {
		t.Fatalf("set values: %v", err)
	}

	result, err := ValidateInstanceAttributes(p, "PI-FAIL-PREFIX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "invalid" {
		t.Fatalf("expected invalid for wrong prefix, got %s", result.Status)
	}
}

func TestValidateInstanceAttributes_Fail_TooLong(t *testing.T) {
	p, bpID, attrID := setupAttrCosmos(t)
	makeInstance(t, p, bpID, "PI-FAIL-LEN")
	// "U" + 9 digits = 10 chars, exceeds max_length 9
	if _, err := SetInstanceAttributeValues(p, "PI-FAIL-LEN", map[string]string{attrID: "U123456789"}); err != nil {
		t.Fatalf("set values: %v", err)
	}

	result, err := ValidateInstanceAttributes(p, "PI-FAIL-LEN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "invalid" {
		t.Fatalf("expected invalid for too-long value, got %s", result.Status)
	}
}

func TestValidateInstanceAttributes_Missing_Required(t *testing.T) {
	p, bpID, _ := setupAttrCosmos(t)
	makeInstance(t, p, bpID, "PI-MISS-001")
	// no attribute values set

	result, err := ValidateInstanceAttributes(p, "PI-MISS-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "invalid" {
		t.Fatalf("expected invalid for missing required attribute, got %s", result.Status)
	}
	if result.Attributes[0].Status != "missing" {
		t.Fatalf("expected attribute status missing, got %s", result.Attributes[0].Status)
	}
}

func TestValidateInstanceAttributes_NoBlueprintAttributes(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	result, err := ValidateInstanceAttributes(p, "PI-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "valid" {
		t.Fatalf("expected valid for blueprint with no attributes, got %s", result.Status)
	}
}

// ── SetInstanceAttributeValues ────────────────────────────────────────────────

func TestSetInstanceAttributeValues_Success(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	got, err := SetInstanceAttributeValues(p, "PI-001", map[string]string{"account_id": "U12345"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AttributeValues["account_id"] != "U12345" {
		t.Fatalf("expected U12345, got %s", got.AttributeValues["account_id"])
	}
}

func TestSetInstanceAttributeValues_Merge(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	if _, err := SetInstanceAttributeValues(p, "PI-001", map[string]string{"a": "1"}); err != nil {
		t.Fatal(err)
	}
	got, err := SetInstanceAttributeValues(p, "PI-001", map[string]string{"b": "2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AttributeValues["a"] != "1" || got.AttributeValues["b"] != "2" {
		t.Fatalf("expected merged values: %+v", got.AttributeValues)
	}
}

func TestSetInstanceAttributeValues_NotFound(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	_, err := SetInstanceAttributeValues(p, "NO-SUCH-INST", map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error for non-existent instance")
	}
}

// suppress unused import
var _ = os.MkdirAll
var _ = filepath.Join

func TestDeleteBlueprintAttribute_RemovesRequirementRefs(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	withAttr, err := AddBlueprintAttribute(p, productID, "Region", "text", true)
	if err != nil {
		t.Fatalf("add attribute: %v", err)
	}
	attrID := withAttr.Attributes[0].ID
	if _, err := AddBlueprintRequirementWithAttributeRefs(p, productID, "Region muss gesetzt sein", []string{attrID}); err != nil {
		t.Fatalf("add requirement: %v", err)
	}

	got, err := DeleteBlueprintAttribute(p, productID, attrID)
	if err != nil {
		t.Fatalf("delete attribute: %v", err)
	}
	if len(got.Requirements) != 1 {
		t.Fatalf("expected requirement to stay, got %d", len(got.Requirements))
	}
	if len(got.Requirements[0].AttributeRefs) != 0 {
		t.Fatalf("expected attribute refs to be removed, got %#v", got.Requirements[0].AttributeRefs)
	}
}
