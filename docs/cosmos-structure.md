# cosmos-structure

## DNS-like namespace and domain tree

Nomos presents domains as a DNS-like tree. The `Cosmos` root contains a `Namespaces` branch. Each namespace is a top-level zone such as `com`, `net`, `ch`, or `cloud`. Domains are label nodes below that zone.

```text
Cosmos
└── Namespaces
    └── <namespace>
        └── <domain-label>
            └── <child-domain-label>
```

The canonical DNS name is derived from the path by reading the labels from the leaf back to the namespace:

- `com/blumer` becomes `blumer.com`.
- `com/blumer/identity` becomes `identity.blumer.com`.
- `cloud/blumer/home` becomes `home.blumer.cloud`.
- `ch/beispiel` becomes `beispiel.ch`.

The tree displays labels only, for example `com → blumer → identity`. Detail panels and APIs continue to use the canonical name such as `identity.blumer.com` as the source of truth. Storage lives under `.nomos/domains/<canonical>` inside the selected Cosmos workspace.

Services are attached to domain nodes, but the UI separates them below a `Services` pseudo-node so that service names are not confused with child domain labels.


## Workspace layout

Nomos source repositories contain code, docs, tests, scripts, templates, and examples only. A concrete Cosmos workspace stores mutable data below `.nomos/`:

```text
my-cosmos/
  .nomos/
    cosmos.yaml
    domains/
    catalog/
      blueprints/
      instances/
      products/
      requirements/
      rules/
      servicegraphs/
    servicegraphs/
    evidence/
    index/
    cache/
  README.md
```

For dedicated Cosmos repositories, commit source-of-truth files under `.nomos/` and ignore only `.nomos/cache/` and `.nomos/index/`. For local experiments inside another project, ignore `.nomos/` entirely.
