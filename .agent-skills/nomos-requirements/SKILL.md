---
name: nomos-requirements
description: >
  Create and manage Requirements and Business Rules in Nomos catalogs.
  Use when adding fachliche Anforderungen, Business Rules, Validierungsszenarien
  or when linking rules to requirements and products.
  Also use when interpreting evidence_requirements or quality_criteria on blueprints.
---

# Nomos Requirements & Rules Skill

## Konzept

In Nomos werden Anforderungen und Regeln als **Referenz-IDs** auf Blueprints deklariert.
Die Artefakte selbst liegen im Python Backend (Product Catalog).
Die Go-CLI / Go-Server zeigt sie als aggregierte Views über alle Blueprints.

### Hierarchie

```
Blueprint (product_blueprint / service_blueprint)
  ├── rules[]                 # Business Rule IDs (e.g. "RULE-GDPR-001")
  ├── quality_criteria[]      # Quality criterion IDs (e.g. "QC-SLA-99.9")
  └── evidence_requirements[] # Compliance evidence IDs (e.g. "EV-GDPR-ART5")
```

## ID-Konventionen

| Typ | Format | Beispiel |
|-----|--------|---------|
| Business Rule | `RULE-<DOMAIN>-<NNN>` | `RULE-GDPR-001` |
| Quality Criterion | `QC-<SHORT>-<VALUE>` | `QC-SLA-99.9` |
| Evidence Requirement | `EV-<REGULATION>-<ARTICLE>` | `EV-GDPR-ART5` |
| Validation Scenario | `VS-<PRODUCT>-<NNN>` | `VS-ACCOUNT-001` |

## Blueprint YAML mit Rules und Requirements

```yaml
id: PB-EXAMPLE-001
type: product_blueprint
name: Example Product Blueprint
version: 0.1.0
status: draft
owner: Compliance Team
summary: Blueprint with rules and requirements.
required_inputs:
  - person_reference
required_service_blueprints:
  - SB-EXAMPLE-001
rules:
  - RULE-GDPR-001
  - RULE-DATA-RETENTION-001
quality_criteria:
  - QC-SLA-99.9
  - QC-AVAILABILITY-TIER1
evidence_requirements:
  - EV-GDPR-ART5
  - EV-GDPR-ART32
```

## Linking-Patterns

### Blueprints mit Rules verknüpfen (CLI)

```bash
# Blueprint erstellen mit Rules (als --summary kann ein verweis angegeben werden)
nomos blueprint create \
  --id PB-GDPR-001 \
  --name "GDPR-compliant Account" \
  --type product_blueprint \
  --path .

# Blueprint YAML direkt editieren und rules[] ergänzen:
# rules:
#   - RULE-GDPR-001
#   - RULE-GDPR-002
```

### Evidence Requirements validieren (CLI)

```bash
# Prüfe ob alle evidence_requirements einer Instance belegt sind
nomos validate --path . --format json | \
  jq '.findings[] | select(.code == "INSTANCE_EVIDENCE_MISSING")'

# Instance als verifiziert markieren (schreibt Evidence-Eintrag)
nomos instance verify <instance-id> --path .
```

### Über REST API

```bash
# Alle Blueprints mit ihren Rules anzeigen
curl http://localhost:8080/api/v1/blueprints | \
  jq '.blueprints[] | {id, rules, quality_criteria, evidence_requirements}'

# Requirements-Übersicht im Browser
open http://localhost:8080/requirements

# Rules-Übersicht im Browser
open http://localhost:8080/rules
```

## Python Backend – Eigenständige CRUD-Endpunkte

Für vollständige Anforderungsverwaltung steht das Python Backend zur Verfügung:

```bash
# Requirement erstellen
curl -X POST http://localhost:8001/api/v1/requirements \
  -H "Content-Type: application/json" \
  -d '{"id": "REQ-001", "name": "GDPR Compliance", "description": "...", "owner": "DPO", "severity": "critical"}'

# Business Rule erstellen
curl -X POST http://localhost:8001/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{"id": "RULE-GDPR-001", "name": "Data Minimisation", "description": "...", "type": "compliance", "severity": "error", "expression": "inputs.has_dpa_agreement == true"}'

# Alle Requirements
curl http://localhost:8001/api/v1/requirements

# Requirement an Produkt verknüpfen
curl -X POST http://localhost:8001/api/v1/products/{product_id}/requirements \
  -H "Content-Type: application/json" \
  -d '{"requirement_id": "REQ-001"}'
```

## Validierungsszenarien

Ein Validierungsszenario beschreibt einen vollständigen Test-Case für ein Produkt:

```yaml
# Datei: .nomos/catalog/scenarios/VS-ACCOUNT-001.yaml
id: VS-ACCOUNT-001
type: validation_scenario
name: Neues Benutzerkonto mit GDPR-Compliance
blueprint_ref: PB-ACC-MBX-001
inputs:
  person_reference: "P-12345"
  first_name: "Max"
  last_name: "Mustermann"
  user_type: "internal"
  tenant_id: "T-001"
expected_rules:
  - RULE-GDPR-001
  - RULE-DATA-RETENTION-001
expected_evidence:
  - EV-GDPR-ART5
```

## Häufige Fehler

| Validierungscode | Bedeutung | Lösung |
|-----------------|-----------|--------|
| `INSTANCE_EVIDENCE_MISSING` | evidence_requirement nicht belegt | `nomos instance verify <id>` oder manuell Evidence-Eintrag hinzufügen |
| `INSTANCE_REQUIRED_INPUT_MISSING` | required_input nicht befüllt | inputs.<key>: <wert> in Instance-YAML ergänzen |
| `BLUEPRINT_SERVICE_REF_MISSING` | namespace_service_ref verweist auf unbekannten Service | Service erstellen: `nomos service add <name> --domain <domain>` |
| `BLUEPRINT_SERVICE_BLUEPRINT_REF_MISSING` | service_blueprint_ref nicht gefunden | Service Blueprint erstellen: `nomos blueprint create --type service_blueprint` |

## Referenz: Validation Finding JSON

```json
{
  "code": "INSTANCE_EVIDENCE_MISSING",
  "severity": "warning",
  "message": "Instance PI-001: evidence_requirement 'EV-GDPR-ART5' nicht belegt",
  "path": ".nomos/catalog/instances/products/pi-001.yaml",
  "artifact_type": "product_instance",
  "artifact_id": "PI-001",
  "suggestion": "Füge Evidence-Eintrag mit ID 'EV-GDPR-ART5' hinzu oder führe 'nomos instance verify PI-001' aus"
}
```
