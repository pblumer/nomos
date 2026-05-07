# Implementation Prompt: Servicegraph for Nomos

## Context

Nomos is a Go-based provisioning knowledge platform. It manages a "Cosmos" repository
on the filesystem consisting of domains, services, blueprints, instances, and rules — all
stored as YAML files. The application exposes a CLI (cobra) and a REST API (net/http).

The codebase is ~5160 lines of Go across these packages:

```
cmd/nomos/main.go          – entry point
internal/model/model.go    – data structs (Cosmos, Domain, Service, Blueprint, Instance, ...)
internal/cosmosfs/tree.go  – filesystem tree loader
internal/fsx/yaml.go       – YAML read/write helpers
internal/app/service.go    – business logic (CRUD for domains, services, blueprints, instances)
internal/app/dto.go        – API response types
internal/server/server.go  – HTTP routes and handlers
internal/validate/validate.go – catalog validation
internal/graph/mermaid.go  – Mermaid graph generation
internal/namespace/         – DNS namespace helpers
internal/cli/root.go       – CLI commands (cosmos, domain, service, blueprint, instance, validate, graph, serve)
```

Filesystem layout of a Cosmos repo:

```
cosmos.yaml
domains/
  <dns-name>/
    domain.yaml
    services/
      <name>/
        service.yaml
catalog/
  blueprints/
    products/   *.yaml
    services/   *.yaml
  instances/
    products/   *.yaml
    services/   *.yaml
  rules/        *.yaml (currently minimal: id, name, description)
  requirements/ *.yaml
  products/     *.yaml
```

---

## Goal

Implement **Servicegraph** as a first-class artifact type in Nomos. A Servicegraph models
a service composition as a directed graph of typed nodes and typed edges, with rules,
variants, and state management.

The implementation must follow strict TDD. Push to origin/main.

---

## Specification

### 1. Data Model (internal/model/)

Add these structs to a new file `internal/model/servicegraph.go`:

```go
package model

type Servicegraph struct {
    ID             string              `yaml:"id" json:"id"`
    Type           string              `yaml:"type" json:"type"`                         // always "servicegraph"
    Name           string              `yaml:"name" json:"name"`
    Version        string              `yaml:"version" json:"version"`
    Status         string              `yaml:"status" json:"status"`
    Owner          string              `yaml:"owner" json:"owner"`
    Summary        string              `yaml:"summary" json:"summary"`
    RelatedProduct string              `yaml:"related_product" json:"related_product"`
    RelatedProcess string              `yaml:"related_process,omitempty" json:"related_process,omitempty"`
    Variants       []GraphVariant      `yaml:"variants,omitempty" json:"variants,omitempty"`
    Nodes          []GraphNode         `yaml:"nodes" json:"nodes"`
    Edges          []GraphEdge         `yaml:"edges" json:"edges"`
    Rules          []GraphRule         `yaml:"rules,omitempty" json:"rules,omitempty"`
}

type GraphNode struct {
    ID             string `yaml:"id" json:"id"`
    Type           string `yaml:"type" json:"type"`              // service|sub_service|activity|precondition|decision|validation|manual|compensation|state|reference
    Name           string `yaml:"name" json:"name"`
    Description    string `yaml:"description,omitempty" json:"description,omitempty"`
    Mandatory      bool   `yaml:"mandatory" json:"mandatory"`
    Reusable       bool   `yaml:"reusable" json:"reusable"`
    Variant        string `yaml:"variant,omitempty" json:"variant,omitempty"`
    ActivationRule string `yaml:"activation_rule,omitempty" json:"activation_rule,omitempty"`
    Owner          string `yaml:"owner,omitempty" json:"owner,omitempty"`
    SkillRef       string `yaml:"skill_ref,omitempty" json:"skill_ref,omitempty"`
    DecisionRef    string `yaml:"decision_ref,omitempty" json:"decision_ref,omitempty"`
    RuleRef        string `yaml:"rule_ref,omitempty" json:"rule_ref,omitempty"`
}

type GraphEdge struct {
    ID          string `yaml:"id" json:"id"`
    Source      string `yaml:"source" json:"source"`
    Target      string `yaml:"target" json:"target"`
    Type        string `yaml:"type" json:"type"`                // composed_of|depends_on|enables|blocks|validates|compensates|alternative_to|requires_condition|produces_state|consumes_state
    Binding     string `yaml:"binding" json:"binding"`          // hard|soft
    Condition   string `yaml:"condition,omitempty" json:"condition,omitempty"`
    Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type GraphRule struct {
    ID          string `yaml:"id" json:"id"`
    Name        string `yaml:"name" json:"name"`
    Type        string `yaml:"type" json:"type"`                // activation|variant|validation|consistency|compensation
    Scope       string `yaml:"scope" json:"scope"`              // node-ID, comma-separated, or "model"
    Binding     string `yaml:"binding" json:"binding"`          // must|should|may
    Expression  string `yaml:"expression,omitempty" json:"expression,omitempty"`
    Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type GraphVariant struct {
    ID      string `yaml:"id" json:"id"`
    Name    string `yaml:"name" json:"name"`
    Context string `yaml:"context,omitempty" json:"context,omitempty"`
}
```

Allowed values (define as constants or use in validation):

```
NodeTypes: service, sub_service, activity, precondition, decision, validation, manual, compensation, state, reference
EdgeTypes: composed_of, depends_on, enables, blocks, validates, compensates, alternative_to, requires_condition, produces_state, consumes_state
EdgeBindings: hard, soft
RuleTypes: activation, variant, validation, consistency, compensation
RuleBindings: must, should, may
```

---

### 2. Filesystem Layout

Servicegraphs are stored in `catalog/servicegraphs/`:

```
catalog/
  servicegraphs/
    <slug>.yaml      (e.g. benutzerkonto-mit-mailbox.yaml)
```

The `type` field in the YAML must be `"servicegraph"`.

---

### 3. Tree Loading (internal/cosmosfs/)

Extend `Tree` struct:

```go
type ServicegraphNode struct {
    Path     string
    Metadata model.Servicegraph
}
```

Add `Servicegraphs []ServicegraphNode` to `Tree`.

In `LoadTree`, scan `catalog/servicegraphs/` for YAML files with `type: servicegraph`.

---

### 4. Validation (internal/validate/)

Create `validateServicegraph(sg model.Servicegraph, path string, res *Result)`.

Validation rules:

| Code | Severity | Condition |
|------|----------|-----------|
| SG_REQUIRED_FIELD | error | id, type, name, version, status, owner empty |
| SG_TYPE_INVALID | error | type != "servicegraph" |
| SG_NO_NODES | error | len(nodes) == 0 |
| SG_NO_EDGES | error | len(edges) == 0 |
| SG_NODE_ID_DUPLICATE | error | duplicate node IDs |
| SG_EDGE_ID_DUPLICATE | error | duplicate edge IDs |
| SG_NODE_TYPE_INVALID | error | node.type not in allowed set |
| SG_EDGE_TYPE_INVALID | error | edge.type not in allowed set |
| SG_EDGE_BINDING_INVALID | error | edge.binding not in {hard, soft} |
| SG_EDGE_SOURCE_MISSING | error | edge.source not found in node IDs |
| SG_EDGE_TARGET_MISSING | error | edge.target not found in node IDs |
| SG_CYCLE_DETECTED | error | cycle in hard depends_on edges (topological sort) |
| SG_NO_SERVICE_NODE | warning | no node with type=service |
| SG_NO_COMPOSED_OF | warning | no edge with type=composed_of |
| SG_STATE_UNBALANCED | warning | consumes_state target has no corresponding produces_state |
| SG_RULE_SCOPE_MISSING | warning | rule scope references non-existent node ID |

Integrate into `validateCatalogArtifacts` by matching `type == "servicegraph"`.

---

### 5. Graph Analysis (internal/graph/)

Create `internal/graph/servicegraph.go`:

```go
package graph

// TopologicalOrder returns node IDs in execution order based on hard depends_on edges.
// Returns error if cycle detected.
func TopologicalOrder(sg model.Servicegraph) ([]string, error)

// ParallelGroups returns groups of node IDs that can execute in parallel.
func ParallelGroups(sg model.Servicegraph) [][]string

// ExecutableNodes returns node IDs whose hard dependencies are all satisfied.
// completedNodes is the set of already-completed node IDs.
func ExecutableNodes(sg model.Servicegraph, completedNodes map[string]bool) []string

// ServicegraphMermaid generates a Mermaid diagram for the servicegraph.
func ServicegraphMermaid(sg model.Servicegraph) string
```

The Mermaid output should:
- Use `graph TD`
- Color-code nodes by type (via CSS classes or shape): service=stadium, activity=rectangle, precondition=hexagon, decision=diamond, validation=circle, state=parallelogram, compensation=trapezoid, manual=subroutine
- Label edges with their type
- Mark hard edges as solid lines (-->), soft edges as dashed (-..->)

---

### 6. Business Logic (internal/app/)

Add to `service.go` (or a new `servicegraph.go`):

```go
func ListServicegraphs(path string) (ServicegraphsDTO, error)
func GetServicegraph(path, id string) (ServicegraphDTO, error)
func CreateServicegraph(path string, sg model.Servicegraph) error
func DeleteServicegraph(path, id string) error
func GetServicegraphMermaid(path, id string) (GraphDTO, error)
func GetServicegraphExecutionOrder(path, id string) (ExecutionOrderDTO, error)
```

DTOs (add to dto.go or new file):

```go
type ServicegraphDTO struct {
    ID             string             `json:"id"`
    Type           string             `json:"type"`
    Name           string             `json:"name"`
    Version        string             `json:"version"`
    Status         string             `json:"status"`
    Owner          string             `json:"owner"`
    Summary        string             `json:"summary"`
    RelatedProduct string             `json:"related_product"`
    Path           string             `json:"path"`
    NodeCount      int                `json:"node_count"`
    EdgeCount      int                `json:"edge_count"`
    RuleCount      int                `json:"rule_count"`
    Variants       []GraphVariantDTO  `json:"variants,omitempty"`
    Nodes          []model.GraphNode  `json:"nodes"`
    Edges          []model.GraphEdge  `json:"edges"`
    Rules          []model.GraphRule  `json:"rules,omitempty"`
}

type ServicegraphsDTO struct {
    Servicegraphs []ServicegraphDTO `json:"servicegraphs"`
    Count         int               `json:"count"`
}

type GraphVariantDTO struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Context string `json:"context,omitempty"`
}

type ExecutionOrderDTO struct {
    ServicegraphID string     `json:"servicegraph_id"`
    Steps          [][]string `json:"steps"`   // parallel groups in order
}
```

---

### 7. REST API (internal/server/)

Add routes:

```
GET    /api/v1/servicegraphs           → list all servicegraphs
GET    /api/v1/servicegraphs/{id}      → get single servicegraph
POST   /api/v1/servicegraphs           → create servicegraph
DELETE /api/v1/servicegraphs/{id}      → delete servicegraph
GET    /api/v1/servicegraphs/{id}/mermaid   → get Mermaid diagram
GET    /api/v1/servicegraphs/{id}/execution → get execution order
```

Update OpenAPI spec (internal/server/openapi.go) with the new endpoints.

---

### 8. CLI Commands (internal/cli/)

Add subcommand `nomos servicegraph`:

```
nomos servicegraph list [--path] [--format json|text]
nomos servicegraph get <id> [--path] [--format json|text]
nomos servicegraph create <file.yaml> [--path]
nomos servicegraph delete <id> [--path]
nomos servicegraph validate <id> [--path]
nomos servicegraph mermaid <id> [--path]
nomos servicegraph execution <id> [--path]
```

Register in `root.go` via `root.AddCommand(servicegraphCmd())`.

---

### 9. Example Servicegraph File

Create `catalog/servicegraphs/benutzerkonto-mit-mailbox.yaml` based on the
instance example in `docs/concepts/servicegraph-instanzbeispiel-benutzerkonto-0.2.md`.

This serves as both documentation and a test fixture.

---

### 10. Test Strategy

All new code requires tests (TDD). Minimum test files:

| File | Tests |
|------|-------|
| `internal/model/servicegraph_test.go` | YAML marshal/unmarshal round-trip |
| `internal/validate/validate_test.go` | All SG_* validation codes (positive + negative) |
| `internal/graph/servicegraph_test.go` | TopologicalOrder (valid, cycle), ParallelGroups, ExecutableNodes, MermaidOutput |
| `internal/app/service_test.go` (or `servicegraph_test.go`) | CRUD operations, error cases |
| `internal/server/server_test.go` | API endpoints (list, get, create, delete, mermaid, execution) |
| `internal/cli/root_test.go` | CLI subcommands basic invocation |

Use `testdata/` directories for fixture YAML files:

```
internal/validate/testdata/
  valid-servicegraph.yaml
  invalid-cycle.yaml
  invalid-missing-fields.yaml
internal/graph/testdata/
  simple-graph.yaml
  parallel-graph.yaml
```

---

### 11. Implementation Order

1. `internal/model/servicegraph.go` + tests (struct + YAML round-trip)
2. `internal/cosmosfs/tree.go` extension (load servicegraphs)
3. `internal/graph/servicegraph.go` + tests (topo-sort, cycle detection, mermaid)
4. `internal/validate/` extension + tests (all SG_* codes)
5. `internal/app/` CRUD + tests
6. `internal/server/` routes + tests
7. `internal/cli/` commands + tests
8. Example `catalog/servicegraphs/benutzerkonto-mit-mailbox.yaml`
9. Update OpenAPI spec
10. Final integration test: `nomos validate --path .` passes with the example file

---

### 12. Constraints

- Go 1.23, no new external dependencies (stdlib + existing: cobra, yaml.v3)
- All code in English (comments, variables, function names)
- Keep file-based architecture (no database)
- Backward-compatible: existing cosmos repos without servicegraphs must still work
- Validation is additive: servicegraph validation only triggers for `type: servicegraph`
- DAG cycle detection via Kahn's algorithm (simple, no recursion stack overflow risk)
- Mermaid output must be valid Mermaid syntax (testable via string assertions)

---

### 13. Acceptance Criteria

- [ ] `nomos servicegraph list` shows all servicegraphs in the cosmos
- [ ] `nomos servicegraph get <id>` returns full details
- [ ] `nomos servicegraph create <file>` creates a new servicegraph YAML
- [ ] `nomos servicegraph delete <id>` removes it
- [ ] `nomos servicegraph mermaid <id>` outputs valid Mermaid diagram
- [ ] `nomos servicegraph execution <id>` outputs parallel execution groups
- [ ] `nomos validate` catches invalid servicegraphs (cycles, missing refs, bad types)
- [ ] All REST endpoints work and return proper JSON
- [ ] All tests pass (`go test ./...`)
- [ ] Example servicegraph file exists and passes validation
- [ ] Existing functionality is not broken (all pre-existing tests still pass)

---

### 14. Reference: Conceptual Documents

The full fachliche Konzepte are in `docs/concepts/`:

- `servicegraph-konzept-0.2.md` – Why graph-based modeling
- `servicegraph-metamodell-0.2.md` – Node types, edge types, rules  
- `servicegraph-modellierungsvorlage-0.2.md` – Template/conventions
- `servicegraph-instanzbeispiel-benutzerkonto-0.2.md` – Full example (use for test fixture)
- `fachliches-metamodell-0.2.md` – Integration with existing Nomos artifacts
