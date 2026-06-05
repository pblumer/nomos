# ADR-0031: Referenz-basierte Form-Komponenten und abhängige Auswahlfelder

## Status

Proposed

## Datum

2026-05-26

## Kontext

Forms in Nomos sind deklarative View-Artefakte mit pluggable Engine
([ADR-0024](ADR-0024-DRAFT-declarative-view-artifacts.md)); die erste integrierte
Engine ist `@bpmn-io/form-js`. Auswahlfelder (`select`, `checkbox group`,
`radio group`, `tag list`) tragen ihre Optionen heute als **statische** `values`-
Liste direkt im Schema (`internal/repo/repo.go`, `internal/viewgen/viewgen.go`).

Auf Typ-Ebene gibt es bereits ein etabliertes Konzept für Beziehungen:
`TypeDependency` und `DataRelation` (`internal/model/model.go`) erlauben es,
dass ein Typ einen anderen referenziert (z. B. `Product` → `Category`). Was fehlt,
ist die Brücke vom Form-Component zu dieser Modellrelation:

- Ein Erfassungsformular für eine Bestellung soll eine Kategorie und ein Produkt
  aus den **vorhandenen Instanzen** der jeweiligen Typen auswählen lassen.
- Die Produktliste soll sich an der gewählten Kategorie ausrichten (kaskadierende
  bzw. abhängige Auswahl).
- Wenn neue Kategorien/Produkte angelegt werden, sollen sie ohne Form-Edit
  verfügbar sein.

ADR-0024 hat „Binding-/Ausdruckssprache für berechnete Felder/Sichtbarkeiten" und
„Mapping von `actions.invokes` auf konkrete Schreib-APIs" explizit als offene
Punkte aufgeführt. Dieser ADR adressiert den Teil **„Optionen aus dem Datenmodell
beziehen, optional gefiltert durch andere Feldwerte"**.

## Entscheidung

### 1. Neuer Component-Subtyp: `selectRef`

Auswahl-Komponenten erhalten einen referenzbasierten Subtyp. Statt einer
statischen `values`-Liste deklariert das Component, **welcher Typ** als Quelle
dient und welche Felder als Label/Wert verwendet werden:

```yaml
- type: selectRef
  key: category
  label: Kategorie
  typeRef: .nomos/types/Category   # kanonische Typ-Referenz
  valueField: id                    # was im Form-Wert landet (default: id)
  labelField: name                  # was im UI angezeigt wird
```

`selectRef` ist eine **Nomos-Erweiterung** der form-js-Komponente und wird im
Web-UI als Custom Component registriert. Das engine-native Schema bleibt damit
gültig (form-js ignoriert unbekannte Typen via Fallback-Renderer, der Nomos-
Renderer übernimmt sie explizit). Analoge Subtypen für Mehrfachauswahl
(`checkboxGroupRef`, `tagListRef`) folgen demselben Muster.

### 2. Abhängige Felder: `dependsOn` + `filter`

Ein referenzbasiertes Feld kann von einem anderen Feld desselben Formulars
abhängen:

```yaml
- type: selectRef
  key: product
  label: Produkt
  typeRef: .nomos/types/Product
  labelField: name
  dependsOn: [category]
  filter: "category == ${category}"
```

- `dependsOn` listet die Form-Keys, deren Änderung eine Neuabfrage auslöst.
- `filter` ist ein **FEEL-Ausdruck** (ADR-0019), ausgewertet im Kontext der
  aktuellen Form-Werte. `${key}` ist Syntaxzucker für FEEL-Variablenzugriff.
- Solange ein `dependsOn`-Feld leer ist, ist das abhängige Feld **disabled** und
  zeigt einen Hinweis („Erst Kategorie wählen").
- Wechselt das Parent-Feld, wird der eigene Wert gelöscht, wenn er nicht mehr im
  gefilterten Ergebnis enthalten ist.

### 3. Auflösung serverseitig über bestehende REST-API

Der Renderer holt die Optionen zur Laufzeit über die bereits bestehende
Data-/Instance-API. Konzeptionell:

```
GET /api/v1/instances?type=<typeRef>&filter=<FEEL>&fields=<valueField,labelField>
```

- Die Filter-Auswertung passiert **serverseitig** (FEEL-Engine, ADR-0019). Der
  Client schickt nur den Ausdruck und die aktuellen Bindings.
- Für entfernte Typen geht der Aufruf über den Mount-Proxy
  ([ADR-0022](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md),
  [ADR-0023](ADR-0023-DRAFT-inter-server-authentication.md)) — derselbe Pfad wie
  für die View-Spec selbst.
- Antworten enthalten ausschließlich die deklarierten Felder; das verhindert
  unbeabsichtigtes Leaken weiterer Properties an das Frontend.

Ergebnisse werden pro Form-Session und `(typeRef, filter-resolved)`-Schlüssel
gecached; eine explizite Invalidierung erfolgt bei Änderung eines `dependsOn`-
Wertes.

### 4. Modell-Verankerung: Relation statt frei wählbarer `typeRef`

Wenn der Form-Key einer **deklarierten Relation** des bindenden Typs entspricht
(`DataRelation` in `internal/model/model.go`), kann `typeRef` weggelassen werden
— `viewgen` füllt ihn aus dem Modell. Damit bleibt das Modell die Quelle der
Wahrheit; der Form-Editor kann referenzbasierte Felder per Klick aus der Relation
generieren („Feld aus Relation einfügen").

Freistehende `typeRef` ohne Modellrelation bleiben erlaubt, sind aber
**Lint-warn**: sie deuten auf eine fehlende Relation im Typmodell hin.

### 5. `viewgen`-Defaults

Der Generator für die vier Standard-Views eines Typs (ADR-0024) erzeugt für jede
`DataRelation` automatisch ein `selectRef`-Feld im `_new`- und `_edit`-Formular,
mit `labelField` aus dem ersten Property des Zieltyps, das als „display name"
markiert ist (Fallback: `id`). Kaskaden (`dependsOn`) werden **nicht**
automatisch erzeugt — sie sind eine bewusste Form-Design-Entscheidung und müssen
manuell ergänzt werden.

## Konsequenzen

### Positiv

- Beziehungen aus dem Typmodell werden im UI **lebendig**, ohne dass Forms bei
  jeder neuen Instanz angepasst werden müssen.
- Kaskadierende Auswahl ist ein deklaratives Feature, kein Code im Form-Editor.
- FEEL als einheitliche Ausdruckssprache (ADR-0019) — derselbe Stack wie
  Decisions; ein Nutzer lernt eine Sprache, nicht zwei.
- Föderierung kostenlos: referenzierte Typen können auf entfernten Servern liegen,
  Auflösung läuft über den vorhandenen Mount-Proxy.
- Form-Artefakt bleibt git-first reviewbar und diffbar (ADR-0001) — nur das
  Schema ändert sich, kein Code.

### Negativ / Risiken

- Form-Vorschau ohne Backend ist nur eingeschränkt möglich (Optionen brauchen
  Server). Der Editor braucht einen Stub-/Mock-Modus.
- FEEL-Auswertung auf jedem Options-Request kostet — bei großen Instanzmengen
  ist Pagination/Suche nötig (siehe Offene Punkte).
- Eingeschränkte Form-Engines (form-js) müssen Custom Components registrieren;
  zusätzliche Engines (forms.js, …) müssen den Subtyp eigenständig implementieren
  oder verlieren das Feature.
- Filter-Ausdrücke sind potentiell schreibsensitiv (z. B. „nur Produkte, die der
  aktuelle Nutzer sieht"). Autorisierung muss serverseitig **vor** Filter-Anwendung
  greifen, nicht nachgelagert.

## Offene Punkte

- Konkrete Form der Options-API: dedizierter `/api/v1/options`-Endpoint vs.
  Wiederverwendung der allgemeinen Instance-Query (ADR-0016).
- Pagination und serverseitige Suche (`?q=…`) für große Ergebnismengen; ab welcher
  Größe wechselt das UI von „alles laden" zu „typeahead"?
- Verhältnis zu `binding.data_object_ref` (ADR-0024): liest ein Edit-Form den
  aktuellen Wert eines Referenzfelds als ID oder als denormalisiertes Snapshot-
  Objekt?
- Schreibverhalten: Was passiert beim Submit, wenn das referenzierte Objekt
  zwischenzeitlich gelöscht wurde? (Hard fail vs. „dangling reference"-Marker.)
- FEEL-Subset für Filter: welche Funktionen sind erlaubt, welche aus Sicherheits-
  oder Performancegründen ausgeschlossen?
- Lokalisierung des `labelField` (i18n) — analog zu ADR-0024 ein offener Punkt
  des Renderers.

## Referenz

- [ADR-0001 — Git-first](ADR-0001-git-first-source-of-truth.md)
- [ADR-0016 (DRAFT) — Graph-Speicherung und Abfragemodell](ADR-0016-DRAFT-graph-storage-and-query-model.md)
- [ADR-0019 — DMN 1.5 / FEEL Evaluation Engine](ADR-0019-dmn-feel-engine-scope.md)
- [ADR-0022 (DRAFT) — Cosmos Explorer Server- und Repository-Mounts](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)
- [ADR-0023 (DRAFT) — Inter-Server-Authentifizierung](ADR-0023-DRAFT-inter-server-authentication.md)
- [ADR-0024 (DRAFT) — Deklarative View-/Form-Artefakte](ADR-0024-DRAFT-declarative-view-artifacts.md)
