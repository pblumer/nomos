# ADR-0020 - System-generierte IDs für Nomos-Artefakte

Status: Draft

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
COS_root     Cosmos (Sonderfall: genau ein Cosmos pro Repo, fixe ID)
```

Ausnahme: die `Cosmos`-Wurzel trägt die feste ID `COS_root`, weil pro Repo definitionsgemäß genau ein Cosmos existiert.

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

## Migrationsplan (Skizze)

1. Präfix-Register und ID-Generator in `internal/idgen/` implementieren.
2. `nomos id` Subcommand (`new`, `assign`, `validate`) + Integration in `nomos validate`.
3. Schema-/Validator-Regel: `id` Pflicht, Format-Regex, Präfix-Passung, Eindeutigkeit.
4. Migrationskommando `nomos id migrate` für bestehende Cosmos-Verzeichnisse: erzeugt neue IDs, schreibt `.nomos/id-history.yaml`, aktualisiert alle Referenzen.
5. Beispiel-Cosmos (`examples/demo-cosmos/`) migrieren — dient als Referenz.
6. Pre-commit-Hook (optional, repo-lokal) der `nomos validate` aufruft.

## Offene Punkte

- Referenz-Auflösung über `id-history` für externe Konsumenten, die alte IDs noch kennen — Verfallszeit?
- Verhalten beim Forken eines Cosmos: bleibt die ID, oder wird im neuen Cosmos rebrandet?
- Sub-Artefakte (z. B. `ProcessStep` innerhalb eines `Process`): eigenständige ID nach diesem Schema oder zusammengesetzt (`PRC_…/STP_…`)? Vorschlag: eigenständig, weil sie individuell referenziert werden.

## Bezüge

- ADR-0001 (Git-first als Quelle der Wahrheit) — motiviert Merge-Sicherheit.
- ADR-0002 (YAML als Artefaktformat) — `id` lebt im YAML-Frontmatter.
- ADR-0005 (Blueprint-/Instance-/Assurance-Modell) — definiert die Artefakttypen, die Präfixe bekommen.
