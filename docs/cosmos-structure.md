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

## Domain-owned product offerings

ADR-0001 introduces explicit product ownership and service-based fulfillment. Product blueprints in the catalog can declare `offered_by`, optional `owning_domain`, and `fulfillment.required_services`. Services below domains can declare `owned_by`, `operated_by`, `capabilities`, and `supported_products`.

The catalog remains a global index view. Semantic ownership is expressed by the domain references in YAML. See [architecture note 016](architecture/016-domain-owned-product-offerings.md) for examples and validation codes.

## Domain-owned product offering workflow

In the Cosmos Explorer, domains are the primary place to create and manage product offerings. Products are still persisted as catalog YAML artifacts so Git remains the source of truth, but the catalog is presented as a global index rather than the semantic owner.

A domain-owned product uses `offered_by` to point to the offering domain. `owning_domain` defaults to that same domain during domain-level product creation. Required fulfillment services use canonical service references and may point to services in the same or another domain.

```yaml
id: PROD-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
offered_by: identity.blumer.cloud
owning_domain: identity.blumer.cloud
fulfillment:
  required_services:
    - service_ref: identity.blumer.cloud/user-account
      role: primary
      required: true
    - service_ref: collaboration.blumer.cloud/mailbox
      role: supporting
      required: true
```
