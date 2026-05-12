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

## Cosmos Explorer product workflow

The Cosmos Explorer mirrors the domain-owned product model directly in the tree. Domain rows show compact product and service counts such as `1 prod · 1 svc`. Domains that offer products contain a `Products` pseudo-node next to their `Services` pseudo-node:

```text
Namespaces
└── cloud
    └── blumer
        └── identity
            ├── Products
            │   └── Benutzerkonto mit Mailbox
            └── Services
                └── user-account
```

When creating a product from a selected domain, the selected canonical domain becomes `offered_by` and `owning_domain` defaults to the same value. Nomos still writes the product as a catalog blueprint under `.nomos/catalog/blueprints/products`; the domain is the semantic home, while the Catalog Index is a secondary index.

Product detail in the Cosmos Explorer shows `Offered by`, `Owning domain`, source path, primary home, `Fulfillment Services`, and an `Add fulfillment service` form. Fulfillment services can be local to the offering domain or cross-domain. Cross-domain fulfillment is valid when the referenced service resolves; only missing domains or missing services are shown as unresolved warnings.

## Moving product offerings between domains

A product offering can be reassigned to another domain by changing its `offered_by` field. The Catalog Index remains the storage and index location; the domain-owned `Products` subtree is derived from `offered_by` and refreshes after the move.

`offered_by` identifies the domain that offers the product in the Cosmos Explorer. `owning_domain` identifies the governance domain responsible for the product definition. During a normal move, when both values were the same before the move, both are updated to the target domain. If `owning_domain` intentionally differed from `offered_by`, it is preserved unless the Product Manager explicitly chooses to update it as well.

Moving a product does not rewrite fulfillment service references. Their local or cross-domain classification is recalculated relative to the new `offered_by` domain, so a fulfillment service can change from cross-domain to local (or the reverse) without changing the YAML reference.

## Visualizing product composition in the tree

Product nodes in the Cosmos Explorer expose a `Fulfillment Services` pseudo-node. This subtree is derived from the product blueprint's `fulfillment.required_services` list and is a visualization of composition, not a separate storage location. The product YAML remains in the Catalog Index, and the Catalog Index remains the secondary index view.

Each fulfillment service row shows the referenced service, role, required/optional state, resolution status, and whether the service is local or cross-domain relative to the product offering. Cross-domain fulfillment is valid when the service resolves; unresolved domains or services remain visible so Product Managers can fix the reference or create the missing service.

SLA/OLA metadata is optional and lightweight. It may be defined inline on a product fulfillment reference or on the resolved service as fallback metadata. The tree shows compact `SLA <target>` and `OLA <target>` badges where available, while the product and fulfillment detail panels can show the fuller name, target, availability, and description.
