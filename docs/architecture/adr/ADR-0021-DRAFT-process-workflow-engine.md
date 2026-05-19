# ADR-0021 - Process/Workflow Engine: Scope und Build-vs-Buy

## Status

Draft

## Datum

2026-05-19

## Kontext

Mit ADR-0018 (Process Trigger Events) ist der **Modell-Layer** für
typisierte Start-Events geklärt: Ein Prozess weiß, *wann* er starten
soll (Cron, Message, Signal, Conditional). Was bisher fehlt — und in
ADR-0018 bewusst Out-of-Scope gestellt wurde — ist die **Runtime**, die
diese Trigger umsetzt und einen BPMN-Prozess tatsächlich *ausführt*:

- den Token von `bpmn:startEvent` über Sequence Flows, Gateways und
  Tasks bis zu `bpmn:endEvent` weiterbewegt,
- bei `bpmn:userTask`/`bpmn:serviceTask` pausiert, bis Eingaben
  vorliegen oder ein externer Service antwortet,
- Timer auslöst, Boundary-Events feuert, Sub-Prozesse instanziiert,
- Instanzzustand persistiert, sodass ein Restart der Nomos-Binary
  laufende Prozesse nicht verliert,
- jeden Schritt revisionssicher in den Decision-/Process-Trace
  (ADR-0017) schreibt.

BPMN-Prozesse in Nomos (`internal/app/process.go`) sind heute reine
Modellierungsartefakte. Mit der Cosmos-Tree-Restrukturierung (Product
Offering → Processes + Business Rules + Service Bindings) werden
Prozesse als zweite gleichrangige Säule neben Decisions sichtbar — und
genau wie bei DMN (siehe ADR-0019) klafft eine Lücke zwischen "im
Editor modellierbar" und "auf der Plattform ausführbar".

Die Frage, die diese ADR beantwortet:

> **Macht es Sinn, eine BPMN-Workflow-Engine selbst in Go zu schreiben,
> eine externe Engine (Camunda Zeebe, Temporal, Flowable, …)
> einzubinden, oder bei reiner Modellierung ohne Runtime zu bleiben?**

Die Antwort muss vier harte Leitplanken respektieren, die durch frühere
ADRs gesetzt sind:

1. **Single-Go-Binary, Single-Process** (ADR-0006)
2. **Keine primäre Datenbank für autoritative Artefakte** (ADR-0003)
3. **Git-first als Quelle der Wahrheit für Modelle** (ADR-0001)
4. **Vollständige Decision-/Process-Traceability mit Nomos-Build-Hash
   als Determinismus-Anker** (ADR-0017)

## Entscheidung

**Option C (Hybrid): Eigene Go-Engine, inkrementell ausgebaut, mit klar
definierter Konformitätsstufe pro Release und Journal-basierter
Persistierung von Instanzzustand auf dem lokalen Dateisystem (nicht im
Git-Repo).**

Konkret werden drei Konformitätsstufen definiert und je Release explizit
deklariert. Modell-Artefakte (Prozess-YAML + BPMN-XML) bleiben in Git
und damit revisionssicher; **Instanz-Zustand** (laufende Token, History)
ist **operativ**, lebt außerhalb von Git in einem lokalen Journal pro
`nomos serve`-Instanz und gehört nicht in die Source of Truth.

### Stufe 1 — Synchronous Happy-Path (Ziel: erstes ausführbares Release)

Konstrukte:

- `bpmn:startEvent` (Plain und Timer mit `iso_duration`/`iso_date` —
  Message-/Signal-/Conditional-Start bleiben Stufe 2)
- `bpmn:endEvent` (terminating, ohne Error/Escalation)
- `bpmn:sequenceFlow` inkl. Condition Expressions (FEEL-Stufe-2 aus
  ADR-0019)
- `bpmn:exclusiveGateway`, `bpmn:parallelGateway`
- `bpmn:task`, `bpmn:scriptTask` (mit FEEL-Expression), `bpmn:userTask`
  (pausiert auf externe Completion via REST), `bpmn:serviceTask` (ruft
  registrierten Go-Handler / MCP-Tool synchron auf)
- `bpmn:businessRuleTask` (Aufruf der DMN-Engine aus ADR-0019)

Persistenz: **Event-Journal als Append-Only-NDJSON** pro Instanz unter
`$NOMOS_STATE_DIR/instances/<process-id>/<instance-uuid>.journal`. Beim
Start liest `nomos serve` alle Journals und rekonstruiert den
Token-Stand durch Replay. Kein externer Store.

Trigger-Runtime: In-Process-Scheduler (`github.com/robfig/cron/v3`
*intern*, ohne öffentliche Cron-Dependency-API) für Timer-Triggers,
in-process Event-Bus für Message/Signal (Webhook-Receiver
`POST /api/v1/events`).

### Stufe 2 — Long-running + asynchrone Events

- Message-/Signal-Start-Events und intermediate Catch-Events
- `bpmn:boundaryEvent` (Timer, Error, Escalation, Message)
- `bpmn:subProcess` (embedded + call activity)
- `bpmn:eventBasedGateway`
- Externes Event-Backend pluggable: in-process (Default) oder
  NATS-JetStream (opt-in über `nomos.yaml`); At-least-once mit
  Idempotenz-Key pro Event.
- Optionaler **externer Journal-Store** (SQLite-File oder Postgres) für
  Multi-Node-Deployment. In-Process bleibt der Default.

### Stufe 3 — Vollständige BPMN-2.0-Executable-Konformität

- Multi-Instance (sequential + parallel)
- Compensation, Transaction Subprocess
- Inclusive Gateway mit konvergenter Synchronisation
- Complex Gateway
- DMN-Decision Service als Lane-Provider

**Nicht-Ziele für absehbare Zeit:** BPMN-2.0-Executable-Audit-Zertifizierung,
Choreography Diagrams, Conversation Diagrams, Collaboration über
mehrere Nomos-Instanzen hinweg, BPMN-DI-Layout-Round-Trip jenseits
dessen, was bpmn-js bereits liefert.

Jeder Release deklariert in `docs/architecture/017-dmn-1.5-mapping.md`
(bzw. einem parallelen `018-bpmn-execution-mapping.md`) die unterstützte
Stufe. Die UI zeigt pro Prozess einen Konformitäts-Badge ("Engine: BPMN
Subset L1"), damit Modellierer wissen, was zur Laufzeit garantiert ist.

## Begründete Alternativen

### Option A — Externe BPMN-Engine als Sidecar/Container

Camunda Zeebe (Java, gRPC), Flowable (Java, REST), jBPM/KIE (Java,
REST), Camunda 7 (EOL/Maintenance, Java).

- **Pro**: Volle BPMN-2.0-Executable-Konformität, mature
  Implementierungen, BPMN-TCK-zertifiziert, große Community, gelöste
  Themen wie Cluster, Backpressure, History-Cleanup.
- **Kontra**:
  - **Bricht ADR-0006**: Nomos wird zur Multi-Process-Anwendung mit
    JVM-Footprint (Zeebe-Broker ≈ 1 GB Heap, dazu Elasticsearch für
    History). Single-Binary-Versprechen entfällt.
  - **Bricht ADR-0003 faktisch**: Zeebe braucht einen
    Exporter+Elasticsearch (oder kompatiblen Store), Flowable einen
    SQL-Store. Ein "kein externer Store"-Versprechen für *Artefakte*
    lässt sich für *Instanzen* schwer aufrechterhalten — aber ein
    JVM+Elasticsearch-Setup ist eine ganz andere Größenordnung als ein
    NDJSON-Journal.
  - **Audit-Lücke gegenüber ADR-0017**: Der Trace-Hash deckt heute
    `engine.commit` ab (Nomos-Build). Bei externer Engine müssten deren
    Version, Build-Hash, Plugin-Versionen und Konfiguration in den Trace
    einfließen. Reproduzierbarkeit hängt am Vorhandensein genau dieser
    externen Engine-Version Jahre später.
  - **Modeler-Lock-in**: Zeebe verlangt `zeebe:`-Extension-Attribute im
    BPMN-XML (Job-Type, Header), Flowable analog `flowable:`. Das
    widerspricht ADR-0018 (Konfiguration in YAML, BPMN bleibt portabel).
  - **Lizenzen**: Camunda 8/Zeebe = SPL (für Production kommerzielle
    Lizenz nötig), Camunda 7 EOL, Flowable = Apache (ok),
    jBPM/KIE = Apache (ok). Aber alles JVM-Stack.
  - **Betriebsaufwand**: Ein zweiter Daemon mit eigener Versionierung,
    eigenen Backup-Pfaden, eigenem Upgrade-Pfad. Für eine
    Governance-Plattform, die "ein YAML-fähiger Mensch betreibt" ein
    erheblicher Bruch.
- **Wann doch sinnvoll**: Wenn ein konkreter Adopter BPMN-2.0-Executable-Vollkonformität
  und Multi-Node-HA als harte Anforderung mitbringt — also Szenarien
  jenseits dessen, was Nomos heute als Plattform-Versprechen ausgibt.

### Option B — Code-driven Workflow-Library (Temporal, Cadence, Conductor)

Temporal.io (Go-SDK, Go-Server), Uber Cadence (Vorgänger), Netflix
Conductor (Java).

- **Pro**: Sehr mature, durable execution as a service, exzellente
  Go-SDKs, gelöste Themen wie Retries, Versioning, Replay-Determinismus.
- **Kontra**:
  - **Workflows sind Code, nicht BPMN-Modelle.** Damit verlässt man das
    gesamte Nomos-Versprechen: "Prozesse werden modelliert und im
    Cosmos versioniert." Temporal-Workflows leben in Go-Quellcode, nicht
    in einem reviewbaren Artefakt — das ist orthogonal zum
    Cosmos-Modell.
  - **Externer Server zwingend**: Temporal-Server + Postgres/Cassandra
    + UI. Selbst die "all-in-one"-Variante ist mehrere Prozesse. Bricht
    ADR-0006 und ADR-0003 ähnlich hart wie Option A.
  - **Audit-Modell passt nicht**: Temporal-History ist ein anderer
    Trace-Begriff als ADR-0017 (Inhalts-adressiert, hash-verkettet,
    Git-projiziert).
- **Wann doch sinnvoll**: Nie für Nomos selbst. Aber: Ein Adopter, der
  Nomos für Governance nutzt und seine *eigentliche* Provisioning-Logik
  in Temporal betreibt, ist ein legitimes Deployment-Muster — Nomos
  triggert dann nur, die Ausführung passiert extern. Dafür reicht
  jedoch der Trigger-Layer aus ADR-0018; Option B muss in Nomos selbst
  nicht eingebaut werden.

### Option C — Eigene Go-Engine, inkrementell *(gewählt)*

- **Pro**:
  - **Konsistenz mit ADR-0001/0003/0006/0017**: Single-Binary,
    Single-Process per Default, lokales Journal, deterministischer
    Trace gekoppelt an Nomos-Build-Hash.
  - **Trace-Determinismus**: Engine-Version = Nomos-Build-Hash.
    Reproduzierbarkeit hängt an genau einem Git-SHA, nicht an
    "Engine-X-Version-Y mit Konfiguration-Z".
  - **Inkrementelle Investition**: Stufe 1 ist mit
    Token-Game-Interpreter + NDJSON-Journal in ~3–4 Wochen erreichbar.
    Stufe 2/3 nur on demand.
  - **In-Process-Embedding für MCP**: Workflow-Aufruf aus
    MCP-Tool-Handler ohne RPC-Round-Trip (ADR-0008). Service-Tasks
    können direkt MCP-Tools aufrufen.
  - **Lizenzklar**: Bleibt unter Nomos-Lizenz, keine JVM, keine
    Sidecars.
  - **Modeler-Portabilität bleibt erhalten**: Wir verlangen *keine*
    Engine-spezifischen Namespaces im BPMN-XML; alles
    Runtime-Spezifische (Service-Task-Binding, User-Task-Form) bleibt
    in YAML neben dem BPMN (ADR-0018-Muster).
  - **Wiederverwendung der DMN-Engine** (ADR-0019): Sequence-Flow-Conditions,
    Script-Tasks und Business-Rule-Tasks teilen sich denselben
    FEEL-Interpreter.
- **Kontra**:
  - **BPMN 2.0 Executable ist groß**: ~30 ausführbare Element-Typen,
    Token-Semantik mit Edge-Cases (Inclusive-Gateway-Synchronisation,
    Compensation, Multi-Instance-Completion-Conditions). Vollkonformität
    wäre 6–12 Monate Vollzeit.
  - **Durable Execution ist anspruchsvoll**: Idempotenz, Crash-Recovery
    aus Journal, Garbage Collection alter Instanzen, Backpressure bei
    Massen-Triggern.
  - **BPMN-TCK gibt es nicht** (anders als DMN-TCK). Konformität bleibt
    eine Selbstaussage, gestützt durch Goldene-Modell-Tests.
  - **Operational Surface erweitert sich**: `$NOMOS_STATE_DIR`, Restart-Replay,
    Journal-Rotation sind neue Betriebskonzepte, die heute nicht
    existieren.
- **Risikomitigation**:
  - Konformitätsstufen pro Release **explizit** dokumentieren; Editor
    zeigt Badge je Prozess.
  - Stufe-1-Scope hart einfrieren, bevor Stufe 2 begonnen wird.
    Insbesondere keine Token-Splits jenseits von Parallel-Gateway, keine
    Boundary-Events in Stufe 1.
  - Journal-Format als versioniertes Schema (`format: nomos.journal.v1`)
    mit expliziter Forward-Migration für Folge-Versionen.
  - Goldene BPMN-Testmodelle (`internal/process/engine/testdata/`)
    decken jedes unterstützte Konstrukt mindestens je einen Happy-Path
    und einen Edge-Case.

### Option D — Status Quo (keine Runtime, nur Modellierung)

Trigger aus ADR-0018 werden gespeichert, aber von Nomos selbst nie
gefeuert; externe Adopter binden ihre eigene Engine an die
Trigger-YAML an.

- **Pro**: Null neuer Code im Kernpfad. Nomos bleibt klar als
  Governance-/Modellierungsschicht positioniert.
- **Kontra**:
  - Process-Trace (ADR-0017) bleibt für tatsächliche Ausführungen
    blind — jeder Adopter implementiert sein eigenes
    Trace-Schreibverfahren, die Hash-Kette ist nicht
    plattformseitig garantiert.
  - Das `nomos process run`-CLI-Versprechen aus dem Self-Model bleibt
    eine Wunschvorstellung.
  - Demos und Onboarding leiden: "Modellieren, aber nicht ausführen
    können" wirkt unfertig, ähnlich wie DMN-Modellieren ohne Eval (siehe
    Begründung ADR-0019).
- **Wann doch sinnvoll**: Wenn Nomos sich strategisch als reine
  *Modellierungs- und Audit-Plattform* positioniert und Ausführung
  konsequent als Adopter-Verantwortung deklariert wird. Dann sollte
  ADR-0018 entsprechend nachgeschärft werden ("Trigger sind reine
  Vertragsbeschreibungen").

### Option E — Hybrid Eigene-Engine + Offload für Long-Running

Stufe 1 wie in Option C, aber statt Stufe 2/3 selbst zu bauen, werden
Long-running-Prozesse über einen `bpmn:callActivity` an eine externe
Engine delegiert.

- **Pro**: Komplexität bleibt in Nomos klein; mature Engines übernehmen
  die harten Teile.
- **Kontra**:
  - Inkonsistenter Trace: ein Prozess teilt sich in einen
    Nomos-getracten Teil und einen extern-getracten Teil. Hash-Kette
    aus ADR-0017 bricht oder muss die externe Engine als
    Trust-Boundary anerkennen.
  - Operationaler Doppel-Stack ist nicht kleiner als Option A — nur
    weniger sichtbar.
- **Verworfen**: Schlechter als Option A *oder* Option C im Reinzustand;
  kombiniert Nachteile beider.

## Begründung der Wahl

Nomos ist seinem Wesen nach **ein in Git lebendes, sich selbst
beschreibendes Cosmos-Modell mit kryptografischer Audit-Spur**
(ADR-0001, ADR-0009, ADR-0017). Prozesse sind dabei — genau wie
Decisions (ADR-0019) — keine periphere Business-Logik, sondern
**First-Class-Governance-Artefakte**. Eine externe Engine als
Ausführungs-Endpunkt zwischenzuschalten verlagert den
Determinismus-Anker von "Git-Repo + Nomos-Binary" auf "Git-Repo +
Nomos-Binary + Engine-X-Version-Y-Konfiguration-Z" — derselbe
Trade-off wie bei DMN, mit denselben Konsequenzen.

Gleichzeitig wäre der Vollausbau einer BPMN-2.0-Executable-Engine ohne
konkreten Anwendungsbedarf reine YAGNI. Die inkrementelle Variante
(Option C) hält die Tür auf, ohne Kosten zu provozieren, die niemand
bezahlt: Sie liefert Stufe 1 in absehbarer Zeit, definiert Stufe 2/3
als optional und macht den jeweils unterstützten Umfang **zur
expliziten, versionierten Aussage** statt zu implizitem Verhalten.

Der wesentliche Unterschied zu ADR-0019 (DMN) ist die **Stateful-Natur**
von Workflow-Ausführung. DMN ist stateless — eine Evaluation hat einen
Input, einen Output, einen Trace. BPMN-Ausführung lebt: Token wandern,
Instanzen pausieren tagelang, externe Events landen. Das Journal-Modell
(Append-Only-NDJSON pro Instanz) ist die kleinste Antwort, die
ADR-0003 respektiert *und* Recovery erlaubt — keine Datenbank, kein
Schema-Migrationszwang, lesbar mit `cat`, diff-bar in Reviews bei
Bedarf.

## Implementierungsskizze Stufe 1

Dieser Abschnitt konkretisiert die gewählte Option C für die erste
Konformitätsstufe. Detaillierter Build-Plan in PR-Granularität liegt
parallel unter
`docs/implementation-prompts/0002-implement-process-workflow-engine-stage-1.md`.

### Paketstruktur

```
internal/process/engine/
├── definition/   — Parsed BPMN-Graph (immutable)
├── runtime/      — Token-Game-Interpreter + Instance-Lifecycle
├── journal/      — Append-Only-NDJSON-Reader/-Writer
├── scheduler/    — Timer-Heap + Cron, hinter Scheduler-Interface
├── eventbus/     — In-Process Pub/Sub, hinter Bus-Interface
├── handlers/     — Service-Task-Handler-Registry
└── manager/      — Top-Level: lädt Definitions, startet/recovered Instanzen
```

### Datenmodell

Drei klar getrennte Schichten:

1. **Definition** — aus BPMN-XML geparst, immutable, identisch über alle
   Instanzen derselben Prozessversion.
2. **Instance** — laufender Zustand, lebt in Memory, Snapshot im Journal.
3. **Event** — atomarer Zustandsübergang, append-only ins Journal.

Skizze der Kernstrukturen (`internal/process/engine/definition`):

```go
type Definition struct {
    ProcessID   string         // BPMN process id
    NomosID     string         // nomos process artefact id (PRC-...)
    Elements    map[string]Element
    Flows       map[string]SequenceFlow
    StartEvents []string       // element ids
}

type Element struct {
    ID            string
    Name          string
    Type          ElementType  // task, userTask, serviceTask, businessRuleTask,
                               // exclusiveGateway, parallelGateway,
                               // startEvent, endEvent, scriptTask
    Incoming      []string     // sequence flow ids
    Outgoing      []string
    TimerDef      *TimerDef    // nil unless startEvent+timer (Stufe 1)
    ScriptBody    string       // for scriptTask: FEEL expression
    DecisionRef   string       // for businessRuleTask: dmn decision id
    ServiceTaskBinding string  // for serviceTask: handler key from task_mapping
}

type SequenceFlow struct {
    ID            string
    Source, Target string
    Condition     string       // FEEL expression, empty = unconditional
}
```

Skizze (`internal/process/engine/runtime`):

```go
type Instance struct {
    UUID        string
    ProcessRef  string                  // nomos process id + version
    State       InstanceState           // running | waiting | completed | failed
    Tokens      map[string]Token        // active tokens by token-id
    Variables   map[string]any          // global instance scope (Stufe 1)
    StartedAt   time.Time
    LastEvent   uint64                  // monotonic seq from journal
    LastHash    []byte                  // ADR-0017 hash chain head
}

type Token struct {
    ID       string
    At       string                     // current element id
    Wait     *WaitState                 // nil = ready to step
    ScopeID  string                     // "root" in Stufe 1
}

type WaitState struct {
    Kind        WaitKind                // waitUserTask | waitTimer | waitSignal
    UserTask    *UserTaskWait
    Timer       *TimerWait              // FireAt time.Time, ScheduleID string
    Signal      *SignalWait             // Stufe 2
    AttemptID   string                  // idempotency key
}
```

### Token-Game-Semantik (Stufe 1)

Die Engine implementiert eine eingeschränkte, aber spec-konforme
Untermenge der BPMN-Token-Semantik (BPMN 2.0 §13). Was Stufe 1 leistet
und was bewusst weggelassen wird:

| Konstrukt                                                | Stufe 1 |
|----------------------------------------------------------|---------|
| Token erzeugen bei `startEvent` (plain + Timer)          | ✅ |
| Token verbrauchen bei `endEvent` (terminating implizit)  | ✅ |
| Sequence Flow ohne Condition: Token wandert              | ✅ |
| Sequence Flow mit FEEL-Condition (ADR-0019 Stufe 2)      | ✅ |
| `exclusiveGateway` Split: erste true-Condition gewinnt + `default` | ✅ |
| `exclusiveGateway` Merge: pass-through (jeder Token unabhängig) | ✅ |
| `parallelGateway` AND-Split: ein Token in → N Token out  | ✅ |
| `parallelGateway` AND-Join: wartet auf N Token in        | ✅ |
| `task` / `scriptTask` / `businessRuleTask` (synchron)    | ✅ |
| `userTask` (Wait-State + REST-Completion)                | ✅ |
| `serviceTask` (synchron, Handler-Registry)               | ✅ |
| `inclusiveGateway` (OR)                                  | ❌ Stufe 3 |
| `eventBasedGateway`                                      | ❌ Stufe 2 |
| `subProcess` / `callActivity`                            | ❌ Stufe 2 |
| `boundaryEvent` (Timer/Error/Message)                    | ❌ Stufe 2 |
| Intermediate `catchEvent` / `throwEvent`                 | ❌ Stufe 2 |
| Multi-Instance (sequential/parallel)                     | ❌ Stufe 3 |
| Compensation, Transaction                                | ❌ Stufe 3 |

**Step-Funktion** (Pseudo-Code, einzelner Mikro-Schritt):

```
func step(inst *Instance) StepResult:
    token := pickReadyToken(inst)
    if token == nil:
        return Idle   // alle Token warten, Instanz pausiert

    elem := def.Elements[token.At]
    switch elem.Type:
      case task, scriptTask, businessRuleTask:
          executeSync(elem, token)        // schreibt task_started + task_completed
          advanceToken(token, elem.Outgoing)
      case serviceTask:
          attemptID := newAttemptID()
          journal.Append(taskStarted{token, elem, attemptID})
          out, err := handlers.Invoke(elem.ServiceTaskBinding, ctx)
          journal.Append(taskCompleted{token, elem, attemptID, out, err})
          advanceToken(token, elem.Outgoing)
      case userTask:
          token.Wait = &WaitState{Kind: waitUserTask, ...}
          journal.Append(taskWaiting{token, elem})
          return Waiting
      case exclusiveGateway:
          flow := pickFirstTrueOrDefault(elem.Outgoing, inst.Variables)
          advanceTokenOver(token, flow)
      case parallelGateway:
          if isJoin(elem):
              if joinReady(inst, elem):
                  consumeIncomingTokens(inst, elem)
                  emitTokenOnAllOutgoing(inst, elem)
              else:
                  parkToken(token)
          else:                            // split
              consumeToken(token)
              for out in elem.Outgoing:
                  emitToken(inst, out)
      case startEvent:
          advanceToken(token, elem.Outgoing)
      case endEvent:
          consumeToken(token)
          if noTokensLeft(inst):
              journal.Append(instanceCompleted{...})
              inst.State = completed
    return Stepped
```

Wichtig: **jeder Mikro-Schritt schreibt sein Ergebnis ins Journal,
bevor er fortfährt.** Das heißt: Stepfunktion und Journal-Append sind
*innerhalb* der Instanz-Goroutine seriell, kein Pipeline-Buffering.

### Concurrency-Modell

- **Eine Goroutine pro aktiver Instanz**, mit einer Channel-Mailbox für
  externe Events (UserTask-Completion, Timer-Fire, externes Signal in
  Stufe 2).
- Die Goroutine ist ein Event-Loop:
  1. `step(inst)` bis `Idle` oder `Waiting`.
  2. Bei `Waiting`: blockierend auf Mailbox; bei eingehender Nachricht
     Wait-State auflösen und zurück zu 1.
- **Manager** (`internal/process/engine/manager`) verwaltet eine Map
  `instanceUUID → mailbox chan`, routet eingehende Events, startet/stoppt
  Instanzen, hält ein `RWMutex` für Definitions-Reload.
- Maximal-Concurrency wird via `engine.max_active_instances` in
  `nomos.yaml` begrenzt (Default 1000), bei Überschreiten landen neue
  Trigger in einer Backpressure-Queue mit Audit-Eintrag.

### Wait-States + Idempotenz

Drei Wait-Kinds in Stufe 1:

| Kind         | Eintritt                          | Austritt                           | Persistenz                      |
|--------------|-----------------------------------|------------------------------------|---------------------------------|
| `waitUserTask` | Token landet auf `userTask`     | `POST /api/v1/instances/{uuid}/signals` mit `task_id` + Variablen | `taskWaiting` im Journal |
| `waitTimer`  | Token landet auf `startEvent` mit Timer *oder* expliziter Timer-Service-Task (Stufe 1 nur Start) | Scheduler feuert zur `FireAt` | `timerScheduled` im Journal mit `fire_at` |
| `waitSignal` | (Stufe 2)                         | (Stufe 2)                          | (Stufe 2) |

**Idempotenz-Protokoll für Service-Tasks:**

- Vor dem Handler-Aufruf wird `taskStarted{attempt_id: <uuid>}`
  geschrieben.
- Beim Replay nach Crash gilt die Regel: Wenn ein `taskStarted` ohne
  zugehöriges `taskCompleted` im Journal liegt, wird der Handler mit
  **demselben** `attempt_id` erneut aufgerufen. Der Handler-Vertrag
  (Stufe-1-Bedingung) lautet: **Handler müssen idempotent gegenüber
  `attempt_id` sein.** Nicht-idempotente Handler werden über die
  Registry mit `Idempotency: HandlerOnly` markiert und im Recovery
  nicht automatisch wiederholt — stattdessen geht die Instanz in
  `failed`-State und braucht einen manuellen `POST .../resume`.
- Für UserTasks ist `attempt_id` der Schlüssel, mit dem der externe
  Completion-Call dedupliziert wird: identischer `attempt_id` +
  identischer Body = idempotent; abweichender Body = Konflikt 409.

**Crash-Recovery-Semantik** (Stufe 1, Single-Node):

1. Beim Start scannt `manager.Recover()` `$NOMOS_STATE_DIR/instances/`.
2. Pro Journal: Replay aller Events → Instanz-Speicher-Zustand
   rekonstruieren. Replay ist deterministisch, weil jedes Event den
   Outcome (Variablen-Diff, Token-Bewegung) bereits enthält — nicht
   nur den Auslöser.
3. Aus den letzten Events werden Wait-States re-armed: Timer landen
   wieder im Scheduler-Heap, UserTask-Mailboxen werden geöffnet,
   unvollendete ServiceTasks lösen Idempotent-Retry oder
   `failed`-State aus.
4. Erst nach abgeschlossenem Recovery werden externe Trigger-Quellen
   (Cron, HTTP-Trigger-Endpoint) angeschaltet — verhindert
   Doppel-Feuer in der Recovery-Phase.

### Service-Task-Handler-Modell

Service-Tasks rufen registrierte Go-Handler in-process auf. Die
Auflösung läuft über das **TaskMapping** aus der Prozess-YAML (heute
schon existent in `internal/app/process.go`):

```yaml
task_mappings:
  - bpmn_element_id: ServiceTask_CreateLDAPAccount
    binding: nomos.identity.ldap.create_account
    inputs:
      uid: "@employee_id"
      ou: "people"
    outputs:
      ldap_dn: "result.dn"
```

`binding` ist ein **stabiler Handler-Key**, kein Funktionspfad. Die
Registry (`internal/process/engine/handlers`) registriert Handler unter
diesem Key:

```go
package handlers

type Handler interface {
    Invoke(ctx context.Context, in HandlerInput) (HandlerOutput, error)
    Metadata() Metadata     // Idempotency, Stage, Description
}

type Metadata struct {
    Idempotency Idempotency  // Idempotent | HandlerOnly | NonIdempotent
    Stage       int          // earliest engine stage supported
    Description string
}

type Registry interface {
    Register(key string, h Handler) error
    Resolve(key string) (Handler, bool)
    List() []RegisteredHandler
}
```

Drei Quellen für Handler in Stufe 1:

1. **Built-in Go-Handler** (`internal/process/engine/handlers/builtin/`):
   `noop`, `log`, `http.call`, `git.commit`, `cosmos.create_artefact`
   — bewusst klein, deckt Self-Model-Use-Cases ab.
2. **MCP-Tools** (`internal/mcpserver/` als Brücke): Jedes registrierte
   MCP-Tool ist automatisch unter `mcp.<tool_name>` als Service-Handler
   verfügbar. Stage `1`, `Idempotency: HandlerOnly` (konservativ).
3. **DMN-Decisions** (`internal/dmn/evaluator.go`): Für
   `businessRuleTask` separater, nicht über die Service-Handler-Registry
   laufender Pfad — direkter In-Process-Call.

**Fehlerverhalten Stufe 1:** Jeder Handler-Fehler setzt die Instanz
auf `failed`. Kein automatisches Retry, kein Compensation. Operator
muss `POST /instances/{uuid}/resume` oder `/abort` aufrufen. Retry-Policies,
Error-Boundary-Events und Compensation sind Stufe-2-/Stufe-3-Themen.

### Variable Scoping + FEEL-Context

Stufe 1 hat **einen einzigen, globalen Scope pro Instanz** —
`inst.Variables` ist eine flache `map[string]any`. Die Begründung:
Sub-Prozesse, die eigenen Scope brauchen, sind ohnehin Stufe 2.

**Variablen-Schreibpfade:**

1. **Bei Instanz-Start:** initiale Variablen aus dem Trigger-Payload
   (Message-Event-Body, Manual-Start-Request-Body) übernommen,
   gefiltert durch `start_variables`-Whitelist im Prozess-YAML.
2. **Nach jedem Task:** Output-Mapping aus `task_mappings.outputs`
   schreibt zurück in `inst.Variables`. Mapping-Quelle ist ein
   FEEL-Pfad (`result.dn` → Wert aus Handler-Output).
3. **Per Script-Task:** `scriptTask` mit FEEL-Body, dessen
   Rückgabewert via `task_mappings.outputs` zugeordnet wird.

**FEEL-Kontext** (Wiederverwendung der ADR-0019-Engine):

- Condition auf Sequence Flow: Expression bekommt
  `{...inst.Variables, _meta: {...}}` als Kontext.
- Input-Mapping (`@employee_id` → wird als FEEL `employee_id`
  ausgewertet — das `@`-Präfix bleibt als ergonomische
  Variablen-Referenz erhalten, intern ist es ein einfacher
  FEEL-Pfad-Eval).
- Output-Mapping: FEEL-Pfad gegen den Handler-Output-Struct,
  Ergebnis landet unter dem Mapping-Key in den Instance-Variablen.

**Variable-Sichtbarkeit im Journal:** Jeder Schreibvorgang ist ein
explizites `variablesPatch`-Event mit Diff statt vollem Snapshot —
kompakt und auditierbar. Replay rekonstruiert den Volltext.

### Journal-Schema (`format: nomos.journal.v1`)

NDJSON, eine Zeile pro Event. Jede Zeile ist self-contained und
trägt die Hash-Kette für ADR-0017.

```json
{"seq":1,"ts":"2026-05-19T08:00:00Z","kind":"instance_created","instance":"i-abc","process_ref":"PRC-ACC-MBX-001@v3","trigger":"timer:StartEvent_NeuerEintritt","engine":{"version":"nomos-1.4.0","commit":"f4a5f42","stage":1},"prev_hash":null,"hash":"sha256:..."}
{"seq":2,"ts":"2026-05-19T08:00:00Z","kind":"token_emitted","instance":"i-abc","token":"t-1","at":"StartEvent_NeuerEintritt","prev_hash":"sha256:...","hash":"sha256:..."}
{"seq":3,"ts":"2026-05-19T08:00:00Z","kind":"variables_patch","instance":"i-abc","set":{"employee_id":"E-42"},"prev_hash":"sha256:...","hash":"sha256:..."}
{"seq":4,"ts":"2026-05-19T08:00:01Z","kind":"task_started","instance":"i-abc","token":"t-1","element":"ServiceTask_CreateLDAPAccount","attempt_id":"a-1","binding":"nomos.identity.ldap.create_account","prev_hash":"sha256:...","hash":"sha256:..."}
{"seq":5,"ts":"2026-05-19T08:00:02Z","kind":"task_completed","instance":"i-abc","token":"t-1","attempt_id":"a-1","output":{"dn":"cn=E-42,ou=people,dc=ex,dc=com"},"prev_hash":"sha256:...","hash":"sha256:..."}
{"seq":6,"ts":"2026-05-19T08:00:02Z","kind":"variables_patch","instance":"i-abc","set":{"ldap_dn":"cn=E-42,ou=people,dc=ex,dc=com"},"prev_hash":"sha256:...","hash":"sha256:..."}
{"seq":7,"ts":"2026-05-19T08:00:02Z","kind":"token_moved","instance":"i-abc","token":"t-1","from":"ServiceTask_CreateLDAPAccount","to":"EndEvent_Done","prev_hash":"sha256:...","hash":"sha256:..."}
{"seq":8,"ts":"2026-05-19T08:00:02Z","kind":"instance_completed","instance":"i-abc","prev_hash":"sha256:...","hash":"sha256:..."}
```

Event-Kinds Stufe 1 (geschlossene Liste):

`instance_created`, `instance_completed`, `instance_failed`,
`token_emitted`, `token_moved`, `token_consumed`,
`task_started`, `task_completed`, `task_failed`,
`task_waiting`, `task_resumed`,
`timer_scheduled`, `timer_fired`,
`gateway_decided`, `variables_patch`.

`hash` ist `sha256(prev_hash || canonical_json(event_without_hash))`.
Damit ist das Journal **gleichzeitig** der Trace-Strom aus ADR-0017 —
kein paralleler Trace-Schreiber.

### Scheduler

`internal/process/engine/scheduler` hinter einem Interface, in Stufe 1
nur In-Process-Implementierung:

```go
type Scheduler interface {
    ScheduleAt(t time.Time, payload TimerPayload) (ScheduleID, error)
    ScheduleCron(spec string, payload CronPayload) (ScheduleID, error)
    Cancel(id ScheduleID) error
    Start(ctx context.Context) error    // startet Wake-Loop
}
```

Implementierung:

- **Min-Heap** indiziert auf `FireAt`, ein `sync.Cond` wacht auf,
  wenn `time.Until(head) <= 0`.
- Bei `Start()` wird zuerst das Journal nach offenen
  `timer_scheduled` ohne `timer_fired` durchsucht und der Heap
  vorbefüllt — verhindert Verlust von Timern über Restart.
- Cron-Triggers (aus ADR-0018) werden als rekursive `ScheduleAt`
  modelliert: nach jedem Feuern wird die nächste Iteration eingeplant.
  Die Cron-Validierung aus ADR-0018 wird zur tatsächlichen
  Cron-Evaluation aufgewertet — hier wird `robfig/cron/v3` als
  *interne* Dependency akzeptiert (nicht öffentliche API).

Stufe 2 ersetzt die In-Process-Impl durch einen K8s-CronJob- oder
externen-Scheduler-Adapter hinter demselben Interface.

### Trace-Integration mit ADR-0017

Da das Journal selbst hash-verkettet ist, fällt der Trace mit dem
Journal zusammen. Zusätzlich:

- Bei `instance_completed` wird **optional** ein verdichtetes
  Trace-Artefakt nach Git geschrieben
  (`.nomos/traces/processes/<process-id>/<yyyy>/<mm>/<instance-uuid>.json`),
  das nur die Element-Sequenz + finale Variablen + Final-Hash
  enthält. Konfiguration: `engine.persist_traces_to_git: false` per
  Default (operativer Lärm), opt-in für Audit-strenge Deployments.
- Das vollständige NDJSON-Journal bleibt **immer** außerhalb von Git
  unter `$NOMOS_STATE_DIR` — operativ, rotierend, nicht
  Quelle-der-Wahrheit.

### Was Stufe 1 explizit NICHT kann

Damit niemand falsche Erwartungen hat:

- Keine Retries auf Service-Task-Fehlern.
- Keine Boundary-Events, keine Error-Handler-Flows.
- Keine Cancellation eines wartenden Tokens durch ein konkurrierendes
  Event (Stufe 2: `eventBasedGateway`).
- Keine Sub-Prozesse, keine Call-Activity.
- Kein Multi-Node-Setup; pro `$NOMOS_STATE_DIR` ein aktiver Writer.
  Mehrere Nomos-Instanzen auf demselben `state_dir` führen zu
  Datenkorruption. Lock-File (`state.lock`) verhindert versehentlichen
  Parallel-Start.
- Keine Migration eines laufenden Prozess-Workflows zwischen
  BPMN-Versionen ("Process Migration"). Eine neue BPMN-Version startet
  neue Instanzen; laufende Instanzen fahren auf ihrer Pin-Version
  weiter.

## Konsequenzen

- Neuer ADR-Pfad `docs/architecture/adr/ADR-0021-...md` ersetzt die
  "Out-of-Scope"-Klausel aus ADR-0018, Abschnitt "Runtime/Scheduler-Architektur".
- Neues Package `internal/process/engine` mit Sub-Paketen:
  - `internal/process/engine/parser` — BPMN-XML → interne Element-Graph-Darstellung
    (wiederverwendet/ergänzt `internal/app/process_triggers.go` für Start-Events)
  - `internal/process/engine/runtime` — Token-Game-Interpreter
  - `internal/process/engine/journal` — NDJSON-Journal-Reader/-Writer mit
    Schema-Version
  - `internal/process/engine/scheduler` — In-Process-Timer/Cron, lebt
    hinter einem `Scheduler`-Interface zwecks Stufe-2-Austausch
  - `internal/process/engine/eventbus` — In-Process-Bus mit `Bus`-Interface
    (Stufe-2-Adapter z. B. für NATS hinter demselben Interface)
- DMN-Aufruf aus `bpmn:businessRuleTask` läuft direkt über
  `internal/dmn/evaluator.go` — kein eigener Klon, kein RPC.
- Neue REST-Endpoints in `internal/server/server.go` (Skizze, Details
  folgen):
  - `POST   /api/v1/processes/{id}/instances`     — Manueller Start
  - `GET    /api/v1/processes/{id}/instances`     — Liste laufender Instanzen
  - `GET    /api/v1/instances/{uuid}`             — Token-Stand + History
  - `POST   /api/v1/instances/{uuid}/signals`     — User-Task-Completion / externes Signal
  - `POST   /api/v1/events`                       — Generischer Event-Receiver (Stufe 2)
- Neue CLI-Subcommands `nomos process run`, `nomos process instances`,
  `nomos process inspect <uuid>` (analog `nomos decision eval` aus
  ADR-0019).
- Neue Konfiguration in `nomos.yaml`:
  ```yaml
  engine:
    state_dir: .nomos/state          # Default ./nomos/state, NICHT in Git
    journal_format: nomos.journal.v1
    scheduler: in-process            # Stufe 2: external
    eventbus: in-process             # Stufe 2: nats
  ```
- `.gitignore` muss `.nomos/state/` per Default ergänzen — Instanzzustand
  ist operativ, gehört nicht in den Quell-of-Truth-Repo.
- Process-Trace-Schema in ADR-0017 wird erweitert um
  `instance_uuid`, `engine.stage`, `engine.commit` (letzteres deckt
  Engine-Versionswechsel implizit ab).
- Editor-Integration in `internal/server/web/templates/cosmos.html`:
  Konformitäts-Badge auf dem Process-Detail-Panel ("Engine: BPMN Subset
  L1") + Live-Instanz-Liste pro Prozess (Stufe 1 reicht für
  Polling-View).
- **Kein JVM** im Distributionspfad, **kein zwingender externer Store**,
  **kein Modeler-spezifischer Namespace** im BPMN — ADR-0001, ADR-0003,
  ADR-0006 und ADR-0018 bleiben unverletzt.

## Offene Punkte

- **Korrelations-Schlüssel**: Wie wird ein eintreffendes Message-Event
  einer wartenden Instanz zugeordnet (Stufe 2)? Vorschlag:
  Correlation-Key als FEEL-Expression über Event-Payload + Instance-Variables,
  analog Zeebe's `correlationKey`. Konkrete Syntax bleibt für ein
  Folge-ADR (Stufe-2-Start).
- **Journal-Rotation und Cleanup**: Wann werden abgeschlossene
  Instanzen aus dem Journal entfernt bzw. archiviert? Vorschlag:
  Retention pro Prozess in YAML (`retention: 30d`), Default 90 Tage,
  Archiv als gzip-NDJSON unter `$NOMOS_STATE_DIR/archive/`.
- **Multi-Node-Deployment**: Stufe 1 ist single-node-only. Stufe 2 müsste
  Leader-Election oder Sharding-by-Process-ID klären — explizit kein
  Stufe-1-Ziel, aber das Journal-Schema sollte das nicht verbauen
  (z. B. durch frühe Node-ID-Felder im Journal-Header).
- **BPMN-Subset-Validator**: Analog zum DMN-Konformitäts-Badge braucht
  es eine Validierung, die BPMN-XML auf "Stufe-N-ausführbar" prüft und
  bei nicht-unterstützten Konstrukten ein Finding setzt (`PROCESS_UNSUPPORTED_ELEMENT`,
  Severity warning oder error je nach Konfiguration).
- **Bestehende externe Go-Bibliotheken**: Eine kurze Marktsichtung
  (`github.com/MarcGrol/go-workflow`, `github.com/bpmn-io/...`,
  `github.com/nitram509/...`) hat keine produktionsreife, BPMN-2.0-konforme
  Go-Engine ergeben. Falls eine reife Apache/MIT-lizenzierte Lib
  auftaucht, ist Option C′ ("forken statt selbst schreiben") eine
  Variante, die diese ADR explizit erlaubt.
- **MCP-Service-Task-Binding**: Wie genau ruft ein `bpmn:serviceTask`
  ein MCP-Tool auf — synchron in-process oder über die MCP-Schnittstelle?
  Empfehlung: in-process direkt, mit demselben Tool-Resolver wie
  `internal/mcpserver/`. Folge-ADR, sobald Service-Plugin-Modell
  (ADR-0011) stabilisiert ist.
- **Race zur Stufe-2-Eventbus-Wahl**: NATS-JetStream als Default-Adapter
  ist eine Vermutung, kein Beschluss. Alternative Kandidaten: Redis
  Streams, eingebettetes nats-server (`github.com/nats-io/nats-server/v2`),
  Webhooks-only. Entscheidung folgt mit dem Stufe-2-Trigger.
- **Backpressure**: Was passiert bei 10 000 gleichzeitigen
  Timer-Triggers um 08:00? Stufe 1 darf "fail loudly" sein
  (Queue-Limit, danach Drop mit Audit-Eintrag); Stufe 2 braucht echte
  Rate-Limit-Politik.
