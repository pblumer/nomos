# ADR-0026 (DRAFT): Domänen als abgeleitetes Naming, Ordner als Nomos-Namensräume

## Status

Proposed

## Datum

2026-05-20

## Kontext

Bisher behandelt Nomos **Domänen als gepflegte Artefakte**: sie werden im Cosmos
Explorer erfasst, bearbeitet, gelöscht, verifiziert und (zwischenzeitlich
angedacht) per Drag&Drop umgehängt. Jede Domäne hat ein `domain.yaml` unter
`.nomos/domains/<canonical>/` mit Owner/Status u. a.

Gleichzeitig ist der Domänenname längst **DNS-artig abgeleitet** (Canonical aus
dem Tree-Pfad, siehe [cosmos-structure](../../cosmos-structure.md)), und
[ADR-0022](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)/
[ADR-0009](ADR-0009-nomos-cosmos-network-and-core-engine.md) legen fest, dass ein
Server „Teil des Domänensystems" ist und **keine eigenen TLDs definiert**.

In der Praxis zeigt sich: Das **Pflegen** von Domänen in Nomos bringt wenig
Mehrwert und erzeugt Reibung (Domäne umbenennen/umhängen kaskadiert auf Kinder
und Referenzen). Domänen sind primär ein **Benennungs-/Adressierungssystem**, kein
fachlich zu pflegendes Artefakt. Was innerhalb von Nomos wirklich gebraucht wird,
ist eine **leichtgewichtige Gruppierung** (Namensräume) für Inhalte.

## Entscheidung

### 1. Domänen sind ein abgeleitetes Naming-System, kein gepflegtes Artefakt

Nomos pflegt Domänen nicht mehr als CRUD-Artefakte:

- **Kein Erfassen/Bearbeiten/Löschen/Umhängen** von Domänen als
  Erstklass-Operation. Die entsprechenden UI-Aktionen und API-Schreibwege werden
  zurückgebaut.
- Der Domänenname ist eine **abgeleitete, DNS-artige Adresse** — er ergibt sich
  aus der Platzierung von Inhalten und der Server-Bindung, nicht aus einem
  editierbaren `domain.yaml`.
- **Domänen-Verschiebung entfällt** (bestätigt die bewusste Auslassung in der
  Move-Implementierung).

### 2. Ordner als in-Nomos-Namensräume

Für Gruppierung **innerhalb** von Nomos wird ein leichtgewichtiges
**Ordner**-Konzept eingeführt:

- Ordner sind reine **organisatorische Namensräume** für Inhalte (Services,
  Products, Decisions, Views …). Sie tragen **keine** DNS-/Ownership-/Proof-
  Semantik.
- Ordner werden von Nutzern erfasst/umbenannt/verschoben (das ist die
  Gruppierungs-Operation, die Domänen vorher fälschlich übernommen haben).
- Die DNS-Domäne bleibt orthogonal und abgeleitet; ein Ordner ist **nicht** eine
  Domäne.

### 3. Was mit Domänen-Bezügen passiert

- `offered_by`, `owned_by`, `service_ref` etc. bleiben **String-Referenzen auf
  einen Domänennamen** (Adressen), nicht auf ein gepflegtes Domänen-Artefakt.
  Sie funktionieren als Naming-Referenz unverändert.
- **Domain-Ownership** ([ADR-0007](ADR-0007-domain-owned-product-offerings.md))
  bleibt als *Bedeutung* erhalten (ein Produkt wird von einer Domäne angeboten),
  wird aber über die Adresse ausgedrückt, nicht über ein editierbares
  Domänen-Objekt.
- **Domain-Proof/Verifizierung** ([ADR-0010](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md))
  betrifft den **Server**, der eine Domäne beweist (well-known/DNS), nicht ein
  Nomos-Domänen-Artefakt. Das bleibt davon unberührt.

## Konsequenzen

### Positiv

- Einfacheres, konsistentes Modell: Domänen können nicht „falsch gepflegt"
  werden, weil sie abgeleitet sind.
- Weniger riskante Operationen (kein kaskadierendes Domänen-Rename/-Move).
- Klare Trennung: **Ordner** = Gruppierung in Nomos, **Domäne** = abgeleitete
  DNS-Adresse, **Server** = beweist/bedient eine Domäne.

### Negativ / Risiken

- **Migration**: Bestehende `.nomos/domains/<canonical>/domain.yaml` und die
  domänenzentrierte Explorer-/REST-Struktur müssen rückgebaut bzw. auf Ordner +
  abgeleitete Namen abgebildet werden. Größerer Umbau.
- Spannung zu [ADR-0007](ADR-0007-domain-owned-product-offerings.md): dort sind
  Domänen der „semantische Heimathafen". Diese ADR re-interpretiert das als reine
  Adressreferenz; ADR-0007 ist entsprechend zu aktualisieren/abzulösen.
- Verifizierungs-UI, die heute „eine Domäne verifiziert", muss als
  **Server-/Adress-Verifizierung** neu verortet werden.

## Offene Punkte

- **Ordner-Modell**: Schema, Speicherort (`.nomos/…`), Verhältnis zum Tree-Pfad
  und zur abgeleiteten Domäne. Ist die heutige Tree-Pfad-Hierarchie bereits „der
  Ordnerbaum"?
- **Wie binden Inhalte an Namen**: Ergibt sich der Domänenname weiterhin aus der
  Ordner-/Pfadposition, oder wird er pro Artefakt als Adresse gesetzt?
- **Rückbau-Umfang**: Welche Domänen-CRUD-Endpunkte/-UI-Aktionen entfallen, was
  bleibt read-only (Anzeige der abgeleiteten Domäne)?
- **Migrationspfad** für bestehende `domain.yaml`-Dateien (Owner/Status →
  wohin?).
- Verhältnis zu Server-Mounts (ADR-0022): ein Server bindet weiterhin an eine
  (abgeleitete) Domäne; Ordner sind innerhalb eines Repositories.

## Referenz

- [ADR-0007 — Domain-owned product offerings](ADR-0007-domain-owned-product-offerings.md)
- [ADR-0009 — Cosmos-Netzwerk und Core Engine](ADR-0009-nomos-cosmos-network-and-core-engine.md)
- [ADR-0010 (DRAFT) — Well-known-Endpoint und Domain-Proof](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md)
- [ADR-0022 (DRAFT) — Cosmos Explorer Server- und Repository-Mounts](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)
- [cosmos-structure](../../cosmos-structure.md)
