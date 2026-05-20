# ADR-0025 (DRAFT): Server-Discovery via PING (1-Hop-Peer-Gossip)

## Status

Proposed

## Datum

2026-05-20

## Kontext

[ADR-0022](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md) führt
manuelle Server-Mounts ein; [ADR-0023](ADR-0023-DRAFT-inter-server-authentication.md)
regelt die Authentifizierung dazwischen. [ADR-0009](ADR-0009-nomos-cosmos-network-and-core-engine.md)
hat ein optionales Discovery-Overlay ausdrücklich offengelassen.

Gewünscht ist ein leichtgewichtiger Weg, **neue Server zu entdecken**, ohne
Endpunkte manuell zu kennen: ein „PING", auf den bekannte Server antworten und
dabei ihre eigenen Peers mitteilen, sodass der Betreiber auswählen kann, welche
neuen Server er aufnimmt.

## Entscheidung

### 1. PING-Endpoint: Identität + Peers

Jeder Server stellt `GET /api/v1/ping` bereit und antwortet mit seiner Identität
und seiner **Peer-Liste** (die Endpunkte seiner konfigurierten Mounts, **ohne**
Tokens):

```json
{
  "server": { "name": "Local Cosmos", "version": "0.1.0", "repositoryCount": 1 },
  "peers":  [ { "endpoint": "nomos.blumer.cloud:7373", "label": "Prod" } ]
}
```

### 2. Discovery: 1-Hop-Gossip über die eigenen Mounts

`GET /api/v1/discover` am lokalen Server: pingt die bekannten Mounts
(authentifiziert via Token, ADR-0023), sammelt deren **advertised peers** und gibt
die **noch nicht eingehängten** Endpunkte als Kandidaten zurück. Genau **ein Hop**
(keine rekursive Verbreitung im MVP). Erreichbarkeit/Identität der Kandidaten wird
best-effort ergänzt (ein direkter PING an den Kandidaten, ohne dessen Peers zu
ziehen).

```json
{ "servers": [
  { "endpoint": "nomos.example.org:7373", "label": "", "name": "Example",
    "reachable": true, "mounted": false, "via": "nomos.blumer.cloud:7373" }
] }
```

### 3. Aufnahme bleibt manuell (opt-in)

Discovery **schlägt nur vor**. Das Aufnehmen eines Servers ist ein expliziter
Schritt des Betreibers (`POST /api/v1/mounts`, ggf. mit Token). Kein Auto-Mount,
kein automatischer Trust.

### 4. Keine Token-Weitergabe

Ein Server gibt in `peers` nur Endpunkte/Labels preis, **nie** Tokens. Ein
entdeckter Kandidat ist zunächst unauthentifiziert; Schreibzugriff erfordert wie
gehabt einen beim Mounten gesetzten Token (ADR-0023).

## Konsequenzen

### Positiv

- Neue Server werden ohne manuelles Endpunkt-Wissen auffindbar.
- Baut nur auf bestehender Mount-/Auth-Schicht auf; kein neues Protokoll.
- Erweiterbar: `.well-known/nomos` / DNS-SD ([ADR-0010](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md))
  können später als weitere Discovery-Quelle ergänzt werden.

### Negativ / Risiken

- **Sichtbarkeit**: Ein Server offenbart authentifizierten Aufrufern seine
  Peer-Endpunkte. Ein Opt-out (`advertise: false`) sollte folgen.
- Nur 1 Hop: weiter entfernte Server werden erst sichtbar, nachdem ein Zwischen-
  server aufgenommen wurde.
- Kein Trust-Beweis für Kandidaten (erst ADR-0010 liefert Domain-Proof).

## Offene Punkte

- Opt-out/Scoping der Peer-Veröffentlichung pro Server/Domain.
- Deduplizierung desselben Servers unter mehreren Endpunkten.
- Optionale größere Hop-Tiefe / rekursive Discovery.

## Referenz

- [ADR-0009 — Cosmos-Netzwerk und Core Engine](ADR-0009-nomos-cosmos-network-and-core-engine.md)
- [ADR-0010 (DRAFT) — Well-known-Endpoint und Domain-Proof](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md)
- [ADR-0022 (DRAFT) — Cosmos Explorer Server- und Repository-Mounts](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)
- [ADR-0023 (DRAFT) — Inter-Server-Authentifizierung](ADR-0023-DRAFT-inter-server-authentication.md)
