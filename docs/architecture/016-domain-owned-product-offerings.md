# 016 Domain-owned product offerings

Status: Implemented foundation step for [ADR-0001](../adr/0001-domain-owned-product-offerings.md).

## Why products are domain-owned offerings

A Nomos domain is an ownership and responsibility space. A product is therefore not just a free-floating catalog file: it is a business-facing offering made by exactly one primary domain. Product blueprints now carry `offered_by` and may also carry `owning_domain` to make that responsibility explicit.

## Blueprints are versioned definitions

A product blueprint is the versioned definition of a product. It describes inputs, requirements, rules, quality criteria, variants and fulfillment. The global catalog can still list all blueprints, but the product's semantic owner is the domain named by `offered_by`.

## Services are domain-owned capabilities

Services are reusable production capabilities below domains. A service can now declare explicit ownership and operating metadata:

```yaml
owned_by: cloud.blumer.identity
operated_by:
  - cloud.blumer.identity
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
offered_by: cloud.blumer.identity
owning_domain: cloud.blumer.identity
fulfillment:
  required_services:
    - service_ref: cloud.blumer.identity/user-account
      role: primary
      required: true
      description: Creates the account in the identity domain.
    - service_ref: cloud.blumer.collaboration/mailbox
      role: supporting
      required: true
      description: Provides the mailbox capability.
```

The app DTO adds a `resolution_status` for each required service. Values are `resolved`, `missing`, `unresolved_domain`, and `unresolved_service`.

## Cross-domain fulfillment

Required services may belong to the offering domain or another domain. For example, an identity product can be offered by `cloud.blumer.identity` while using mailbox and license services from `cloud.blumer.collaboration`. This makes dependencies explicit without moving service ownership.

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
