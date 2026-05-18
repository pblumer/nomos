# ADR-0018 - Process Trigger Events: Modell-Layer für Timer- und Message-Starts

## Status

Draft

## Datum

2026-05-18

## Kontext

BPMN-Prozesse in Nomos (`internal/app/process.go`) sind heute reine
Modellierungsartefakte. Ein Prozess hat genau ein implizites Start-Event,
und wann er „ausgeführt“ wird, steht nirgends — weder im Modell noch in
einer Runtime.

Der bpmn-js-Editor erlaubt allerdings bereits, einen Plain Start Event in
einen **Message-, Timer-, Conditional- oder Signal-Start** umzuwandeln
(`bpmn:timerEventDefinition` etc. als Kind-Element). Diese Information
landet zwar im BPMN-XML, aber:

- **Es gibt keinen Ort für die Konfiguration:** Eine Timer-Start braucht
  einen Cron-Ausdruck oder ISO-8601-Intervall, eine Message-Start einen
  Event-Topic. BPMN-XML kennt dafür nur untypisierte
  Implementer-Attribute (z.B. `camunda:`, `zeebe:`-Namespaces).
- **Es gibt keine Validierung:** Falsche Cron-Syntax, fehlender Event-Ref
  oder ein YAML-Trigger, der auf einen nicht-existierenden Start-Event
  zeigt, bleibt unentdeckt.
- **Es gibt keine Runtime-Abstraktion:** Es ist nicht klar, ob Nomos
  selbst einen Scheduler/Event-Bus betreibt oder Trigger an eine externe
  Plattform delegiert.

Ziel dieses ADR: den **Modell-Layer** so festzulegen, dass Trigger
- explizit, reviewbar und Git-first im Cosmos liegen,
- vom BPMN-Element-Typ konsistent abgeleitet (nicht widersprüchlich) sind,
- eine **stabile Schnittstelle** für eine spätere Runtime bilden, ohne
  diese vorwegzunehmen.

Die Runtime selbst (Scheduler, Event-Store-Anbindung) bleibt **explizit
Out-of-Scope dieses ADR** und folgt in einem separaten ADR (siehe
„Offene Punkte“).

## Entscheidung

### Trigger-Modell

Jeder Prozess kann pro Start-Event genau einen `ProcessTrigger`
deklarieren. Triggers werden als `triggers:` Block in der Process-YAML
gespeichert (analog zu `task_mappings`). Der Block referenziert ein
`bpmn_element_id` aus dem zugehörigen `.bpmn`-File.

```yaml
id: PRC-ACC-MBX-001
type: process
name: Provision Benutzeraccount
related_product: PROD-ACC-MBX-001
bpmn:
  file: prc-acc-mbx-001.bpmn
  process_id: Process_ProvisionAccount
  primary: true

triggers:
  - bpmn_element_id: StartEvent_NeuerEintritt
    type: message
    message:
      event_ref: identity.user.created
      topic: identity/user-events
      filter: "tenant_id == 'main'"

  - bpmn_element_id: StartEvent_WoechentlicherCheck
    type: timer
    timer:
      cron: "0 8 * * 1"            # Montags 08:00

  - bpmn_element_id: StartEvent_NachAblauf
    type: timer
    timer:
      iso_duration: PT24H          # 24h relativ
```

### Trigger-Typen

| `type`        | BPMN-Element                                              | Pflichtfelder           | Optionale Felder            |
|---------------|-----------------------------------------------------------|-------------------------|-----------------------------|
| `none`        | `bpmn:startEvent` ohne Definition                         | —                       | —                           |
| `timer`       | `bpmn:startEvent` mit `bpmn:timerEventDefinition`         | einer von `cron`, `iso_duration`, `iso_date` | `timezone` |
| `message`     | `bpmn:startEvent` mit `bpmn:messageEventDefinition`       | `event_ref`             | `topic`, `filter`, `correlation_key` |
| `signal`      | `bpmn:startEvent` mit `bpmn:signalEventDefinition`        | `signal_ref`            | —                           |
| `conditional` | `bpmn:startEvent` mit `bpmn:conditionalEventDefinition`   | `expression`            | —                           |

### Quelle der Wahrheit

- **Trigger-Typ** ergibt sich aus dem BPMN-XML
  (`<bpmn:timerEventDefinition/>` etc.). Die YAML wiederholt diesen Typ
  redundant, damit Reviews ohne BPMN-XML-Diff lesbar bleiben — bei
  Abweichung gewinnt das BPMN-XML, die YAML produziert ein Finding.
- **Trigger-Konfiguration** (Cron, Event-Ref, Filter) lebt
  ausschließlich in der YAML — nicht in implementer-spezifischen
  Attributen des BPMN-XML. Damit bleibt das BPMN portabel zwischen
  Modelern (bpmn.io, Camunda, Zeebe, …) und Konfiguration in YAML
  reviewbar.

### Extraktion und API

- `ExtractBPMNStartEvents(xml)` → `[]StartEventInfo{ID, Name, Type}`.
  Wird beim Lesen und beim `UpdateProcessBPMN` aufgerufen.
- `ProcessDTO.Triggers` enthält die effektive, gemerkte Sicht
  (BPMN-Element ∪ YAML-Konfiguration ∪ Validierungs-Status).
- `PUT /api/v1/processes/{id}/triggers` mit Body
  `{triggers: [ProcessTrigger,…]}` schreibt die YAML.
- `GET  /api/v1/processes/{id}/triggers` listet sowohl konfigurierte als
  auch unkonfigurierte (im BPMN existierende) Trigger.

### Validierung

| Code                              | Severity | Bedingung                                                    |
|-----------------------------------|----------|--------------------------------------------------------------|
| `TRIGGER_UNKNOWN_ELEMENT`         | error    | YAML zeigt auf BPMN-Element-ID, die im XML nicht existiert   |
| `TRIGGER_TYPE_MISMATCH`           | error    | YAML-`type` ≠ aus BPMN-XML abgeleiteter Typ                  |
| `TRIGGER_TIMER_NO_SCHEDULE`       | warning  | Timer-Trigger ohne `cron`/`iso_duration`/`iso_date`          |
| `TRIGGER_TIMER_INVALID_CRON`      | error    | `cron` lässt sich nicht in 5 oder 6 Felder parsen            |
| `TRIGGER_MESSAGE_NO_REF`          | warning  | Message-Trigger ohne `event_ref`                             |
| `TRIGGER_SIGNAL_NO_REF`           | warning  | Signal-Trigger ohne `signal_ref`                             |
| `TRIGGER_CONDITIONAL_NO_EXPR`     | warning  | Conditional-Trigger ohne `expression`                        |
| `TRIGGER_TYPED_START_UNCONFIGURED`| warning  | BPMN hat typisiertes Start-Event, aber keine Trigger-YAML    |

### Cron-Format

Akzeptiert wird das **klassische 5-Felder-Cron-Format** (`min hour dom
mon dow`) und optional ein 6-Felder-Format mit führendem Sekunden-Feld
(Quartz-kompatibel). Pro Feld:
- `*`, `*/N`, `A-B`, `A-B/N`, `A,B,C`
- Ganze Zahlen im jeweils gültigen Bereich
- Macros (`@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly`) werden
  als Single-Token erkannt und intern auf 5-Felder expandiert

Eine separate Cron-Library wird **bewusst nicht** als Dependency
eingeführt; die Validierung ist syntaktisch, nicht Schedule-evaluierend.
Erst die Runtime (Folge-ADR) entscheidet, ob `github.com/robfig/cron/v3`
oder ein externer Scheduler genutzt wird.

## Begründung

- **Git-first**: Trigger-Konfiguration ist reviewbar in PRs, ohne dass
  ein BPMN-Modeller läuft. Eine Änderung von `cron: "0 8 * * 1"` auf
  `cron: "0 9 * * 1"` ist ein 1-Zeilen-YAML-Diff.
- **Modeler-Portabilität**: Wir vermeiden bewusst die
  `camunda:`/`zeebe:`-Extension-Attribute im BPMN. Damit bleibt das XML
  zwischen Modelern austauschbar (siehe ADR-0002, ADR-0001).
- **Stabile Runtime-Schnittstelle**: Eine spätere Runtime (Cron-Daemon,
  Event-Store-Abonnent) konsumiert `ProcessTrigger` direkt — ohne
  BPMN-XML-Parsing zu wiederholen.
- **Keine Runtime-Vorwegnahme**: Wir entscheiden **nicht** in diesem
  ADR, ob Timer in-process via Goroutine, als K8s-CronJob, oder via
  externem Scheduler laufen. Auch der Event-Store (NATS, Kafka,
  in-process Bus, Webhook-Empfänger) bleibt offen. Dies isoliert die
  Modell-Entscheidung von Infrastruktur-Trade-offs.
- **Lokale Cron-Validierung**: Verhindert offensichtliche Tippfehler,
  ohne eine externe Lib einzuziehen oder Scheduling-Semantik
  vorwegzunehmen.

## Konsequenzen

- Neue Felder in `internal/model/model.go`: `ProcessTrigger`,
  `TimerTriggerConfig`, `MessageTriggerConfig`, `SignalTriggerConfig`,
  `ConditionalTriggerConfig` und `Triggers []ProcessTrigger` auf
  `Process`.
- Neuer Block in `internal/app/process.go`: `ExtractBPMNStartEvents`,
  `UpdateProcessTriggers`, Trigger-Validierungen, Integration in
  `processDTO` und `validateProcessDTO`.
- Neue REST-Endpoints in `internal/server/server.go`:
  - `GET /api/v1/processes/{id}/triggers`
  - `PUT /api/v1/processes/{id}/triggers`
- DTOs in `internal/app/dto.go`: `ProcessTriggerDTO`,
  `UpdateProcessTriggersRequest`.
- Tests in `internal/app/process_triggers_test.go` decken Extraktion,
  Round-Trip, Validierung und Cron-Syntax ab.
- Bestehende Prozesse bleiben kompatibel: Ohne `triggers:`-Block hat ein
  Plain-Start-Event Typ `none` und erzeugt kein Finding.

## Out-of-Scope (für separates ADR)

- **Runtime/Scheduler-Architektur** (in-process vs. K8s-CronJob vs.
  externer Scheduler, Idempotenz, Catch-up nach Downtime, Backfill).
- **Event-Store-Abstraktion** (Protokoll, Topic-Naming-Konventionen,
  At-least-once vs. Exactly-once, Replay).
- **Korrelation** zwischen mehreren Triggers eines Prozesses und
  laufenden Instanzen.
- **Authorization**: wer darf welche Trigger feuern bzw. Cron-Ausdrücke
  ändern.

## Offene Punkte

- **UI**: Ein Trigger-Editor im Side-Panel des BPMN-Editors, der beim
  Auswählen eines Start-Events Cron-/Event-Ref-Felder anzeigt, ist in
  einem Folge-PR vorgesehen. Bis dahin können Trigger über die REST-API
  bzw. direkten YAML-Edit gepflegt werden.
- **CLI**: `nomos process trigger set …` analog zu den anderen
  Process-Subcommands ist optional und folgt zusammen mit der UI.
- **OpenAPI**: Die neuen Endpoints sind aktuell nicht in
  `internal/server/openapi.go` dokumentiert; nachzuführen, sobald die
  Shape stabil ist.
- **Mehrere Triggers pro Element**: Aktuell maximal einer pro Start-Event
  (1:1 mit BPMN-Definition). Falls später mehrere Cron-Ausdrücke
  derselben Logik nötig werden, ist eine `schedules: []`-Liste innerhalb
  von `TimerTriggerConfig` der natürliche Erweiterungspfad.
