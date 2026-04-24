# 010 - Implementierungsroadmap für Hermes Agenten

## Zweck
Phasenplan für die schrittweise Umsetzung nach der Architekturgrundlage. Dieses Dokument enthaelt keine Implementierung, sondern den Arbeitsrahmen.

## Phase 0 - Dokumentations- und Architekturgrundlage
- Ziel: Gemeinsamer Rahmen für MVP 0.1.
- Scope: Scope, Git-first, Artefaktmodell, Governance, Validierung, UI, REST API Konzept.
- Non-Scope: Codeimplementierung.
- Deliverables: Architekturpaket + ADRs.
- Abhängigkeiten: keine.
- Risiken: Interpretationsspielraum.
- Akzeptanz: konsistente, vollstaendige Doku.
- Handover: Startpunkt für Phase 1.

## Phase 1 - Repository- und Artefaktgrundlage
- Ziel: Strukturiertes Kataloggeruest.
- Scope: Ordnerstruktur, initiale Schemas, Beispielartefakte, Strukturvalidierung, Doku-Beispiele.
- Non-Scope: Runtime Services.
- Deliverables: befuellbares Artefaktfundament.
- Abhängigkeiten: Phase 0.
- Risiken: inkonsistente Konventionen.
- Akzeptanz: Artefakte parsebar, Konventionen eingehalten.
- Handover: Basis für Backend-Lesezugriff.

## Phase 2 - Backend Foundation
- Ziel: API-Grundgeruest und Artefaktlesen aus Git Working Copy.
- Scope: Read Endpunkte, Parsing, Basisauswertung.
- Non-Scope: Vollstaendige Governance-Automatisierung.
- Deliverables: lauffähiger Service-Rahmen.
- Abhängigkeiten: Phase 1.
- Risiken: fruehe Übermodellierung.
- Akzeptanz: Produkte/Regeln lesbar über API.
- Handover: Vorbereitung für Workspace-Flows.

## Phase 3 - Git Workspace Handling
- Ziel: Änderungsvorschlagsprozess abbilden.
- Scope: Branch/Workspace, Speichern, Commit, Review-Request, Diff.
- Non-Scope: Vollautomatisches Freigabe-Management.
- Deliverables: End-to-End Change Proposal Flow.
- Abhängigkeiten: Phase 2.
- Risiken: Konfliktbehandlung.
- Akzeptanz: nachvollziehbarer Branch-zu-PR Prozess.
- Handover: UI kann sichere Änderungen steuern.

## Phase 4 - Produktmanagement UI
- Ziel: Businessfähige Bearbeitungsoberflaeche.
- Scope: Produktübersicht, Detail, Anforderung/Regeleditor, Validierungstest, Findings, Änderungsvorschlag.
- Non-Scope: Vollumfaengliche Admin- und Betriebsoberflaechen.
- Deliverables: nutzbarer MVP UI-Flow.
- Abhängigkeiten: Phase 3.
- Risiken: UX-Komplexitaet.
- Akzeptanz: Fachnutzer können Artefakte ohne Git Know-how pflegen.
- Handover: Vorbereitung für Validierungsengine.

## Phase 5 - Validierungsengine MVP
- Ziel: Deterministische Regelpruefung.
- Scope: Regelauswertung, Szenarien, Findings, Erklärung.
- Non-Scope: Provisionierungsruntime.
- Deliverables: reproduzierbare Validierungsergebnisse.
- Abhängigkeiten: Phase 2-4.
- Risiken: unklare Regelsemantik.
- Akzeptanz: definierte Beispielcases laufen deterministisch.
- Handover: Basis für optionale KI-Unterstuetzung.

## Phase 6 - KI Assistenz MVP
- Ziel: Sichere Draft-Unterstuetzung.
- Scope: Rule/Requirement Drafts, Szenariovorschläege.
- Non-Scope: autonome Änderungen oder Hauptbranch-Schreiben.
- Deliverables: assistiver KI-Flow mit Governance.
- Abhängigkeiten: Phase 4-5.
- Risiken: Überschreitung der Assistenzgrenzen.
- Akzeptanz: alle KI-Vorschlaege bleiben reviewpflichtig.
- Handover: Vorbereitung für Hardening.

## Phase 7 - Hardening und Governance
- Ziel: Betriebssicherheit und Prozessreife.
- Scope: CI Validierung, Reviewchecks, Doku-Updates, Releaseprozess, Audit-Feinschliff.
- Non-Scope: komplette Enterprise-Skalierung.
- Deliverables: stabiler MVP Releaseprozess.
- Abhängigkeiten: Phase 1-6.
- Risiken: Prozessüberladung.
- Akzeptanz: reproduzierbarer, governance-konformer Releaseablauf.
- Handover: Basis für Ausbau nach MVP.
