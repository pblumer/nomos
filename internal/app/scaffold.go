package app

import (
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

// RequiredFieldScaffold inspects a YAML artifact's content and returns a snippet
// of the required fields it is still missing, each as a placeholder template the
// user can fill in. It never modifies or deletes existing content — the caller
// appends the snippet and the user reviews it before saving. Only artifact types
// with a known required schema are handled; others yield an empty result.
func RequiredFieldScaffold(content string) (string, []string, error) {
	var m map[string]any
	if err := yaml.Unmarshal([]byte(content), &m); err != nil {
		return "", nil, Error(CodeInvalidInput, "content is not valid YAML: "+err.Error(), http.StatusUnprocessableEntity, err)
	}
	if m == nil {
		m = map[string]any{}
	}
	typ, _ := m["type"].(string)

	// hasScalar reports whether key holds a non-empty string.
	hasScalar := func(key string) bool {
		s, ok := m[key].(string)
		return ok && strings.TrimSpace(s) != ""
	}
	// hasList reports whether key holds a non-empty sequence.
	hasList := func(key string) bool {
		v, ok := m[key]
		if !ok {
			return false
		}
		s, ok := v.([]any)
		return ok && len(s) > 0
	}
	hasNested := func(parent, child string) bool {
		p, ok := m[parent].(map[string]any)
		if !ok {
			return false
		}
		s, ok := p[child].([]any)
		return ok && len(s) > 0
	}

	var (
		missing []string
		b       strings.Builder
	)
	add := func(key, template string) {
		missing = append(missing, key)
		b.WriteString(template)
	}

	switch typ {
	case "product_blueprint", "service_blueprint":
		for _, f := range []struct{ key, tmpl string }{
			{"id", "id: PB-TODO\n"},
			{"name", "name: TODO\n"},
			{"version", "version: 0.1.0\n"},
			{"status", "status: draft\n"},
			{"owner", "owner: TODO\n"},
		} {
			if !hasScalar(f.key) {
				add(f.key, f.tmpl)
			}
		}
		if typ == "product_blueprint" {
			if !hasList("required_inputs") {
				add("required_inputs", "required_inputs:\n  - TODO   # ID eines erwarteten Eingabewerts\n")
			}
			if !hasList("required_service_blueprints") && !hasNested("fulfillment", "required_services") {
				add("required_service_blueprints", "required_service_blueprints:\n  - SB-TODO   # oder fulfillment.required_services angeben\n")
			}
		}
	default:
		// No known required schema for this type — nothing to scaffold.
		return "", []string{}, nil
	}

	if len(missing) == 0 {
		return "", []string{}, nil
	}
	snippet := "# --- Fehlende Pflichtfelder (bitte ausfüllen) ---\n" + b.String()
	return snippet, missing, nil
}
