# ADR-0007: Domain-owned product offerings and service-based fulfillment

## Status

Accepted

## Date

2026-05-11

## Context

Nomos currently represents product blueprints in a largely independent "Blueprint Catalog" structure, while services are represented separately below domains and namespaces.

This creates a conceptual gap:

- Product blueprints appear to be detached from ownership.
- It is unclear which domain offers a product.
- It is unclear which services are required to produce or fulfill a product.
- It is unclear whether a domain owns a product, operates a service, or both.
- The relationship between products, services, processes, skills and target systems is not explicit enough.

The current Cosmos Explorer tree structure does not express this ownership and fulfillment model clearly enough.

The intended model requires a clearer distinction between:

- domains as ownership and responsibility spaces
- products as business-facing offerings
- product blueprints as versioned definitions of products
- services as reusable production capabilities
- fulfillment as the explicit relationship between products and required services
- processes as orchestration models
- skills as executable task units
- target systems as technical systems used during fulfillment

## Decision

Nomos will distinguish explicitly between the following concepts:

1. Domain
2. Product
3. Product Blueprint
4. Service
5. Fulfillment
6. Process
7. Skill
8. Target System

A domain is a responsibility and ownership space.

A product is a business-facing offering that can be requested, consumed, provisioned or governed.

A product blueprint is the versioned definition of a product. It describes requirements, rules, decisions, variants, processes, quality criteria and fulfillment references.

A service is a reusable capability or production building block. Services are owned or operated by domains.

A product is offered by exactly one primary domain.

A product can be fulfilled by one or more services. These services may belong to the same domain or to other domains.

The relationship between product and service fulfillment must be explicit.

The global blueprint catalog may still exist as an index or library view, but it must not be the primary semantic ownership structure.

The primary semantic structure is:

```text
Cosmos
└── Namespaces
    └── <namespace>
        └── <domain>
            ├── Products
            │   └── <product>
            │       ├── Blueprint
            │       ├── Variants
            │       ├── Requirements
            │       ├── Rules
            │       ├── Decisions
            │       ├── Process
            │       └── Fulfillment
            └── Services
                └── <service>
```

Products must support at least the following ownership and fulfillment fields:

```yaml
offered_by: identity.blumer.cloud

fulfillment:
  required_services:
    - service_ref: identity.blumer.cloud/user-account
      role: primary
    - service_ref: collaboration.blumer.cloud/mailbox
      role: supporting
    - service_ref: collaboration.blumer.cloud/license-assignment
      role: supporting
```

Services must support at least the following ownership fields:

```yaml
owned_by: identity.blumer.cloud
operated_by:
  - identity.blumer.cloud
```

## Consequences

### Positive consequences

- Products no longer appear as free-floating blueprints.
- The offering domain of a product becomes explicit.
- The production or fulfillment model of a product becomes explicit.
- Cross-domain product fulfillment becomes possible.
- The UI can answer important architectural questions:
  - Who offers this product?
  - Which services produce this product?
  - Which domains are involved?
  - Which services are missing or unresolved?
  - Which domain must approve a change?
- The model becomes better aligned with governance, versioning and later orchestration.
- The model becomes easier to use for Codex, AI agents and automation.

### Negative consequences

- Existing blueprint catalog structures must be migrated or reinterpreted.
- The UI must be adjusted to show products under domains.
- Validation logic must check product ownership and service references.
- Existing demo data must be updated.
- API and DTOs may need changes.

## Implementation notes

The implementation should not remove the global catalog immediately.

Instead, Nomos should introduce two complementary views:

1. Domain ownership view  
   The primary model view. Products and services are shown below their owning or offering domain.

2. Catalog index view  
   A secondary index view for searching and browsing all products, blueprints, services, rules and decisions.

The Cosmos Explorer should prefer the domain ownership view.

A product detail page should show:

- product ID
- name
- version
- status
- offered by
- owning domain
- product blueprint
- variants
- requirements
- rules
- decisions
- process
- fulfillment
- required services
- operating domains
- target systems
- unresolved references
- validation findings

A service detail page should show:

- owning domain
- operating domain
- provided capabilities
- supported products
- dependent products
- target systems
- skills or execution hooks, if available

## Validation rules

Nomos should validate at least the following:

- Every product must have `offered_by`.
- `offered_by` must reference an existing domain.
- Every required service reference must resolve to an existing service.
- Required services may belong to the same or a different domain.
- Every service must have `owned_by`.
- `owned_by` must reference an existing domain.
- A product may not reference itself as a service.
- Missing fulfillment references must be reported as findings, not silently ignored.

## Migration strategy

Existing product blueprints in the catalog should be migrated by adding an `offered_by` field.

If the offering domain cannot be derived automatically, the migration should mark the product as unresolved and produce a validation finding.

Existing services under domains should keep their canonical names.

Existing blueprint catalog entries should remain addressable through the catalog index view.

## Decision outcome

Nomos will model products as domain-owned offerings and services as domain-owned production capabilities.

Blueprints are no longer treated as free-floating top-level semantic objects. They become the versioned definition of a product.

Fulfillment becomes an explicit relationship between products and services.

## Related

- [ADR-0001 - Git-first als Quelle der Wahrheit](ADR-0001-git-first-source-of-truth.md)
- [ADR-0005: Blueprint-/Instance-/Assurance-Modell](ADR-0005-blueprint-instance-assurance-model.md)
- [016 - Domain-owned product offerings (implementation)](../016-domain-owned-product-offerings.md)
- [Implementation prompt](../../implementation-prompts/0001-implement-domain-owned-product-offerings.md)
