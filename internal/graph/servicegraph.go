package graph

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/nomos/nomos/internal/model"
)

// ErrCycleDetected is returned when a cycle is found in hard depends_on edges.
var ErrCycleDetected = errors.New("cycle detected in hard depends_on edges")

// TopologicalOrder returns node IDs in execution order based on hard depends_on edges.
// Uses Kahn's algorithm. Returns error if a cycle is detected.
func TopologicalOrder(sg model.Servicegraph) ([]string, error) {
	nodeIDs := make(map[string]bool)
	for _, n := range sg.Nodes {
		nodeIDs[n.ID] = true
	}

	// Build adjacency: source -> targets (source must complete before target can start)
	// depends_on edge: source is the dependency, target depends on source
	// So target cannot start until source is done => source -> target in execution order
	inDegree := make(map[string]int)
	adj := make(map[string][]string)
	for id := range nodeIDs {
		inDegree[id] = 0
	}

	for _, e := range sg.Edges {
		if e.Type == "depends_on" && e.Binding == "hard" {
			if !nodeIDs[e.Source] || !nodeIDs[e.Target] {
				continue
			}
			// Source is the prerequisite, Target depends on it
			// Execution order: Source before Target
			adj[e.Source] = append(adj[e.Source], e.Target)
			inDegree[e.Target]++
		}
	}

	var queue []string
	for id := range nodeIDs {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	var order []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)
		neighbors := adj[current]
		sort.Strings(neighbors)
		for _, next := range neighbors {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
				sort.Strings(queue)
			}
		}
	}

	if len(order) != len(nodeIDs) {
		return nil, ErrCycleDetected
	}
	return order, nil
}

// ParallelGroups returns groups of node IDs that can execute in parallel.
// Each group contains nodes whose dependencies have all been met by prior groups.
func ParallelGroups(sg model.Servicegraph) [][]string {
	nodeIDs := make(map[string]bool)
	for _, n := range sg.Nodes {
		nodeIDs[n.ID] = true
	}

	inDegree := make(map[string]int)
	adj := make(map[string][]string)
	for id := range nodeIDs {
		inDegree[id] = 0
	}
	for _, e := range sg.Edges {
		if e.Type == "depends_on" && e.Binding == "hard" {
			if !nodeIDs[e.Source] || !nodeIDs[e.Target] {
				continue
			}
			adj[e.Source] = append(adj[e.Source], e.Target)
			inDegree[e.Target]++
		}
	}

	var groups [][]string
	for {
		var group []string
		for id := range nodeIDs {
			if inDegree[id] == 0 {
				group = append(group, id)
			}
		}
		if len(group) == 0 {
			break
		}
		sort.Strings(group)
		groups = append(groups, group)
		for _, id := range group {
			delete(nodeIDs, id)
			for _, next := range adj[id] {
				inDegree[next]--
			}
		}
		// Remove processed from inDegree
		for _, id := range group {
			delete(inDegree, id)
		}
	}
	return groups
}

// ExecutableNodes returns node IDs whose hard dependencies are all satisfied.
func ExecutableNodes(sg model.Servicegraph, completedNodes map[string]bool) []string {
	nodeIDs := make(map[string]bool)
	for _, n := range sg.Nodes {
		nodeIDs[n.ID] = true
	}

	// Build reverse dep map: which nodes does each node depend on?
	deps := make(map[string][]string) // target -> list of sources it depends on
	for _, e := range sg.Edges {
		if e.Type == "depends_on" && e.Binding == "hard" {
			if !nodeIDs[e.Source] || !nodeIDs[e.Target] {
				continue
			}
			deps[e.Target] = append(deps[e.Target], e.Source)
		}
	}

	var executable []string
	for id := range nodeIDs {
		if completedNodes[id] {
			continue
		}
		allMet := true
		for _, dep := range deps[id] {
			if !completedNodes[dep] {
				allMet = false
				break
			}
		}
		if allMet {
			executable = append(executable, id)
		}
	}
	sort.Strings(executable)
	return executable
}

// ServicegraphMermaid generates a Mermaid diagram for the servicegraph.
func ServicegraphMermaid(sg model.Servicegraph) string {
	var b strings.Builder
	b.WriteString("graph TD\n")

	// Node shapes by type
	for _, n := range sg.Nodes {
		id := sgMermaidID(n.ID)
		label := mermaidLabel(n.Name)
		switch n.Type {
		case "service":
			fmt.Fprintf(&b, "  %s([\"%s\"])\n", id, label) // stadium
		case "activity":
			fmt.Fprintf(&b, "  %s[\"%s\"]\n", id, label) // rectangle
		case "precondition":
			fmt.Fprintf(&b, "  %s{{\"%s\"}}\n", id, label) // hexagon
		case "decision":
			fmt.Fprintf(&b, "  %s{\"%s\"}\n", id, label) // diamond
		case "validation":
			fmt.Fprintf(&b, "  %s((\"%s\"))\n", id, label) // circle
		case "state":
			fmt.Fprintf(&b, "  %s[/\"%s\"/]\n", id, label) // parallelogram
		case "compensation":
			fmt.Fprintf(&b, "  %s[/\"%s\\\\]\n", id, label) // trapezoid
		case "manual":
			fmt.Fprintf(&b, "  %s[[\"%s\"]]\n", id, label) // subroutine
		default:
			fmt.Fprintf(&b, "  %s[\"%s\"]\n", id, label)
		}
	}

	// Edges
	for _, e := range sg.Edges {
		src := sgMermaidID(e.Source)
		tgt := sgMermaidID(e.Target)
		label := e.Type
		if e.Binding == "hard" {
			fmt.Fprintf(&b, "  %s -->|%s| %s\n", src, label, tgt)
		} else {
			fmt.Fprintf(&b, "  %s -.->|%s| %s\n", src, label, tgt)
		}
	}

	return b.String()
}

func sgMermaidID(id string) string {
	// Replace dashes and dots with underscores for valid mermaid IDs
	r := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	return strings.ToLower(r.Replace(id))
}
