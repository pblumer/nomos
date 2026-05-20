# ADR-0023 (DRAFT): Inter-Server-Authentifizierung für Cosmos-Mounts

## Status

Proposed

## Datum

2026-05-20

## Kontext

[ADR-0022](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md) führt
Server-Mounts in den Cosmos Explorer ein: ein lokaler Server hängt entfernte
Nomos-Server (host:7373) ein und liest deren Repositories über REST. Lesen
funktioniert; das **Schreiben** auf entfernten Repositories wurde explizit
offengelassen, weil die Authentifizierung zwischen Servern ungeklärt war
(offener Punkt in ADR-0022 §„Offene Punkte"; [ADR-0010](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md)
spezifiziert Domain-Proof erst als Draft).

Es existiert bereits ein einfacher, funktionierender Auth-Mechanismus:

- `internal/server/auth.go` validiert den Header `X-API-Key` gegen einen
  Cosmos-Key-Store (`.nomos/keys.yaml`).
- `internal/app/keys.go` bietet `CreateKey`, `ValidateKey`, `RevokeKey`,
  `ListKeys`; die CLI hat `nomos key`.
- Requests **ohne** `X-API-Key` werden durchgelassen (ein vorgelagerter Proxy
  wie Traefik kann deren Auth übernehmen). Faktisch sind GET-Reads heute offen.

Diese ADR entscheidet die **MVP-Auth zwischen Servern**, damit das
Cross-Server-CRUD aus ADR-0022 §4 implementiert werden kann.

## Entscheidung

### 1. Per-Mount API-Key über den bestehenden `X-API-Key`

Inter-Server-Auth nutzt den vorhandenen `X-API-Key`-Mechanismus. Ein Mount kann
optional einen **Token** tragen, der der API-Key des entfernten Servers ist
(dort via `nomos key create` erzeugt). Der lokale Server sendet diesen Token als
`X-API-Key`, wenn er den entfernten Server aufruft.

Kein neues Auth-Protokoll im MVP. mTLS und Domain-Proof
([ADR-0010](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md)) bleiben die
spätere, stärkere Schicht für die föderierte Topologie.

### 2. Lokaler Server als Write-Proxy

Die UI ruft **immer den lokalen Server**. Für Operationen auf einem entfernten
Repository leitet der lokale Server den Request an den Besitzer-Server weiter
(REST-Proxy auf `…/api/v1/repositories/{repo}/…`) und hängt den Mount-Token als
`X-API-Key` an. Damit:

- bleibt das Single-Origin-Modell der UI erhalten (keine CORS-/Token-Probleme im
  Browser, Token verlässt nie den Server),
- ist der Besitzer-Server autoritativ und git-first für seine Repositories
  (ADR-0022 §4, [ADR-0001](ADR-0001-git-first-source-of-truth.md)).

### 3. Mount-Token-Speicherung

`.nomos/mounts.yaml` (nicht-autoritativ, git-ignored, ADR-0022 §5) wird um ein
optionales `token`-Feld je Mount erweitert. Das Feld wird in API-Antworten
**nicht** zurückgegeben (nur ein `authenticated: true|false`-Hinweis).

### 4. Read offen, Write tokenpflichtig

- **Reads** (`GET`) auf entfernte Repositories funktionieren weiterhin ohne
  Token (konsistent mit dem heutigen Pass-Through-Verhalten von `apiKeyAuth`).
- **Writes** (`POST`/`PUT`/`PATCH`/`DELETE`) auf entfernte Repositories
  erfordern einen Mount-Token. Ohne Token ist das entfernte Repository in der UI
  **read-only**; Schreibversuche werden klar abgewiesen
  (`MOUNT_NOT_AUTHENTICATED`).

### 5. Lokales Repository unverändert

Das lokale Default-Repository wird wie bisher direkt in-process geschrieben
(kein Proxy, kein Token).

## Konsequenzen

### Positiv

- Schreiben auf entfernte Repositories wird möglich, ohne ein neues
  Auth-Protokoll zu erfinden — der vorhandene Key-Store wird wiederverwendet.
- Token verlässt den Browser nicht (Proxy-Modell).
- Sauberer Upgrade-Pfad: mTLS/Domain-Proof können den Token später ersetzen oder
  ergänzen, ohne UI/Tree-Builder zu ändern.

### Negativ / Risiken

- **Token im Klartext** in `.nomos/mounts.yaml`. Mitigation: git-ignored;
  Empfehlung, Dateirechte zu beschränken; Secret-Store/ENV-Quelle als Folgeschritt.
- **Keine Domain-Bindung**: Ein Token beweist nicht, dass der entfernte Server
  wirklich die beanspruchte Domain besitzt (das leistet erst ADR-0010).
- **Kein Token-Scoping** pro Repository/Operation im MVP — der Key gilt für den
  ganzen entfernten Server (wie heute `X-API-Key`).

## Implementierungshinweise (für das Folge-CRMD von ADR-0022 §4)

- `mount.Mount` um `Token string` erweitern; `MountDTO` gibt nur `Authenticated bool` aus.
- Schreib-Proxy in `internal/app`: für nicht-lokale Repositories Requests an
  `http://{endpoint}/api/v1/repositories/{repo}/…` mit `X-API-Key: {token}`
  weiterleiten.
- Fehlercode `CodeMountNotAuthenticated` (HTTP 401/403) wenn Token fehlt.
- Der Explorer markiert entfernte Knoten ohne Token als read-only.

## Referenz

- [ADR-0001 — Git-first](ADR-0001-git-first-source-of-truth.md)
- [ADR-0009 — Cosmos-Netzwerk und Core Engine](ADR-0009-nomos-cosmos-network-and-core-engine.md)
- [ADR-0010 (DRAFT) — Well-known-Endpoint und Domain-Proof](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md)
- [ADR-0022 (DRAFT) — Cosmos Explorer Server- und Repository-Mounts](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)
