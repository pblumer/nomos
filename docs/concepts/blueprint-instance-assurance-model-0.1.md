# Blueprint-/Instance-/Assurance-Modell 0.1

## Zweck

Dieses Konzept beschreibt das Git-first- und YAML-first-Kernmodell fuer provisionierbare Nomos Artefakte.

## Begriffe

### Namespace Service

Ein Namespace Service ist ein logischer oder organisatorischer Service im Cosmos- bzw. Domain-Baum. Beispiele:

- `identity.blumer.cloud/user-account`
- `collaboration.blumer.cloud/mailbox`
- `collaboration.blumer.cloud/license-assignment`

### Service Blueprint

Ein Service Blueprint ist eine provisionierbare Vorlage fuer genau einen solchen Namespace Service. Service Blueprints koennen deshalb `namespace_service_ref` setzen, zum Beispiel:

```yaml
id: SB-MAILBOX-001
type: service_blueprint
namespace_service_ref: collaboration.blumer.cloud/mailbox
```

### Product Blueprint

Ein Product Blueprint ist eine hoehere provisionierbare Vorlage, die mehrere benoetigte Namespace Services buendelt. Er behaelt aus Kompatibilitaetsgruenden `required_service_blueprints`, beschreibt die konkrete Orchestrierungsbeziehung aber mit `required_services`:

```yaml
id: PB-ACC-MBX-001
type: product_blueprint
required_service_blueprints:
  - SB-IDENTITY-ACCOUNT-001
  - SB-MAILBOX-001
  - SB-LICENSE-ASSIGNMENT-001
required_services:
  - service_ref: identity.blumer.cloud/user-account
    service_blueprint_ref: SB-IDENTITY-ACCOUNT-001
    purpose: Erstellt und verwaltet das Benutzerkonto.
    required: true
```

Damit ist sichtbar, welcher konkrete Namespace Service benoetigt wird und welcher Service Blueprint fuer seine Provisionierung verwendet werden soll.

## Assurance

Instances verweisen auf Blueprints. Assurance vergleicht den beobachteten Ist-Zustand der Instances mit den Blueprint-Anforderungen und dokumentiert Compliance-Status, Evidence und Findings.
