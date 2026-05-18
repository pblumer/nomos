# ADR-0014 - Nomos-Selbstmodellierung als versioniertes Cosmos-Bundle (Nomos-as-Nomos)

## Status

Draft

## Datum

2026-05-18

## Kontext

ADR-0009 hat als Leitprinzip festgehalten: *"Nomos-Applikationen — einschliesslich
der eigenen Tooling-Oberflaeche — sollen selbst als Nomos-Konfiguration und
Nomos-Services modelliert werden (Nomos betreibt sich mit Nomos)"*.

Dieses Prinzip ist bisher rein aspirativ. In der Praxis existiert Nomos heute als:

- Go-Binary (`cmd/nomos`) mit CLI-Kommandos, REST-Server, MCP-Server, Web-UI.
- Code-internem Wissen ueber die eigenen Services, Funktionen, Endpunkte und UIs.
- Keiner maschinell konsumierbaren Selbstbeschreibung im Nomos-Artefaktformat.

Konkret fehlt:

1. Eine **versionierte Beschreibung** der Services, die Nomos selbst erbringt
   (Git-Repo-Manager, Validator Engine, Service-Router, REST-/MCP-/Web-Adapter,
   Catalog-Manager, UCI-Manager etc.) im selben YAML-Format wie alle anderen
   Cosmos-Artefakte.
2. Eine **Auflistung der Funktionen / Capabilities**, die diese Services
   bereitstellen — inklusive ihrer REST-Endpunkte, MCP-Tools und CLI-Kommandos
   als Connector-Metadaten.
3. Eine Verknuepfung zu den **UCIs** (ADR-0013), die Nomos selbst mitliefert
   (Cosmos Explorer, Domain Explorer, Blueprint Editor, Validation View, …).
4. Ein **Bootstrap-Mechanismus**, der diese Selbstbeschreibung beim Anlegen
   eines neuen Cosmos ohne Mehraufwand verfuegbar macht.

Solange Nomos sich nicht selbst modelliert, ist das Selbstanwendungs-Prinzip
nicht ueberpruefbar: Es gibt keinen Weg, in einer Cosmos-Explorer-UI auf eine
Nomos-eigene Funktion zu zeigen und zu sehen, *wie* sie modelliert ist.

## Entscheidung

Nomos liefert sich selbst als **Self-Model-Bundle**: ein versionierter, im
Binary eingebetteter Satz von Cosmos-Artefakten, der die Services, Capabilities
und UCIs der jeweiligen Nomos-Version beschreibt und bei Bedarf in jeden
Cosmos-Workspace materialisiert werden kann.

### 1. Reservierter Namespace und kanonische Domain

Das Self-Model lebt unter einem reservierten Namespace, getrennt von
benutzereigenen Domains:

```text
Namespaces
└── nomos                       # reservierter Namespace
    └── core                    # kanonische Domain: core.nomos
        ├── Products
        │   └── nomos-core-engine
        └── Services
            ├── git-repo-manager
            ├── validator-engine
            ├── service-router
            ├── rest-adapter
            ├── mcp-adapter
            ├── web-adapter
            ├── catalog-manager
            └── uci-manager
```

- Der Namespace `nomos` ist **reserviert** und darf in der CLI/UI nicht ueber
  `domain add` erweitert werden.
- Aenderungen an Artefakten unter `nomos/` durch den Endnutzer sind nicht
  vorgesehen; sie werden ueberschrieben, wenn ein neueres Bundle importiert
  wird (siehe Abschnitt 4).

### 2. Bundle-Struktur (im Binary eingebettet)

Das Self-Model-Bundle liegt im Source-Tree unter
`internal/selfmodel/bundle/` und wird via Go `embed` in das Nomos-Binary
kompiliert:

```text
internal/selfmodel/bundle/
  bundle.yaml                    # Bundle-Metadaten (Version, Checksumme)
  domains/
    core.nomos/
      domain.yaml
      services/
        git-repo-manager/service.yaml
        validator-engine/service.yaml
        ...
  catalog/
    blueprints/
      products/nomos-core-engine.yaml
      services/<service-blueprint>.yaml
  uci/
    cosmos-explorer/
      uci.yaml
      schema.yaml
    domain-explorer/uci.yaml
    blueprint-editor/uci.yaml
    validation-view/uci.yaml
    uci-explorer/uci.yaml
```

Die Bundle-Version ist an die Nomos-Binary-Version gebunden:

```yaml
# bundle.yaml
id: nomos-self-model
version: "0.1.0"          # entspricht der CLI-Version
nomos_version: "0.1.0"
generated: 2026-05-18
checksum: sha256:<…>
```

### 3. Connector-Metadaten in Service-YAMLs

Jeder Self-Model-Service deklariert seine Funktionen als
`capabilities`-Eintrag mit Connector-Block. Das ist eine schlanke,
optionale Erweiterung des bestehenden Service-Schemas:

```yaml
id: service-validator-engine
type: service
name: validator-engine
version: "0.1.0"
status: stable
owned_by: core.nomos
operated_by:
  - core.nomos
capabilities:
  - id: cap-validate-cosmos
    name: Validate Cosmos
    summary: Deterministische Vollvalidierung eines Cosmos-Workspaces.
    connectors:
      - type: cli
        invocation: "nomos validate"
      - type: rest
        method: POST
        path: /api/v1/validate
      - type: mcp
        tool: nomos.validate
    inputs_schema_ref: schemas/validate-input.yaml
    outputs_schema_ref: schemas/validate-output.yaml
    related_uci:
      - uci-validation-view
summary: Strukturelle und referentielle Validierung von Cosmos-Artefakten.
```

Connector-Typen fuer MVP: `cli`, `rest`, `mcp`. Weitere (z. B. `webhook`,
`grpc`) koennen spaeter ergaenzt werden, ohne das Schema umzustossen.

### 4. Bootstrap und Import

Das Bundle wird auf zwei Wegen in einen Cosmos materialisiert:

**Implizit beim `cosmos init`:**

```bash
nomos cosmos init ./my-cosmos --git
# Materialisiert zusaetzlich .nomos/domains/core.nomos/, das
# zugehoerige Catalog-Bundle und die UCIs unter .nomos/uci/.
```

Das Default-Verhalten **importiert das Self-Model**. Wer das nicht moechte
(z. B. fuer Minimal-Tests), kann `--without-self` setzen:

```bash
nomos cosmos init ./my-cosmos --without-self
```

**Explizit fuer bestehende Cosmos:**

```bash
nomos self import           # importiert die im Binary eingebettete Version
nomos self status           # zeigt installierte vs. Binary-Version
nomos self diff             # Diff Binary-Bundle vs. Workspace
nomos self upgrade          # ueberschreibt Workspace-Bundle mit Binary-Version
```

Import schreibt unter `.nomos/domains/core.nomos/`, `.nomos/catalog/` und
`.nomos/uci/` — also dieselbe Struktur wie fuer jedes andere Artefakt. Der
Cosmos enthaelt damit sein eigenes Werkzeug als regulaere, validierbare
Artefakte.

### 5. Versionsbindung und Upgrade

- Eine Nomos-Binary bringt **genau eine** Self-Model-Version mit.
- Ein Cosmos haelt die Bundle-Version in `.nomos/cosmos.yaml` fest:
  ```yaml
  self_model:
    version: "0.1.0"
    imported_at: 2026-05-18T10:11:12Z
  ```
- `nomos self status` warnt, wenn Binary-Version und Workspace-Bundle
  abweichen. Der Upgrade-Pfad ist explizit (`nomos self upgrade`) und
  laeuft ueber den normalen Git-Workflow (Branch + Commit + PR), konsistent
  mit ADR-0001.
- Breaking Changes am Self-Model-Schema benoetigen ein eigenes ADR und einen
  Migrationspfad.

### 6. Validierung des Self-Models

Das eingebettete Bundle wird beim Build mit denselben Regeln validiert wie
jeder andere Cosmos:

- Build-Schritt im `Makefile`: `make validate-self-model` materialisiert das
  Bundle in einen Temp-Cosmos und ruft `nomos validate` auf.
- CI-Quality-Gate scheitert, wenn das Self-Model nicht zu seinem eigenen
  Validator passt — das ist die strikteste Form der Dogfooding-Pruefung.

### 7. Web-UI

Die bestehende Cosmos-Explorer-UI rendert `core.nomos` ohne Sonderfall,
markiert die Domain aber visuell als **System**/**Read-only**. UCIs aus dem
Bundle werden im UCI-Explorer (ADR-0013) wie normale UCIs gelistet, jedoch
ebenfalls als Read-only gekennzeichnet.

## Verworfene Alternativen

**Self-Model nur als statisches Markdown / OpenAPI-Doc**: Wuerde Nomos
nicht selbst modellieren, sondern nur dokumentieren. Verfehlt das Prinzip
aus ADR-0009. Abgelehnt.

**Self-Model in einem separaten Repository nachladen**: Erzeugt eine
Netz-/Konnektivitaetsabhaengigkeit beim `cosmos init` und entkoppelt
Binary-Version und Modellversion. Abgelehnt — Embed sichert die
Versionsgleichheit by construction.

**Eigentliche Implementierung aus dem Self-Model generieren (code-gen)**:
Wuerde die Konsequenzen umkehren — der Code haengt am Modell. Fuer den
aktuellen Reifegrad zu invasiv; bleibt als zukuenftige Option offen.
Abgelehnt fuer dieses ADR.

**Self-Model unter Benutzerdomain ablegen (z. B. `nomos.example.com`)**:
Kollidiert mit Domain-Ownership (ADR-0007) und macht Cross-Cosmos-Vergleiche
schwierig. Abgelehnt — reservierter `nomos`-Namespace ist eindeutig.

## Konsequenzen

### Positiv

- Das Prinzip aus ADR-0009 wird konkret und ueberpruefbar.
- Jeder neue Cosmos hat ohne Zusatzaufwand eine maschinenlesbare
  Beschreibung der Nomos-Funktionen, ihrer Connector-Endpunkte und der
  zugehoerigen UCIs.
- KI-Agenten (ADR-0008, MCP) koennen ueber den normalen Artefakt-Zugriff
  herausfinden, *wie* sie mit Nomos sprechen — keine separate API-Doku noetig.
- UCIs aus ADR-0013 bekommen mit `core.nomos`-UCIs ihren ersten realen
  Anwendungsfall.
- Validierung des Self-Models im Build ist ein hartes Quality-Gate gegen
  Drift zwischen Code und Modell.

### Negativ / Risiken

- Pflegeaufwand: Jede neue CLI-Funktion / jeder neue Endpunkt muss im
  Bundle nachgezogen werden, sonst schlaegt `make validate-self-model`
  fehl. Bewusst akzeptiert — das ist genau der Drift-Schutz.
- Bundle-Groesse erhoeht die Binary-Groesse leicht (im MVP unkritisch).
- `cosmos init` schreibt mehr Dateien — kann bei Demo-Skripten ueberraschen;
  `--without-self` mildert das ab.

### Offene Punkte

- Genaue Schema-Definition fuer `capabilities[].connectors[]` → eigener
  Schema-Entwurf im PR zur Implementierung dieses ADR.
- Verhaeltnis zum geplanten Service-Plugin-Modell (ADR-0011-DRAFT): Self-Model
  beschreibt Built-in-Services; Plugins haetten analoge Beschreibungen.
- Mehrsprachigkeit der UCI-Labels — vorerst nur Deutsch, i18n als Backlog.
- Verhalten bei `nomos cosmos init` in bestehenden Self-Model-Cosmos
  (Idempotenz) — im Implementierungs-PR zu klaeren.

## Referenz

- ADR-0001 — Git-first als Quelle der Wahrheit
- ADR-0005 — Blueprint-/Instance-/Assurance-Modell
- ADR-0007 — Domain-owned product offerings
- ADR-0008 — MCP Server als KI-Agenten-Schnittstelle
- ADR-0009 — Nomos Cosmos Netzwerkarchitektur und Core Engine
  (Prinzip "Nomos betreibt sich mit Nomos")
- ADR-0013 — User Contact Interfaces (UCI)
