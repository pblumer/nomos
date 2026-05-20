# ADR-0022 (DRAFT): Cosmos Explorer — Server- und Repository-Mounts im Namespace-Tree

## Status

Proposed

## Datum

2026-05-20

## Kontext

Der Cosmos Explorer (`/cosmos`, gerendert aus `internal/server/web/templates/cosmos.html`)
zeigt den Cosmos heute als reinen lokalen Baum:

```text
Local Cosmos
├── Catalog Index
└── Namespaces
    └── <TLD>            (ch, cloud, nomos …)
        └── <Domain>
            └── <Subdomain>
                ├── Products
                ├── Services
                └── Decisions
```

Der Baum wird von `app.BuildNamespaceTree()` (`internal/app/service.go:126`) als
rekursive `NamespaceTreeNodeDTO` (`internal/app/dto.go:196`) aufgebaut. Die Daten
stammen **ausschliesslich aus dem lokalen Dateisystem** (`cosmosfs.LoadTree` über
`.nomos/…`). Ein laufender Prozess (`nomos serve`, Default-Port `7373` laut
[ADR-0015](ADR-0015-default-http-port-7373.md)) bedient **genau einen** Cosmos-Workspace.

[ADR-0009](ADR-0009-nomos-cosmos-network-and-core-engine.md) hat die Zielarchitektur
bereits skizziert: ein föderierter Cosmos aus autonomen Core-Knoten, jeder autoritativ
für seine eigenen Domains/Services, jeder mit eigenem Git-Repository (git-first,
[ADR-0001](ADR-0001-git-first-source-of-truth.md)), erreichbar über REST/MCP/Web aus
einer Engine. [ADR-0010](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md) spezifiziert
(als Draft) das `.well-known/nomos`-Discovery samt Domain-Proof.

Was bisher fehlt, ist die **Repräsentation dieser Topologie im Explorer-Baum** und die
Schnittstelle dahinter:

- Es gibt keinen Knotentyp „Server" und keinen Knotentyp „Repository".
- Ein Server kann seine Repositories nicht ausgeben (kein REST-Endpoint dafür).
- Der Baum kann keine Daten von einem entfernten oder einem zweiten lokalen Server
  einlesen — er liest nur das eine lokale Dateisystem.

Das Domänenkonzept (DNS-artige Namensräume, Domain-Ownership nach
[ADR-0007](ADR-0007-domain-owned-product-offerings.md)) soll dabei **unverändert gültig**
bleiben.

## Entscheidung

### 1. Erweiterte Baum-Topologie

Der Explorer-Baum wird um zwei Ebenen ergänzt. Die **Subdomäne wird zum Mount-Punkt
eines Servers**; unter dem Server hängen dessen **Repositories**, und der bisherige
fachliche Teilbaum (Domains/Services/Products) wird **pro Repository** dargestellt:

```text
<TLD>
└── <Domain>
    └── <Subdomain>            ← bindet optional einen Server (host:7373)
        └── server:7373        Kind "server" (Nomos Core-Knoten)
            ├── <Repository A>  Kind "repository" (git-first: filesystem | github | gitbucket)
            │   └── <Cosmos-Inhalt des Repos: Namespaces → Domain → Service/Product …>
            └── <Repository B>
                └── …
```

- Der **lokale Server** (das laufende `nomos serve`, bzw. `localhost:7373` beim
  Zugriff über das Web-UI aus einer lokal gestarteten CLI) ist **immer als Mount
  vorhanden** (Default-Mount) und benötigt keine Konfiguration.
- Ein Server kann **mehrere Repositories** verwalten. Jedes Repository ist ein
  eigenständiger git-first Cosmos/Katalog (filesystem-Pfad, GitHub- oder
  GitBucket-Remote).
- Das Domänenmodell gilt **innerhalb jedes Repositories** unverändert weiter
  (TLD → Domain → Subdomain → Services/Products). Die neue Server-/Repository-Ebene
  ist eine **Mount- und Herkunftsebene**, kein Ersatz für das Domänenmodell.

### 2. Server gibt seine Repositories selbst aus (neue REST-Schnittstelle)

Ein Nomos Core-Knoten exponiert seine Repositories über einen neuen Endpoint:

```text
GET  /api/v1/repositories            → Liste der vom Server verwalteten Repositories
GET  /api/v1/repositories/{repo}     → Metadaten eines Repositories
```

Repository-Deskriptor (Vorschlag):

```json
{
  "id": "nomos-core",
  "name": "Nomos Core",
  "kind": "filesystem",            // filesystem | github | gitbucket
  "location": "/srv/nomos/core",   // Pfad oder Remote-URL
  "default_branch": "main",
  "status": "clean",               // clean | dirty | unreachable
  "head": "b6c7516"
}
```

Die bestehenden Cosmos-Lesezugriffe werden **repository-scoped** angeboten, damit der
Baum den Inhalt jedes Repositories über REST einlesen kann:

```text
GET  /api/v1/repositories/{repo}/cosmos
GET  /api/v1/repositories/{repo}/namespaces
GET  /api/v1/repositories/{repo}/domains[/{domain}]
…
```

Die heutigen, nicht-scoped Endpoints (`/api/v1/cosmos`, `/api/v1/namespaces` …) bleiben
als **Alias auf das Default-Repository** des Servers erhalten (Abwärtskompatibilität,
siehe Abschnitt 6).

### 3. Repositories werden git-first eingelesen

Jedes Repository wird wie heute git-first behandelt ([ADR-0001](ADR-0001-git-first-source-of-truth.md)):
Quelle der Wahrheit sind die YAML-Artefakte unter `.nomos/`. `filesystem`-Repositories
werden direkt gelesen; `github`/`gitbucket`-Repositories werden geklont/gefetcht und
lokal gecacht (`.nomos/cache/`, nicht autoritativ). Die Repository-Liste eines Servers
ergibt sich aus seiner Konfiguration, nicht aus dem Dateisystem-Walk eines einzelnen
Workspaces.

### 4. Lesen, Ändern, Löschen über REST

Sämtliche CRUD-Operationen auf Repository-Inhalten laufen über die REST-Schnittstelle
des **besitzenden** Servers (das Repository ist dort autoritativ, analog
[ADR-0009](ADR-0009-nomos-cosmos-network-and-core-engine.md)). Schreibende Operationen
sind git-first (Commit gegen das Repo). Der Explorer richtet seine Aktionen (Domain
anlegen, Service hinzufügen, Produkt verschieben …) an `…/repositories/{repo}/…` des
Servers, der das Repository hält — nicht zwangsläufig an den lokalen Server.

### 5. Mounts: manuell jetzt, Federation-ready

Eingehängte Server stammen aus einer **manuellen Mount-Liste** (MVP):

```text
GET    /api/v1/mounts          → konfigurierte Server-Mounts
POST   /api/v1/mounts          → Server einhängen { endpoint: "host:7373", label? }
DELETE /api/v1/mounts/{id}     → Server aushängen
```

- Der lokale Server ist ein impliziter Default-Mount und kann nicht ausgehängt werden.
- Persistenz der Mount-Liste in einer Runtime-Konfiguration (z. B.
  `.nomos/mounts.yaml`, **nicht** autoritativer Fachartefakt — analog Index/Cache).
- Die Mount-Auflösung wird hinter einer Schnittstelle (`MountResolver`) gekapselt, sodass
  später ein **`.well-known/nomos`-Discovery + Domain-Proof**
  ([ADR-0009](ADR-0009-nomos-cosmos-network-and-core-engine.md) /
  [ADR-0010](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md)) als zweite Quelle
  ergänzt werden kann, ohne den Explorer oder die Tree-Erzeugung zu ändern.

### 6. Abwärtskompatibilität

- Bestehende Single-Repo-Deployments funktionieren unverändert: der lokale Server
  präsentiert genau **ein** Repository (den heutigen `.nomos/`-Workspace) als
  Default-Repository.
- Die nicht-scoped REST-Endpoints bleiben als Alias auf das Default-Repository bestehen.
- Wenn nur der lokale Server mit einem Repository gemountet ist, **kann das UI die
  Server-/Repository-Ebene optisch kollabieren** (transparent), sodass der Baum für den
  einfachen Fall aussieht wie heute.

## Konsequenzen

### Positiv

- Der Explorer bildet die föderierte Zieltopologie aus
  [ADR-0009](ADR-0009-nomos-cosmos-network-and-core-engine.md) sichtbar ab.
- Mehrere Repositories pro Server und mehrere Server pro Baum werden möglich, ohne das
  Domänenmodell zu verändern.
- Klare Herkunft jedes Artefakts (welcher Server, welches Repository) — Voraussetzung
  für korrektes, autoritatives Schreiben über REST.
- Federation bleibt nachrüstbar (gleiche Tree-/REST-Form, andere Mount-Quelle).

### Negativ / Risiken

- **Performance/Latenz**: Der Baum aggregiert jetzt potenziell über mehrere
  Netzwerk-REST-Aufrufe. Erfordert Caching/Lazy-Loading pro Server-/Repository-Knoten.
- **Fehlertoleranz**: Ein nicht erreichbarer Server darf den Gesamtbaum nicht brechen
  (Status `unreachable`, Teilbaum als degradiert markieren).
- **Authentifizierung** zwischen Servern ist offen (siehe Offene Punkte).
- **Repository-Identität/Benennung** muss server-übergreifend eindeutig referenzierbar
  sein (Mount-ID + Repo-ID).
- Die bestehende `cosmos.html` (~3950 Z. mit eingebettetem JS) muss um zwei Knotentypen,
  Mount-/Repo-Aktionen und asynchrones Nachladen erweitert werden.

## Offene Punkte (eigene Klärung / ggf. ADR)

- **Auth/Trust für REST zwischen Servern** — Token, mTLS, oder an Domain-Proof
  ([ADR-0010](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md)) gekoppelt?
- **Bindung Subdomäne ↔ Server**: rein lokale Mount-Zuordnung (MVP) vs. autoritativ über
  `.well-known/nomos` der Domäne (Federation).
- **Repository-Quelle GitHub/GitBucket**: Klon-/Cache-Strategie, Branch-/Tag-Pinning
  (vgl. [014 – Katalog-Repository und Release-Modell](../014-katalog-repository-und-release-modell.md)).
- **Schreibpfad cross-server**: Reicht ein REST-Proxy, oder braucht es PR-/Review-Flows
  pro Repository?
- **Lazy-Loading-Vertrag**: Welche Tree-Ebenen werden eager, welche on-expand geladen?

## Referenz

- [ADR-0001 — Git-first als Quelle der Wahrheit](ADR-0001-git-first-source-of-truth.md)
- [ADR-0007 — Domain-owned product offerings](ADR-0007-domain-owned-product-offerings.md)
- [ADR-0009 — Nomos Cosmos Netzwerkarchitektur und Core Engine](ADR-0009-nomos-cosmos-network-and-core-engine.md)
- [ADR-0010 (DRAFT) — Well-known-Endpoint und Domain-Proof](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md)
- [ADR-0012 (DRAFT) — Globaler Cosmos-Index-Service](ADR-0012-DRAFT-global-cosmos-index-service.md)
- [ADR-0015 — Default-Port 7373](ADR-0015-default-http-port-7373.md)
- [014 — Katalog-Repository und Release-Modell](../014-katalog-repository-und-release-modell.md)
- [cosmos-structure](../../cosmos-structure.md)
