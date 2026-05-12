# 016 Domain-owned product offerings

Status: Implemented foundation step for [ADR-0001](../adr/0001-domain-owned-product-offerings.md).

## Why products are domain-owned offerings

A Nomos domain is an ownership and responsibility space. A product is therefore not just a free-floating catalog file: it is a business-facing offering made by exactly one primary domain. Product blueprints now carry `offered_by` and may also carry `owning_domain` to make that responsibility explicit.

## Blueprints are versioned definitions

A product blueprint is the versioned definition of a product. It describes inputs, requirements, rules, quality criteria, variants and fulfillment. The global catalog can still list all blueprints, but the product's semantic owner is the domain named by `offered_by`.

## Services are domain-owned capabilities

Services are reusable production capabilities below domains. A service can now declare explicit ownership and operating metadata:

```yaml
owned_by: identity.blumer.cloud
operated_by:
  - identity.blumer.cloud
capabilities:
  - user-account-management
supported_products:
  - PROD-ACC-MBX-001
```

Older services with only `owner` still load. DTOs can derive `owned_by` from the containing domain for compatibility, but validation reports `SERVICE_OWNED_BY_MISSING` until the YAML is explicit.

## Fulfillment connects products to services

Product fulfillment is represented as required service references. References use `<domain-canonical>/<service-name>`:

```yaml
id: PROD-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
owner: Identity & Collaboration
offered_by: identity.blumer.cloud
owning_domain: identity.blumer.cloud
fulfillment:
  required_services:
    - service_ref: identity.blumer.cloud/user-account
      role: primary
      required: true
      description: Creates the account in the identity domain.
    - service_ref: collaboration.blumer.cloud/mailbox
      role: supporting
      required: true
      description: Provides the mailbox capability.
```

The app DTO adds a `resolution_status` for each required service. Values are `resolved`, `missing`, `unresolved_domain`, and `unresolved_service`.

## Cross-domain fulfillment

Required services may belong to the offering domain or another domain. For example, an identity product can be offered by `identity.blumer.cloud` while using mailbox and license services from `collaboration.blumer.cloud`. This makes dependencies explicit without moving service ownership.

## Global catalog as index view

The catalog remains available as a global index for browsing and search. It is not removed and existing catalog paths keep loading. The primary semantic ownership is expressed by fields in the YAML, not by moving files during this foundation step.

## Validation rules and finding codes

Nomos validates ownership and fulfillment without failing YAML loading:

| Code | Severity | Meaning |
| --- | --- | --- |
| `PRODUCT_OFFERED_BY_MISSING` | error | Product blueprint has no `offered_by`. |
| `PRODUCT_OFFERED_BY_UNRESOLVED` | error | `offered_by` does not match an existing domain. |
| `SERVICE_OWNED_BY_MISSING` | warning | Service YAML does not explicitly set `owned_by`. |
| `SERVICE_OWNED_BY_UNRESOLVED` | error | `owned_by` does not match an existing domain. |
| `FULFILLMENT_SERVICE_REF_MISSING` | error | A required service entry has no `service_ref`. |
| `FULFILLMENT_SERVICE_DOMAIN_UNRESOLVED` | error or warning | The referenced domain is missing; optional services warn. |
| `FULFILLMENT_SERVICE_UNRESOLVED` | error or warning | The domain exists but the service is missing; optional services warn. |
| `PRODUCT_SELF_SERVICE_REFERENCE` | error | A product appears to reference itself as a service. |

## Compatibility notes

- Existing product blueprints without `offered_by` still load; validation reports an error.
- `owning_domain` is optional. DTO output falls back to `offered_by` when it is empty.
- Legacy `required_services` and `required_service_blueprints` remain readable.
- If the new `fulfillment.required_services` is absent, DTOs can expose legacy `required_services` as fulfillment references.
- Existing services with `owner` but no `owned_by` still load; validation encourages explicit service ownership.
- No files are rewritten automatically and no database is introduced.

## Follow-up implementation steps

1. Add a domain-oriented Products subtree to the Cosmos Explorer UI.
2. Add product detail pages that show ownership, fulfillment and unresolved references.
3. Add service detail pages showing dependent products and capabilities.
4. Add guided migration or authoring helpers for old catalog files.

## Domain-owned product offering workflow

Domains are the primary ownership context for product offerings in the Cosmos Explorer. Product YAML artifacts continue to be stored in the catalog (`.nomos/catalog/blueprints/products`) so the repository remains file-first and Git-friendly, but the catalog is an index, not the semantic owner of a product.

A product offering is linked to its domain with `offered_by`. When a product is created from a selected domain in the Explorer, `offered_by` is set automatically and `owning_domain` defaults to the same canonical domain. Fulfillment services are referenced by canonical service refs in the form `<domain>/<service>`, and cross-domain fulfillment is allowed.

The UI therefore shows products below their offering domain under `Products`, while the global Catalog Index remains useful for search and overview. Fulfillment references are resolved in the app layer and shown as `resolved`, `missing`, `unresolved_domain`, or `unresolved_service`.

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

## Corrected Cosmos Explorer parity

The Cosmos Explorer now supports the same domain-owned product offering workflow as the Domains Explorer:

1. Select a domain in the namespace tree.
2. Review `Products / Offerings` before the domain's services.
3. Use `Produkt hinzufügen` / product creation from either the domain detail panel or the domain context menu.
4. Nomos sets `offered_by` to the selected canonical domain and defaults `owning_domain` to the same domain.
5. The product blueprint is written to `.nomos/catalog/blueprints/products`, because storage remains catalog-based.
6. The new product appears under the selected domain's `Products` pseudo-node and also remains visible in the secondary `Catalog Index`.
7. Product detail in the Cosmos Explorer shows ownership, source path, primary home, fulfillment resolution, and an add-fulfillment form.

This keeps a single app-layer workflow for domain product lookup, product creation, service-reference listing, product detail DTOs, and fulfillment resolution. Templates and handlers do not infer ownership themselves; they render the DTOs returned by the app layer.

## Local and cross-domain fulfillment in the UI

Fulfillment service references resolve to canonical domain service refs. A resolved reference is shown as either `local service` or `cross-domain service`:

- `local service` means the resolved service domain matches the product's `offered_by` domain.
- `cross-domain service` means the product is validly offered by one domain while using a service owned by another domain.
- unresolved refs remain warnings/errors such as `unresolved domain` or `missing service`.

For example, a product offered by `blumer.net` may use `identity.blumer.cloud/user-account`. This is valid when the service exists, and the UI explains that the product is offered by `blumer.net` while the fulfillment service is owned by `identity.blumer.cloud`.

## Moving product offerings between domains

A product offering can be reassigned to another domain by changing its `offered_by` field. The Catalog Index remains the storage and index location; the domain-owned `Products` subtree is derived from `offered_by` and refreshes after the move.

`offered_by` identifies the domain that offers the product in the Cosmos Explorer. `owning_domain` identifies the governance domain responsible for the product definition. During a normal move, when both values were the same before the move, both are updated to the target domain. If `owning_domain` intentionally differed from `offered_by`, it is preserved unless the Product Manager explicitly chooses to update it as well.

Moving a product does not rewrite fulfillment service references. Their local or cross-domain classification is recalculated relative to the new `offered_by` domain, so a fulfillment service can change from cross-domain to local (or the reverse) without changing the YAML reference.

## Visualizing product composition in the tree

Product nodes in the Cosmos Explorer expose a `Fulfillment Services` pseudo-node. This subtree is derived from the product blueprint's `fulfillment.required_services` list and is a visualization of composition, not a separate storage location. The product YAML remains in the Catalog Index, and the Catalog Index remains the secondary index view.

Each fulfillment service row shows the referenced service, role, required/optional state, resolution status, and whether the service is local or cross-domain relative to the product offering. Cross-domain fulfillment is valid when the service resolves; unresolved domains or services remain visible so Product Managers can fix the reference or create the missing service.

SLA/OLA metadata is optional and lightweight. It may be defined inline on a product fulfillment reference or on the resolved service as fallback metadata. The tree shows compact `SLA <target>` and `OLA <target>` badges where available, while the product and fulfillment detail panels can show the fuller name, target, availability, and description.
