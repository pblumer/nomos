---
name: nomos-cosmos
description: Create and manage Nomos Cosmos repositories — initialize repos, manage domain hierarchies (TLD, domains, subdomains), services, blueprints, and instances. Use when setting up or restructuring a Nomos cosmos.
---

# Nomos Cosmos Management

## Domain Hierarchy

Domains use DNS-style reverse naming. The tree groups them by TLD:

```
Cosmos
  └── Domains
       └── cloud                        (namespace)
            └── blumer.cloud            (domain)
                 ├── home.blumer.cloud  (subdomain)
                 └── identity.blumer.cloud (subdomain)
                      └── services/
```

## Creating a Cosmos

```
nomos cosmos init /path/to/repo
```

Creates cosmos.yaml, domains/, catalog/ structure.

## Managing Domains

### Add a domain (full DNS name)
```
nomos domain add blumer.cloud --owner "Platform Team" --path .
```

### Add a subdomain
```
nomos domain add identity.blumer.cloud --owner "IAM Team" --path .
```

Or via API:
```
POST /api/v1/domains/{parent}/children
Content-Type: application/x-www-form-urlencoded
segment=identity&owner=IAM+Team
```

### List and delete
```
nomos domain list --path .
nomos domain delete identity.blumer.cloud --path .
```

## Managing Services

```
nomos service add user-provisioning --domain identity.blumer.cloud --owner "IAM Team" --path .
```

Creates: `domains/identity.blumer.cloud/services/user-provisioning/service.yaml`

## Product Blueprint

```yaml
id: PB-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: "0.1.0"
status: draft
owner: Identity Team
summary: Provisioning blueprint for user accounts
required_inputs:
  - person_reference
  - user_type
  - tenant_id
required_service_blueprints:
  - SB-IDENTITY-001
  - SB-MAILBOX-001
required_services:
  - service_ref: identity.blumer.cloud/account-service
    service_blueprint_ref: SB-IDENTITY-001
    purpose: Account creation
    required: true
quality_criteria:
  - Account exists in IAM
evidence_requirements:
  - Account ID
```

## Service Blueprint

```yaml
id: SB-IDENTITY-001
type: service_blueprint
name: Identity Account Service
version: "0.1.0"
status: draft
owner: IAM Team
summary: Creates identity accounts
namespace_service_ref: identity.blumer.cloud/account-service
capabilities:
  - create-account
target_systems:
  - Active Directory
providers:
  - IAM Operations
quality_criteria:
  - Account created
evidence_requirements:
  - Account ID
```

## Instance

```yaml
id: PI-ACC-MBX-001
type: product_instance
name: Example Account
blueprint_ref: PB-ACC-MBX-001
blueprint_version: "0.1.0"
status: active
owner: IAM Operations
inputs:
  person_reference: "P-12345"
  user_type: internal
observed_state:
  account_id: "ACC-67890"
compliance_status: compliant
evidence:
  - id: EV-001
    type: creation_proof
    summary: Account created
findings: []
```

## Web UI (Cosmos Explorer)

Available at `/cosmos`:
- Collapsible tree of domains, services, blueprints
- Right-click context menu:
  - On "Domains" node → Add Domain (full name)
  - On namespace node (e.g. "cloud") → Add Domain (segment only)
  - On domain → Add Subdomain, Add Service, Delete
  - On service → Delete

## Validation

```
nomos validate --path .
```

Checks all artifacts for required fields, valid references, and structural consistency.
