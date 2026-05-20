# ADR-0028 (DRAFT): ID-basierte Referenzen statt Adress-Strings

## Status

Proposed

## Datum

2026-05-20

## Kontext

[ADR-0026](ADR-0026-DRAFT-domains-as-derived-naming-and-folders.md) macht Domänen
zu abgeleitetem Naming, [ADR-0027](ADR-0027-DRAFT-folder-namespace-model.md)
führt Ordner als git-getrackte Verzeichnisse ein, deren Pfad die DNS-artige
Adresse ableitet. Damit wird **Verschieben/Umbenennen von Ordnern** zur normalen
Operation.

Heute referenzieren Artefakte einander über **Adress-Strings**:
`offered_by: identity.blumer.cloud`, `service_ref: identity.blumer.cloud/user-account`,
`owned_by`, Decision-Refs usw. Diese Adressen sind aus dem Ordnerpfad abgeleitet
— wird ein Ordner verschoben/umbenannt, ändert sich die Adresse und **alle
Adress-Referenzen brechen** (oder müssten kaskadierend umgeschrieben werden).

ADR-0027 hat dies als zentrale offene Frage benannt. [ADR-0020](ADR-0020-DRAFT-artefact-id-generation.md)
legt bereits **system-generierte, stabile IDs** für Artefakte fest — die Basis
für eine stabile Referenzierung.

## Entscheidung

### 1. Querverweise erfolgen über stabile IDs

Verweise zwischen Artefakten (Produkt→Service, Produkt→Decision, Prozess→Service-
Method, View→DataObject …) referenzieren das Ziel über seine **stabile ID**
([ADR-0020](ADR-0020-DRAFT-artefact-id-generation.md)), nicht über den
abgeleiteten Adresspfad.

### 2. Die Adresse ist eine abgeleitete Sicht

Der Adress-/Domänenpfad (`identity.blumer.cloud/user-account`) bleibt als
**menschenlesbare, abgeleitete Anzeige** und als Auflösungs-Komfort erhalten,
ist aber **nicht** die maßgebliche Referenz. UI/Doku zeigen die Adresse; der
gespeicherte Verweis ist die ID.

### 3. Auflösung per ID-Index

Referenzen werden über einen **ID-Index** des Repositories aufgelöst (ID →
Artefakt + aktuelle Adresse). Über Servergrenzen hinweg löst der **besitzende
Server** die ID auf (konsistent mit dem Mount-/Proxy-Modell aus
[ADR-0022](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)).

### 4. Moves brechen keine Verweise

Da Verweise IDs nutzen, bleiben sie bei **Ordner-/Inhalts-Verschiebungen
unverändert gültig** — die abgeleitete Adresse ändert sich, die ID nicht. Kein
kaskadierendes Ref-Rewriting mehr.

### 5. Migration

Bestehende Adress-String-Referenzen werden einmalig **zu ID-Referenzen
aufgelöst** (Adresse → ID via Index). Nicht auflösbare Adressen bleiben als
unaufgelöste Referenz sichtbar (wie heute fehlende Fulfillment-Refs).

## Konsequenzen

### Positiv

- Ordner-/Inhalts-Moves sind sicher; keine brechenden oder kaskadierenden Refs.
- Klare Trennung: **ID** = stabile Identität, **Adresse** = abgeleitete Anzeige.
- Baut direkt auf [ADR-0020](ADR-0020-DRAFT-artefact-id-generation.md) auf.

### Negativ / Risiken

- **Lesbarkeit roher YAML**: `service_ref: SVC-AB12` ist weniger sprechend als
  ein Adresspfad. Mitigation: UI/Tools zeigen die aufgelöste Adresse; ggf.
  Adress-Hinweis als Kommentar/Sekundärfeld.
- **ID-Eindeutigkeit & Auflösung**: IDs müssen pro Repository (und für Federation
  pro Server) eindeutig auflösbar sein; der Index muss gepflegt werden.
- **Migration** aller bestehenden Adress-Referenzen; cross-server Auflösung.
- Spannung zu Artefakten, die bewusst eine Adresse meinen (z. B. eine externe,
  noch nicht existierende Domäne) — solche bleiben als Adresse referenzierbar.

## Offene Punkte

- **ID-Geltungsbereich**: pro Repository eindeutig genügt lokal; für Federation
  Server-qualifiziert (`{repo|server}:{id}`)?
- **Index-Speicherung**: abgeleitet/cached (`.nomos/index/`, nicht autoritativ)
  vs. zur Laufzeit gescannt.
- **Anzeige/Eingabe**: wie referenziert ein Mensch beim Erfassen — Adresse
  wählen, intern als ID speichern?
- **Externe/unaufgelöste Referenzen**: weiterhin als Adress-String zulässig?
- Schrittweise Migration vs. harter Schnitt; Übergangszeit mit beidem.

## Referenz

- [ADR-0020 (DRAFT) — System-generierte IDs für Nomos-Artefakte](ADR-0020-DRAFT-artefact-id-generation.md)
- [ADR-0022 (DRAFT) — Cosmos Explorer Server- und Repository-Mounts](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)
- [ADR-0026 (DRAFT) — Domänen als abgeleitetes Naming](ADR-0026-DRAFT-domains-as-derived-naming-and-folders.md)
- [ADR-0027 (DRAFT) — Ordner-Modell als Namensraum](ADR-0027-DRAFT-folder-namespace-model.md)
