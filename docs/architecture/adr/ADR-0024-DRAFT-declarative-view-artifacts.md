# ADR-0024 (DRAFT): Deklarative View-/Form-Artefakte und lokaler Renderer

## Status

Proposed

## Datum

2026-05-20

## Kontext

User Interfaces sollen als Nomos-Artefakte im Repository gespeichert werden. Heute
ist `ServiceUserInterface` (im Service-Modell, siehe auch
[ADR-0013](ADR-0013-user-contact-interfaces.md)) im Wesentlichen **Metadaten**
(Name, Channel, URL). Es fehlt ein Modell, mit dem eine *erfassende* Oberfläche
(z. B. ein Formular „XY erfassen") versioniert, validiert und — auch über den
föderierten Cosmos — ausgeliefert und dargestellt werden kann.

Es gibt im Repo bereits ein etabliertes Muster für genau dieses Problem: **BPMN**
und **DMN**. Der Server liefert die **deklarative Quelle** (BPMN-/DMN-XML) über
REST; der Client rendert sie **lokal** mit einer mitgelieferten Renderer-Bibliothek
(bpmn-js/dmn-js). Der Server schickt nie ausführbares UI, sondern ein Modell.

Mit den Server-Mounts ([ADR-0022](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md))
und dem authentifizierten Proxy ([ADR-0023](ADR-0023-DRAFT-inter-server-authentication.md))
stellt sich die Frage konkret: Wie liefert ein **entfernter** Server eine View?
Als gerendertes HTML/JS oder als Modell, das lokal gerendert wird?

## Entscheidung

### 1. Views sind deklarative Artefakte (git-first)

Eine View wird als deklaratives **View-/Form-Artefakt** (YAML/JSON) im Repository
gespeichert — analog zu allen anderen Nomos-Artefakten und git-first
([ADR-0001](ADR-0001-git-first-source-of-truth.md)). Das Artefakt beschreibt
**was** dargestellt/erfasst wird, nicht **wie** es technisch gerendert wird:

- Felder (id, label, Typ, Pflicht, Default, Hilfetext),
- Validierungen (wiederverwendbar, vgl. Blueprint-Attribut-Regeln),
- Daten-Bindings (Referenz auf `DataObject`/Service-Method/Decision per kanonischer Ref),
- Layout/Gruppierung (Sektionen, Reihenfolge),
- Aktionen (Submit-Ziel als Service-Method-/REST-Referenz).

Skizze (nicht final):

```yaml
id: VIEW-ACC-CAPTURE-001
type: view
name: Benutzerkonto erfassen
binding:
  data_object_ref: identity.blumer.cloud/user-account/PersonData
sections:
  - title: Stammdaten
    fields:
      - id: given_name
        label: Vorname
        type: text
        required: true
      - id: birthdate
        label: Geburtsdatum
        type: date
        validation: { type: manual }
actions:
  - id: submit
    label: Erfassen
    invokes: identity.blumer.cloud/user-account#createAccount
```

### 2. REST liefert die Spec, nicht gerendertes UI

Die View wird als **Spec** über REST ausgeliefert — lokal, repository-scoped und
über den Mount-Proxy (`/api/v1/mounts/{id}/r/…`). Ein entfernter Server liefert
seine View also als **Modell**, exakt wie BPMN/DMN ihr XML liefern. Es wird **kein**
server-gerendertes HTML/JS ausgeliefert.

### 3. Generischer Renderer im Nomos-Web-UI

Ein generischer, mit dem Nomos-Web-UI ausgelieferter **View-Renderer** (analog
bpmn-js/dmn-js) baut aus der Spec das Formular. Derselbe Renderer stellt lokale
**und** entfernte Views dar; entfernte werden über den Proxy als Spec geholt und
lokal gerendert. Die Renderer-Bibliothek wird **nicht pro Server** ausgeliefert.

### 4. Bindings und Aktionen laufen über REST

Lese-Bindings und Submit-Aktionen referenzieren Nomos-Artefakte per kanonischer
Referenz (DataObject, Service-Method, Decision). Zur Laufzeit gehen sie über REST:
lokal direkt, für entfernte Views über den Mount-Proxy zum Besitzer-Server
(Token/Auth nach [ADR-0023](ADR-0023-DRAFT-inter-server-authentication.md); Lesen
offen, Schreiben tokenpflichtig).

### 5. Kein vom Server geliefertes ausführbares UI

Server-gerendertes HTML/JS, das im lokalen UI ausgeführt wird, wird **verworfen**:
es ist ein XSS-/Trust-Risiko im föderierten Cosmos und bricht die git-first
Reviewbarkeit. Wenn eine **externe/Legacy-Oberfläche** eingebunden werden muss,
geschieht das ausschließlich als explizit deklarierter externer Link (UCI-Metadaten,
[ADR-0013](ADR-0013-user-contact-interfaces.md)), bei Bedarf in einer iframe-Sandbox,
klar getrennt von deklarativen Views.

## Konsequenzen

### Positiv

- **git-first**: Views sind reviewbar, diffbar, validierbar wie jedes andere Artefakt.
- **Föderierung kostenlos**: Cross-Server-Views funktionieren über den vorhandenen
  Mount-Proxy, ohne neuen Mechanismus.
- **Kein Fremdcode** im lokalen UI — kein Trust-/XSS-Problem zwischen Servern.
- **Ein Renderer** für alle Server; konsistentes Look & Feel.
- Gleiches mentales Modell wie BPMN/DMN (Modell + lokaler Renderer).

### Negativ / Risiken

- Der Renderer deckt nur einen **definierten Widget-/Feature-Satz** ab; hochgradig
  maßgeschneiderte UIs sind eingeschränkt. Ein bewusster Escape-Hatch (externer Link)
  ist nötig.
- Spec-Schema und Versionierung müssen sorgfältig spezifiziert und stabil gehalten
  werden (Breaking Changes = eigene ADR).
- Binding-/Ausdruckssprache muss definiert werden (ggf. FEEL aus dem DMN-Umfeld
  wiederverwenden, ADR-0019).

## Offene Punkte

- Genaues JSON-Schema des View-Artefakts und der unterstützte Widget-/Feldtyp-Satz.
- Binding-/Ausdruckssprache für berechnete Felder/Sichtbarkeiten (FEEL wiederverwenden?).
- Mapping von `actions.invokes` auf konkrete Schreib-APIs (Service-Method-Aufruf vs. REST).
- Verhältnis zu UCI ([ADR-0013](ADR-0013-user-contact-interfaces.md)): ist eine View
  ein UCI-Subtyp oder ein eigenständiges Artefakt?
- Theming, i18n, Barrierefreiheit des Renderers.
- Caching/Invalidierung entfernter View-Specs.

## Referenz

- [ADR-0001 — Git-first](ADR-0001-git-first-source-of-truth.md)
- [ADR-0013 — User Contact Interfaces (UCI)](ADR-0013-user-contact-interfaces.md)
- [ADR-0019 — DMN 1.5 / FEEL Evaluation Engine](ADR-0019-dmn-feel-engine-scope.md)
- [ADR-0022 (DRAFT) — Cosmos Explorer Server- und Repository-Mounts](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)
- [ADR-0023 (DRAFT) — Inter-Server-Authentifizierung](ADR-0023-DRAFT-inter-server-authentication.md)
