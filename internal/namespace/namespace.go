package namespace

import "strings"

// NamespaceView exposes canonical DNS-like namespaces alongside tree-oriented display metadata.
type NamespaceView struct {
	Canonical   string   `json:"canonical"`
	Parts       []string `json:"parts"`
	TreeParts   []string `json:"treeParts"`
	TreePath    string   `json:"treePath"`
	DisplayPath string   `json:"displayPath"`
	Leaf        string   `json:"leaf"`
}

func Canonical(name string) string { return strings.TrimSpace(name) }

func Parts(canonical string) []string {
	canonical = Canonical(canonical)
	if canonical == "" {
		return []string{}
	}
	return strings.Split(canonical, ".")
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

func Leaf(canonical string) string {
	parts := Parts(canonical)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func ParentTreePath(canonical string) string {
	parts := TreeParts(canonical)
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], " / ")
}

func View(canonical string) NamespaceView {
	c := Canonical(canonical)
	return NamespaceView{Canonical: c, Parts: Parts(c), TreeParts: TreeParts(c), TreePath: TreePath(c), DisplayPath: DisplayPath(c), Leaf: Leaf(c)}
}
