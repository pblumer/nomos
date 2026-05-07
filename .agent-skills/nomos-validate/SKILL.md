---
name: nomos-validate
description: >
  Run validation, interpret findings, and fix common validation errors in Nomos.
  Use when executing 'nomos validate', reading validation output, debugging
  BLUEPRINT_*, INSTANCE_*, SG_* finding codes, or setting up CI quality gates.
---

# Nomos Validation Skill

## Schnellstart

```bash
# Vollständige Validierung (Exit 0 = keine Errors; Warnings OK)
nomos validate --path .

# Maschinenlesbares Output für CI / Scripts
nomos validate --path . --format json

# Nur Error-Findings anzeigen
nomos validate --path . --format json | \
  jq '.findings[] | select(.severity == "error")'

# Einzelnen Blueprint validieren
nomos blueprint validate <blueprint-id> --path .
nomos blueprint validate <blueprint-id> --path . --format json
```

## Exit-Code-Semantik

| Exit Code | Bedeutung |
|-----------|-----------|
| `0` | Keine Error-Findings (Warnings sind OK) |
| `1` | Mindestens ein Finding mit `severity: error` |

## Finding-Struktur (JSON)

```json
{
  "code":          "INSTANCE_BLUEPRINT_REF_MISSING",
  "severity":      "error",
  "message":       "Instance PI-001: blueprint_ref nicht gefunden: PB-XYZ",
  "path":          ".nomos/catalog/instances/products/pi-001.yaml",
  "artifact_type": "product_instance",
  "artifact_id":   "PI-001",
  "suggestion":    "Erstelle den Blueprint mit 'nomos blueprint create --id PB-XYZ'"
}
```

## Finding-Code-Referenz

### Blueprint-Findings

| Code | Severity | Ursache | Lösung |
|------|----------|---------|--------|
| `BLUEPRINT_REQUIRED_FIELD` | error | Pflichtfeld (id/type/name/version/status/owner) fehlt | Feld in Blueprint-YAML ergänzen |
| `PRODUCT_BLUEPRINT_INPUTS_EMPTY` | error | `required_inputs` fehlt beim product_blueprint | `required_inputs` Liste ergänzen |
| `PRODUCT_BLUEPRINT_SERVICES_EMPTY` | error | `required_service_blueprints` fehlt | Service Blueprints verlinken |
| `PRODUCT_BLUEPRINT_REQUIRED_SERVICES_RECOMMENDED` | warning | `required_services` leer | Service-Namespace-Refs ergänzen |
| `SERVICE_BLUEPRINT_CAPABILITY_EMPTY` | error | Service Blueprint ohne Fähigkeiten | `capabilities` oder `target_systems` hinzufügen |
| `SERVICE_BLUEPRINT_NAMESPACE_SERVICE_REF_SHAPE` | warning | `namespace_service_ref` hat kein `domain/service`-Format | Format korrigieren: `domain.tld/service-name` |
| `BLUEPRINT_SERVICE_REF_MISSING` | error | `namespace_service_ref` verweist auf unbekannten Service | Service erstellen oder Ref korrigieren |
| `BLUEPRINT_SERVICE_BLUEPRINT_REF_MISSING` | error | `service_blueprint_ref` nicht gefunden | Service Blueprint erstellen |
| `BLUEPRINT_REQUIRED_SERVICE_REF_MISSING` | warning | `required_services[].service_ref` nicht im Cosmos | Service erstellen |

### Instance-Findings

| Code | Severity | Ursache | Lösung |
|------|----------|---------|--------|
| `INSTANCE_REQUIRED_FIELD` | error | Pflichtfeld fehlt | Feld ergänzen |
| `INSTANCE_COMPLIANCE_STATUS_INVALID` | error | Ungültiger `compliance_status` | Erlaubte Werte: `unknown`, `compliant`, `warning`, `non_compliant`, `blocked` |
| `INSTANCE_BLUEPRINT_REF_MISSING` | error | Blueprint nicht gefunden | Blueprint erstellen |
| `INSTANCE_TYPE_MISMATCH` | error | Instance-Type passt nicht zum Blueprint-Type | Type korrigieren (`product_instance` ↔ `product_blueprint`) |
| `INSTANCE_REQUIRED_INPUT_MISSING` | warning | `required_input` nicht belegt | `inputs.<key>: <wert>` ergänzen |
| `INSTANCE_EVIDENCE_MISSING` | warning | `evidence_requirement` nicht abgedeckt | `nomos instance verify <id>` oder manuell belegen |

### Servicegraph-Findings

| Code | Severity | Ursache | Lösung |
|------|----------|---------|--------|
| `SG_REQUIRED_FIELD` | error | Pflichtfeld fehlt | Feld ergänzen |
| `SG_NO_NODES` / `SG_NO_EDGES` | error | Leerer Graph | Nodes und Edges definieren |
| `SG_NODE_TYPE_INVALID` | error | Unbekannter Node-Type | Erlaubt: service, sub_service, activity, precondition, decision, validation, manual, compensation, state, reference |
| `SG_EDGE_TYPE_INVALID` | error | Unbekannter Edge-Type | Erlaubt: composed_of, depends_on, enables, blocks, validates, compensates, alternative_to, requires_condition, produces_state, consumes_state |
| `SG_EDGE_BINDING_INVALID` | error | Ungültige Verbindlichkeit | Erlaubt: `hard`, `soft` |
| `SG_EDGE_SOURCE_MISSING` / `SG_EDGE_TARGET_MISSING` | error | Node-ID nicht gefunden | Node-ID in edge.source/edge.target korrigieren |
| `SG_CYCLE_DETECTED` | error | Zyklus in `depends_on`-Kanten | Zyklus auflösen (harte Abhängigkeiten müssen azyklisch sein) |
| `SG_NO_SERVICE_NODE` | warning | Kein Node mit `type: service` | Haupt-Service-Node definieren |
| `SG_NO_COMPOSED_OF` | warning | Keine `composed_of`-Beziehung | Strukturkanten ergänzen |
| `SG_RULE_SCOPE_MISSING` | warning | Regel referenziert unbekannten Node | Node-ID in Rule.scope korrigieren |
| `SG_STATE_UNBALANCED` | warning | `consumes_state` ohne passendes `produces_state` | State-Flow prüfen |

## Typische Fix-Workflows

### Blueprint-Fehler beheben

```bash
# 1. Fehler anzeigen
nomos blueprint validate PB-001 --path . --format json | jq .

# 2. Blueprint-YAML öffnen und korrigieren
# Pfad aus Finding.path entnehmen, z.B. .nomos/catalog/blueprints/products/pb-001.yaml

# 3. Erneut validieren
nomos blueprint validate PB-001 --path .
```

### Instance-Inputs belegen

```bash
# Fehlende required_inputs anzeigen
nomos validate --path . --format json | \
  jq '.findings[] | select(.code == "INSTANCE_REQUIRED_INPUT_MISSING")'

# Instance-YAML manuell editieren:
# inputs:
#   person_reference: "P-12345"
#   first_name: "Max"
```

### Instance verifizieren (Evidence schreiben)

```bash
# Prüfen welche Evidence fehlt
nomos validate --path . --format json | \
  jq '.findings[] | select(.code == "INSTANCE_EVIDENCE_MISSING") | {artifact_id, suggestion}'

# Manual verification evidence schreiben
nomos instance verify PI-001 --path .

# Danach: Instance zeigt compliance_status: compliant
nomos instance show PI-001 --path . --format json | jq '{compliance_status, evidence}'
```

### Servicegraph-Zyklen finden

```bash
# Execution Order zeigt Fehler wenn Zyklen vorhanden
nomos servicegraph execution <sg-id> --path .

# Mermaid-Diagram zur visuellen Inspektion
nomos servicegraph mermaid <sg-id> --path .
```

## CI-Integration

### GitHub Actions Quality Gate

```yaml
# .github/workflows/ci.yml (Ausschnitt)
- name: Nomos Validation Quality Gate
  run: |
    ./nomos validate --path examples/demo-cosmos --format json | tee /tmp/findings.json
    # Exit-Code 1 wenn Errors vorhanden (Warnings OK)
    ./nomos validate --path examples/demo-cosmos
```

### Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit
nomos validate --path . || { echo "Nomos validation failed. Run 'nomos validate --path .' to see findings."; exit 1; }
```

## REST API

```bash
# Validation via REST (immer Exit 0, Ergebnis im Body)
curl http://localhost:8080/api/v1/validate | jq .

# Blueprint-Validation via REST
curl http://localhost:8080/api/v1/blueprints/PB-001/validate | jq .

# Dashboard im Browser
open http://localhost:8080/validate
```

## Validierungsscope

`nomos validate` prüft:
1. **Cosmos-Datei** — `.nomos/cosmos.yaml` vorhanden
2. **Blueprint-Felder** — Pflichtfelder, Typ-spezifische Regeln
3. **Instance-Felder** — Pflichtfelder, erlaubte compliance_status-Werte
4. **Servicegraph-Struktur** — Node/Edge-Typen, Zyklen, Scope-Referenzen
5. **Cross-Artifact-Referenzen** (neu):
   - Blueprint `namespace_service_ref` → Service existiert im Cosmos
   - Blueprint `required_services[].service_blueprint_ref` → Service Blueprint existiert
   - Instance `blueprint_ref` → Blueprint existiert
   - Instance `type` passt zu Blueprint-Type
   - Instance `required_inputs` vollständig belegt
   - Instance `evidence_requirements` abgedeckt
