package graph

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/namespace"
)

func Mermaid(tree cosmosfs.Tree) string {
	cosmosName := tree.Cosmos.Name
	if cosmosName == "" {
		cosmosName = tree.Path
	}
	cosmosID := mermaidID("cosmos", cosmosName)
	var b strings.Builder
	fmt.Fprintf(&b, "graph TD\n")
	fmt.Fprintf(&b, "  %s[\"Cosmos: %s\"]\n", cosmosID, mermaidLabel(cosmosName))
	for _, d := range tree.Domains {
		domainID := mermaidID("domain", d.Name)
		fmt.Fprintf(&b, "  %s --> %s[\"Domain: %s\"]\n", cosmosID, domainID, mermaidLabel(d.Name+"\n"+namespace.DisplayPath(d.Name)))
		for _, s := range d.Services {
			serviceID := mermaidID("service", d.Name, s.Name)
			fmt.Fprintf(&b, "  %s --> %s[\"Service: %s\"]\n", domainID, serviceID, mermaidLabel(s.Name))
		}
	}
	return b.String()
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func mermaidID(parts ...string) string {
	base := strings.ToLower(strings.Join(parts, "_"))
	base = nonAlphaNum.ReplaceAllString(base, "_")
	base = strings.Trim(base, "_")
	if base == "" {
		base = "node"
	}
	if unicode.IsDigit(rune(base[0])) {
		base = "n_" + base
	}
	return base
}
func mermaidLabel(label string) string {
	label = strings.ReplaceAll(label, "\"", "\\\"")
	label = strings.ReplaceAll(label, "\n", " ")
	label = strings.ReplaceAll(label, "\r", " ")
	return label
}
