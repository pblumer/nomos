---
name: nomos-servicegraph
description: Create and manage Servicegraph artifacts in Nomos — graph-based service composition with typed nodes, edges, rules, variants, topological execution order, and Mermaid visualization. Use when modeling service dependencies, provisioning order, or service compositions.
---

# Nomos Servicegraph

A Servicegraph models a service composition as a directed graph of typed nodes and typed edges with rules, variants, and state management.

## When to use

- Modeling how a service is composed of sub-services and activities
- Defining execution dependencies and parallel execution groups
- Documenting provisioning order, validations, compensations
- Generating Mermaid diagrams of service compositions

## Servicegraph YAML Schema

```yaml
id: SG-ACC-001
type: servicegraph
name: Service Name
version: "0.2.0"
status: draft
owner: Team Name
summary: Description of the service composition
related_product: PROD-ID

variants:
  - id: V-001
    name: Variant A
    context: when this variant applies

nodes:
  - id: N-001
    type: service
    name: Main Service
    mandatory: true
    reusable: false
    owner: Responsible Role
  - id: N-002
    type: activity
    name: Step Name
    mandatory: true
    reusable: true
    skill_ref: SKILL-ID
    activation_rule: G-001

edges:
  - id: E-001
    source: N-001
    target: N-002
    type: composed_of
    binding: hard
    description: Explanation

rules:
  - id: G-001
    name: Rule Name
    type: activation
    scope: N-002
    binding: must
    expression: condition expression
```

## Node Types

| Type | Mermaid Shape | Purpose |
|------|--------------|---------|
| service | stadium | Top-level orderable service |
| sub_service | rectangle | Part of a service |
| activity | rectangle | Executable step |
| precondition | hexagon | Must be fulfilled first |
| decision | diamond | Decision point |
| validation | circle | Checks result |
| manual | subroutine | Human task |
| compensation | trapezoid | Rollback on failure |
| state | parallelogram | Named state |
| reference | rectangle | Reuse pointer |

## Edge Types

| Type | Meaning |
|------|---------|
| composed_of | Hierarchical containment (structure) |
| depends_on | Must complete before target starts (DAG!) |
| enables | Unlocks target |
| blocks | Prevents target execution |
| validates | Checks target result |
| compensates | Rollback relationship |
| alternative_to | Alternative paths |
| requires_condition | Conditional activation |
| produces_state | Creates a state node |
| consumes_state | Requires an existing state |

## Rule Types

| Type | Purpose |
|------|---------|
| activation | When a node becomes active |
| variant | Variant-specific activation |
| validation | Success criteria |
| consistency | Model quality check |
| compensation | Compensation logic |

## Constraints

1. Hard `depends_on` edges must form a DAG (no cycles)
2. Every `consumes_state` source needs a matching `produces_state` target
3. All IDs (nodes, edges, rules) must be unique within the graph
4. Binding: `hard` (blocking) or `soft` (advisory)

## CLI

```
nomos servicegraph create graph.yaml --path .
nomos servicegraph list --path .
nomos servicegraph get <id> --path .
nomos servicegraph mermaid <id> --path .
nomos servicegraph execution <id> --path .
nomos servicegraph delete <id> --path .
```

## Execution Order

`/api/v1/servicegraphs/{id}/execution` returns parallel groups from topological sort of hard depends_on edges. Same-group nodes can run concurrently.

## Validation Codes

| Code | Severity | Meaning |
|------|----------|---------|
| SG_REQUIRED_FIELD | error | Mandatory field empty |
| SG_TYPE_INVALID | error | type != servicegraph |
| SG_NO_NODES | error | No nodes |
| SG_NO_EDGES | error | No edges |
| SG_NODE_ID_DUPLICATE | error | Duplicate node ID |
| SG_EDGE_ID_DUPLICATE | error | Duplicate edge ID |
| SG_NODE_TYPE_INVALID | error | Unknown node type |
| SG_EDGE_TYPE_INVALID | error | Unknown edge type |
| SG_EDGE_BINDING_INVALID | error | Not hard/soft |
| SG_EDGE_SOURCE_MISSING | error | Source not found |
| SG_EDGE_TARGET_MISSING | error | Target not found |
| SG_CYCLE_DETECTED | error | Cycle in hard depends_on |
| SG_NO_SERVICE_NODE | warning | No service node |
| SG_NO_COMPOSED_OF | warning | No composed_of edge |
| SG_STATE_UNBALANCED | warning | consumes without produces |
| SG_RULE_SCOPE_MISSING | warning | Rule references unknown node |

## ID Conventions

- Servicegraph: `SG-<PREFIX>-<NNN>`
- Nodes: `<PREFIX>-N-<NNN>`
- Edges: `<PREFIX>-R-<NNN>`
- Rules: `<PREFIX>-G-<NNN>`

## File Location

`catalog/servicegraphs/<slug>.yaml`
