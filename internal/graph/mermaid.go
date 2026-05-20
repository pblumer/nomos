package graph

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/nomos/nomos/internal/cosmosfs"
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
	for _, bpn := range tree.Blueprints {
		bp := bpn.Metadata
		bpID := mermaidID("blueprint", bp.ID)
		fmt.Fprintf(&b, "  %s --> %s[\"Blueprint: %s\"]\n", cosmosID, bpID, mermaidLabel(firstNonEmpty(bp.ID, bp.Name)))
		if len(bp.RequiredServices) > 0 {
			for _, svc := range bp.RequiredServices {
				serviceID := mermaidServiceRefID(svc.ServiceRef)
				fmt.Fprintf(&b, "  %s -. requires .-> %s[\"Service: %s\"]\n", bpID, serviceID, mermaidLabel(svc.ServiceRef))
				if svc.ServiceBlueprintRef != "" {
					fmt.Fprintf(&b, "  %s -. uses .-> %s[\"Service Blueprint: %s\"]\n", serviceID, mermaidID("blueprint", svc.ServiceBlueprintRef), mermaidLabel(svc.ServiceBlueprintRef))
				}
			}
			continue
		}
		for _, svc := range bp.RequiredServiceBlueprints {
			fmt.Fprintf(&b, "  %s -. requires .-> %s[\"Service Blueprint: %s\"]\n", bpID, mermaidID("blueprint", svc), mermaidLabel(svc))
		}
	}
	for _, instn := range tree.Instances {
		inst := instn.Metadata
		instID := mermaidID("instance", inst.ID)
		fmt.Fprintf(&b, "  %s --> %s[\"Instance: %s\"]\n", cosmosID, instID, mermaidLabel(inst.Name))
		if inst.BlueprintRef != "" {
			fmt.Fprintf(&b, "  %s -. conforms .-> %s\n", instID, mermaidID("blueprint", inst.BlueprintRef))
		}
	}
	for _, s := range tree.Services {
		serviceID := mermaidID("service", s.Name)
		fmt.Fprintf(&b, "  %s --> %s[\"Service: %s\"]\n", cosmosID, serviceID, mermaidLabel(s.Name))
	}
	for _, dn := range tree.Decisions {
		decID := mermaidID("decision", dn.Metadata.ID)
		fmt.Fprintf(&b, "  %s --> %s[\"Decision: %s\"]\n", cosmosID, decID, mermaidLabel(firstNonEmpty(dn.Metadata.Name, dn.Metadata.ID)))
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

func mermaidServiceRefID(ref string) string {
	domain, service, ok := strings.Cut(ref, "/")
	if !ok {
		return mermaidID("service", ref)
	}
	return mermaidID("service", domain, service)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
