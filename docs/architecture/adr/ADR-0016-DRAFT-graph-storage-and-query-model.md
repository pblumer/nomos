# ADR-0016 - Graph-Speicherung und Abfragemodell des Cosmos (Git-projizierter openCypher-Graph)

## Status

Draft

## Datum

2026-05-18

## Kontext

Der Nomos-Cosmos ist konzeptionell ein **Graph**: Produkte, Anforderungen,
Business Rules, Entscheidungen, Prozesse, Tasks, Skills, Servicegraph-Knoten,
UCIs, Findings und ihre Beziehungen bilden ein zusammenhängendes Netzwerk
(vgl. [ADR-0009](ADR-0009-nomos-cosmos-network-and-core-engine.md) und das
Fachliche Metamodell 0.2 unter `docs/concepts/fachliches-metamodell-0.2.md`,
das mit Servicegraph, Graph-Node, Graph-Edge und Graph-Rule bereits explizit
graph-basiert modelliert ist).

Heute existiert dieser Graph nur **implizit**:

- Artefakte liegen als YAML-Dateien in Git ([ADR-0002](ADR-0002-yaml-artefacts-for-mvp.md)).
- Beziehungen entstehen über `id`-Referenzen in Feldern wie `related_product`,
  `related_process`, `nodes[].edges[]`, `inputs/outputs`, `producedBy` etc.
- Es gibt **keine standardisierte Abfragesprache**, um Fragen wie
  *„welche Entscheidungen hängen an Rule X?"*, *„welche Prozesse nutzen
  Skill Y?"*, *„welche Findings betreffen ein bestimmtes Servicegraph-Variant?"*
  einheitlich zu beantworten.
- Insbesondere **Prozesse und Entscheidungen** sind als Bäume / Ketten von
  Schritten und Bewertungen modelliert und sollen wie der Git-Tree
  versionsgetreu, mergebar und auditierbar bleiben.

[ADR-0003 (kein Primary Database im MVP)](ADR-0003-no-primary-database-in-mvp.md)
hat festgehalten, dass Nomos keine autoritative Datenbank einführt. Diese
Entscheidung soll erhalten bleiben: Git ist und bleibt **Source of Truth**
([ADR-0001](ADR-0001-git-first-source-of-truth.md)). Gleichzeitig wächst der
Bedarf, den Cosmos **als Graph abzufragen**, ohne die Source-of-Truth-Semantik
zu verletzen.

Etablierte Standards für Graph-Abfragesprachen sind:

- **openCypher** (ISO/IEC GQL-nah, Neo4j, Memgraph, KuzuDB, AGE, RedisGraph).
- **Gremlin / Apache TinkerPop** (JanusGraph, CosmosDB, Neptune).
- **SPARQL / RDF** (Triple-Stores, semantische Reasoning-Stacks).

Eine Festlegung fehlt — sie ist aber Voraussetzung dafür, dass Prozesse,
Entscheidungen und Servicegraph konsistent gespeichert, abgefragt und
versioniert werden.

## Entscheidung

1. **Git bleibt Source of Truth.** Der Cosmos-Graph wird **nicht primär** in
   einer Graphdatenbank gespeichert, sondern als versionierter Satz von
   YAML-Artefakten mit expliziten Knoten- und Kanten-Feldern.
2. **Der Git-Tree ist das Graph-Versionsmodell.** Commits sind unveränderliche
   Snapshots des gesamten Cosmos-Graphen; Branches sind parallele
   Graph-Welten; Merges sind Graph-Joins. Es gibt keinen separaten
   Graph-Versionierungsmechanismus.
3. **openCypher wird die kanonische Abfragesprache** für den Cosmos-Graph.
   Gremlin und SPARQL werden für MVP 0.1 ausdrücklich **nicht** unterstützt.
4. **Die Graph-Engine ist eine Projektion**, kein autoritativer Speicher.
   Beim Start (oder on-demand) baut Nomos aus den YAML-Artefakten einen
   in-memory bzw. embedded Graph auf, gegen den Cypher-Queries laufen.
   Standardimplementierung für MVP: **KuzuDB embedded** (Apache-2.0, Cypher,
   eingebettet, ohne separaten Server). Optional kann eine externe
   Neo4j-Instanz als Read-Modell angebunden werden — die Quelle bleibt Git.
5. **Knoten- und Kantentypen folgen dem Fachlichen Metamodell 0.2**
   (Servicegraph) und werden um die in 0.1 etablierten Artefakttypen
   ergänzt (Product, Rule, Decision, Process, Task, Skill, Finding, UCI).

## Begründung

- **openCypher statt Gremlin**: deutlich grösseres Tooling-Ökosystem
  (Neo4j Browser, Bloom, Memgraph Lab, KuzuDB Explorer), bessere
  Lesbarkeit für Fachpersonen, kompatibel zum kommenden ISO-Standard GQL.
  Gremlin ist mächtiger für komplexe Traversals, aber für die hier
  vorherrschenden Pfad- und Mustersuchen (Produkt → Rule → Decision →
  Process → Task → Skill) ist Cypher ausreichend und didaktisch klarer.
- **Embedded statt Server (MVP)**: KuzuDB läuft im selben Prozess wie
  `nomos serve`, kein Operations-Overhead, kein zusätzlicher
  Deployment-Schritt — passt zur Git-first-Linie aus
  [ADR-0001](ADR-0001-git-first-source-of-truth.md) und zur
  „kein Primary Database"-Linie aus
  [ADR-0003](ADR-0003-no-primary-database-in-mvp.md), weil die DB-Datei
  jederzeit aus Git rekonstruierbar und damit **nicht autoritativ** ist.
- **Projektion statt Persistenz**: Die Graphdatei (`.nomos/graph.kuzu/` o.ä.)
  ist ein Cache. Sie liegt im `.gitignore` und wird beim Wechsel von Branch
  oder Commit neu aufgebaut. Damit bleibt der Git-Tree der eindeutige
  Versions- und Audit-Pfad.
- **Standard-konform abfragbar**: Tools von Drittanbietern, die openCypher
  sprechen (Neo4j-Cypher-Shell, Memgraph Lab, KuzuDB-CLI, BI-Connectoren),
  funktionieren auf der Projektion ohne weitere Adapter.
- **Determinismus**: Da die Projektion vollständig aus YAML rekonstruiert
  wird, ist sie für denselben Commit reproduzierbar und damit für
  Validierung und CI-Quality-Gates geeignet.

## Konsequenzen

### Artefaktmodell

- YAML-Artefakte erhalten ein konsistentes Schema:
  - `id` (Cosmos-eindeutig)
  - `type` (einer der erlaubten Knotentypen)
  - `edges:` Liste von `{ to: <id>, kind: <kantentyp>, attrs: {…} }`
- Bestehende Felder wie `related_product`, `related_process`,
  `nodes/edges` im Servicegraph werden auf das einheitliche Kantenmodell
  abgebildet (Migrationspfad, kein Breaking Change in 0.1).

### Knoten- und Kantentypen (initial)

Knoten: `Product`, `ProductVariant`, `Requirement`, `Rule`, `Decision`,
`Process`, `Task`, `Skill`, `Servicegraph`, `GraphNode`, `Finding`,
`UCI`, `Service` (Nomos-Self-Model, [ADR-0014](ADR-0014-DRAFT-nomos-self-model-bundle.md)).

Kanten (Beispiele, vollständige Liste folgt im Implementierungs-PR):
`REALIZES`, `REFERENCES`, `VALIDATES`, `DECIDES_ON`, `TRIGGERS`,
`COMPENSATES`, `PRODUCES_STATE`, `CONSUMES_STATE`, `BELONGS_TO_VARIANT`,
`OWNED_BY`, `FOUND_BY`.

### Komponenten

- Neues internes Paket `internal/graph/` mit:
  - `Loader` — liest YAML aus dem Cosmos-Workspace und baut die Projektion.
  - `Store` — abstrahiert KuzuDB (Default), optional Neo4j-Adapter.
  - `Query` — Cypher-Eingang, getypte Ergebnisstrukturen.
- Neuer CLI-Befehl `nomos graph` (Subkommandos `build`, `query`, `stats`).
- Neuer REST-Endpoint `POST /api/v1/graph/cypher` (read-only, parametrisierte
  Queries, parameterisierte Whitelist optional konfigurierbar).
- Web-UI-Erweiterung: Cosmos-Explorer kann gespeicherte Cypher-Snippets
  ausführen und Ergebnisse als Tabelle/Graph anzeigen.

### Versionierung & Reviews

- Da Git die Versionierung übernimmt, ist `git log -- <pfad>` automatisch
  die Audit-Spur einzelner Knoten / Kanten.
- Prozess- und Entscheidungs-Historie ergibt sich aus dem Commit-Graphen
  selbst; ein Merge-Commit ist ein expliziter Knoten-Punkt mit zwei
  Eltern-Versionen des Cosmos-Graphen.
- Diffs zwischen zwei Commits können auf Graph-Ebene als
  „added/removed/changed nodes & edges" visualisiert werden (Folge-Story).

### Performance & Limits (MVP)

- Initiale Zielgrösse: bis ca. 100.000 Knoten / 500.000 Kanten — passt
  bequem in KuzuDB embedded.
- Rebuild-Zeit beim Branch-Wechsel sollte unter 2 s für typische
  Cosmos-Grössen liegen; sonst inkrementelles Update auf Basis
  `git diff --name-only`.

### Nicht-Ziele

- Keine Schreiboperationen über Cypher in den Cosmos. Mutationen laufen
  ausschliesslich über YAML-Editierung + Git ([ADR-0001](ADR-0001-git-first-source-of-truth.md),
  [ADR-0002](ADR-0002-yaml-artefacts-for-mvp.md)).
- Kein Multi-User-Locking auf Graph-Ebene; Konfliktlösung ist Git-Merge.
- Keine eigene Authentifizierung der Graph-Engine — die REST-Schicht
  setzt die bestehenden Auth-Konzepte fort.

## Betrachtete Alternativen

- **Gremlin / TinkerPop**: Mächtiger für tiefe Traversals, aber schwergewichtigerer
  Stack (JanusGraph braucht Cassandra/HBase, TinkerPop-in-process via TinkerGraph
  ist möglich, aber Tooling-Ökosystem schwächer). Verworfen für MVP, kann später
  zusätzlich angeboten werden.
- **SPARQL / RDF**: Stark für Ontologie und Reasoning, aber Modellierungsaufwand
  (Triples, Namespaces, Ontologien) ist für den fachlichen Cosmos überzogen.
  Optional als Export-Format für Interop denkbar (JSON-LD), nicht als primäre
  Abfrageschnittstelle.
- **Neo4j als Primary Database**: Würde Source-of-Truth-Linie aus ADR-0001
  brechen und Operations-Aufwand erzeugen. Verworfen. Bleibt als optionaler
  Read-Adapter denkbar.
- **Eigene Mini-Query-DSL**: Schnell gebaut, aber kein Standard, keine Tools,
  Lock-in. Verworfen.
- **SQLite mit rekursiven CTEs**: Funktioniert für flache Graphen, aber kein
  Standard für Graph-Queries und unleserlich bei Mustern wie
  `MATCH (p:Product)-[:REALIZES*1..3]->(t:Task)`. Verworfen.

## Offene Punkte

- **Konkrete Bibliothek**: Endgültige Wahl zwischen KuzuDB (Go-Bindings via
  CGo) und einer reinen Go-Implementierung (z. B. `cayley` mit eigenem
  Cypher-Frontend) wird im Implementierungs-Spike entschieden.
- **Inkrementeller Rebuild**: Algorithmus für Diff-basierten Update der
  Projektion bei `git checkout` / `git pull`.
- **Schema-Versionierung**: Wie wird der Knoten-/Kanten-Typkatalog selbst
  versioniert? Vorschlag: als Nomos-Artefakt im Self-Model-Bundle
  ([ADR-0014](ADR-0014-DRAFT-nomos-self-model-bundle.md)).
- **Export-Formate**: GraphML, JSON-LD, Cypher-Dump — welche garantieren
  wir als stabile Schnittstelle?
- **Read-Adapter Neo4j**: Genaue Sync-Strategie (Push beim Commit?
  Pull beim Branch-Wechsel?) klären, sobald reale Nachfrage besteht.

## Related

- [ADR-0001 - Git-first Source of Truth](ADR-0001-git-first-source-of-truth.md)
- [ADR-0002 - YAML-Artefakte für MVP](ADR-0002-yaml-artefacts-for-mvp.md)
- [ADR-0003 - Kein Primary Database im MVP](ADR-0003-no-primary-database-in-mvp.md)
- [ADR-0009 - Nomos-Cosmos-Netzwerk und Core-Engine](ADR-0009-nomos-cosmos-network-and-core-engine.md)
- [ADR-0014 - Nomos Self-Model Bundle](ADR-0014-DRAFT-nomos-self-model-bundle.md)
- `docs/concepts/fachliches-metamodell-0.2.md` (Servicegraph-Metamodell)
- `docs/concepts/servicegraph-metamodell-0.2.md`
