# ADR-0020 - System-generierte IDs für Nomos-Artefakte

Status: Proposed

## Kontext

Nomos-Artefakte (Cosmos, Domain, Service, Variant, Blueprint, Product, Requirement, Rule, Validation Scenario, Finding, Process, Step, Instance, Evidence, Decision, …) tragen ein `id`-Feld, das sie eindeutig identifiziert und für Querverweise dient (z. B. `process_id`, `decision_id`, `bpmn_element_id`).

Heute werden IDs durch den Autor frei gewählt. Das hat in der Praxis zu drei Problemklassen geführt:

1. **Inkonsistenz.** Stilbrüche zwischen Artefakten (Beispiele aus `examples/demo-cosmos/`: `produkt-1`, `rule-password-policy`, `PI-ACC-MBX-EXAMPLE-001`, `SVC-001`, `cosmos-local`). Es ist nicht erkennbar, ob ein Stil verbindlich ist.
2. **Merge-Konflikte.** Fortlaufend nummerierte IDs (`PROD-001`, `PROD-002`) kollidieren strukturell, sobald zwei Branches gleichzeitig die nächste Nummer vergeben. Im Git-first-Modell (ADR-0001) ist genau dieser Fall der Normalfall.
3. **Instabilität.** Slug-basierte IDs (`req-login-required`) koppeln den Identifier an einen mutierbaren Namen. Umbenennen bricht alle Referenzen; Nicht-Umbenennen führt zu irreführenden IDs.

Das ADR legt fest, wie IDs erzeugt, validiert und gepflegt werden.

## Entscheidung

IDs werden **systemgeneriert** und folgen einem festen Format:

```
<TYP-PREFIX>_<6 Zeichen Crockford-Base32>
```

- **Typ-Präfix**: drei Großbuchstaben, fix pro Artefakttyp (siehe Präfix-Register unten).
- **Trenner**: Unterstrich `_`.
- **Suffix**: 6 Zeichen aus dem Crockford-Base32-Alphabet (`0123456789ABCDEFGHJKMNPQRSTVWXYZ`), kryptografisch zufällig erzeugt.
- **Gesamtlänge**: 10 Zeichen.
- **Validierungs-Regex**: `^[A-Z]{3}_[0-9A-HJKMNP-TV-Z]{6}$`.

Beispiele:

```
PRD_A7K3M2   Product (Blueprint)
PRV_X9P4N1   Product Variant
REQ_H2M8Q5   Requirement
RUL_T4R7W3   Business Rule
VSC_B6N2K9   Validation Scenario
FND_D1F5H8   Finding
SVC_C3J7L2   Service (Blueprint)
SVI_M5Q8N4   Service Instance
PRI_F2K6R9   Product Instance
PRC_W7X3Y1   Process
STP_N4M6B2   Process Step
DEC_R9S5T7   Decision
EVD_V2W8X4   Evidence
DOM_K1L3M5   Domain
SVR_J8H4K2   Server (Nomos Core-Knoten / Mount, ADR-0022)
REP_P3Q7N5   Repository (git-first Cosmos-Mount, ADR-0022)
COS_root     Cosmos (Sonderfall: repo-lokale opake ID, siehe unten)
```

### Cosmos-Identität: opake ID plus weltweiter Handle

Frühere Entwürfe behandelten den Cosmos als „genau einen pro Repo" mit der
einzigen fixen ID `COS_root`. Das kollidiert mit der föderierten Zielarchitektur:

- [ADR-0009](ADR-0009-nomos-cosmos-network-and-core-engine.md) beschreibt einen
  föderierten Cosmos aus vielen autonomen Knoten mit Cross-Domain-Referenzen.
- [ADR-0022](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md) hält
  fest, dass ein **Server mehrere Repositories** verwaltet und *jedes Repository
  ein eigenständiger git-first Cosmos* ist. Es gibt also mehrere Cosmos pro Server.

Trügen alle Repos dieselbe ID `COS_root`, ließe sich „der Cloud-Cosmos" nicht
von „dem Local-Cosmos" per ID unterscheiden. Daher trägt der Cosmos **zwei
Identifier-Ebenen**:

1. **Opake ID `COS_root`** — bleibt erhalten, ist aber ausdrücklich nur
   **repo-lokal** (genau ein Cosmos pro Repository). Sie dient internen
   Referenzen innerhalb desselben Repositories.
2. **Weltweiter Handle `<authority>/<slug>`** — der publizierte, global
   eindeutige Identifier. `authority` ist die DNS-artige besitzende Domain
   (provable über `.well-known/nomos`, [ADR-0010](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md)),
   `slug` ein cosmos-lokaler, mutierbarer Kurzname. Global eindeutig, weil
   DNS-Domains global eindeutig sind und der Domaininhaber die Slugs in seinem
   Namensraum kontrolliert. Beispiel: `blumer.net/nomos`.

Cross-Cosmos- und Föderations-Referenzen gehen über den **Handle**, nicht über
die opake `COS_root`-ID. Server und Repository (die Mount-/Topologie-Ebene aus
ADR-0022) bekommen reguläre systemgenerierte IDs (`SVR_…`, `REP_…`).

### Vergabe

IDs werden **ausschließlich von der `nomos`-CLI bzw. der Server-API** beim Anlegen eines Artefakts vergeben. Konkret:

- `nomos new <typ> …` erzeugt YAML inklusive `id`.
- Wird ein Artefakt ohne `id` angelegt, ergänzt `nomos validate --fix` die ID automatisch und schreibt die Datei zurück.
- `nomos id assign <pfad>` füllt eine fehlende ID gezielt nach.

### Validierung und Unveränderbarkeit

- `nomos validate` prüft pro Artefakt: `id` vorhanden, Format korrekt, Präfix passt zum Artefakttyp.
- IDs sind nach Vergabe **unveränderlich**. Eine Änderung wird vom Validator als Fehler gemeldet (mit Verweis auf alle Stellen, die referenzieren).
- Eindeutigkeit wird repo-weit geprüft. Bei Kollision (Wahrscheinlichkeit pro Typ bei <10.000 Artefakten praktisch null) generiert die CLI erneut.

### Anzeige und Suche

Da IDs selbst keinen semantischen Inhalt tragen, bekommen Artefakte zusätzlich:

- `name:` (Pflicht) — fachlicher Klartextname, mutierbar.
- `slug:` (optional, generiert aus `name`) — URL-/CLI-freundlicher Kurzname, mutierbar.

UI, CLI-Listings und Suche zeigen `name` (ggf. `slug`) prominent, die ID nur als Sekundärinfo. Referenzen zwischen Artefakten gehen **immer** über `id`, nie über `name`/`slug`.

## Begründung

| Eigenschaft | Wie erreicht |
| --- | --- |
| Merge-sicher | 6 Crockford-Zeichen ≈ 10⁹ Werte → praktisch kollisionsfrei ohne Koordination zwischen Branches. |
| Lesbar | Typ-Präfix sagt sofort, was es ist. 10 Zeichen sind kürzer als heutige Praxis (`PI-ACC-MBX-EXAMPLE-001` = 21 Zeichen). |
| Robust beim Abtippen | Crockford-Base32 vermeidet `0/O`, `1/I/L`-Verwechslung, ist case-insensitive lesbar. |
| Stabil | ID hat keinen semantischen Inhalt — Umbenennen ändert die ID nie. |
| Maschinell validierbar | Eine einzige Regex deckt das Format ab. Präfix-Mapping ist im Code zentral. |

## Verworfene Alternativen

- **UUID v4** — 36 Zeichen, kein Typ-Präfix, schwer lesbar.
- **ULID** — 26 Zeichen, Zeit-Prefix bringt im Git-first-Kontext keinen Mehrwert.
- **Fortlaufende Sequenz (`PRD-0001`)** — strukturell merge-unsicher.
- **Slug aus Name (`req-login-required`)** — koppelt ID an mutierbares Feld; bricht Stabilität.
- **Inhalts-Hash** — instabil bei jeder Bearbeitung.

## Konsequenzen

### Positive

- Eindeutige, deterministisch validierbare IDs.
- Keine Merge-Konflikte mehr durch ID-Vergabe.
- UI/CLI können IDs verlässlich verkürzt darstellen (Präfix verrät den Typ).

### Negative / Aufwand

- **Migration des Bestands.** Vorhandene Artefakte in `examples/` und ggf. Produktivkosmen müssen einmalig auf das neue Format gehoben werden. Migrationsskript erzeugt neue IDs, schreibt eine `id-history:`-Mapping-Tabelle pro Cosmos und ersetzt Referenzen.
- **Kein manuelles „sprechendes" ID-Wählen mehr.** Wer heute `PROD-CLOUD-MAILBOX-001` mochte, verliert das. Kompensiert durch `name`/`slug`.
- **Tooling-Pflicht.** Artefakte können nicht mehr sinnvoll per Hand ohne CLI angelegt werden. `validate --fix` mildert das ab, ist aber Voraussetzung.

## Kollisionsverhalten (wichtig)

6 Crockford-Base32-Zeichen ergeben 32⁶ ≈ 1,07 × 10⁹ mögliche Werte. Nach
dem Geburtstagsproblem liegt die Kollisionswahrscheinlichkeit bei ~5 ×
10⁻⁴ pro 1000 Artefakten desselben Typs und steigt bei ca. 46.000
Artefakten desselben Typs auf 50 %. Für realistische Cosmos-Größen
(<10.000 Artefakte je Typ) ist eine zufällige Kollision selten, aber nicht
ausgeschlossen.

Konsequenz: `CreateBlueprint`/`CreateInstance` prüfen Eindeutigkeit
explizit und liefern bei Konflikt einen 409 — der Aufrufer (CLI/MCP)
kann transparent neu generieren. Falls langfristig nötig, ohne ADR-
Bruch: Suffix-Länge auf 8 Zeichen erhöhen (Gesamt 12 Zeichen, ~10¹²
Werte) — kompatibel zur bestehenden Regex.

## Umsetzung (Stand)

1. **`internal/idgen`** — Präfix-Register (inkl. `server`/`SVR`,
   `repository`/`REP`), `New`, `NewForType`, `IsValid`, `IsValidForType`,
   `IsLegacy` sowie der weltweite Cosmos-Handle (`CosmosHandle`,
   `IsValidHandle`). Vollständige Unit-Tests.
2. **`internal/idmigrate`** — `Scan`, `BuildPlan`, `Apply`, History-
   Persistenz (`.nomos/id-history.yaml`), transparente `Resolve(path,
   anyID)`-Auflösung mit In-Memory-Cache.
3. **Auto-Vergabe in `app.Create*`** — Leere `id` triggert
   `idgen.NewForType(type)`; vorhandene IDs werden weiterhin akzeptiert
   (Backward-Compat während der Migration).
4. **MCP-Tools** — `id` ist optional bei `nomos_product_create`,
   `nomos_decision_create`, `nomos_instance_create`. Bei Auslassung
   generiert der Server.
5. **CLI** — `nomos id new <type>`, `nomos id check`, `nomos id migrate
   [--dry-run]`. `--id` bei `blueprint create` und `instance create`
   optional.
6. **Transparente Auflösung** — `GetBlueprint`, `GetInstance`,
   `GetProcess`, `GetDecision` konsultieren die History und finden
   Artefakte auch über ihre alte ID. Externe Konsumenten, die alte
   Referenzen halten, können weiter lesen.

## Offene Punkte

- Verfallszeit von `id-history.yaml`-Einträgen (heute: forever; akzeptabel
  weil append-only und sehr klein).
- Verhalten beim Forken eines Cosmos: die opake `COS_root`-ID bleibt
  (repo-lokal); die globale Identität wird über einen **neuen Handle**
  (`<authority>/<slug>` der forkenden Domain) vergeben, damit Fork und Original
  weltweit unterscheidbar bleiben.
- Persistenz und Vergabe des Cosmos-Handles: heute über `CosmosHandle` aus
  Domain + Slug formbar; wo `authority`/`slug` autoritativ gepflegt werden
  (cosmos.yaml-Feld vs. abgeleitet aus `.well-known/nomos`) ist noch offen.
- Server-/Repository-IDs (`SVR_`/`REP_`): Vergabezeitpunkt — beim Mount
  ([ADR-0022](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md))
  vs. beim Anlegen eines Repositories — ist noch zu klären.
- Sub-Artefakte (`ProcessStep`): eigenständige ID-Vergabe (`STP_…`) wurde
  implementiert. Verbund-IDs (`PRC_…/STP_…`) wurden verworfen, weil
  Steps individuell referenzierbar sein müssen.
- Optional: Suffix-Länge konfigurierbar machen — derzeit Konstante
  `idgen.SuffixLength = 6`.

## Bezüge

- ADR-0001 (Git-first als Quelle der Wahrheit) — motiviert Merge-Sicherheit.
- ADR-0002 (YAML als Artefaktformat) — `id` lebt im YAML-Frontmatter.
- ADR-0005 (Blueprint-/Instance-/Assurance-Modell) — definiert die Artefakttypen, die Präfixe bekommen.
- ADR-0009 (Föderierter Cosmos / Core Engine) — motiviert den weltweiten Cosmos-Handle.
- ADR-0010 (Well-known-Endpoint und Domain-Proof) — liefert die Autorität des Handles.
- ADR-0022 (Cosmos Explorer — Server- und Repository-Mounts) — führt die Typen Server/Repository ein.
