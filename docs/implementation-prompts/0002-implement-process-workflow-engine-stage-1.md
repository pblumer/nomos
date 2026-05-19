# Implementation Plan: ADR-0021 Process/Workflow Engine — Stufe 1

You are working in the Nomos repository.

## Goal

Implement Stage 1 of the BPMN process/workflow engine as decided in
`docs/architecture/adr/ADR-0021-DRAFT-process-workflow-engine.md`. The
engine executes the synchronous happy-path subset of BPMN 2.0 in-process,
persists instance state via an append-only NDJSON journal under
`$NOMOS_STATE_DIR`, and integrates with the existing DMN engine
(ADR-0019), the trigger model (ADR-0018) and the hash-chained trace
(ADR-0017).

This document breaks the build down into PR-sized increments. Each PR
ends in a green `make test` and a runnable demo.

## Read first

Before changing code, read:

- `docs/architecture/adr/ADR-0021-DRAFT-process-workflow-engine.md` — full design
- `docs/architecture/adr/ADR-0018-DRAFT-process-trigger-events.md` — trigger model
- `docs/architecture/adr/ADR-0019-dmn-feel-engine-scope.md` — DMN engine + FEEL stages
- `docs/architecture/adr/ADR-0017-DRAFT-decision-traceability.md` — trace hash chain
- `internal/app/process.go` and `internal/app/process_triggers.go` — current process model
- `internal/dmn/evaluator.go` — DMN evaluator entry point
- `internal/mcpserver/` — MCP tool registry (for service-task bridge)

## Conventions

- New code lives under `internal/process/engine/`. Do not extend the
  monolithic `internal/app/process.go`; keep modeling and runtime
  cleanly separated.
- All public types in `internal/process/engine` carry doc comments that
  reference the ADR section they implement.
- Test data (golden BPMN files, expected journals) lives in
  `internal/process/engine/<pkg>/testdata/`.
- No new third-party dependencies *except* `github.com/robfig/cron/v3`,
  and only inside `internal/process/engine/scheduler` (never exposed in
  public APIs).
- All journal/trace timestamps use UTC RFC3339Nano. All `seq` values
  are monotonically increasing per-instance starting at 1.

## PR 1 — Engine skeleton and definition parser

**Scope.** Create the package layout, build the BPMN → `Definition`
parser, expose no runtime yet.

- `internal/process/engine/definition/` with `Definition`, `Element`,
  `ElementType`, `SequenceFlow`, `TimerDef` (signatures as in ADR-0021
  "Datenmodell").
- `Parse(xml []byte) (*Definition, []Finding, error)` reusing the
  `ExtractBPMNStartEvents` helper from
  `internal/app/process_triggers.go` (move it here if cleaner, with a
  shim to keep the existing trigger code working).
- Recognize Stage 1 element types only. Anything else is parsed into a
  generic `unsupported` element with a `Finding` of code
  `PROCESS_UNSUPPORTED_ELEMENT_STAGE1` (severity warning).
- Unit tests with at least: linear flow, exclusive gateway with
  default, parallel split+join, user task, service task,
  business-rule task, plain start, timer start, end event.

**Done when.** `go test ./internal/process/engine/definition/...`
passes; existing process tests still green; no new public types
exposed from `internal/app`.

## PR 2 — Token-game runtime (in-memory, no journal)

**Scope.** Implement the step function described in ADR-0021
"Token-Game-Semantik". Pure in-memory; one goroutine per instance;
synchronous service-task handlers via a stub registry.

- `internal/process/engine/runtime/` with `Instance`, `Token`,
  `WaitState`, `State` types.
- `Step(inst *Instance, def *Definition) StepResult` implementing the
  switch from the ADR.
- Reuse `internal/dmn` for `businessRuleTask` and for sequence-flow
  FEEL conditions. Wire a minimal `Evaluator` interface so the runtime
  package does not import `internal/dmn` directly (avoid cycles).
- `internal/process/engine/handlers/` with the `Registry`, `Handler`,
  `Metadata` interfaces from ADR-0021. Add three built-in handlers in
  `handlers/builtin/`: `noop`, `log`, `cosmos.read_artefact` (the
  read-only one is safe and useful for the demo).
- Golden tests in `runtime/testdata/`: each test is a BPMN file +
  initial variables + expected element-visit sequence + expected
  final variables. Aim for ≥ 12 cases covering all Stage 1 elements
  and exclusive-/parallel-gateway edge cases (default flow, all-false,
  join with extra incoming tokens).

**Done when.** Tests green; a single test can drive an instance to
completion with at least one each of `serviceTask`, `userTask`
(externally completed via direct `Resume()` call on the instance),
`businessRuleTask`, `parallelGateway` split+join,
`exclusiveGateway`.

## PR 3 — Journal: append, replay, recovery

**Scope.** Make the engine durable. Every state transition becomes an
event in `$NOMOS_STATE_DIR/instances/<process-id>/<instance-uuid>.journal`.

- `internal/process/engine/journal/` with `Writer`, `Reader`, event
  struct types matching the ADR-0021 "Journal-Schema" exactly. Encode
  with canonical JSON (`encoding/json` with a stable key order helper)
  so the hash chain is reproducible.
- Hash-chain implementation: `Hash(prevHash []byte, eventBytes []byte) []byte`
  using SHA-256. The hash field on each event is the hash *of* that
  event including its `prev_hash`, computed before append.
- `Append(event Event) error` is the only write path. It fsyncs by
  default (config: `engine.journal_fsync: true`).
- Format version `nomos.journal.v1` recorded in the first event
  (`instance_created.format`). Reader rejects unknown formats with a
  clear error.
- Runtime is changed so that each `Step` decision routes through
  `journal.Append` *before* visible state changes commit in memory.
  Replay path reads events and applies them to a fresh `Instance`.
- `Recover(ctx) ([]*Instance, error)` on the manager scans the state
  dir, replays each journal, returns reconstructed instances.
- Tests: round-trip (run → snapshot journal → replay → identical
  state), hash-chain integrity (mutating any event byte breaks the
  chain at that line and onwards), partial-write tolerance (truncated
  last line is dropped with a warning, not a fatal).

**Done when.** A test runs an instance, kills the engine mid-flow
(simulated by closing the writer), restarts via `Recover`, finishes
the instance, asserts the journal is internally consistent and the
final hash matches a recomputed chain.

## PR 4 — Service-task handler registry and MCP bridge

**Scope.** Make service tasks useful by wiring the registry to real
handlers, including MCP tools.

- Flesh out `internal/process/engine/handlers/` with:
  - Built-in handlers: `noop`, `log`, `http.call` (with allow-list
    config), `cosmos.read_artefact`. Each declares Stage 1 +
    appropriate `Idempotency` value.
  - `internal/process/engine/handlers/mcp/` bridge: enumerates the
    MCP tool registry at engine boot, exposes each tool as
    `mcp.<tool_name>` with `Idempotency: HandlerOnly`.
- Honor the existing `task_mappings.inputs` / `task_mappings.outputs`
  FEEL mapping shape. The mapping evaluator lives next to the DMN
  evaluator (shared FEEL Stage 2).
- Idempotency protocol from ADR-0021 implemented end-to-end:
  `task_started` carries `attempt_id`; on recovery, an in-flight
  service-task is retried with the *same* `attempt_id` if its handler
  declares `Idempotent`, otherwise the instance transitions to
  `failed` and waits for operator action.
- Validation finding `TASK_MAPPING_UNKNOWN_BINDING` when a process
  YAML references a handler key that is not in the registry.
- Tests: a service task wired to `noop` succeeds; one wired to a
  handler that returns an error transitions the instance to `failed`
  and writes a `task_failed` + `instance_failed` event pair; recovery
  of an idempotent in-flight call replays with the same `attempt_id`.

**Done when.** Demo flow with an MCP-tool-backed service task runs
end-to-end. `nomos process inspect <uuid>` (added in PR 6) shows the
handler key and the resolved output.

## PR 5 — Scheduler and trigger wiring

**Scope.** Make timer-start events from ADR-0018 actually fire, and
make manual + REST-triggered starts work.

- `internal/process/engine/scheduler/` with the `Scheduler` interface
  from ADR-0021 and an in-process implementation using `container/heap`
  + `time.Timer`. Cron expressions resolved with
  `github.com/robfig/cron/v3` *internally*; no public dependency leak.
- Recovery loads outstanding `timer_scheduled` events whose
  `timer_fired` does not yet exist and re-arms them.
- Trigger pipeline: at engine start, walk all processes via the
  existing process loader, register each configured trigger from
  ADR-0018 with the scheduler / event bus.
- In-process event bus (`internal/process/engine/eventbus/`) with the
  `Bus` interface. Stage 1 implementation is a single-process pubsub;
  Stage 2 will swap in NATS behind the same interface.
- `POST /api/v1/events` endpoint accepts external message-style events
  and routes them through the bus (Stage 1: events with no matching
  process trigger are logged and dropped, no queueing).
- Tests: a process with a `cron: "*/1 * * * *"` trigger fires once
  inside a fake-clock harness; a timer scheduled for `+10s` survives
  a simulated restart at `+5s` and still fires at `+10s` of the new
  process's clock.

**Done when.** Running `nomos serve` with a process whose start event
has `iso_duration: PT5S` produces an instance after five seconds and
records the full trigger → instance → completion chain in the journal.

## PR 6 — REST API and CLI

**Scope.** Make the engine observable and steerable from outside.

- New endpoints in `internal/server/server.go`:
  - `POST   /api/v1/processes/{id}/instances`             — start
  - `GET    /api/v1/processes/{id}/instances`             — list
  - `GET    /api/v1/instances/{uuid}`                     — detail
  - `GET    /api/v1/instances/{uuid}/journal`             — raw NDJSON
    stream (auth-gated, opt-in via `engine.expose_journal_api: false`
    default)
  - `POST   /api/v1/instances/{uuid}/signals`             — user-task
    completion (`task_id`, `variables`, `attempt_id`)
  - `POST   /api/v1/instances/{uuid}/resume`              — re-arm a
    `failed` instance after operator fix
  - `POST   /api/v1/instances/{uuid}/abort`               — terminal
- Matching DTOs in `internal/app/dto.go`.
- CLI subcommands under `cmd/nomos/`:
  - `nomos process run <process-id> [--var k=v ...]`
  - `nomos process instances [<process-id>]`
  - `nomos process inspect <uuid>`
  - `nomos process signal <uuid> <task-id> [--var k=v ...]`
- OpenAPI updates in `internal/server/openapi.go`.
- Tests: REST + CLI integration tests that boot the engine against a
  temp `state_dir`, run a process end-to-end and tear down cleanly.

**Done when.** `nomos process run` from the demo cosmos drives an
instance to completion with visible journal output via `nomos process
inspect`.

## PR 7 — UI integration: conformance badge and live instance list

**Scope.** Surface the engine in the Cosmos Explorer so modelers see
what is actually executable.

- Conformance badge on the process detail panel in
  `internal/server/web/templates/cosmos.html`: "Engine: BPMN Subset L1"
  (read from a single `engineStage()` helper, sourced from a build-time
  constant in `internal/version`).
- Per-process "Active instances" sub-panel: polling every 5 s on
  `GET /api/v1/processes/{id}/instances`, table with UUID, state,
  current element, started-at.
- Per-instance detail view (linked from the table) showing the journal
  in a human-readable form (collapsed JSON, with element names resolved
  from the definition).
- If a process contains unsupported elements, the badge degrades to
  "Engine: BPMN Subset L1 (with unsupported elements)" and a finding
  list appears.

**Done when.** Opening the demo process in the Explorer shows a green
"Subset L1" badge and a live instance after starting one via CLI.

## PR 8 — Stage-1 conformance validator and documentation

**Scope.** Make Stage 1 a clearly bounded, validated contract.

- `internal/process/engine/definition` exposes
  `ValidateStage1(*Definition) []Finding` returning all
  `PROCESS_UNSUPPORTED_ELEMENT_STAGE1` plus three new findings:
  - `PROCESS_UNREACHABLE_END` — control-flow analysis finds an end
    event that cannot be reached.
  - `PROCESS_DEAD_TASK` — task with no outgoing flow (other than end).
  - `PROCESS_PARALLEL_JOIN_UNBALANCED` — parallel join with more
    incoming flows than any reaching split provides.
- Wire validator into the existing `validateProcessDTO` chain.
- New companion architecture doc
  `docs/architecture/018-bpmn-execution-mapping.md` listing per-element
  what Stage 1 implements (mirroring
  `017-dmn-1.5-mapping.md` style).
- Update `docs/cli-usage.md` and `README.md` engine section.

**Done when.** A deliberately broken BPMN (e.g., with an `inclusiveGateway`)
produces a clear finding in the API response and a visible warning
badge in the UI.

## Non-goals for Stage 1 (do not implement)

These are explicitly out of scope and must remain unimplemented to keep
Stage 1 reviewable:

- Retries, error boundary events, compensation.
- Multi-instance markers (sequential/parallel).
- `subProcess`, `callActivity`, `eventBasedGateway`,
  `inclusiveGateway`.
- Intermediate catch/throw events (timer/message/signal).
- Multi-node deployment; HA; leader election.
- Variable scoping beyond a single flat instance map.
- Process migration of running instances across BPMN versions.

Anything in this list belongs to Stage 2 or Stage 3 and gets its own
implementation prompt at that time.

## Definition of done for the whole stage

- All eight PRs merged.
- `make test` green, including new packages.
- `internal/process/engine` has no import of `internal/app` (one-way
  dependency: app depends on engine, never the reverse).
- A demo cosmos under `examples/` contains at least one process whose
  trigger fires a real instance through the engine, with the resulting
  journal checked into `examples/<x>/expected-journal/` as a golden
  artefact (regenerable via a `go test -update` flag).
- `nomos serve` start-up time on an empty `state_dir` does not regress
  by more than 100 ms compared to pre-engine baseline.
- `docs/architecture/018-bpmn-execution-mapping.md` reflects the exact
  Stage 1 contract; the engine validates every process against it.
