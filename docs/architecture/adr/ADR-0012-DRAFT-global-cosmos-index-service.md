# ADR-0012 (DRAFT): Globaler Cosmos-Index-Service (opt-in Crawler)

## Status

Draft

## Datum

2026-05-16

## Kontext

ADR-0009 legt fest, dass es im foederierten Cosmos keinen globalen Sync-Mechanismus gibt.
Das bedeutet: Cross-Domain-Suche ("Welche Domains bieten diesen Service an?") ist ohne
zusaetzlichen Mechanismus nicht moeglich. Dieses ADR spezifiziert einen opt-in Index-Service.

## Offene Fragen

- Welche Entitaeten werden indiziert (Domains, Produkte, Services, Blueprints)?
- Wie registrieren sich Knoten am Index (Push vs. Pull/Crawl)?
- Wie oft wird der Index aktualisiert?
- Wie wird sichergestellt, dass der Index keine autoritative Quelle wird (ADR-0001 muss Trumpf bleiben)?
- Braucht es Zugangskontrolle — koennen Domains entscheiden, ob sie indiziert werden wollen?
- Kann der Index dezentralisiert sein (mehrere Index-Knoten, gegenseitige Synchronisierung)?

## Kandidaten-Optionen (noch nicht entschieden)

**Option A — Zentraler Opt-in-Crawler:**
Ein einzelner, von Nomos betriebener Index-Dienst crawlt alle bekannten `.well-known/nomos`-Endpunkte
und baut einen Suchindex. Domains koennen sich via `crawl: false` in `cosmos.yaml` ausschliessen.

**Option B — Verteilte Index-Knoten:**
Mehrere Index-Knoten synchronisieren sich gegenseitig (z.B. via Gossip). Kein Single Point of Failure.
Hoehere Implementierungskomplexitaet.

**Option C — Domain-pushed Index:**
Domains pushen ihre Metadaten aktiv an einen oder mehrere Index-Knoten. Vorteil: Keine Crawl-Latenz.
Nachteil: Domains muessen aktiv publizieren.

## Naechste Schritte

- Entscheidung ob ein globaler Index im MVP ueberhaupt noetig ist (Kosten-Nutzen).
- Option A als einfachster Einstieg evaluieren.
- Datenschutz- und Governance-Anforderungen klaeren (wer betreibt den Index?).

## Referenz

- ADR-0007 — Domain-owned product offerings
- ADR-0009 — Nomos Cosmos Netzwerkarchitektur und Core Engine
- ADR-0010 (DRAFT) — Well-known-Endpoint-Spezifikation
