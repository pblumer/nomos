# ADR-0030 (DRAFT): Dezentrales, selbstheilendes Verzeichnis (Membership-Gossip mit Leaf-Knoten)

## Status

Draft

## Datum

2026-05-21

## Kontext

[ADR-0009](ADR-0009-nomos-cosmos-network-and-core-engine.md) wählt eine föderierte
Topologie und lässt ein P2P-Discovery-Overlay ausdrücklich offen. [ADR-0025](ADR-0025-DRAFT-server-discovery-ping.md)
liefert dazu einen ersten Schritt: `GET /api/v1/ping` (Identität + Peer-Liste) und
`GET /api/v1/discover` (1-Hop-Gossip über die eigenen Mounts, Aufnahme bleibt manuell).
[ADR-0012](ADR-0012-DRAFT-global-cosmos-index-service.md) hält die Frage eines
verteilten Verzeichnisses als offenen Punkt fest (u.a. Option B: verteilte,
sich gegenseitig synchronisierende Index-Knoten).

Gewünscht ist ein **übergeordnetes, nicht zentralisiertes Verzeichnis**, das die
Knoten **untereinander synchronisieren, ergänzen und am Leben halten** — ohne
Single Point of Failure. Eine zusätzliche Anforderung: Teilnehmer werden häufig
**private Notebooks** sein, die hinter NAT/Firewall stehen und keine stabile,
eingehend erreichbare Adresse besitzen.

Die ursprüngliche Idee schlug vor, dafür **IPv6-Adressinformationen** zur
Kommunikation zu nutzen, und fragte nach **Portwahl**, um nicht an Firewalls zu
scheitern oder als Schadsoftware zu gelten.

Dieses ADR dokumentiert die Designentscheidung. Es ist bewusst Draft und dient
als Diskussionsgrundlage; offene Punkte sind unten gelistet.

## Entscheidung

### 1. Zwei Knotenklassen statt „jeder ist ein Peer"

Wir trennen das Netz in zwei Rollen, weil private Notebooks nicht zuverlässig
eingehend erreichbar sind:

- **Backbone-Knoten** — Knoten mit stabiler, öffentlich erreichbarer Adresse
  (die `.well-known/nomos`-Knoten aus ADR-0009). Sie bilden das Verzeichnis und
  gossippen untereinander.
- **Leaf-Knoten** — Notebooks und sonstige Clients hinter NAT/Firewall. Sie
  **verbinden ausgehend** zu einem oder mehreren Backbone-Knoten, tragen sich
  dort ein und abonnieren das Verzeichnis. Sie sind **nicht** eingehend
  erreichbare Peers.

**Verworfen — vollständiges P2P / direkte Notebook-zu-Notebook-Verbindungen:**
Erfordert NAT-Traversal (STUN/TURN/ICE, Hole-Punching, Relays; libp2p/WebRTC).
Großer Komplexitätssprung und in ADR-0009 für diesen Reifegrad bereits abgelehnt.

### 2. Verzeichnis als Membership-Gossip zwischen Backbone-Knoten

Das Verzeichnis ist eine Erweiterung von ADR-0025 vom einmaligen 1-Hop-Vorschlag
zu einem **selbstheilenden Membership-/Gossip-Protokoll** (Stichwort SWIM bzw.
epidemisches Gossip) zwischen Backbone-Knoten:

- Periodischer Austausch von Verzeichnis-Einträgen (Endpunkt, Label, Identität,
  Liveness/Heartbeat, Versionsstand).
- **Liveness-Erkennung**: ausgefallene Knoten werden als `suspect`/`dead` markiert
  und propagiert; wiederkehrende Knoten reaktiviert.
- **Konfliktfreies Mergen** über monoton steigende Versionszähler/Inkarnationen pro
  Eintrag (kein globaler Konsens nötig).
- Aufnahme/Trust bleibt wie in ADR-0025 **opt-in** und manuell; Gossip schlägt nur
  vor und hält bekannte Knoten aktuell.

### 3. Leaf-Knoten-Teilnahme nur ausgehend

- Leaf-Knoten halten eine **ausgehende, langlebige Verbindung** zu einem
  Backbone-Knoten (Registrierung + Heartbeat + Subscription auf
  Verzeichnis-Updates).
- Kein eingehender Listener auf dem Notebook erforderlich.
- Fällt der gewählte Backbone-Knoten aus, verbindet der Leaf-Knoten zum nächsten
  bekannten — das Verzeichnis selbst bleibt durch Gossip am Leben.

### 4. IPv6: Komfort für Backbone, kein Ersatz für Erreichbarkeit

- **Backbone-Knoten**: IPv6 ist willkommen (global routbare Adressen, kein NAT),
  ändert das Modell aber nicht.
- **Leaf-Knoten/Notebooks**: IPv6 löst das Problem **nicht**. Adressierbarkeit ≠
  Erreichbarkeit — Host-Firewall und Router blockieren eingehend per Default auch
  unter IPv6; Privacy Extensions (RFC 4941) und Präfixwechsel machen Adressen
  zudem instabil. Das ausgehende-Verbindung-Modell aus Punkt 3 bleibt nötig.

### 5. Transport und Ports

- **Backbone↔Backbone**: Default `7373` (ADR-0015), TLS.
- **Leaf→Backbone**: ausgehend, **TLS auf Port 443** als Fallback, damit auch
  restriktive Client-Netze passiert werden. Keinen Inbound-Port öffnen.
- Es gibt **keinen** Port, der eingehende Firewall-Regeln umgeht; der zuverlässige
  Weg ist die ausgehende Verbindung.
- **Reputations-/AV-Verhalten**: Schadsoftware-Erkennung reagiert auf *Verhalten*,
  nicht auf Portnummern. Daher: ausgehende TLS-Verbindungen zu wenigen, benannten
  Endpunkten; **kein** eingehender Listener auf Leaf-Knoten; **kein** Port-/
  Adress-Scanning; moderate, vorhersehbare Gossip-Frequenz.

## Konsequenzen

### Positiv

- Kein Single Point of Failure; Verzeichnis überlebt Ausfall einzelner Knoten.
- Private Notebooks können teilnehmen, ohne eingehend erreichbar zu sein.
- Baut auf bestehender Mount-/Auth-/PING-Schicht auf (ADR-0022/0023/0025);
  kein fremdes P2P-Framework.
- Opt-in-Trust und Governance-Nachvollziehbarkeit aus ADR-0025 bleiben erhalten.

### Negativ / Risiken

- Backbone-Knoten tragen die Verfügbarkeitslast; ihre Auswahl/Betrieb braucht eine
  Konvention (wer ist „stabil genug"?).
- Gossip-Protokoll muss gegen Fehlinformationen/Spoofing abgesichert werden —
  Trust-Beweis liefert erst ADR-0010 (Domain-Proof).
- Liveness-/Inkarnations-Logik braucht sorgfältige Spezifikation gegen Flapping.
- Datenschutz: ein Verzeichnis offenbart Teilnehmer-Endpunkte; Opt-out/Scoping
  (analog ADR-0025 `advertise: false`) ist nötig.

## Offene Punkte (zu diskutieren)

- Gossip-Mechanik konkret: SWIM vs. einfacheres periodisches Anti-Entropy-Pull.
- Verhältnis zu ADR-0012 — ist dieses Verzeichnis *der* verteilte Index (Option B)
  oder eine darunterliegende Membership-Schicht?
- Wie wird ein Knoten zum Backbone-Knoten (Selbstdeklaration, Domain-Proof,
  Mindest-Uptime)?
- Datenmodell des Verzeichniseintrags und Abgrenzung zu Mounts (ADR-0022).
- Brauchen Leaf-Knoten überhaupt Sichtbarkeit im Verzeichnis, oder nur Lesezugriff?
- Relay-/Hole-Punching als spätere Option für echte Leaf-zu-Leaf-Kommunikation?

## Referenz

- [ADR-0009 — Cosmos-Netzwerk und Core Engine](ADR-0009-nomos-cosmos-network-and-core-engine.md)
- [ADR-0010 (DRAFT) — Well-known-Endpoint und Domain-Proof](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md)
- [ADR-0012 (DRAFT) — Globaler Cosmos-Index-Service](ADR-0012-DRAFT-global-cosmos-index-service.md)
- [ADR-0015 — Default-Port 7373](ADR-0015-default-http-port-7373.md)
- [ADR-0022 (DRAFT) — Cosmos Explorer Server- und Repository-Mounts](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)
- [ADR-0023 (DRAFT) — Inter-Server-Authentifizierung](ADR-0023-DRAFT-inter-server-authentication.md)
- [ADR-0025 (DRAFT) — Server-Discovery via PING](ADR-0025-DRAFT-server-discovery-ping.md)
