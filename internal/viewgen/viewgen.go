// Package viewgen generates default form-js view schemas (ADR-0024) from a
// type's properties. It is a leaf package (stdlib only) so both the repo seeder
// and the app layer can produce the same starter forms without an import cycle.
//
// Each type ships with four standard views following the
// <typeID>_<variant>.frm convention:
//
//	new   — capture form for a new instance (editable)
//	edit  — edit form for an existing instance (editable)
//	list  — table/list overview of the type's key fields (read-only)
//	short — compact single-line summary of the type (read-only)
package viewgen

// Variant identifies one of the four standard view forms a type ships with.
type Variant string

const (
	VariantNew   Variant = "new"
	VariantEdit  Variant = "edit"
	VariantList  Variant = "list"
	VariantShort Variant = "short"
)

// Variants is the canonical ordered set of standard view variants.
var Variants = []Variant{VariantNew, VariantEdit, VariantList, VariantShort}

// IsVariant reports whether s names a known standard view variant.
func IsVariant(s string) bool {
	switch Variant(s) {
	case VariantNew, VariantEdit, VariantList, VariantShort:
		return true
	}
	return false
}

// FormName returns the .frm filename for a type's view variant, e.g.
// FormName("decision", VariantEdit) -> "decision_edit.frm".
func FormName(typeID string, v Variant) string { return typeID + "_" + string(v) + ".frm" }

// Prop is the minimal property description the generator needs. It mirrors the
// shared subset of model.TypeProperty and the repo seeder's typeProp.
type Prop struct {
	Name        string
	Type        string
	Label       string
	Required    bool
	Description string
}

// Schema builds the form-js schema (ADR-0024 engine-native content) for the
// given variant from a type's label and properties. The returned map is ready
// to store under a View envelope's `schema` field.
func Schema(typeID, label string, props []Prop, v Variant) map[string]any {
	if label == "" {
		label = typeID
	}
	switch v {
	case VariantEdit:
		return editableSchema(label+" bearbeiten", props)
	case VariantList:
		return readonlySchema(label+" – Liste", listProps(props), "Felder dieser Listen-/Tabellenansicht. Eine Zeile je Eintrag.")
	case VariantShort:
		return readonlySchema(label, shortProps(props), "Kompakte Kurzansicht eines Eintrags.")
	default: // VariantNew
		return editableSchema(label+" erfassen", props)
	}
}

func editableSchema(title string, props []Prop) map[string]any {
	return schemaWith(title, props, false)
}

func readonlySchema(title string, props []Prop, hint string) map[string]any {
	s := schemaWith(title, props, true)
	if hint != "" {
		comps := s["components"].([]map[string]any)
		// Insert the hint right after the title text component.
		note := map[string]any{"type": "text", "text": "*" + hint + "*"}
		comps = append(comps[:1], append([]map[string]any{note}, comps[1:]...)...)
		s["components"] = comps
	}
	return s
}

func schemaWith(title string, props []Prop, readonly bool) map[string]any {
	comps := []map[string]any{{"type": "text", "text": "## " + title}}
	for _, p := range props {
		if p.Name == "" {
			continue
		}
		comps = append(comps, component(p, readonly))
	}
	return map[string]any{"type": "default", "components": comps}
}

// component maps one property to a form-js field component. readonly fields are
// rendered disabled so list/short views stay non-editable previews.
func component(p Prop, readonly bool) map[string]any {
	label := p.Label
	if label == "" {
		label = p.Name
	}
	c := map[string]any{"key": p.Name, "label": label}
	if p.Description != "" {
		c["description"] = p.Description
	}
	switch p.Type {
	case "number", "int", "integer", "float", "double":
		c["type"] = "number"
	case "boolean", "bool":
		c["type"] = "checkbox"
	case "text":
		c["type"] = "textarea"
	case "list":
		c["type"] = "textarea"
		if p.Description == "" {
			c["description"] = "Ein Eintrag pro Zeile."
		}
	default:
		c["type"] = "textfield"
	}
	if readonly {
		c["disabled"] = true
	} else if p.Required {
		c["validate"] = map[string]any{"required": true}
	}
	return c
}

// listProps returns the properties shown as columns in a list view: all named
// properties, since the list is a flat overview.
func listProps(props []Prop) []Prop {
	out := make([]Prop, 0, len(props))
	for _, p := range props {
		if p.Name != "" {
			out = append(out, p)
		}
	}
	return out
}

// shortProps returns the identifying properties for a compact summary: the
// required ones, or the first property when none are required.
func shortProps(props []Prop) []Prop {
	var out []Prop
	for _, p := range props {
		if p.Name != "" && p.Required {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		for _, p := range props {
			if p.Name != "" {
				out = append(out, p)
				break
			}
		}
	}
	return out
}
