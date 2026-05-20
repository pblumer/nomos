# ADR-0027 (DRAFT): Ordner-Modell als Namensraum in Nomos

## Status

Proposed

## Datum

2026-05-20

## Kontext

[ADR-0026](ADR-0026-DRAFT-domains-as-derived-naming-and-folders.md) entscheidet,
dass Domänen ein **abgeleitetes Naming** sind (nicht mehr als CRUD-Artefakte
gepflegt) und führt **Ordner** als in-Nomos-Namensräume ein. Diese ADR
spezifiziert das Ordner-Modell.

Heute liegen Inhalte bereits in einer Verzeichnishierarchie unter
`.nomos/domains/<tree-path>/…`, und der DNS-artige Domänenname wird aus dem
Tree-Pfad abgeleitet (Labels umgekehrt: Pfad `cloud/blumer/identity` →
`identity.blumer.cloud`, siehe [cosmos-structure](../../cosmos-structure.md)).
Faktisch ist diese Hierarchie **schon ein Ordnerbaum** — sie wurde nur als
„Domänen" bezeichnet und mit `domain.yaml` als Artefakt gepflegt.

## Entscheidung

### 1. Ein Ordner ist ein git-first Verzeichnisknoten

Ein **Ordner** ist ein Verzeichnis im Content-Baum eines Repositories. Optional
liegt darin eine `folder.yaml` mit menschlichen Metadaten (Label, Beschreibung).
Ohne `folder.yaml` ist der Ordner trotzdem gültig — sein Segmentname genügt.

```text
<repo>/.nomos/tree/
  cloud/
    blumer/
      identity/
        folder.yaml          # optional: { label: "Identity", description: "…" }
        services/…
        products/…
        decisions/…
```

(Der konkrete Wurzelpfad — `tree/` vs. das bisherige `domains/` — wird in der
Migration festgelegt; siehe Offene Punkte.)

### 2. Der Domänenname ist aus dem Ordnerpfad abgeleitet

Der DNS-artige Domänenname ergibt sich **ausschließlich** aus dem Ordnerpfad
(Labels umgekehrt). Es gibt kein separates Domänen-Artefakt mehr. Ein Ordner
trägt **keine** DNS-/Ownership-/Proof-Semantik; er ist reine Gruppierung. Die
abgeleitete Domäne wird nur **angezeigt** (read-only).

### 3. Ordner-Operationen ersetzen die Domänen-CRUD

Die Explorer-Operationen auf der Namensraum-Ebene sind **Ordner**-Operationen:

- **Anlegen** eines Unterordners (ersetzt „Domäne/Subdomäne erfassen").
- **Umbenennen** eines Ordners.
- **Verschieben** eines Ordners (Teilbaum) — die Gruppierungs-Operation, die für
  Domänen verworfen wurde, ist hier legitim, weil sie Organisation und nicht
  Identität betrifft.
- **Löschen** eines (leeren) Ordners.

Inhalte (Services, Products, Decisions, Views) liegen **in** Ordnern; ihre
Adresse (Domäne) ergibt sich aus dem Ordnerpfad.

### 4. Querverweise: stabile IDs bevorzugen

Damit Ordner-Umbenennungen/-Verschiebungen Verweise nicht brechen, gilt als
Richtlinie: **Verweise zwischen Artefakten erfolgen über stabile IDs**, nicht
über den abgeleiteten Adresspfad. Adress-Strings (`offered_by`, `service_ref`)
sind eine abgeleitete Sicht und werden bei Ordner-Operationen neu berechnet bzw.
sollten mittelfristig durch ID-Referenzen ersetzt werden.

Dies ist die zentrale, noch nicht vollständig ausentschiedene Frage und erhält
ggf. eine eigene ADR (Referenz-Stabilität / ID-Modell).

## Konsequenzen

### Positiv

- Ein einziges, klares Gruppierungsprimitiv; Domäne = abgeleitete Anzeige.
- Ordner-Verschiebung ist eine normale Organisations-Operation (kein Identitäts-
  Risiko wie bei „Domäne umhängen").
- Die bestehende Tree-Pfad-Hierarchie wird konzeptionell sauber als Ordnerbaum
  benannt, statt als gepflegte Domänen.

### Negativ / Risiken

- **Referenz-Stabilität** ist der Knackpunkt: solange Adress-Strings verwendet
  werden, brechen Ordner-Moves Verweise (Rewrite nötig). Erst ID-Referenzen
  lösen das sauber.
- **Migration** der bestehenden `domain.yaml`-Artefakte und des `domains/`-Pfads
  auf das Ordnermodell (`folder.yaml`/Pfadwurzel).
- Spannung zu [ADR-0007](ADR-0007-domain-owned-product-offerings.md) bleibt
  (Domäne als „Heimathafen" → jetzt abgeleitete Adresse).

## Offene Punkte

- **Referenz-Modell**: ID-basierte Verweise vs. abgeleitete Adress-Strings
  (eigene ADR). Bestimmt, wie schmerzhaft Ordner-Moves sind.
- **Pfadwurzel & `folder.yaml`-Schema**: `domains/` → `tree/`? Welche Felder hat
  `folder.yaml` (nur Label/Beschreibung, oder mehr)?
- **Migration** existierender `domain.yaml` (Owner/Status → `folder.yaml` oder
  entfallen).
- **Server-Bindung** (ADR-0022): ein Server bindet an eine abgeleitete Domäne;
  wie wird diese Bindung relativ zu Ordnern gesetzt?
- Verhältnis der `services/`, `products/`, `decisions/`-Unterordner zu „echten"
  Inhaltsordnern (sind das reservierte Namen oder ebenfalls Ordner?).

## Referenz

- [ADR-0007 — Domain-owned product offerings](ADR-0007-domain-owned-product-offerings.md)
- [ADR-0022 (DRAFT) — Cosmos Explorer Server- und Repository-Mounts](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)
- [ADR-0026 (DRAFT) — Domänen als abgeleitetes Naming, Ordner als Nomos-Namensräume](ADR-0026-DRAFT-domains-as-derived-naming-and-folders.md)
- [cosmos-structure](../../cosmos-structure.md)
