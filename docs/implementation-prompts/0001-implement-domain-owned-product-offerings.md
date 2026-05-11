# Implementation Prompt: ADR-0001 Domain-owned product offerings and service-based fulfillment

You are working in the Nomos repository.

## Goal

Implement ADR-0001: Domain-owned product offerings and service-based fulfillment.

Nomos must no longer treat product blueprints as free-floating semantic objects in the primary Cosmos Explorer model.

Products must be shown as offerings owned by domains.

Services remain domain-owned production capabilities.

Product fulfillment must explicitly reference the services required to produce or provision the product.

## Read first

Before changing code, read:

- `docs/adr/0001-domain-owned-product-offerings.md`
- existing domain, namespace, service, blueprint and catalog model code
- existing API DTOs
- existing Cosmos Explorer tree builder
- existing demo cosmos generation script
- existing validation logic

## Required conceptual changes

Implement the following model semantics:

1. A domain is an ownership and responsibility space.
2. A product is a business-facing offering.
3. A product is offered by exactly one primary domain.
4. A product blueprint is the versioned definition of a product.
5. A service is a reusable production capability.
6. A service is owned by a domain.
7. A product is fulfilled by one or more required services.
8. Required services may belong to the same or to another domain.
9. The global catalog remains available as an index view, not as the primary ownership structure.

## Data model changes

Extend product blueprint or product artefact DTOs with:

```yaml
offered_by: string
owning_domain: string optional
fulfillment:
  required_services:
    - service_ref: string
      role: string
      required: boolean optional
      description: string optional
```

Extend service DTOs with:

```yaml
owned_by: string
operated_by:
  - string
capabilities:
  - string
supported_products:
  - string
```

Use existing naming and canonical-name conventions where possible.

Do not introduce database persistence. Keep the existing file-first and Git-friendly model.

## Tree model changes

Update the Cosmos Explorer tree builder so that the primary tree is domain-oriented:

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

The existing Blueprint Catalog may remain, but only as a secondary catalog/index view.

## Product detail UI

Update the product detail view to show:

- product ID
- name
- version
- status
- offered by
- owning domain, if different
- summary
- variants
- requirements
- rules
- decisions
- process
- fulfillment
- required services
- target systems
- validation findings

Required services should be shown with their resolution status:

- resolved
- missing
- unresolved domain
- unresolved service

## Service detail UI

Update the service detail view to show:

- service ID
- name
- canonical name
- owned by
- operated by
- capabilities
- supported products
- dependent products
- target systems
- validation findings

## Validation

Add or update validation rules:

- A product must define `offered_by`.
- `offered_by` must resolve to an existing domain.
- A service must define `owned_by`.
- `owned_by` must resolve to an existing domain.
- Product fulfillment service references must resolve to existing services.
- Missing services must produce validation findings.
- Missing offering domain must produce validation findings.
- Do not silently ignore unresolved references.

## API changes

Update relevant API responses so that product and service DTOs expose ownership and fulfillment information.

Keep CLI and REST JSON consistent.

If existing endpoints return products, include:

```json
{
  "offered_by": "...",
  "fulfillment": {
    "required_services": []
  }
}
```

If existing endpoints return services, include:

```json
{
  "owned_by": "...",
  "operated_by": [],
  "capabilities": []
}
```

## CLI changes

Update CLI output where useful so product ownership and fulfillment are visible.

At minimum, `nomos cosmos doctor` or equivalent validation output must report missing product ownership and unresolved fulfillment services.

## Demo data

Update the demo cosmos data so that it demonstrates cross-domain fulfillment.

Example:

```yaml
product:
  name: Benutzerkonto mit Mailbox
  offered_by: cloud.blumer.identity
  fulfillment:
    required_services:
      - service_ref: cloud.blumer.identity/user-account
        role: primary
      - service_ref: cloud.blumer.collaboration/mailbox
        role: supporting
      - service_ref: cloud.blumer.collaboration/license-assignment
        role: supporting
```

Ensure the demo contains the referenced domains and services.

## Tests

Add or update tests for:

- parsing product ownership
- parsing service ownership
- resolving product `offered_by`
- resolving required services
- reporting unresolved required services
- rendering domain-oriented product tree
- preserving catalog index behavior
- CLI/API DTO consistency

## Compatibility

Do not break existing catalog files unnecessarily.

If existing product files do not yet contain `offered_by`, load them but mark them with validation findings.

Avoid destructive migrations.

## Expected outcome

After implementation:

- Products are no longer primarily shown as free-floating blueprints.
- Products appear under their offering domain.
- Services appear under their owning domain.
- Product fulfillment explicitly shows which services produce the product.
- Cross-domain fulfillment is visible.
- Validation reports missing ownership or unresolved service references.
- The global catalog remains usable as an index view.
