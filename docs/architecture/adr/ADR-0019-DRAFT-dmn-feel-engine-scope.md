# ADR-0019 - DMN 1.5 / FEEL Evaluation Engine: Scope und Build-vs-Buy

## Status

Draft

## Datum

2026-05-19

## Kontext

Nomos kann DMN-Dateien **modellieren** (dmn-js v17 als eingebetteter Editor, ADR-Mapping in `docs/architecture/017-dmn-1.5-mapping.md`) und in Teilen **ausführen** (`internal/dmn/evaluator.go`). Der heutige Stand:

| DMN-Konstrukt                     | Modellieren | Ausführen |
|-----------------------------------|-------------|-----------|
| Decision Tables (alle Hit Policies, Aggregationen) | ✅ | ✅ |
| FEEL Unary Tests (`<`, `<=`, `=`, `!=`, Ranges, Listen, Datum, Zahl, String, Bool) | ✅ | ✅ |
| DRD (InputData, Decision, BKM, KnowledgeSource, DecisionService) | ✅ | parsed only |
| Literal Expressions / volle FEEL-Ausdrücke (Arithmetik, Built-ins, Pfadnavigation) | ✅ | ❌ |
| Boxed Expressions (Context, Invocation, Function, List, Relation, Conditional, For, Every, Some, Filter) | ✅ | ❌ |
| BKM-Invocation, Decision Services | ✅ | ❌ |

Die heutige MVP-Beschränkung ist in einem Mapping-Dokument festgehalten ("Out of scope: Full FEEL expression evaluation beyond unary tests"), **nicht** in einer begründeten ADR. Mit der Cosmos-Tree-Restrukturierung (Product Offering → Processes + Business Rules) wird Decisions als gleichrangige Säule neben Prozessen sichtbar — die Lücke zwischen "modellierbar" und "ausführbar" wird damit erklärungsbedürftig.

Die Frage, die diese ADR beantwortet:

> **Macht es Sinn, eine saubere DMN-1.5- bzw. FEEL-Engine selbst in Go zu schreiben, eine externe Engine einzubinden, oder bei der heutigen Tabellen-Engine zu bleiben?**

## Entscheidung

**Option C (Hybrid): Eigene Go-Engine, inkrementell ausbauen, mit klar definierter Konformitätsstufe pro Release.**

Konkret:

1. **Stufe 1 (heute, erledigt):** Decision Tables mit FEEL Unary Tests. Bleibt der schnelle, deterministische Hauptpfad.
2. **Stufe 2 (vorgeschlagen):** *Common FEEL Subset* — Arithmetik, Vergleiche, Logik, Pfadnavigation (`a.b.c`), grundlegende String-/Listen-/Datums-Built-ins (`contains`, `count`, `sum`, `today`, `date`, `years and months duration`, ~25 Funktionen statt der vollen ~80), `if/then/else`, Context-Literale `{a: 1, b: 2}`. Damit werden Literal Expressions in Decision-Outputs und in Decision-Table-Input-Expressions echte FEEL-Ausdrücke statt String-Matches.
3. **Stufe 3 (bedarfsgetrieben):** Boxed Function, Invocation, BKM-Invocation, Decision-Service-Aufrufe, For/Every/Some/Filter über Relations, vollständiger Built-in-Katalog.
4. **Nicht-Ziele für absehbare Zeit:** Externe FEEL-Imports zwischen DMN-Dateien, eingebettete Java-Skripte, DMN 1.4-only-Konstrukte (Decision Logic Levels), volle DMN-TCK-Konformität.

Jeder Release deklariert in `docs/architecture/017-dmn-1.5-mapping.md` explizit die unterstützte Stufe und ob er **DMN 1.5 partial-conformance Level "Decision Table"** oder höher anstrebt (`Conformance Levels` aus DMN-Spec §11.3).

## Begründete Alternativen

### Option A — Externe Engine als Sidecar/Container

Camunda Zeebe DMN (Java) oder KIE/Drools (Java) per gRPC/REST anbinden.

- **Pro**: Reifer Code, hohe DMN-TCK-Abdeckung, FEEL inkl. aller Built-ins, aktive Community.
- **Kontra**:
  - **Bricht ADR-0006** (Single-Go-Binary, Cobra-CLI). Nomos wird zur Multi-Process-Anwendung mit JVM-Footprint (≈200 MB Heap minimum).
  - **Bricht das Versprechen "kein externer Store"** (ADR-0003) faktisch, weil ein zweiter Runtime-Prozess mit eigener Konfiguration entsteht.
  - **Audit-Lücke gegenüber ADR-0017**: Der Trace-Hash deckt heute `engine.commit` ab (Nomos-Build). Bei externer Engine müsste deren Version, Build-Hash und Konfiguration in den Trace einfließen; Reproduzierbarkeit hängt am Vorhandensein genau dieser externen Engine-Version Jahre später.
  - **MCP-Integration** (ADR-0008): MCP-Tools laufen heute in-process. Decision-Eval würde zu RPC-im-RPC.
  - Lizenz: Camunda 8 = SPL (kommerzielle Lizenz für Production), Camunda 7 EOL, KIE = Apache aber Java-Stack.
- **Wann doch sinnvoll**: Wenn Anwender konkret Boxed-Function-Invocation oder externe DMN-Imports bräuchten — also Konstrukte, deren Eigenentwicklung mehrwöchig ist und die im Nomos-Self-Model (siehe `internal/selfmodel/bundle/domains/nomos/core/decisions`) keinerlei Verwendung haben.

### Option B — Status Quo (nur Decision Tables + Unary Tests)

- **Pro**: Null neuer Code. Decision Tables decken empirisch ~80 % realer Governance-Entscheidungen.
- **Kontra**: Der dmn-js-Editor erlaubt das Modellieren von Konstrukten, die zur Laufzeit dann mit "unsupported" failen — schlechte UX, da der User keinen Hinweis bekommt, was ausführbar ist und was nicht. Außerdem: jede Erweiterung muss dann wieder ad hoc entschieden werden.
- **Wann doch sinnvoll**: Wenn Decisions in Nomos nie über Tabellen hinausgehen sollen. Dann sollte der Editor entsprechend auf Tabellen-Modus beschränkt werden (View-Switcher ohne FEEL/DRD).

### Option C — Eigene Go-Engine, inkrementell *(gewählt)*

- **Pro**:
  - **Konsistenz mit ADR-0006, 0003, 0017, 0008**: Single-Binary, Single-Process, deterministisch, in-process für MCP.
  - **Trace-Determinismus**: Engine-Version = Nomos-Build-Hash. Reproduzierbarkeit über Jahre an genau ein Git-SHA gebunden.
  - **Inkrementelle Investition**: Stufe 2 ist mit einem klaren Pratt-Parser + AST-Walker in ~2–3 Wochen erreichbar (FEEL-Grammatik in DMN-Spec §10 ist überschaubar). Stufe 3 nur on demand.
  - **Lizenzklar**: Bleibt unter Nomos-Lizenz, keine viralen Klauseln.
  - **Embedding in Workflows**: Decision-Eval kann inline in BPMN-Engine, in MCP-Tool-Aufrufen und in CLI-Subcommands genutzt werden, ohne Round-Trip-Latenz.
- **Kontra**:
  - **DMN 1.5 + FEEL ist groß**: ~80 Built-ins, Date-/Duration-/Time-Arithmetik mit allen ISO-Edge-Cases, Boxed Expressions, Decision Services. Volle Konformität wäre 3–6 Monate Vollzeit.
  - **DMN-TCK-Tests** (~1500 Cases) zu durchlaufen ist nicht trivial; ohne Suite-Lauf ist "Konformität" eine Behauptung.
  - **Wartung dauerhaft**: DMN-Spec entwickelt sich (1.6 in Arbeit), Nomos muss mitziehen.
- **Risikomitigation**:
  - Konformitätsstufen pro Release explizit dokumentieren, damit Erwartung und Ausführbarkeit übereinstimmen.
  - Editor-View-Switcher (DRD/Tabelle/FEEL) je nach Stufe selektiv aktivieren.
  - DMN-TCK als externer Testlauf in `make test-dmn-tck` integrieren, sobald Stufe 2 steht — auch wenn anfangs nur ein Bruchteil grün ist, ist die Lücke dann *messbar*.

## Begründung der Wahl

Nomos ist seinem Wesen nach **ein in Git lebendes, sich selbst beschreibendes Cosmos-Modell** (ADR-0001, ADR-0009). Decisions sind dabei nicht periphere Business-Logik, sondern **First-Class-Governance-Artefakte mit kryptografischer Audit-Spur** (ADR-0017). Eine externe Engine als Audit-Endpunkt zwischenzuschalten widerspricht dem Anspruch, dass Nomos selbst die Wahrheit über seine Entscheidungen führt — der Determinismus verlagert sich von "Git-Repo + Nomos-Binary" auf "Git-Repo + Nomos-Binary + Engine-X-Version-Y-Konfiguration-Z".

Gleichzeitig wäre der Vollausbau einer DMN-1.5-Engine ohne konkreten Anwendungsbedarf reine YAGNI. Die inkrementelle Variante (Option C) hält die Tür auf, ohne Kosten zu provozieren, die niemand bezahlt: Sie liefert Stufe 2 in absehbarer Zeit, definiert Stufe 3 als optional und macht den jeweils unterstützten Umfang **zur expliziten, versionierten Aussage** statt zu implizitem Verhalten.

## Konsequenzen

- Neuer ADR-Pfad `docs/architecture/adr/ADR-0019-...md` ersetzt die implizite Scope-Aussage in `017-dmn-1.5-mapping.md`. Das Mapping-Dokument bleibt das **Was**, diese ADR ist das **Warum** und **Wie weit**.
- `internal/dmn` bekommt für Stufe 2 ein neues Sub-Paket `internal/dmn/feel` (Lexer, Pratt-Parser, AST, Interpreter, Built-in-Katalog). Der bestehende Unary-Test-Code in `evaluator.go` wird darauf umgebaut, um Doppelimplementierung zu vermeiden.
- Decision-Trace-Schema bleibt unverändert; `engine.version` und `engine.commit` decken Engine-Stufenwechsel implizit ab.
- Editor-Integration in `internal/server/web/templates/cosmos.html`: ein Konformitätsbadge auf dem Decision-Detail-Panel ("Engine: Tables + FEEL Subset L2") macht für den Anwender sichtbar, was zur Laufzeit garantiert ist.
- Tests: Goldene Testdateien unter `internal/dmn/feel/testdata/` für Stufe 2; spätere TCK-Integration über `make test-dmn-tck` als optionales Target.
- CLI-Subcommand `nomos decision eval <domain> <id> --inputs=...` profitiert direkt von der Stufe-2-Engine (heute scheitert er an allem jenseits von Unary Tests).
- **Kein Lizenz-Sidecar** und **keine JVM** im Distributionspfad — ADR-0006 bleibt unverletzt.

## Offene Punkte

- **TCK-Bootstrap**: Klären, ob die DMN-TCK-Test-Suite (XML-Cases + Erwartungsergebnisse) ohne Java-Runner direkt aus Go ladbar ist. Andernfalls einen schlanken Go-Runner schreiben, der die TCK-XML-Cases als Tabelle abarbeitet.
- **Externe FEEL-Bibliotheken**: Eine schnelle Marktsichtung in Go (`go-feel`, `dmn-go`, etc.) hat keine produktionsreife Implementierung gefunden. Falls eine reife Apache/MIT-lizenzierte Go-Lib auftaucht, ist Option C′ ("forken statt selbst schreiben") eine Variante, die diese ADR explizit erlaubt.
- **Editor-Gating**: Soll der dmn-js-View-Switcher Konstrukte ausblenden, die Nomos nicht ausführt? Pro: ehrliche UX. Kontra: Modellieren als Dokumentation hat eigenen Wert. Empfehlung: nicht ausblenden, aber Konformitätsbadge zeigen.
- **Stufe-3-Trigger**: Welche Anwendung im Nomos-Self-Model oder bei externen Adoptern verlangt zuerst Boxed Functions / BKM-Invocation? Ohne diesen Trigger bleibt Stufe 3 explizit offen.
- **Performance-Budget**: Decision-Eval gehört nicht in Hot-Loops; trotzdem Zielzeit dokumentieren (z. B. < 1 ms für Tabellen, < 5 ms für Stufe-2-Ausdrücke auf typischer Hardware).
