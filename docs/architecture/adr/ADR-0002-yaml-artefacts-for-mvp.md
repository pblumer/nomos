# ADR-0002 - YAML als primäres Artefaktformat in MVP 0.1

## Kontext
Artefakte muessen von Fachseite lesbar und von Technikseite maschinell verarbeitbar sein.

## Entscheidung
YAML ist das primäre, menschenlesbare Artefaktformat für MVP 0.1.

## Warum YAML
- Gute Lesbarkeit für Produktmanagement und Business Analyse.
- Strukturierbar für deterministische Verarbeitung.
- Diff-freundlich im Pull Request.

## Warum nicht nur JSON
JSON ist strikt, aber für Fachanwender oft weniger lesbar. YAML reduziert Huerden in fruehen Modellierungsphasen.

## Bezug zu JSON Schema
JSON Schema kann zur strukturellen Validierung genutzt werden, auch wenn Artefakte in YAML gepflegt werden.

## Risiken
- Formatierungsfehler durch Einrückung.
- Uneinheitliche Stilkonventionen ohne Linting.

## Formatierungsanforderungen
- Einheitliche Einrückung und Schlüsselkonvention.
- Stabile Feldreihenfolge für bessere Diffs.
- Keine Secrets in YAML Artefakten.

## Zukunftsoptionen
- Teilmigration zu JSON für Maschinenpfade.
- Zusatzausleitungen in Read Model oder Suchindex.
