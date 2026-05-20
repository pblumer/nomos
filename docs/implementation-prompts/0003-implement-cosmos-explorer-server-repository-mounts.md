# Implementation Plan: ADR-0022 Cosmos Explorer — Server- und Repository-Mounts

You are working in the Nomos repository.

## Goal

Implement the redesign decided in
`docs/architecture/adr/ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md`.

Target tree shape in the Cosmos Explorer:

```text
<TLD> → <Domain> → <Subdomain> → server:7373 → <Repository> → <existing cosmos subtree>
```

A Nomos server exposes its own repositories over REST (git-first). The Explorer
aggregates over mounted servers (the local server is always mounted; remote servers are
mounted manually now, with the resolver designed so `.well-known/nomos` federation can be
added later). Repository content is read/created/updated/deleted via the owning server's
REST API.

This document breaks the build into PR-sized increments. Each PR ends in a green
`make test` and a runnable demo via `nomos serve`.

## Read first

Before changing code, read:

- `docs/architecture/adr/ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md` — full design
- `docs/architecture/adr/ADR-0009-nomos-cosmos-network-and-core-engine.md` — federation target
- `docs/architecture/adr/ADR-0001-git-first-source-of-truth.md` — git-first invariant
- `internal/app/service.go:126` — `BuildNamespaceTree`; `:891` — `insertDomain`
- `internal/app/dto.go:5` — `CosmosDTO`; `:196` — `NamespaceTreeNodeDTO`
- `internal/server/server.go:32-70` — route table; `:1606` — `cosmosPage`
- `internal/server/domain_explorer.go` — `buildDomainsExplorer` (the `/domains` view variant)
- `internal/server/web/templates/cosmos.html` — the explorer template (recursive node
  define + embedded CSS/JS for search, expand/collapse, context menu)
- `internal/cosmosfs/tree.go` — `LoadTree` (current single-workspace filesystem walker)
- `internal/storage/storage.go` — `.nomos/` path layout
- `internal/cli/root.go` (around the `serve` command, default `127.0.0.1:7373`)

## Conventions

- Server-/Repository-/Mount logic lives in new packages under `internal/` — do **not**
  bloat `internal/app/service.go`. Suggested: `internal/repo` (repository registry +
  git-first loaders), `internal/mount` (mount resolver + REST client).
- Mount config and remote caches are **non-authoritative** (treat like `.nomos/index/`,
  `.nomos/cache/`): never commit them as fachartefakte.
- New public types carry doc comments referencing the ADR-0022 section they implement.
- Keep the non-scoped REST endpoints working as aliases to the default repository.
- The local single-repo case must visually collapse the server/repository levels so the
  default UX is unchanged.

## Increments (PR-sized)

### PR 1 — Repository registry + `GET /api/v1/repositories` (read-only, local only)

- New `internal/repo`: `Repository` descriptor (`ID`, `Name`, `Kind`
  filesystem|github|gitbucket, `Location`, `DefaultBranch`, `Status`, `Head`) and a
  `Registry` that, for the local server, returns a single default repository pointing at
  the current `.nomos/` workspace (the existing `cosmosPath`).
- New DTO `RepositoryDTO` in `internal/app/dto.go`; app function `ListRepositories(path)`.
- New handler `apiRepositories` + route `GET /api/v1/repositories` in
  `internal/server/server.go`.
- Tests: registry returns the default repo; endpoint shape matches the ADR descriptor.
- Demo: `curl localhost:7373/api/v1/repositories`.

### PR 2 — Repository-scoped read endpoints + alias

- Add routes `GET /api/v1/repositories/{repo}/{cosmos|namespaces|domains…}` that resolve
  `{repo}` via the registry and delegate to the existing app read functions scoped to that
  repository's path.
- Keep `/api/v1/cosmos`, `/api/v1/namespaces`, … as aliases to the default repository.
- Tests: scoped and alias endpoints return identical payloads for the default repo.

### PR 3 — Server/Repository node types in the tree (local, eager)

- Extend `NamespaceTreeNodeDTO` (`internal/app/dto.go:196`) with `Server *ServerDTO` and
  `Repository *RepositoryDTO` and accept new `Kind` values `"server"` and `"repository"`.
- In `BuildNamespaceTree` (`internal/app/service.go:126`): wrap the existing namespace
  subtree under a `server` node (local) → `repository` node (default). For the
  single-local-repo case, mark these nodes so the template can collapse them.
- Update `cosmos.html`: add `data-kind="server"` / `data-kind="repository"` icons + CSS
  (mirror the existing `.cn-row[data-kind="…"]` blocks), and collapse logic for the
  trivial case.
- Tests: tree snapshot includes server→repository→namespaces; trivial case collapses.

### PR 4 — Manual mounts (`/api/v1/mounts`) + multi-server aggregation

- New `internal/mount`: `MountResolver` interface; a `ConfigMountResolver` backed by
  `.nomos/mounts.yaml` (non-authoritative). Local server = implicit, non-removable mount.
- Routes: `GET/POST /api/v1/mounts`, `DELETE /api/v1/mounts/{id}`.
- A REST client that fetches `…/repositories` and repository content from a remote
  server endpoint (`host:7373`). Handle `unreachable` gracefully (degraded subtree, never
  break the whole tree).
- `BuildNamespaceTree` aggregates: for each mounted server → its repositories → content.
  Add lazy-loading where eager fan-out is too expensive (see PR 5).
- Tests: two mounts (one stubbed remote) produce two server subtrees; unreachable remote
  yields a degraded node, not an error.

### PR 5 — Lazy-loading + UI mount/repo actions

- Load server → repository levels eagerly, but defer each repository's cosmos subtree until
  its repository node is expanded (new `GET …/repositories/{repo}/namespaces` call from the
  template JS).
- `cosmos.html` context menu: "Mount server…", "Unmount server", per-repository refresh.
- Tests: expansion triggers exactly one scoped fetch; collapse/expand is idempotent.

### PR 6 — CRUD over REST, git-first, owning-server aware

- Route write actions (create domain, add service, move product, …) to
  `…/repositories/{repo}/…` on the **owning** server, not implicitly the local one.
- Ensure writes remain git-first (commit against the target repo). For remote repos via
  REST proxy to the owning server.
- Tests: a write targeting a non-default repository hits the correct scoped path and
  results in a commit in that repo.

## Out of scope (separate ADRs / later)

- Inter-server authentication/trust (token/mTLS/domain-proof) — open point in ADR-0022.
- `.well-known/nomos` federation as a mount source (ADR-0009/0010) — the `MountResolver`
  interface must allow adding it without touching the tree builder or template.
- GitHub/GitBucket clone/cache and branch/tag pinning strategy (cf. architecture note 014).
- Cross-server PR/review write flows.
