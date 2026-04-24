# 011 - Hermes Agent Handover

## Zweck
Dieses Dokument gibt Folgeagenten verbindliche Leitlinien für die kontrollierte Umsetzung nach MVP 0.1 Architekturrahmen.

## Aktuelle Architekturentscheidungen
- Git-first als Quelle der Wahrheit.
- YAML als primäres Artefaktformat.
- Keine primäre Datenbank für autoritative Artefakte in MVP 0.1.
- KI nur assistiv für Drafts, nicht autonom entscheidend.

## MVP Grenzen
- Fokus auf Produktkatalog, Anforderungen, Business Rules, Validierung, Findings, Versionierung, Governance.
- Keine Provisionierungsruntime.
- Keine Workflow Engine.
- Keine autonome KI-Entscheidung.

## Dokumentationsstruktur
- docs/architecture/000-011 für Rahmen und Umsetzungspfad.
- docs/architecture/adr für Architekturentscheidungen.

## Arbeitsweise für Folgeagenten
- Immer zuerst Scope und ADRs lesen.
- Änderungen in kleinen, reviewbaren Schritten liefern.
- Keine Vermischung von Architektur- und Laufzeitentscheidungen ohne Doku-Update.
- Bei Abweichungen: zuerst ADR vorschlagen, dann umsetzen.

## Erlaubte Änderungen ohne Sonderfreigabe
- Erweiterung von Dokumentation im bestehenden Rahmen.
- Umsetzung der jeweils nächsten Roadmap-Phase.
- Ergänzung von validierbaren Beispielartefakten (ab Phase 1).

## Nicht erlaubt ohne explizite Anweisung
- Abkehr von Git-first.
- Einfuehrung einer primären Artefaktdatenbank im MVP 0.1.
- Autonome KI-Entscheidungslogik.
- Überspringen von Governance für kritische Rule-Änderungen.

## Golden Rules for Hermes Agents
1. Respect Git as the source of truth for business artefacts.
2. Do not introduce a primary database for artefacts in MVP 0.1.
3. Do not implement autonomous AI decisions.
4. Do not mix product, rule, decision, process and skill concepts.
5. Do not hide governance-relevant changes.
6. Do not store secrets in artefacts.
7. Keep examples based on "Benutzerkonto mit Mailbox".
8. Keep documentation in German.
9. Use ä, ö und ü.
10. Never use the sharp s character.

## Pflegehinweise
- Jede inhaltliche Änderung an Prinzipien in ADR dokumentieren.
- Terminologie konsistent halten.
- Bei neuen Artefakttypen MVP-Relevanz markieren.
