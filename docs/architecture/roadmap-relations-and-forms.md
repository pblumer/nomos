# Roadmap: Beziehungen und Form-Komposition

Stand: 2026-05-26 · Status: Vorschlag

Diese Roadmap fasst die Umsetzung der vier zusammenhängenden ADRs in
Phasen mit klaren Meilensteinen zusammen. Sie ersetzt keine
PR-genauen Implementierungs-Prompts (siehe `docs/implementation-prompts/`),
sondern legt **Reihenfolge und Schnitt** fest.

## Beteiligte ADRs

- **ADR-0024** — Deklarative View-/Form-Artefakte *(Baseline, bereits in
  Umsetzung)*
- **ADR-0031** — Referenz-basierte Form-Komponenten (`selectRef`, Kaskade)
- **ADR-0032** — Form-Komposition über eingebettete Subforms (`formRef`)
- **ADR-0033** — Beziehungs-Editor für Typen und Propagation in Standard-Views

## Leitprinzipien für die Reihenfolge

1. **Modell vor UI vor Generator** — eine Beziehung muss erst sauber
   persistiert sein, bevor irgendwas anderes davon abhängen darf.
2. **Statisch vor dynamisch** — `selectRef` mit fixer Optionsliste vor
   Kaskade mit FEEL-Filter; macht jede Phase isoliert testbar.
3. **Read-only vor editierbar** — `formRef` mit `action: get` (Anzeige) vor
   `bind` (verschachtelte Bearbeitung).
4. **Auto-Generate beim Neu-Anlegen vor Propagations-Dialog** — der
   Suggest-Flow ist die teurere UX und kommt zuletzt.

## Phasen

### Phase 1 — Beziehungs-Editor end-to-end *(ADR-0033, Abschnitt 1 + 6)*

Ziel: „Verbinden" im Typ-Detail-Panel persistiert eine `TypeDependency` und
eine `DataRelation` in den entsprechenden YAMLs; Bearbeiten und Löschen
funktionieren; das ERD spiegelt den Stand.

**Meilenstein 1.1**: Backend-API für `TypeDependency`/`DataRelation` —
Anlegen, Aktualisieren, Löschen über die bestehende Repo-Save-Schicht.

**Meilenstein 1.2**: Picker-Dialog im UI (Ziel-Typ, Relations-Label,
Required-Flag, optional Feldname); Liste rechts neben dem ERD; Kanten-
Klick für Edit/Delete.

**Meilenstein 1.3**: Validierung (Ziel existiert, keine Duplikate);
Föderation: Ziel-Typ darf in einem Mount liegen (kanonische Referenz nach
ADR-0028).

**Akzeptanz**: Neue Demo „Person → Address" lässt sich vollständig im UI
anlegen, im Git-Diff erscheinen ausschließlich die zwei erwarteten
YAML-Änderungen. Forms werden in dieser Phase noch **nicht** angefasst.

---

### Phase 2 — `selectRef`, statisch *(ADR-0031, Abschnitt 1 + 4 ohne Kaskade)*

Ziel: Eine Form kann ein referenz-basiertes Auswahlfeld haben, das alle
Instanzen eines Zieltyps zeigt — ohne Filter, ohne Abhängigkeiten.

**Meilenstein 2.1**: REST-Options-Endpoint, der `(typeRef, valueField,
labelField)` zu einer Optionsliste auflöst (lokal + Mount).

**Meilenstein 2.2**: Custom Component `selectRef` im form-js-Renderer,
inkl. Property-Panel im Editor.

**Meilenstein 2.3**: `viewgen` erzeugt für jede `DataRelation` beim
**Erstanlegen** eines Typs automatisch ein `selectRef`-Feld in `_new`,
`_edit`, `_list` (statt eines stumpfen Textfelds). Bestehende Forms
werden noch nicht angefasst.

**Akzeptanz**: Bestellung-Demo zeigt eine Dropdown-Liste mit allen
Kategorien aus dem lokalen Repo; Wechsel speichert die Kategorie-ID; im
Edit-Form ist die vorhandene Auswahl korrekt vorbelegt.

---

### Phase 3 — Kaskade *(ADR-0031, Abschnitt 2 + 3)*

Ziel: Ein `selectRef` filtert seine Optionen abhängig von anderen
Feldwerten desselben Forms (FEEL).

**Meilenstein 3.1**: Filter-Ausdruck (`filter:`) wird serverseitig via
FEEL-Engine (ADR-0019) ausgewertet; Optionen-Endpoint nimmt Ausdruck +
Bindings entgegen, prüft Permissions vor Filter-Anwendung.

**Meilenstein 3.2**: `dependsOn` triggert Re-Fetch im Renderer; bei
geleertem Parent werden Child-Werte verworfen, wenn nicht mehr im
Resultat enthalten.

**Meilenstein 3.3**: Editor-Support — Property-Panel zeigt
`dependsOn`/`filter` mit FEEL-Linter und „Felder dieses Forms"-Hinweis.

**Akzeptanz**: Bestellformular zeigt Produkte gefiltert nach gewählter
Kategorie; bei Kategorienwechsel passt sich die Liste ohne Reload an;
ein vorher gewähltes Produkt wird zurückgesetzt, falls es nicht zur
neuen Kategorie gehört.

---

### Phase 4 — `formRef`, read-only *(ADR-0032, Abschnitt 1–5, 7 (`get`-Teil), 8)*

Ziel: Eine Form kann eine andere Form als Subform einbetten, befüllt per
`get` über eine ID.

**Meilenstein 4.1**: Custom Component `formRef` mit `source.action: get`
und `dependsOn`; rendert das Sub-`.frm` an Komponentenposition (Header/
Submit unterdrückt).

**Meilenstein 4.2**: Editor-UX (ADR-0032 Abschnitt 9) — Toolbox-Eintrag
„Subform", Form-Picker mit Filter nach Typ + Variante, WYSIWYG-Render im
Canvas, Property-Panel-Bindings.

**Meilenstein 4.3**: Preview-Modi im Editor: Struktur (Default), Live
(erste Instanz fetchen), Mock (JSON im Editor-Setting); Zyklenschutz +
Tiefenlimit greifen auch im Editor-Rendering.

**Meilenstein 4.4**: Föderation — `formRef` auf entfernten Typ läuft über
Mount-Proxy ohne neuen Code-Pfad.

**Akzeptanz**: Bestellung-Demo zeigt nach Auswahl der Kundin deren
`Person_short.frm` inline; Auswahlwechsel aktualisiert den Block; ein
absichtlich gebauter Zyklus (A→B→A) wird vom Lint blockiert.

---

### Phase 5 — `formRef`, editierbar *(ADR-0032, Abschnitt 3 (`bind`), 7 (`bind`/`inline`))*

Ziel: Subforms können Teil des Eltern-Submits sein.

**Meilenstein 5.1**: `source.action: bind` — Subform-Werte werden im
JSON des Eltern-Submits unter `key` verschachtelt; Validitäts-Beitrag
zum Eltern-Form.

**Meilenstein 5.2**: `source.action: inline` für Seed-Daten /
Default-Werte.

**Meilenstein 5.3**: Fehler-/Validierungspropagation vom Sub- ins
Eltern-Form-Panel (Anker auf Fehlerfeld klickbar).

**Akzeptanz**: Eine neue Person kann direkt im Bestellformular angelegt
werden, ohne den Kontext zu verlassen; der Submit speichert beide
Artefakte konsistent oder bricht atomar ab.

---

### Phase 6 — Propagation in bestehende Forms *(ADR-0033, Abschnitt 2 + 4 + 5)*

Ziel: Wenn eine Beziehung **nach** Erstanlage des Typs hinzukommt oder
sich ändert, schlägt das System die passenden Form-Updates vor — ohne
User-Edits zu überschreiben.

**Meilenstein 6.1**: Diff-Engine im `viewgen` — vergleicht aktuellen
Relations-Stand mit dem Stand der Standard-Forms und erzeugt eine Liste
„soll eingefügt / aktualisiert werden".

**Meilenstein 6.2**: Suggest-Dialog im UI (Vorschau pro Variante,
einzeln an-/abwählbar, Einfügen am Form-Ende, „Ignorieren"-Stand wird
im Sidecar gemerkt).

**Meilenstein 6.3**: Dangling-Lint für gelöschte Relationen
(`view`-Lint markiert betroffene `selectRef`/`formRef`-Komponenten;
Form-Editor zeigt Warn-Marker).

**Meilenstein 6.4**: Match-Heuristik — neue `relation:`-Annotation an
referenz-basierten Komponenten, damit der Diff erkennt, ob ein Feld
„schon zu dieser Relation gehört" und nicht doppelt einfügt.

**Akzeptanz**: An einem bestehenden, vom User layoutmäßig angepassten
Form wird nach Hinzufügen einer neuen Relation der Vorschlag korrekt
gezeigt; nach Bestätigung enthält die `.frm` exakt das vorgeschlagene
Diff, sonst nichts.

## Abhängigkeitsgraph (Kurzform)

```
Phase 1 (Beziehungs-Editor)
   └── Phase 2 (selectRef statisch)
          ├── Phase 3 (Kaskade)
          └── Phase 4 (formRef read-only)
                 ├── Phase 5 (formRef editierbar)
                 └── Phase 6 (Propagation)  ← braucht 2 + 4 als Insertion-Targets
```

Phase 1 ist isoliert nutzbar (Demo-Wert sofort). Phasen 3 und 4 können
parallel laufen, sobald 2 steht. Phase 5 und 6 hängen beide an Phase 4
(Propagation muss `formRef` als möglichen Vorschlag erzeugen können) und
sollten in dieser Reihenfolge angegangen werden, weil 5 die einfachere
Lerneinheit ist.

## Querschnitts-Themen

Diese Punkte werden phasenübergreifend mitgeführt, nicht in einer
Einzelphase „abgehakt":

- **FEEL-Subset** für Filter — Definition in Phase 3, Wiederverwendung
  in allen späteren Phasen.
- **Paginierung / Typeahead** für große Optionsmengen — frühestens in
  Phase 3 nötig; konkrete Form (Endpoint-Shape, UI-Schwellwerte) in
  einem separaten Mini-ADR fixieren, sobald die ersten realistischen
  Datenmengen vorliegen.
- **Permissions** — der Options-Endpoint muss Autorisierung **vor**
  Filter-Anwendung greifen; gilt ab Phase 2.
- **Dangling-Verhalten** — gelöschte Beziehungen / gelöschte
  Zielobjekte: konsequent als sichtbare Warnung, nie als stille
  Bereinigung. In Phase 6 für Beziehungen, in Phase 4 für
  Zielinstanzen.
- **i18n des `labelField`** — offener Punkt aus ADR-0031, wird wenn
  möglich gemeinsam mit der i18n-Strategie für Forms (ADR-0024 offener
  Punkt) entschieden, nicht in dieser Roadmap blockierend.

## Was diese Roadmap explizit **nicht** abdeckt

- Konkrete API-Shapes, Datei-Pfade, Funktions-Signaturen → kommen in
  PR-genauen Implementierungs-Prompts unter `docs/implementation-prompts/`.
- UX-Detail-Design (Icons, Wording, Farbcodierung der Lint-Warnungen) →
  begleitend in Phase 1 und Phase 6, nicht als eigener Meilenstein.
- Migrationspfad für vorhandene `.frm`-Dateien mit Textfeldern, die
  „eigentlich" Relationen sein sollten → optional in Phase 6 als
  Lint-Vorschlag.
