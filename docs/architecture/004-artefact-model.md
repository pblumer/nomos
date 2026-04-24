# 004 - Artefaktmodell und Ablagestruktur

## Zweck
Definiert die logische Struktur für fachliche Artefakte im Git Repository.

## Vorgeschlagene Ordnerstruktur
```text
catalog/
  products/
  requirements/
  rules/
  decisions/
  processes/
  skills/
  quality-criteria/
  target-systems/
  validation-scenarios/

schemas/
  product.schema.json
  requirement.schema.json
  rule.schema.json
  decision.schema.json
  process.schema.json
  skill.schema.json
  validation-scenario.schema.json

docs/
  architecture/
```

## Designprinzipien
- Trennung nach Artefakttyp für Klarheit und Ownership.
- Referenzen statt Duplikation.
- Lesbare IDs mit stabilem Praefix.
- Einheitliche Statuswerte und Versionierung.
- Tags für Suche und AI Nutzbarkeit.

## Gemeinsame Basisfelder für alle Artefakte
- id
- type
- name
- version
- status
- owner
- summary
- description
- tags

## Statuswerte
- draft
- in_review
- approved
- deprecated

## ID-Konvention (empfohlen)
- PROD-* für Produkte
- VAR-* für Produktvarianten
- ANF-* für Anforderungen
- RULE-* für Regeln
- DEC-* für Entscheidungen
- PROC-* für Prozesse
- TASK-* für Tasks
- SKILL-* für Skills
- QK-* für Qualitätskriterien
- SYS-* für Zielsysteme
- FIND-* für Findings

## Referenzierung
Artefakte referenzieren einander über IDs und optional über versionierte Referenzen (z. B. `RULE-123@1.2.0`).

## Ownership
`owner` beschreibt fachliche Verantwortung. Kritische Artefakte erfordern zusätzliche Reviewer/Approver gemaess Governance.

## Beispielnamen für Referenzprodukt (nur Dokumentationsbeispiele)
```yaml
# Produkt
id: PROD-BKM
name: Benutzerkonto mit Mailbox

# Produktvariante
id: VAR-BKM-EXTERN
name: Benutzerkonto extern mit Mailbox

# Anforderung
id: ANF-BKM-EXTERN-ENDDATUM
name: Externe Nutzer benötigen Enddatum

# Business Rule
id: RULE-BKM-EXTERN-ENDDATUM-PFLICHT
name: Enddatum muss gesetzt sein, wenn Nutzertyp extern ist
```

Hinweis: Diese Snippets sind Dokumentation und keine produktiven Artefaktdateien.
