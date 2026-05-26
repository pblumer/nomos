# ADR-0033 (DRAFT): Beziehungs-Editor für Typen und Propagation in Standard-Views

## Status

Draft

## Datum

2026-05-26

## Kontext

Das Modell kennt Beziehungen zwischen Typen bereits in zwei Schichten
(`internal/model/model.go`):

- `TypeDependency` (auf `TypeDef.Dependencies`): „Typ X darf/muss auf Typ Y
  verweisen", semantisches Verhältnis (`uses`, `has`, …), `Required`-Flag.
- `DataRelation` (auf `DataObject.Relations`): das FK-Pendant auf Datenebene
  mit `Target` und `Field`, damit Generatoren konkrete Felder kennen.

Das ERD-Modell speichert nur die **Layout-Boxen** (`.erd`-Dateien); Kanten
werden live aus den Type-YAMLs abgeleitet — die Type-YAMLs sind die Quelle
der Wahrheit (`model.go:84-85`).

In der UI sichtbar ist heute der Tab **„Beziehungen"** in der Typ-Detailansicht
mit einem ERD-Ausschnitt für den aktuellen Typ und einem **„Verbinden"**-Button.
Was fehlt:

1. **Persistenz-Pfad**: „Verbinden" hat keinen End-to-End-Flow, der eine neue
   `TypeDependency` (und ggf. `DataRelation`) in die Typ-YAML schreibt.
2. **Propagation in Forms**: Selbst wenn die Relation existierte, würden die
   Standard-Views des Typs (`_new`, `_edit`, `_list`, `_short` — ADR-0024)
   sie nicht zeigen. Das Modell wirkt tot.

ADR-0031 (`selectRef`) und ADR-0032 (`formRef`) liefern die Laufzeit-
Bausteine, mit denen Beziehungen in Forms dargestellt werden — sie sagen
aber **nicht**, wie eine neu definierte Beziehung dort landet, ohne
manuell editierte Forms zu überschreiben. ADR-0024 § Standard-Views deutet
nur grob an: „Ein erneutes Speichern der Typdefinition ergänzt nur
fehlende Varianten und überschreibt vom Nutzer bearbeitete Forms nicht."

Dieser ADR schließt beide Lücken.

## Entscheidung

### 1. Beziehungs-Editor schreibt in die Typ-YAML

Das Beziehungs-Panel in der Typ-Detailansicht erhält einen vollständigen
CRUD-Pfad:

**Anlegen** (über „Verbinden" oder Drag im ERD):
1. Ziel-Typ wählen (Picker mit lokalen + gemounteten Typen).
2. Relations-Label (`uses`, `has`, `belongs_to`, …; freie Eingabe mit
   Vorschlägen) und `Required`-Flag.
3. Optional: Feldname auf Datenebene (Default abgeleitet aus Ziel-Typ +
   Konvention, z. B. `customer_id` → ein lower_snake_case-Default).
4. Speichern → `TypeDependency` wandert in `TypeDef.Dependencies`,
   `DataRelation` (mit `Target` und `Field`) in das Haupt-DataObject des
   Quelltyps. Beides in einem Commit über die bestehende Repo-Save-API.

**Bearbeiten / Löschen**: Klick auf Kante im ERD oder Eintrag in einer
Liste → Inline-Edit oder Entfernen.

**Validierung beim Speichern**:
- Ziel-Typ muss existieren (lokaler Repo oder erreichbarer Mount).
- Keine duplizierten Relationen mit identischem `Target` + `Relation`-Label.
- Selbstreferenzen sind erlaubt (z. B. „Person ist Vorgesetzter von Person"),
  aber als solche markiert.

Damit ist das Beziehungs-Panel nicht länger reine Anzeige, sondern der
**bevorzugte Editor** für Type-Relations; direktes YAML-Editing bleibt
selbstverständlich möglich (git-first).

### 2. Propagation als Vorschlag, nicht als Überschreiben

Wenn das Speichern einer Typdefinition eine **Differenz** in den
Relationen ergibt (neu, geändert, gelöscht), berechnet `viewgen` einen
**Propagations-Vorschlag** für die betroffenen Standard-Views — er führt
ihn aber nicht automatisch aus.

Die User-Interaktion:

```
Neue Beziehung Person → Address (has, required) erkannt.
Soll sie in die Standard-Views eingefügt werden?

  [x] Person_new.frm   – Feld "address" (selectRef)
  [x] Person_edit.frm  – Feld "address" (selectRef) + Detail (formRef short)
  [ ] Person_short.frm – nur Name anzeigen
  [x] Person_list.frm  – Spalte "address"

  [Vorschau aller Änderungen]  [Einfügen]  [Ignorieren]
```

- **Auto-Vorschlag**: `selectRef` ist Default für jede neue Relation;
  `formRef` (Variante `short`, read-only) wird vorgeschlagen, wenn die
  Variante des Eltern-Forms `edit` oder `new` ist.
- **Position**: Neue Felder werden ans **Ende** des Forms angehängt,
  damit bestehendes Layout nicht aufgerissen wird. Die Nutzerin kann sie
  danach im Form-Editor verschieben.
- **Idempotenz**: Wer „Ignorieren" wählt, bekommt denselben Vorschlag
  nicht wieder, bis sich die Relation erneut ändert (Stand wird im
  ERD-/Typ-Sidecar gemerkt — kein Eingriff in die Form-Artefakte).

Damit überschreibt das System **nie** ein vom Nutzer bearbeitetes Form;
ADR-0024 § Standard-Views wird konkretisiert: Ergänzungen passieren nur
auf explizite Bestätigung, niemals stillschweigend beim Type-Save.

### 3. Erstanlage eines Typs: weiterhin Auto-Generierung

Beim **erstmaligen** Generieren der vier Standard-Views (Typ ist neu, noch
keine `.frm` vorhanden) verhält sich `viewgen` wie heute: Forms werden aus
Properties **und** Relationen direkt erzeugt — ohne Dialog. Der
Vorschlags-Dialog aus Abschnitt 2 greift nur, wenn die `.frm`-Dateien
bereits existieren.

So bleibt der „neuer Typ"-Pfad reibungslos und das Suggest-Verhalten
beschränkt sich auf Änderungen am bestehenden Bestand.

### 4. Löschen einer Relation: Lint statt automatischer Eingriff

Wird eine Relation entfernt, **bleiben** Felder in Forms, die sie nutzten,
zunächst stehen. Stattdessen:

- `view`-Lint markiert die betroffenen `selectRef`/`formRef`-Komponenten
  als **dangling**: „Feld `address` in `Person_edit.frm` verweist auf eine
  nicht mehr existierende Beziehung."
- Im Form-Editor erscheint das Feld mit Warn-Markierung; das
  Property-Panel bietet „Entfernen" oder „Auf andere Relation umbiegen" an.

Begründung: ein gelöschtes Form-Feld kann später nicht ohne weiteres
rekonstruiert werden (Reihenfolge, individuelle Anpassungen) — eine
sichtbare Warnung ist sicherer als stille Bereinigung.

### 5. Föderation

Wenn die Beziehung auf einen **entfernten** Typ zeigt (Mount, ADR-0022),
verweist die `TypeDependency` einfach über die kanonische Referenz
(ADR-0028). Der Beziehungs-Editor löst Ziel-Typen über Mount-Proxy auf;
beim Propagations-Vorschlag werden die entsprechenden `selectRef`-/
`formRef`-Bindings (siehe ADR-0031/0032) bereits mit der föderierten
Referenz vorbefüllt — kein zusätzlicher Mechanismus nötig.

### 6. UI-Skizze: Beziehungs-Panel

| Element | Verhalten |
| --- | --- |
| ERD-Ausschnitt | Aktueller Typ in der Mitte, direkte Nachbarn drumherum; Kanten beschriftet mit Relations-Label. |
| Liste rechts | Tabellarische Sicht aller Relationen mit Spalten `Ziel`, `Relation`, `Required`, `Datenfeld`. |
| Button „Verbinden" | Öffnet Picker (Schritt 1–3 aus Abschnitt 1). |
| Kante / Listen-Zeile anklicken | Inline-Edit oder „Entfernen". |
| Save-Button | Persistiert in Typ-YAML; löst danach automatisch den Propagations-Vorschlag aus Abschnitt 2 aus, sofern Forms vorhanden sind. |

## Konsequenzen

### Positiv

- Beziehungen sind **End-to-End** editierbar — vom UI bis ins Form.
- Forms bleiben **Single-Source** auf dem Stand des Typmodells, ohne dass
  Nutzer manuell Felder synchronisieren müssen.
- User-Edits an Forms werden nie überschrieben; das System schlägt vor,
  entscheidet aber nicht autonom (passt zu ADR-0004 — „KI-Assistenz nur
  für Drafts").
- ADR-0031 und ADR-0032 bekommen einen klar definierten Auslöser-Pfad
  („wer und wann erzeugt eigentlich diese Komponenten?").
- Föderierte Beziehungen funktionieren ohne neuen Mechanismus.

### Negativ / Risiken

- Der Propagations-Dialog ist ein zusätzlicher UI-Schritt — manche Nutzer
  werden „immer einfügen" wollen; eine User-Preference dafür ist ein
  späteres Polish-Ticket, nicht Kern dieses ADR.
- Dangling-Felder bleiben sichtbar liegen, bis jemand sie quittiert; bei
  schlechter Disziplin sammeln sich Warnings.
- Konvention für Default-Feldnamen (Abschnitt 1) ist eine kleine, aber
  echte Designentscheidung — falsche Defaults erzeugen schiefe Datenmodelle.
- Der Propagations-Vorschlag muss diffen können, was an einem Form schon
  bearbeitet wurde, ohne sich darin zu verheddern (z. B. wenn der Nutzer
  ein selectRef-Feld zur gleichen Relation bereits per Hand eingefügt hat —
  doppelt einfügen wäre falsch).

## Offene Punkte

- Konvention für Default-Datenfeldnamen aus Relationen
  (`customer_id` vs. `customer` vs. `customer_ref`).
- Wie genau erkennt der Propagator, dass ein User-Edit-Feld „schon zu
  dieser Relation gehört" (Match per `relation`-Annotation am Component?
  Match heuristisch über Feldnamen?). Vermutlich braucht der
  `selectRef`-Component ein optionales `relation: <name>`-Tag.
- Wenn ein Typ extern liegt und der Mount nicht erreichbar ist — wie
  präsentiert der Beziehungs-Editor ihn? Read-only mit Hinweis? Sperren?
- Verhalten, wenn `cosmos.yaml` die Sichtbarkeit eines referenzierten Typs
  einschränkt (Permissions auf Typ-Ebene gibt es heute noch nicht
  durchgängig).
- Beziehungs-Editor und freies ERD-Diagramm: bleibt die „freie"
  ERD-Bearbeitung (`ERD`-Artefakt) der Ort für Multi-Typ-Layouts, oder
  konvergieren beide auf eine Oberfläche?

## Referenz

- [ADR-0001 — Git-first](ADR-0001-git-first-source-of-truth.md)
- [ADR-0004 — KI-Assistenz nur für Drafts](ADR-0004-ai-assistance-only-no-autonomous-decisions.md)
- [ADR-0022 (DRAFT) — Cosmos Explorer Server- und Repository-Mounts](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)
- [ADR-0024 (DRAFT) — Deklarative View-/Form-Artefakte](ADR-0024-DRAFT-declarative-view-artifacts.md)
- [ADR-0028 (DRAFT) — ID-basierte Referenzen](ADR-0028-DRAFT-id-based-references.md)
- [ADR-0031 (DRAFT) — Referenz-basierte Form-Komponenten](ADR-0031-DRAFT-reference-form-components.md)
- [ADR-0032 (DRAFT) — Form-Komposition über eingebettete Subforms](ADR-0032-DRAFT-form-composition-subforms.md)
