package namespace

import (
	"fmt"
	"strings"
)

// NamespaceView exposes canonical DNS-like namespaces alongside tree-oriented display metadata.
type NamespaceView struct {
	Canonical       string   `json:"canonical"`
	Namespace       string   `json:"namespace"`
	Labels          []string `json:"labels"`
	Label           string   `json:"label"`
	ParentCanonical string   `json:"parentCanonical"`
	Parts           []string `json:"parts"`
	TreeParts       []string `json:"treeParts"`
	TreePath        string   `json:"treePath"`
	DisplayPath     string   `json:"displayPath"`
	Leaf            string   `json:"leaf"`
}

func Canonical(name string) string { return strings.TrimSpace(name) }

func Parts(canonical string) []string {
	canonical = Canonical(canonical)
	if canonical == "" {
		return []string{}
	}
	return strings.Split(canonical, ".")
}

// Namespace returns the top-level DNS-like zone for a canonical name.
func Namespace(canonical string) string {
	parts := Parts(canonical)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// Labels returns labels from namespace inward, excluding the namespace itself.
func Labels(canonical string) []string {
	tree := TreeParts(canonical)
	if len(tree) <= 1 {
		return []string{}
	}
	labels := make([]string, len(tree)-1)
	copy(labels, tree[1:])
	return labels
}

// Label returns the current node label.
func Label(canonical string) string {
	parts := Parts(canonical)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

// ParentCanonical returns the canonical DNS name of the parent domain, or an empty string
// when the parent is the namespace pseudo-parent.
func ParentCanonical(canonical string) string {
	parts := Parts(canonical)
	if len(parts) <= 2 {
		return ""
	}
	return strings.Join(parts[1:], ".")
}

func TreeParts(canonical string) []string {
	parts := Parts(canonical)
	out := make([]string, len(parts))
	for i := range parts {
		out[i] = parts[len(parts)-1-i]
	}
	return out
}

func TreePath(canonical string) string { return strings.Join(TreeParts(canonical), "/") }

func DisplayPath(canonical string) string { return strings.Join(TreeParts(canonical), " / ") }

func Leaf(canonical string) string { return Label(canonical) }

func ParentTreePath(canonical string) string {
	parts := TreeParts(canonical)
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], " / ")
}

// ComposeCanonical derives a canonical DNS name from a namespace and labels ordered
// from namespace inward, e.g. ("com", "blumer", "identity") => "identity.blumer.com".
func ComposeCanonical(ns string, labels ...string) (string, error) {
	ns = strings.TrimSpace(strings.ToLower(ns))
	if !validLabel(ns) {
		return "", fmt.Errorf("invalid namespace: %s", ns)
	}
	clean := make([]string, 0, len(labels)+1)
	for _, label := range labels {
		label = strings.TrimSpace(strings.ToLower(label))
		if !validLabel(label) {
			return "", fmt.Errorf("invalid label: %s", label)
		}
		clean = append(clean, label)
	}
	parts := make([]string, 0, len(clean)+1)
	for i := len(clean) - 1; i >= 0; i-- {
		parts = append(parts, clean[i])
	}
	parts = append(parts, ns)
	return strings.Join(parts, "."), nil
}

// ComposeChildCanonical derives a child canonical name from a canonical parent domain
// or from a namespace pseudo-parent such as "cloud".
func ComposeChildCanonical(parentCanonical, childLabel string) (string, error) {
	parent := Canonical(parentCanonical)
	childLabel = strings.TrimSpace(strings.ToLower(childLabel))
	if !validLabel(childLabel) {
		return "", fmt.Errorf("invalid label: %s", childLabel)
	}
	if parent == "" || strings.ContainsAny(parent, `/\\`) || strings.Contains(parent, " ") || strings.HasPrefix(parent, ".") || strings.HasSuffix(parent, ".") {
		return "", fmt.Errorf("invalid parent: %s", parent)
	}
	for _, part := range strings.Split(strings.ToLower(parent), ".") {
		if !validLabel(part) {
			return "", fmt.Errorf("invalid parent: %s", parent)
		}
	}
	return childLabel + "." + parent, nil
}

func View(canonical string) NamespaceView {
	c := Canonical(canonical)
	label := Label(c)
	return NamespaceView{
		Canonical:       c,
		Namespace:       Namespace(c),
		Labels:          Labels(c),
		Label:           label,
		ParentCanonical: ParentCanonical(c),
		Parts:           Parts(c),
		TreeParts:       TreeParts(c),
		TreePath:        TreePath(c),
		DisplayPath:     DisplayPath(c),
		Leaf:            label,
	}
}

func validLabel(label string) bool {
	if label == "" || len(label) > 63 || strings.ContainsAny(label, `. /\\`) || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
		return false
	}
	for _, r := range label {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return false
	}
	return true
}
