# CLI Usage (`nomos`)

Diese Seite dokumentiert die aktuelle Nutzung der Nomos-CLI auf Basis der implementierten Commands in `internal/cli/root.go`.

## Schnellstart

```bash
go run ./cmd/nomos --help
```

Version anzeigen:

```bash
go run ./cmd/nomos version
```

## Command-Übersicht

- `nomos version`
- `nomos cosmos init <path>`
- `nomos cosmos info [--path <dir>]`
- `nomos cosmos doctor [--path <dir>]`
- `nomos domain add <dns> [--path <dir>] [--owner <owner>] [--force]`
- `nomos domain list [--path <dir>]`
- `nomos service add <name> --domain <dns> [--path <dir>] [--owner <owner>] [--force]`
- `nomos validate [--path <dir>] [--format <text|json>]`
- `nomos graph [--path <dir>]`
- `nomos verify domain <dns> [--path <dir>]`
- `nomos serve [--path <dir>] [--listen <host:port>]`

---

## `nomos cosmos`

### `nomos cosmos init <path>`
Erzeugt eine neue lokale Cosmos-Struktur.

**Erstellt u. a. folgende Pfade:**
- `<path>/cosmos.yaml`
- `<path>/README.md`
- `<path>/domains/`
- `<path>/.nomos/cache/`
- `<path>/.nomos/index/`
- `<path>/.nomos/evidence/`

**Flags:**
- `--force`: erlaubt Initialisierung in nicht-leerem Zielverzeichnis.
- `--git`: führt `git init` im Zielverzeichnis aus und erzeugt zusätzlich `.gitkeep`.

**Beispiel:**
```bash
go run ./cmd/nomos cosmos init ./my-cosmos --force
```

### `nomos cosmos info`
Liest `cosmos.yaml` und gibt Kernfelder aus (`id`, `name`, `version`, `status`, `owner`, `domains`).
Die Domain-Anzahl wird aus dem Dateisystem ermittelt (`domains/*/domain.yaml`) und nicht aus einem statischen Feld in `cosmos.yaml`.

**Flag:**
- `--path` (Default: `.`)

### `nomos cosmos doctor`
Prüft Basiszustand:
- ob `cosmos.yaml` existiert
- ob ein `.git`-Ordner vorhanden ist

**Ausgabeverhalten:**
- Fehler bei fehlender `cosmos.yaml` (Exit-Code 1)
- Warnung, wenn Git nicht initialisiert ist

---

## `nomos domain`

The CLI keeps the existing explicit canonical-name workflow. In the web UI, normal domain creation is more guided: select a parent node and enter only the new segment. Nomos composes the canonical namespace. Advanced mode allows full canonical input.

Example web flow:

```text
Selected parent: blumer.cloud
New segment: test2
Created domain: test2.blumer.cloud
```

### `nomos domain add <dns>`
Erzeugt eine Domain unter `domains/<dns>/`.

**Validierung:**
- DNS-Name muss mindestens einen Punkt enthalten (`.`), sonst Fehler.

**Erstellt:**
- `domains/<dns>/domain.yaml`
- `domains/<dns>/README.md`
- `domains/<dns>/services/`

**Flags:**
- `--path` (Default: `.`)
- `--owner` (Default: `unknown`)
- `--force` (überschreibt vorhandene Domain-Struktur)

### `nomos domain list`
Listet alle Unterordner unter `domains/`.

Flag: `--path` (Default: `.`).

---

## `nomos service`

The CLI keeps the explicit `--domain` flag. In the web UI, service creation is contextual: select a domain, choose **Add service**, enter the service name, and review the resulting `domain / services / name` preview.

### `nomos service add <name> --domain <dns>`
Erzeugt einen Service unter `domains/<dns>/services/<name>/`.

**Erstellt Unterordner:**
- `capabilities/`
- `requirements/`
- `rules/`
- `processes/`
- `skills/`
- `findings/`
- `evidence/`

**Zusätzlich:**
- `service.yaml`
- `README.md`

**Flags:**
- `--domain` (**pflichtig**)
- `--path` (Default: `.`)
- `--owner` (Default: `unknown`)
- `--force`

---

## `nomos validate`

Führt eine minimale Validierung aus.

Aktuell wird geprüft:
- Existenz von `cosmos.yaml`

**Ausgabe:**
- Text (Standard)
- JSON mit `--format json`

**Exit-Codes:**
- `0`: keine Findings
- `1`: mindestens ein Finding mit Fehler

Beispiel:

```bash
go run ./cmd/nomos validate --path . --format json
```

---

## `nomos graph`

Liest die Cosmos-Struktur aus dem Dateisystem und gibt eine Mermaid-Graph-Definition auf stdout aus:
- Cosmos-Knoten
- Domain-Knoten aus `domains/*/domain.yaml`
- Service-Knoten aus `domains/<domain>/services/*/service.yaml`

Beispielausgabe:

```mermaid
graph TD
  cosmos_local_cosmos["Cosmos: Local Cosmos"]
  cosmos_local_cosmos --> domain_identity_blumer_cloud["Domain: identity.blumer.cloud"]
  domain_identity_blumer_cloud --> service_identity_blumer_cloud_user_account["Service: user-account"]
```

---

## `nomos verify`

### `nomos verify domain <dns>`
Prüft DNS-TXT-Record `_nomos.<dns>` auf den Inhalt `nomos-domain=<dns>`.

**Nebenwirkung:**
- schreibt ein Evidence-File nach `.nomos/evidence/<dns>-dns.yaml`

**Statuslogik:**
- `verified`, wenn erwarteter TXT-Inhalt gefunden wurde
- sonst `failed` und Befehl endet mit Fehler

**Flag:**
- `--path` (Default: `.`)

---

## `nomos serve`

Startet einen HTTP-Server.

**Flags:**
- `--path` (Default: `.`)
- `--listen` (Default: `127.0.0.1:8080`)

**Endpoints:**
- `GET /health` → `{ "status": "ok", "service": "nomos", "version": "..." }`
- `GET /api/v1/validate` → `{ "status": "ok" }`

---

## Typische Workflow-Beispiele

### 1) Neues Cosmos-Repository erzeugen
```bash
go run ./cmd/nomos cosmos init ./demo-cosmos
cd ./demo-cosmos
go run ../cmd/nomos cosmos doctor --path .
```

### 2) Domain und Service anlegen
```bash
go run ./cmd/nomos domain add example.com --path . --owner platform-team
go run ./cmd/nomos service add identity-api --domain example.com --path . --owner iam-team
```

### 3) Validieren und Graph ausgeben
```bash
go run ./cmd/nomos validate --path .
go run ./cmd/nomos graph --path .
```

## `nomos version`

Supports multiple output modes:

```bash
nomos version
nomos version --short
nomos version --format text
nomos version --format json
```

Notes:
- `--format` supports `text` and `json`.
- Invalid formats (for example `--format xml`) return an error.
- If `--short` and `--format` are both set, `--short` wins and only the version string is printed.

## Web UI
Run `./bin/nomos serve --path /tmp/nomos-demo --listen 127.0.0.1:8080`.

### Go Web UI (read-only)

`nomos serve` also provides a read-only web interface rendered by embedded Go templates and static assets. It intentionally mirrors the Nomos frontend design language and does not require a separate Node/Vite build chain.

```bash
make build
COSMOS_PATH=/tmp/nomos-demo NOMOS_BIN=./bin/nomos ./scripts/create-demo-cosmos.sh
./bin/nomos serve --path /tmp/nomos-demo --listen 127.0.0.1:8080
```

---

## CLI and REST API consistency

The read-only CLI and REST API are adapters over the same internal application DTOs for Cosmos, domains, services, validation, graph output and namespace trees.

| CLI command | REST endpoint |
| --- | --- |
| `nomos cosmos info --format json` | `GET /api/v1/cosmos` |
| `nomos domain list --format json` | `GET /api/v1/domains` |
| `nomos domain get <domain> --format json` | `GET /api/v1/domains/{domain}` |
| `nomos service get <service> --domain <domain> --format json` | `GET /api/v1/domains/{domain}/services/{service}` |
| `nomos graph` | `GET /api/v1/graph` |
| `nomos graph --format json` | `GET /api/v1/graph?format=json` |
| `nomos validate --format json` | `GET /api/v1/validate` |
| `nomos namespace tree --format json` | `GET /api/v1/namespaces` |

`--format` supports `text` and `json` where available. Invalid values return the shared `INVALID_FORMAT` error code.

## Canonical namespaces and tree display

Nomos keeps DNS-like domain identifiers as the canonical technical namespace:

```text
identity.blumer.cloud
```

The canonical name remains the filesystem and route identifier:

```text
domains/identity.blumer.cloud/
domains/identity.blumer.cloud/services/user-account/
GET /api/v1/domains/identity.blumer.cloud
```

For human presentation, explorer UIs and client navigation metadata reverse only the domain namespace into tree order:

```text
cloud / blumer / identity
```

Example namespace JSON metadata:

```json
{
  "canonical": "identity.blumer.cloud",
  "parts": ["identity", "blumer", "cloud"],
  "treeParts": ["cloud", "blumer", "identity"],
  "treePath": "cloud/blumer/identity",
  "displayPath": "cloud / blumer / identity",
  "leaf": "identity"
}
```

Services are not reversed. They are displayed below their canonical domain leaf in namespace tree output.

## Read-only API endpoints

- `GET /api/v1/cosmos` returns Cosmos summary metadata and deterministic domain/service counts.
- `GET /api/v1/domains` returns sorted domains with namespace metadata.
- `GET /api/v1/domains/{domain}` returns one domain plus sorted services.
- `GET /api/v1/domains/{domain}/services` returns the services for one canonical domain.
- `GET /api/v1/domains/{domain}/services/{service}` returns one service.
- `GET /api/v1/namespaces` returns the tree-oriented namespace representation rooted at the local Cosmos.
- `GET /api/v1/graph` returns Mermaid as `text/plain` by default.
- `GET /api/v1/graph?format=json` returns `{ "format": "mermaid", "content": "..." }`.
- `GET /api/v1/validate` returns the validation DTO.

Shared error codes include `COSMOS_MISSING`, `COSMOS_LOAD_FAILED`, `DOMAIN_NOT_FOUND`, `SERVICE_NOT_FOUND`, `VALIDATION_FAILED`, `INVALID_FORMAT`, `INVALID_NAMESPACE` and `INTERNAL_ERROR`.

## Web UI coverage

The Go-served web UI is available through `nomos serve` and is intended to be the primary human-facing interface for the local Cosmos repository:

```bash
./bin/nomos serve --path /tmp/nomos-demo --listen 127.0.0.1:8080
```

The UI is server-rendered from embedded Go templates and static assets. It reuses `internal/app` DTOs and application functions for reads, creation, validation, graph, namespace, blueprint, instance, doctor, and verification operations.

| CLI command | API endpoint | Web page |
|---|---|---|
| `nomos cosmos info` | `/api/v1/cosmos` | `/cosmos` |
| `nomos cosmos doctor` | TBD/app doctor endpoint | `/cosmos` |
| `nomos domain list` | `/api/v1/domains` | `/domains` |
| `nomos domain get` | `/api/v1/domains/{domain}` | `/domains/{domain}` |
| `nomos domain add` | `POST /api/v1/domains` | `/domains` form |
| `nomos service list` | `/api/v1/domains/{domain}/services` | `/services` |
| `nomos service get` | `/api/v1/domains/{domain}/services/{service}` | `/services` detail |
| `nomos service add` | `POST /api/v1/domains/{domain}/services` | `/services` form |
| `nomos validate` | `/api/v1/validate` | `/validate` |
| `nomos graph` | `/api/v1/graph` | `/graph` |
| `nomos namespace tree` | `/api/v1/namespaces` | `/namespaces` |
| `nomos verify domain` | TBD | `/verify` |
| `nomos blueprint list` | `/api/v1/blueprints` | `/blueprints` |
| `nomos blueprint show` | `/api/v1/blueprints/{id}` | `/blueprints/{id}` |
| `nomos instance list` | `/api/v1/instances` | `/instances` |
| `nomos instance show` | `/api/v1/instances/{id}` | `/instances/{id}` |
| `nomos instance compliance` | `/api/v1/instances/{id}/compliance` | `/instances/{id}` |

### Read-only and write operations

Read-only pages: dashboard, Cosmos metadata/doctor, namespace tree, graph, validation, blueprints, instances, API index, and verification evidence listing.

Write/create operations: domain creation and service creation are supported through UI forms and explicit POST endpoints. Domain verification can be triggered from `/verify`; it performs DNS TXT lookup and writes evidence in `.nomos/evidence` using the CLI-compatible format.

Known limitations: the doctor function currently lives in the application layer rather than as a dedicated public API endpoint, and verification does not yet expose a stable read endpoint beyond the `/verify` page.
