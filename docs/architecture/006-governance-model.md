# 006 - Governance-Modell

## Zweck
Dieses Dokument definiert Rollen, Reviewstrenge und Freigabeprinzipien für MVP 0.1.

## Warum Governance von Beginn an nötig ist
Business Rules und Produktvarianten haben direkte fachliche Wirkung. Ohne fruehe Governance entstehen inkonsistente Kataloge, unklare Verantwortung und hohes Audit-Risiko.

## Rollen
- Produktverantwortung
- Fachdomäne
- Business Analyse
- Architektur
- Plattformteam
- Delivery Team
- Betrieb
- Reviewer
- Approver

## Verantwortungsgrundsatz
- Produktverantwortung: Produkt und Varianten.
- Business Analyse/Fachdomäne: Anforderungen.
- Architektur/Fachdomäne: Business Rules und kritische Logik.
- Plattformteam: technische Guardrails, CI, Branchschutz.

## Änderungstypen und empfohlene Reviewstrenge
| Änderungstyp | Reviewstrenge |
|---|---|
| minor textual change | 1 Reviewer |
| requirement change | 1 Fachreview + 1 Produktreview |
| rule logic change | 2 Reviews inkl. Architektur/Fachdomäne |
| severity change | 2 Reviews, mindestens 1 Approver |
| product variant change | 2 Reviews |
| validation scenario change | 1-2 Reviews je Kritikalitaet |
| process/skill preparation change | 1 Architekturreview |

## Kritische Änderungen
Kritisch sind insbesondere:
- Rule-Logik,
- Severity,
- Blocking-Bedingungen,
- Freigaberegeln,
- Zielsystem-Referenzen,
- Variantenlogik,
- spaetere Entscheidungslogik.

## Git-basierte Governanceinstrumente
- Branchschutz für Hauptbranch.
- Pflicht-Reviews vor Merge.
- CI als automatisches Gate.
- Pull Request als nachvollziehbarer Review-Antrag.

## Einfacher MVP Workflow
1. Änderungsvorschlag erstellen.
2. Branch erstellen.
3. Artefakte bearbeiten.
4. Lokale oder CI-Validierung ausfuehren.
5. Commit erstellen.
6. Pull Request eroeffnen.
7. Review durchfuehren.
8. Findings beheben.
9. Freigabe erteilen.
10. In Hauptbranch mergen.
