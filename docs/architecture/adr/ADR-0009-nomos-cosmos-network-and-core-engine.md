# ADR-0009: Nomos Cosmos Netzwerkarchitektur und Core Engine

## Status

Proposed

## Datum

2026-05-16

## Kontext

Mit ADR-0007 wurde das Domain-Ownership-Modell etabliert: Domains besitzen Produkte und Services,
und der Cosmos ist der semantische Raum aller Domains. Bisher war der Cosmos implizit ein einzelnes
Git-Repository auf einem zentralen Server.

Nun stellen sich zwei zusammenhaengende Fragen:

**Frage 1 — Topologie:** Soll der Nomos Cosmos zentralisiert (ein autoritatives Repository) oder
dezentralisiert (P2P-Netzwerk autonomer Cosmos-Knoten) betrieben werden?

**Frage 2 — Core Engine:** Wie soll die laufende Infrastruktur pro Knoten aufgebaut sein, damit
Nomos als System moeglichst stabil, selten neu zu starten und durch Konfiguration statt
Code-Aenderungen erweiterbar ist?

Beide Fragen haengen zusammen: Die Wahl der Topologie bestimmt, welche Faehigkeiten die Core Engine
eines einzelnen Knotens mitbringen muss.

### Anforderungen

- Eine bestaetigte Domain soll ihre Produkte und Services im Cosmos publizieren koennen.
- Die Core Engine muss einen **MCP-Server**, einen **REST-Server** und einen **Web-Server** bereitstellen.
- Die Core Engine soll das **Git-Repository verwalten** (ADR-0001).
- Aenderungen an Produkten, Services und Regeln sollen moeglichst ohne Neustart des Prozesses wirksam werden.
- Die Core Engine selbst soll so **stabil** sein, dass Breaking Changes extrem selten sind.
- Nomos-Applikationen — einschliesslich der eigenen Tooling-Oberflaeche — sollen **selbst als
  Nomos-Konfiguration und Nomos-Services** modelliert werden (Nomos betreibt sich mit Nomos).

## Entscheidung

### 1. Netzwerktopologie: Federated Cosmos mit optionalem P2P-Overlay

Wir waehlen weder ein einzelnes zentrales Repository noch ein vollstaendiges P2P-Netzwerk, sondern
eine **foederierte Topologie**:

- Jede **bestaetigte Domain** betreibt einen eigenen Nomos Core-Knoten (oder delegiert an einen
  vertrauenswuerdigen Operator).
- Jeder Knoten ist **autoritativ fuer seine eigenen Domains und Services**. Er verwaltet sein lokales
  Git-Repository als Quelle der Wahrheit (ADR-0001).
- Knoten koennen **Verweise auf externe Domains** aufloesen, indem sie den Core-Knoten der
  Fremd-Domain direkt abfragen (REST oder MCP). Es gibt keinen globalen Sync-Mechanismus.
- Ein optionales **P2P-Discovery-Overlay** (z.B. via DNS-SD oder einem leichtgewichtigen
  Gossip-Protokoll) kann spaeter eingefuehrt werden — ist aber kein MVP-Bestandteil.
- Ein **Well-known-Endpoint** pro Domain (`https://<domain>/.well-known/nomos`) gibt den
  Core-Knoten-Endpunkt bekannt.

**Verworfen — Zentrales Repository:** Ein einzelnes Repo fuer alle Domains ist ein Single Point of
Failure, skaliert schlecht und verletzt das Domain-Ownership-Prinzip (ADR-0007). Abgelehnt.

**Verworfen — Reines P2P-Netzwerk:** Ein vollstaendiges P2P-Netz (z.B. CRDTs, distributed consensus)
erzeugt erhebliche Implementierungskomplexitaet und macht Governance schwerer nachvollziehbar.
Abgelehnt fuer diesen Reifegrad.

### 2. Domain-Bestaetigung (Trust)

Eine Domain gilt im Cosmos als **bestaetigt**, wenn:

1. Sie einen erreichbaren Core-Knoten unter `https://<domain>/.well-known/nomos` betreibt.
2. Ein kryptografischer **Domain-Proof** vorliegt (signiertes `cosmos.yaml` mit DNS-TXT-Record
   oder X.509-Zertifikat des Domaininhabers).

Unbestaetigte Domains koennen lokal modelliert werden, sind aber im foederierten Cosmos nicht
oeffentlich auffindbar und koennen nicht als `offered_by` in Produkten externer Domains referenziert
werden.

### 3. Nomos Core Engine

Jeder Core-Knoten besteht aus einer einzigen, langlebigen **Nomos Core Engine** mit folgendem Aufbau:

```
┌─────────────────────────────────────────────────────┐
│                  Nomos Core Engine                  │
│                                                     │
│  ┌────────────┐  ┌────────────┐  ┌───────────────┐  │
│  │ MCP Server │  │ REST Server│  │  Web Server   │  │
│  │  (stdio /  │  │ (HTTP/JSON)│  │  (UI + static)│  │
│  │   SSE)     │  │            │  │               │  │
│  └─────┬──────┘  └─────┬──────┘  └───────┬───────┘  │
│        └───────────────┴──────────────────┘          │
│                        │                             │
│              ┌─────────▼──────────┐                  │
│              │   Service Router   │                  │
│              │  (config-driven)   │                  │
│              └─────────┬──────────┘                  │
│                        │                             │
│        ┌───────────────┼───────────────┐             │
│        │               │               │             │
│  ┌─────▼─────┐  ┌──────▼──────┐  ┌────▼──────┐      │
│  │  Git Repo │  │  Validator  │  │  Service  │      │
│  │  Manager  │  │  Engine     │  │  Registry │      │
│  └───────────┘  └─────────────┘  └───────────┘      │
└─────────────────────────────────────────────────────┘
```

**Stabiler Kern — minimale Verantwortlichkeiten:**

Die Core Engine stellt ausschliesslich bereit:

- **Git Repo Manager**: Lesen, Schreiben, Branching, PR-Erstellung (ADR-0001).
- **Validator Engine**: Strukturvalidierung von Artefakten (ADR-0004).
- **Service Router**: Leitet eingehende Requests an registrierte Services weiter.
- **Service Registry**: Verwaltet registrierte Services zur Laufzeit (hot-reload-faehig).
- **MCP-Server, REST-Server, Web-Server**: Protokoll-Adapter; transportieren Requests zum Service Router.

**Erweiterung durch Services — nicht durch Core-Aenderungen:**

Alle fachlichen Faehigkeiten (Domain-Verwaltung, Produkt-Publizierung, Federated Lookup, Assurance,
Governance-Workflows) werden als **registrierte Services** implementiert. Services sind:

- In YAML konfiguriert (analog zu Nomos-Artefakten).
- Hot-reload-faehig: Aenderungen an Service-Konfigurationen werden ohne Core-Neustart wirksam.
- Selbst als Nomos-Artefakte versioniert und validiert.

**Nomos betreibt sich mit Nomos:**

Das Nomos-Tooling (Web-UI, MCP-Integration, REST-API-Dokumentation) wird als eigene Nomos-Domain
(`core.nomos`) mit Produkten und Services modelliert.

### 4. Hot-Reload und Stabilitaet

| Ebene | Reload-Strategie |
|---|---|
| Core Engine (Go-Binary) | Neustart nur bei sicherheitskritischen Fixes oder Breaking-API-Changes |
| Service-Konfiguration (YAML) | Hot-reload via inotify / Git-Webhook; kein Prozess-Neustart |
| Service-Implementierung | Graceful restart des betroffenen Services; Core laeuft weiter |
| Git-Repository-Inhalt | Immer sofort wirksam; kein Prozess-Neustart |

Ziel: Die Core Engine laeuft **monate- bis jahrelang ohne Neustart**. Breaking Changes an der
Core-API erhalten eine eigene ADR und einen Migrationspfad.

### 5. Publizierung von Produkten und Services

Eine bestaetigte Domain publiziert ihre Produkte und Services durch:

1. **Commit in das lokale Git-Repository** (bestehender Mechanismus, ADR-0001).
2. **Registrierung im foederierten Cosmos** durch einen `publish`-Service-Call, der das Artefakt
   unter `/.well-known/nomos/catalog` erreichbar macht.
3. Andere Core-Knoten koennen fremde Domains via Well-known-Endpoint abfragen und Referenzen
   aufloesen.

## Konsequenzen

### Positiv

- **Keine Zentralinstanz**: Jede Domain ist autonom. Ausfall eines Knotens betrifft nur diese Domain.
- **Klare Eigentumsverhaeltnisse**: Domain-Ownership gilt technisch und organisatorisch (ADR-0007).
- **Hohe Core-Stabilitaet**: Fachliche Erweiterungen erfolgen als Services, nicht als Core-Aenderungen.
- **Self-hosting by design**: Nomos modelliert sich selbst als Nomos-Artefakt.
- **Schrittweise Foederierung**: Bestehende Single-Node-Deployments bleiben unberuehrt.
- **MCP, REST, Web aus einer Engine**: KI-Agenten, Programme und Menschen haben je einen optimalen Zugang.

### Negativ / Risiken

- **Kein globaler Namespace-Index out-of-the-box**: Cross-Domain-Suche erfordert spaeter einen
  optionalen Crawler/Index-Service (siehe ADR-0012-DRAFT).
- **Domain-Proof-Mechanismus ist neu**: Muss sorgfaeltig spezifiziert werden (siehe ADR-0010-DRAFT).
- **Hot-reload-Stabilitaet**: Service-Registry muss gegen Race Conditions abgesichert werden
  (siehe ADR-0011-DRAFT).
- **Bootstrapping**: Ein neuer Knoten muss seine eigene Domain zuerst selbst bestaetigten koennen.

## Offene Punkte (eigene ADRs)

| ADR | Thema | Status |
|---|---|---|
| ADR-0010 | Well-known-Endpoint-Spezifikation und Domain-Proof-Format | Draft |
| ADR-0011 | Service-Plugin-Modell und Hot-Reload-Isolation | Draft |
| ADR-0012 | Globaler Cosmos-Index-Service (opt-in Crawler) | Draft |

## Referenz

- [ADR-0001 — Git-first als Quelle der Wahrheit](ADR-0001-git-first-source-of-truth.md)
- [ADR-0004 — KI-Assistenz nur fuer Drafts](ADR-0004-ai-assistance-only-no-autonomous-decisions.md)
- [ADR-0005 — Blueprint-/Instance-/Assurance-Modell](ADR-0005-blueprint-instance-assurance-model.md)
- [ADR-0007 — Domain-owned product offerings](ADR-0007-domain-owned-product-offerings.md)
- [ADR-0008 — MCP Server als KI-Agenten-Schnittstelle](ADR-0008-mcp-server-ai-agent-interface.md)
