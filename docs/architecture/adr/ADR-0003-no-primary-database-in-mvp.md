# ADR-0003 - Keine primäre Datenbank für autoritative Artefakte in MVP 0.1

## Kontext
Das MVP fokussiert auf Governance, Traceability und deterministische Validierung von Fachartefakten.

## Entscheidung
Es wird keine primäre Datenbank für autoritative Business-Artefakte in MVP 0.1 eingefuehrt.

## Begruendung
- Transactional CRUD Persistenz ist für den MVP-Kern nicht notwendig.
- Git liefert Versionierung, Review und Historie ohne zusätzliche Persistenzschicht.

## Abgrenzung
- Read Model, Suchindex oder Runtime Store können spaeter als abgeleitete Komponenten eingefuehrt werden.
- Diese Komponenten sind nicht Quelle der Wahrheit.

## Konsequenzen
- Fokus auf dateibasierte Artefaktqualität.
- API und UI arbeiten branchbewusst.
- Governance bleibt zentral im Git-Workflow.
