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

### 1. Views sind deklarative Artefakte mit pluggable Engine (git-first)

Eine View wird als deklaratives **View-Artefakt** im Repository gespeichert —
git-first ([ADR-0001](ADR-0001-git-first-source-of-truth.md)). Entscheidend:
**welcher Editor/Renderer ein Artefakt verarbeitet, bleibt offen.** Das Artefakt
ist daher **engine-getaggt** und trägt das **engine-native Schema** als Inhalt —
genau wie BPMN/DMN das bpmn.io-native XML als Quelle ablegen und mit
bpmn-js/dmn-js rendern.

Ein dünner Nomos-**Envelope** (Metadaten + Engine-Discriminator) umschließt das
engine-native Schema:

```yaml
id: VIEW-ACC-CAPTURE-001
type: view
name: Benutzerkonto erfassen
engine: form-js            # Renderer-Discriminator (form-js | forms-js | …)
engine_version: "1"
schema:                    # engine-natives Schema, unverändert (hier: @bpmn-io/form-js)
  type: default
  components:
    - { type: textfield, key: given_name, label: Vorname, validate: { required: true } }
    - { type: datetime,  key: birthdate, label: Geburtsdatum, subtype: date }
binding:                   # optionale Nomos-Verknüpfung (engine-unabhängig)
  data_object_ref: identity.blumer.cloud/user-account/PersonData
  submit: identity.blumer.cloud/user-account#createAccount
```

`engine` ist ein offener Discriminator: pro Engine gibt es einen registrierten
Renderer im Web-UI. Das `schema` wird **unverändert** gespeichert und ausgeliefert;
Nomos interpretiert nur Envelope + optionale `binding`-Verknüpfung. So lassen sich
mehrere Editoren (form-js, forms.js, künftige) nebeneinander betreiben, ohne das
Artefaktmodell zu ändern.

### 2. REST liefert die Spec, nicht gerendertes UI

Die View wird als **Spec** über REST ausgeliefert — lokal, repository-scoped und
über den Mount-Proxy (`/api/v1/mounts/{id}/r/…`). Ein entfernter Server liefert
seine View also als **Modell**, exakt wie BPMN/DMN ihr XML liefern. Es wird **kein**
server-gerendertes HTML/JS ausgeliefert.

### 3. Renderer-Registry im Nomos-Web-UI, erste Engine: @bpmn-io/form-js

Das Web-UI hält eine **Renderer-Registry**: `engine` → registrierter Renderer.
Beim Öffnen einer View wird die Spec geholt und an den zur `engine` passenden
Renderer übergeben. Derselbe Mechanismus stellt lokale **und** entfernte Views dar
(entfernte über den Proxy als Spec geholt, lokal gerendert). Renderer-Bibliotheken
werden mit dem Web-UI ausgeliefert (vendored, wie bpmn-js/dmn-js) — **nicht pro
Server**.

**Erste integrierte Engine: `@bpmn-io/form-js`** (bpmn.io, konsistent mit den
bestehenden bpmn-js/dmn-js-Integrationen). Weitere Engines (z. B. forms.js) können
später als zusätzliche Registry-Einträge ergänzt werden, ohne Artefaktmodell oder
REST zu ändern.

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

- Genaues Schema des Nomos-**Envelope** (Pflichtfelder, `engine`-Werteliste); das
  engine-native `schema` validiert die jeweilige Engine selbst.
- Versionierung/Kompatibilität pro Engine (`engine_version`) und Vendoring-Pinning.
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
