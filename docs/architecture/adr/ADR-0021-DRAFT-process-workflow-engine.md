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
