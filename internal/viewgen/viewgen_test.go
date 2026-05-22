package viewgen

import "testing"

func TestFormName(t *testing.T) {
	cases := map[Variant]string{
		VariantNew:   "decision_new.frm",
		VariantEdit:  "decision_edit.frm",
		VariantList:  "decision_list.frm",
		VariantShort: "decision_short.frm",
	}
	for v, want := range cases {
		if got := FormName("decision", v); got != want {
			t.Errorf("FormName(decision, %s) = %q, want %q", v, got, want)
		}
	}
}

func TestIsVariant(t *testing.T) {
	for _, v := range Variants {
		if !IsVariant(string(v)) {
			t.Errorf("IsVariant(%q) = false, want true", v)
		}
	}
	if IsVariant("bogus") {
		t.Error("IsVariant(bogus) = true, want false")
	}
}

func TestSchemaEditableFieldsAndRequired(t *testing.T) {
	props := []Prop{
		{Name: "name", Type: "string", Required: true},
		{Name: "count", Type: "number"},
		{Name: "active", Type: "boolean"},
	}
	s := Schema("decision", "Decision", props, VariantNew)
	comps, ok := s["components"].([]map[string]any)
	if !ok {
		t.Fatalf("components has unexpected type %T", s["components"])
	}
	// First component is the intro title text; the rest are fields.
	if comps[0]["type"] != "text" {
		t.Errorf("expected leading text component, got %v", comps[0]["type"])
	}
	byKey := map[string]map[string]any{}
	for _, c := range comps {
		if k, _ := c["key"].(string); k != "" {
			byKey[k] = c
		}
	}
	if byKey["name"]["type"] != "textfield" {
		t.Errorf("name field type = %v, want textfield", byKey["name"]["type"])
	}
	if byKey["name"]["validate"] == nil {
		t.Error("required field name should carry validate.required")
	}
	if byKey["count"]["type"] != "number" {
		t.Errorf("count field type = %v, want number", byKey["count"]["type"])
	}
	if byKey["active"]["type"] != "checkbox" {
		t.Errorf("active field type = %v, want checkbox", byKey["active"]["type"])
	}
}

func TestSchemaListAndShortReadonly(t *testing.T) {
	props := []Prop{
		{Name: "name", Type: "string", Required: true},
		{Name: "note", Type: "text"},
	}
	for _, v := range []Variant{VariantList, VariantShort} {
		s := Schema("decision", "Decision", props, v)
		comps := s["components"].([]map[string]any)
		for _, c := range comps {
			if _, isField := c["key"]; !isField {
				continue
			}
			if c["disabled"] != true {
				t.Errorf("variant %s field %v should be disabled (read-only)", v, c["key"])
			}
		}
	}
	// short keeps only the required field; list keeps all.
	short := Schema("decision", "Decision", props, VariantShort)
	if got := countFields(short); got != 1 {
		t.Errorf("short field count = %d, want 1 (required only)", got)
	}
	list := Schema("decision", "Decision", props, VariantList)
	if got := countFields(list); got != 2 {
		t.Errorf("list field count = %d, want 2 (all)", got)
	}
}

func countFields(s map[string]any) int {
	n := 0
	for _, c := range s["components"].([]map[string]any) {
		if _, ok := c["key"]; ok {
			n++
		}
	}
	return n
}
