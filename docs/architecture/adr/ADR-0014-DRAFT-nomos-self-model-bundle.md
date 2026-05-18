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
optionale Erweiterung des bestehenden Service-Schemas. Beispiel:

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
    stability: stable
    side_effect: read_only
    connectors:
      - type: cli
        invocation: "nomos validate"
        args:
          - name: path
            flag: "--path"
            required: false
            default: "."
          - name: format
            flag: "--format"
            required: false
            enum: [text, json]
            default: text
        exit_codes:
          - code: 0
            meaning: ok
          - code: 1
            meaning: findings_present
          - code: 2
            meaning: invocation_error
      - type: rest
        method: POST
        path: /api/v1/validate
        request_content_type: application/json
        response_content_type: application/json
        auth: none
      - type: mcp
        tool: nomos.validate
        kind: tool          # tool | resource | prompt
        idempotent: true
    inputs_schema_ref: schemas/validate-input.yaml
    outputs_schema_ref: schemas/validate-output.yaml
    related_uci:
      - uci-validation-view
summary: Strukturelle und referentielle Validierung von Cosmos-Artefakten.
```

#### 3.1 Schema fuer `capabilities[].connectors[]`

Pflicht- und optionale Felder pro Connector-Typ. Unbekannte Felder werden
beim Parsen ignoriert (forward-compatible); fehlende Pflichtfelder fuehren
zu Validator-Findings (`code: SELF.CONNECTOR.MISSING_FIELD`).

**Gemeinsame Felder (alle Connector-Typen):**

| Feld | Typ | Pflicht | Bedeutung |
|---|---|---|---|
| `type` | enum `cli` \| `rest` \| `mcp` | ja | Connector-Klasse. Weitere Werte ohne Schema-Bruch nicht erlaubt — neue Typen brauchen einen Schema-Bump. |
| `description` | string | nein | Kurzbeschreibung; wird in der UI gerendert. |

Felder, die fuer **alle Capabilities** (nicht pro Connector) gelten,
stehen auf der Capability-Ebene:

| Feld | Typ | Pflicht | Bedeutung |
|---|---|---|---|
| `stability` | enum `experimental` \| `beta` \| `stable` \| `deprecated` | nein, Default `stable` | Reifegrad. `deprecated` erzeugt Validator-Warnung. |
| `side_effect` | enum `read_only` \| `mutates_workspace` \| `mutates_git` \| `network_egress` | nein, Default `read_only` | Beeinflusst UI-Bestaetigungsdialoge und MCP-`idempotent`-Default. |

**`type: cli`:**

| Feld | Typ | Pflicht | Bedeutung |
|---|---|---|---|
| `invocation` | string | ja | Vollstaendiges CLI-Pattern, beginnend mit `nomos`. |
| `args[].name` | string | ja | Logischer Argumentname (identisch zum REST-Body-Feld, sofern moeglich). |
| `args[].flag` | string | nein | Konkreter CLI-Flag (z. B. `--path`). Fehlt bei Positional-Args. |
| `args[].positional` | bool | nein, Default `false` | Positional vs. Flag. |
| `args[].required` | bool | nein, Default `false` | |
| `args[].default` | scalar | nein | Default-Wert; nur dokumentarisch. |
| `args[].enum` | list | nein | Erlaubte Werte. |
| `exit_codes[].code` | int | ja | |
| `exit_codes[].meaning` | string | ja | Symbolischer Name (Snake-Case). |

**`type: rest`:**

| Feld | Typ | Pflicht | Bedeutung |
|---|---|---|---|
| `method` | enum `GET` \| `POST` \| `PUT` \| `PATCH` \| `DELETE` | ja | |
| `path` | string | ja | Pfad relativ zur Server-Basis, inkl. Pfadparameter (`/api/v1/foo/{id}`). |
| `request_content_type` | string | nein, Default `application/json` | |
| `response_content_type` | string | nein, Default `application/json` | |
| `auth` | enum `none` \| `session` \| `token` | nein, Default `session` | Erwartetes Auth-Schema (rein dokumentarisch in MVP). |

**`type: mcp`:**

| Feld | Typ | Pflicht | Bedeutung |
|---|---|---|---|
| `tool` | string | ja | Vollqualifizierter MCP-Tool-Name (Dot-Notation, z. B. `nomos.validate`). |
| `kind` | enum `tool` \| `resource` \| `prompt` | nein, Default `tool` | MCP-Primitive (ADR-0008). |
| `idempotent` | bool | nein, Default abgeleitet von `side_effect == read_only` | Steuert, ob MCP-Hosts den Aufruf cachen/retry-en duerfen. |

**Schema-Identitaet und -Versionierung:** Das vollstaendige JSON-Schema liegt
unter `internal/selfmodel/schema/capability-connector.schema.json` im
Repository und wird in `bundle.yaml` via `connector_schema_version: "1"`
referenziert. Schema-Bumps (`"2"`, …) erfolgen ueber ein eigenes ADR.

Connector-Typen fuer MVP sind damit abschliessend: `cli`, `rest`, `mcp`.
Weitere (z. B. `webhook`, `grpc`) erfordern einen Schema-Bump.

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

### 8. Idempotenz beim Re-Init und Re-Import

`cosmos init` und `nomos self import` muessen wiederholbar sein, ohne den
Workspace zu beschaedigen oder benutzereigene Aenderungen unbemerkt zu
ueberschreiben. Es gelten folgende Regeln:

**Erkennung des Workspace-Zustands.** `.nomos/cosmos.yaml` enthaelt nach
einem erfolgreichen Import den Block (siehe Abschnitt 5):

```yaml
self_model:
  version: "0.1.0"
  bundle_checksum: sha256:<…>     # Checksumme aus bundle.yaml zum Importzeitpunkt
  imported_at: 2026-05-18T10:11:12Z
```

Beim Re-Init/Re-Import vergleicht Nomos drei Werte:

1. `binary_bundle_checksum`  — Checksumme des Bundles im laufenden Binary.
2. `workspace_recorded_checksum` — Wert aus `.nomos/cosmos.yaml`.
3. `workspace_actual_checksum` — neu berechnet ueber alle Dateien unter
   `.nomos/domains/core.nomos/`, `.nomos/catalog/blueprints/{products,services}/nomos-*`
   (Self-Model-Anteil im Katalog) und `.nomos/uci/` (nur Self-Model-UCIs,
   per Manifest aufgelistet).

**Entscheidungs-Matrix:**

| Fall | Binary == Recorded | Recorded == Actual | Verhalten |
|---|---|---|---|
| A — Frisch | n/a (kein `self_model`-Block) | n/a | Import wie gehabt. |
| B — No-op | ja | ja | Nichts schreiben, `imported_at` nicht aktualisieren. Exit 0. |
| C — Binary-Upgrade noetig | nein | ja | Hinweis: "Binary-Version X, Workspace Y. Nutze `nomos self upgrade`." Kein automatisches Schreiben. |
| D — Lokale Aenderungen | ja | nein | Validator-Finding `SELF.WORKSPACE.MODIFIED` mit Liste der abweichenden Dateien. Kein Schreiben ohne `--force-overwrite-self`. |
| E — Drift + Version | nein | nein | Wie D, zusaetzlich Hinweis aus C. |

**`cosmos init` auf bereits initialisiertem Verzeichnis:** Verhaelt sich
weiterhin wie bisher (Fehler ohne `--force`); zusaetzlich aktiviert
`--force` den Pfad fuer Faelle B/C/D/E gemaess Matrix — nicht ein
blindes Ueberschreiben.

**`nomos self import` auf existierendem Bundle:** Default ist Fall B/D
ohne Schreiboperation; `--force-overwrite-self` ueberschreibt
Self-Model-Dateien (Fall D/E). Benutzer-Artefakte ausserhalb des
Self-Model-Scopes werden niemals angefasst.

**Git-Verhalten.** Der Importer erzeugt nie Commits selbst — er schreibt
nur ins Working-Tree. So bleibt der bestehende Workflow (Branch → Commit
→ PR) intakt; bei Re-Imports ist ein leerer `git diff` der Beweis fuer
Fall B.

### 9. i18n der UCI- und Capability-Labels

Self-Model-Artefakte werden in englischen IDs und mit lokalisierbaren
Labels ausgeliefert. Lokalisierung ist ein separater, optionaler Layer
und keine Schema-Erweiterung pro Sprache.

**Source-Sprache.** Alle YAML-Felder mit Anzeigetext (`name`, `summary`,
`description`, `title`) werden im Bundle in **Deutsch** als Source-Locale
gepflegt — analog zum aktuellen Repo-Bestand und ADR-0013. Die Source-Locale
ist in `bundle.yaml` markiert:

```yaml
source_locale: de
supported_locales: [de, en]
```

**Uebersetzungsablage.** Uebersetzungen liegen pro Artefakt-Verzeichnis
unter `i18n/<bcp47>.yaml` neben dem Hauptartefakt, nicht im Hauptartefakt
selbst:

```text
.nomos/uci/cosmos-explorer/
  uci.yaml                      # source (de)
  i18n/
    en.yaml
    fr.yaml
```

Eine `i18n/en.yaml` enthaelt nur die uebersetzbaren Schluessel:

```yaml
title: Cosmos Explorer
description: |
  Browse the namespace tree, domains, services, and products of a Nomos cosmos.
```

Fuer Services und Capabilities analog unter
`.nomos/domains/core.nomos/services/<svc>/i18n/<bcp47>.yaml`:

```yaml
name: Validator Engine
summary: Structural and referential validation of cosmos artefacts.
capabilities:
  cap-validate-cosmos:
    name: Validate Cosmos
    summary: Deterministic full validation of a cosmos workspace.
```

**Aufloesung zur Laufzeit.** UI und REST/MCP-Adapter ermitteln die
gewuenschte Locale via Accept-Language (Web) bzw. CLI-Flag
`--locale <bcp47>` (CLI/MCP). Aufloesung: angefragte Locale → Fallback-Kette
→ `source_locale`. Fehlende Schluessel fallen pro Feld auf die Source zurueck;
keine harte Fehlerbedingung.

**Validierung.** Der Self-Model-Validator prueft, dass jede
`i18n/<locale>.yaml` ausschliesslich Schluessel aus dem Source-Artefakt
enthaelt (`SELF.I18N.UNKNOWN_KEY`) und dass fuer in `supported_locales`
gelistete Sprachen Pflichtfelder (`name`, `title`) vorhanden sind
(`SELF.I18N.MISSING_REQUIRED`). Fehlende optionale Felder erzeugen nur
Info-Findings.

**Out of Scope fuer MVP.** Pluralregeln, ICU-Message-Format und
Right-to-Left-spezifische Layout-Hinweise werden nicht im Bundle
abgelegt; das bleibt Verantwortung der UCI-Implementierungen.

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

- Verhaeltnis zum geplanten Service-Plugin-Modell (ADR-0011-DRAFT): Self-Model
  beschreibt Built-in-Services; Plugins haetten analoge Beschreibungen.
  Konkrete Abstimmung erfolgt im PR zu ADR-0011.
- Auslieferungsumfang der initialen `supported_locales`: MVP startet mit
  `[de]`; `en` ist im Schema vorgesehen, aber inhaltlich noch nicht
  ausgeliefert.
- Format der Capability-Inputs-/Outputs-Schemas (`inputs_schema_ref`,
  `outputs_schema_ref`) — JSON Schema vs. eigenes Schlankformat, zu
  entscheiden im Implementierungs-PR.

## Referenz

- ADR-0001 — Git-first als Quelle der Wahrheit
- ADR-0005 — Blueprint-/Instance-/Assurance-Modell
- ADR-0007 — Domain-owned product offerings
- ADR-0008 — MCP Server als KI-Agenten-Schnittstelle
- ADR-0009 — Nomos Cosmos Netzwerkarchitektur und Core Engine
  (Prinzip "Nomos betreibt sich mit Nomos")
- ADR-0013 — User Contact Interfaces (UCI)
