# Nomos Konzeptdokumentation

Diese Markdown-Dateien sind aus den bereitgestellten PDF-Konzepten abgeleitet und dienen als versionierbare Quellgrundlage im Repo `nomos`.

## Version 0.2 – Servicegraph-Integration

Mit Version 0.2 wird das Nomos-Modell um eine graphbasierte Service-Modellierung erweitert.
Der Servicegraph positioniert sich zwischen Product und Process und beschreibt:
- Woraus ein Service besteht (Hierarchie via composed_of)
- Welche Abhängigkeiten gelten (Graph via depends_on, enables, blocks, ...)
- Welche Zustände erzeugt/konsumiert werden
- Wie Kompensation modelliert wird
- Wie Varianten ohne Modellduplikation abgebildet werden

## Dokumente

### Fachkonzept und Einführung (0.1)

- `fachkonzept-provisionierungsplattform-0.1.md` – Fachliches Zielbild der Plattform für produktbezogene Provisionierungsregeln, Entscheidungslogik und skillbasierte Ausführung.
- `einfuehrungsmodell-und-mvp-scope-0.1.md` – Schrittweises Einführungsmodell und MVP-Scope 0.1.
- `artefaktmodell-0.1.md` – Artefaktmodell mit Ablagestruktur und Schemas für Product, Rule, Decision, Skill und Process.

### Metamodell

- `fachliches-metamodell-0.1.md` – Metamodell 0.1 mit zentralen Artefakttypen und Referenzprodukt Benutzerkonto mit Mailbox.
- **`fachliches-metamodell-0.2.md`** – **Integriertes Metamodell 0.2 (Nomos + Servicegraph): Mapping, Schema, Konsistenzregeln.**

### Servicegraph (0.2)

- **`servicegraph-konzept-0.2.md`** – Fachliches Konzept der graphbasierten Service-Modellierung für Komposition und Provisionierung.
- **`servicegraph-metamodell-0.2.md`** – Servicegraph-Metamodell: Knotentypen, Beziehungstypen, Regeln, Varianten.
- **`servicegraph-modellierungsvorlage-0.2.md`** – Vorlage zur standardisierten Erfassung eines Servicegraphen.
- **`servicegraph-instanzbeispiel-benutzerkonto-0.2.md`** – Vollständiges Instanzbeispiel: Benutzerkonto mit Exchange-Mailbox als Servicegraph.

### Sonstige

- `blueprint-instance-assurance-model-0.1.md` – Blueprint/Instance/Assurance-Modell.

## Zusammenhang

```
Product (0.1)
     │
     ▼
Servicegraph (0.2)  ◄── Graphbasierte Komposition, Abhängigkeiten, Zustände
     │
     ├── Rules (0.1)         ← formalisieren Vorbedingungen/Validierung
     ├── Decisions (0.1)     ← bewerten Entscheidungsknoten
     │
     ▼
Process (0.1)        ← Ausführungsreihenfolge, ableitbar aus Graph
     │
     ▼
Skills (0.1)         ← operative Ausführung der Aktivitätsknoten
```

## Hinweis

Die PDFs im Verzeichnis sind die Originalquellen. Die Markdown-Dateien sollen als fachliche Source-of-Truth-Dokumentation genutzt und iterativ verbessert werden.
