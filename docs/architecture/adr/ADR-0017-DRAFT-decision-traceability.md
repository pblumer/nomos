# ADR-0017 - Decision Traceability via inhaltsadressierte, hash-verkettete Trace-Artefakte

## Status

Draft

## Datum

2026-05-18

## Kontext

Mit ADR-0007 (Domain-owned product offerings) und der DMN-Engine (`internal/dmn`) trifft Nomos automatisierte Entscheidungen, deren Output Folgeprozesse in BPMN-Gateways steuert. Beispiel im Demo-Cosmos: `com.blumer.governance/decisions/Provisioning Eligibility` entscheidet via `decision.dmn`, ob ein Benutzeraccount provisioniert wird.

Bisher war eine Auswertung *flüchtig*: `dmn.Evaluate` liefert ein `Result`, aber **es entstand kein persistentes Artefakt**, mit dem sich nachträglich beantworten lässt:

- *Wer* hat *welche Decision* mit *welchen Inputs* ausgewertet?
- *Welche Version der Regeln* (DMN-Inhalt) war zum Auswertungszeitpunkt aktiv?
- *Welches Output* ist entstanden — und ist das stored Output identisch mit dem, das die Regeln für diese Inputs heute liefern würden?
- Ist die Reihe aller Auswertungen *manipulationssicher*?

Audit- und Compliance-Anforderungen (insbesondere für Provisionierungs- und Berechtigungsentscheidungen) verlangen eine vollständige, unveränderbare Spur. Eine Identität pro Auswertung allein per Zufalls-UUID reicht nicht: *gleiche Inputs gegen gleiche Regeln müssen denselben Inhalts-Fingerabdruck ergeben*, damit Reproduzierbarkeit und Diff-Erkennung möglich sind.

## Entscheidung

Jede Decision-Auswertung erzeugt ein **DecisionTrace**-Artefakt, das alle für Nachvollziehbarkeit relevanten Felder enthält und durch einen Content-Hash sowie `parent_trace_id` zu einer Hash-Chain verkettet ist. Traces werden als YAML im Cosmos-Repository persistiert und folgen damit ADR-0001 (Git-first) und ADR-0002 (YAML als primäres Artefaktformat).

### Trace-Struktur

```yaml
schema_version: 1
trace_id: sha256:<hex>                       # content hash über alle anderen Felder
parent_trace_id: sha256:<hex>                # leer für Genesis
timestamp: 2026-05-18T14:30:21.123456789Z    # RFC3339Nano UTC
domain: com.blumer.governance
decision_id: DEC-001
decision_name: Provisioning Eligibility
decision_version: 0.1.0
rule_hash: sha256:<hex>                      # sha256 der decision.dmn Bytes
inputs:                                       # vollständiger Input-Snapshot
  category: premium
outputs:                                      # vollständiger Output-Snapshot
  eligible: true
matched_rules: [r1]
hit_policy: FIRST
evaluator:
  id: alice@example.com                       # optional via X-Nomos-User
  ip: 10.0.0.5                                # aus X-Forwarded-For oder RemoteAddr
  user_agent: curl/8.0
engine:
  name: nomos
  version: 0.1.0
  commit: <git-sha>
```

### Eindeutigkeits-Definition

`trace_id` = sha256 über kanonisches JSON aller Felder mit `trace_id = ""`. Die kanonische JSON-Form sortiert Map-Keys auf allen Ebenen (`internal/dmn/trace.go::canonicalJSON`).

Damit gilt: zwei Auswertungen sind **inhaltsidentisch** genau dann, wenn Domain, Decision, Decision-Version, Regel-Hash, Inputs, Outputs, matched_rules, hit_policy, Evaluator, Engine, Timestamp und `parent_trace_id` übereinstimmen. Reihenfolge der Input-Keys ist irrelevant.

### Hash-Chain

`parent_trace_id` jedes neuen Traces zeigt auf den `trace_id` des letzten Traces im Decision-Verzeichnis (chronologisch). Tampering an einem mittleren Eintrag macht die Chain-Verifikation für alle Folgeeinträge schlagen — gleichwertig zum Git-Commit-DAG, aber pro Decision lokal.

### Speicherort

```
<cosmos>/.nomos/domains/<domain>/decisions/<id>/traces/
  2026-05-18T14-30-21.123Z__<short-hash>.yaml
  2026-05-18T14-31-04.987Z__<short-hash>.yaml
  ...
```

Dateinamen sind zeitlich sortierbar; der Short-Hash macht parallele Auswertungen kollisionsfrei.

### API

- `POST /api/v1/domains/{domain}/decisions/{id}/evaluate` schreibt zusätzlich einen Trace und gibt `{ "result": ..., "trace": ... }` zurück. Evaluator-Felder werden serverseitig aus dem Request (RemoteAddr, X-Forwarded-For, X-Nomos-User, User-Agent) abgeleitet.
- `GET  /api/v1/domains/{domain}/decisions/{id}/traces` listet alle Traces (aufsteigend).
- `GET  /api/v1/domains/{domain}/decisions/{id}/traces/{trace_id}` liefert einen einzelnen Trace.
- `POST /api/v1/domains/{domain}/decisions/{id}/traces/verify` validiert Hash und Kettenintegrität pro Trace.

## Begründung

- **Inhaltsadressierung statt Random-UUID**: Reproduzierbar, deduplizierbar und prüfbar. Wer den Hash kennt, kennt den Inhalt.
- **Git-first, kein externer Store**: Konsistent mit ADR-0001, ADR-0002, ADR-0003. Signierte Commits liefern Nicht-Abstreitbarkeit ohne zusätzliche Infrastruktur.
- **Keine Blockchain**: Eine Blockchain löst *distributed consensus zwischen sich misstrauenden Parteien ohne Notar* — ein Problem, das Nomos auf Cosmos-Ebene nicht hat. Eine Hash-Chain mit signierten Git-Commits liefert Tamper-Evidence ohne Konsens-Overhead, Mining/Staking, Gas-Kosten oder Reorgs. Sollte später cross-Cosmos-Verifikation gefordert sein, ist der natürliche Pfad ein **Transparency Log (Rekor/Sigstore)** statt einer Blockchain — leichter, etabliert, append-only, ohne Token-Ökonomie.
- **YAML statt Binär**: Traces sind menschenlesbar, diffbar, in PRs reviewbar. Größere Datenmengen sind im Decision-Audit-Kontext nicht zu erwarten (eine Auswertung ≪ 10 KB).
- **Datei pro Trace statt Append-Log**: Vereinfacht Verifikation, Concurrent-Writes (über atomare File-Renames in einer Folge-Iteration) und erlaubt Pre-Commit-Hooks pro Eintrag. Bei >100 evals/s sollte auf eine append-only-Log-Datei umgestellt werden — siehe „Offene Punkte".

## Konsequenzen

- Neue Felder in `internal/model/model.go`: `DecisionTrace`, `Evaluator`, `EngineInfo`.
- Neues Paket-File `internal/dmn/trace.go` mit `BuildTrace`, `ComputeTraceID`, `VerifyTrace` und kanonischem JSON-Hashing.
- `internal/app/decision_trace.go` kapselt Persistierung und Verifikation (`EvaluateDecisionWithTrace`, `ListDecisionTraces`, `GetDecisionTrace`, `VerifyDecisionTraces`).
- `internal/server/server.go` ergänzt drei neue Endpoints (siehe oben). Die `/evaluate`-Response-Shape ändert sich auf `{ "result": ..., "trace": ... }`. Bestehende Clients, die direkt `dmn.Result` erwarten, müssen angepasst werden — vertretbar, da die API noch nicht im stabilen Bereich liegt.
- Schreiben des Traces ist Best-Effort: schlägt das Persistieren fehl, wird das Eval-Ergebnis trotzdem zurückgeliefert (mit `trace_error`-Hinweis), damit die Engine nicht durch Disk-Probleme blockiert.
- Bestehende `dmn.Evaluate`-Tests bleiben unverändert; neue Tests in `internal/dmn/trace_test.go` und `internal/app/decision_trace_test.go` decken Hash-Determinismus, Reihenfolge-Unabhängigkeit, Parent-Chain und Tamper-Detection ab.

## Offene Punkte

- **CLI**: `nomos decision verify <domain> <id>` als Pendant zum REST-Endpoint folgt in einem separaten PR.
- **UI**: Trace-Tab am Decision-Node im Cosmos-Explorer (Liste, Detail, Verify-Status) — noch nicht enthalten.
- **Append-Log-Variante**: Für Hot-Path-Decisions evaluieren, ob ein einzelnes `traces.ndjson` mit Hash-Chain-Eintrag pro Zeile besser skaliert. Stellt sich aktuell nicht.
- **Commit-Signaturen**: Ob/wie Nomos Trace-Schreibvorgänge automatisch in signierten Commits bündelt, ist Sache eines Folge-ADRs zur Repo-Workflow-Automation.
- **OpenAPI**: Die Endpoints sind aktuell nicht in `internal/server/openapi.go` dokumentiert; nachzuführen, sobald die Response-Shape stabil ist.
