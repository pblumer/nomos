package namespace

import (
	"fmt"
	"path"
	"strings"
)

// DomainIdentity centralizes the canonical DNS-like name and tree-ordered storage identity.
type DomainIdentity struct {
	Namespace       string   `json:"namespace"`
	Labels          []string `json:"labels"`
	Label           string   `json:"label"`
	TreePath        string   `json:"treePath"`
	GitPath         string   `json:"gitPath"`
	CanonicalName   string   `json:"canonicalName"`
	ParentTreePath  string   `json:"parentTreePath"`
	ParentCanonical string   `json:"parentCanonical"`
}

// NamespaceView exposes canonical DNS-like namespaces alongside tree-oriented display metadata.
type NamespaceView struct {
	Canonical       string   `json:"canonical"`
	CanonicalName   string   `json:"canonicalName"`
	Namespace       string   `json:"namespace"`
	Labels          []string `json:"labels"`
	Label           string   `json:"label"`
	ParentCanonical string   `json:"parentCanonical"`
	Parts           []string `json:"parts"`
	TreeParts       []string `json:"treeParts"`
	TreePath        string   `json:"treePath"`
	GitPath         string   `json:"gitPath"`
	ParentTreePath  string   `json:"parentTreePath"`
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

func TreePath(canonical string) string { return "/" + strings.Join(TreeParts(canonical), "/") }

func DisplayPath(canonical string) string { return strings.Join(TreeParts(canonical), " / ") }

func Leaf(canonical string) string { return Label(canonical) }

func ParentTreePath(canonical string) string {
	parts := TreeParts(canonical)
	if len(parts) <= 1 {
		return ""
	}
	return "/" + strings.Join(parts[:len(parts)-1], "/")
}

func TreePathToCanonical(treePath string) (string, error) {
	parts := cleanTreePathParts(treePath)
	if len(parts) < 2 {
		return "", fmt.Errorf("tree path must contain namespace and at least one label: %s", treePath)
	}
	return ComposeCanonical(parts[0], parts[1:]...)
}

func CanonicalToTreePath(canonical string) (string, error) {
	if _, err := validateCanonical(canonical); err != nil {
		return "", err
	}
	return TreePath(canonical), nil
}

func TreePathToGitPath(treePath string) (string, error) {
	parts := cleanTreePathParts(treePath)
	if len(parts) < 2 {
		return "", fmt.Errorf("tree path must contain namespace and at least one label: %s", treePath)
	}
	for _, part := range parts {
		if !validLabel(strings.ToLower(part)) || part != strings.ToLower(part) {
			return "", fmt.Errorf("invalid tree path label: %s", part)
		}
	}
	return path.Join(append([]string{".nomos", "domains"}, parts...)...) + "/domain.yaml", nil
}

func Identity(canonical string) (DomainIdentity, error) {
	c, err := validateCanonical(canonical)
	if err != nil {
		return DomainIdentity{}, err
	}
	gitPath, err := TreePathToGitPath(TreePath(c))
	if err != nil {
		return DomainIdentity{}, err
	}
	return DomainIdentity{
		Namespace:       Namespace(c),
		Labels:          Labels(c),
		Label:           Label(c),
		TreePath:        TreePath(c),
		GitPath:         gitPath,
		CanonicalName:   c,
		ParentTreePath:  ParentTreePath(c),
		ParentCanonical: ParentCanonical(c),
	}, nil
}

func cleanTreePathParts(treePath string) []string {
	trimmed := strings.Trim(strings.TrimSpace(treePath), "/")
	if trimmed == "" {
		return []string{}
	}
	raw := strings.Split(trimmed, "/")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
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
	gitPath, _ := TreePathToGitPath(TreePath(c))
	return NamespaceView{
		Canonical:       c,
		CanonicalName:   c,
		Namespace:       Namespace(c),
		Labels:          Labels(c),
		Label:           label,
		ParentCanonical: ParentCanonical(c),
		Parts:           Parts(c),
		TreeParts:       TreeParts(c),
		TreePath:        TreePath(c),
		GitPath:         gitPath,
		ParentTreePath:  ParentTreePath(c),
		DisplayPath:     DisplayPath(c),
		Leaf:            label,
	}
}

func validateCanonical(canonical string) (string, error) {
	c := Canonical(strings.ToLower(canonical))
	parts := Parts(c)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid canonical name: %s", canonical)
	}
	for _, part := range parts {
		if !validLabel(part) {
			return "", fmt.Errorf("invalid canonical name: %s", canonical)
		}
	}
	return c, nil
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
