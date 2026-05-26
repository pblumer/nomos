# ADR-0032 (DRAFT): Form-Komposition über eingebettete Subforms

## Status

Draft

## Datum

2026-05-26

## Kontext

Mit ADR-0024 sind Forms deklarative Artefakte (`.frm`), und jeder Typ bringt
vier Standard-Varianten mit: `_new`, `_edit`, `_list`, `_short`. ADR-0031
ergänzt referenz-basierte Auswahlfelder (`selectRef`), mit denen ein Form auf
eine **ID** eines anderen Typs verweist.

Was fehlt, ist der nächste Schritt der Komposition: das **referenzierte Objekt
inline darstellen oder bearbeiten**, ohne den Inhalt im aktuellen Form neu
modellieren zu müssen. Konkretes Beispiel:

- Ein Bestellformular hat ein `selectRef` für die Kundin (Typ `Person`).
- Direkt darunter soll die ausgewählte Person mit ihrer **eigenen
  `Person_short.frm`** angezeigt werden — dieselbe Form, die im Person-
  Kontext überall sonst auch verwendet wird.
- Wechselt die Auswahl, aktualisiert sich der eingebettete Bereich.

Dieselbe Form (`Person_short.frm`) ist damit an mehreren Stellen wiederverwendbar
— sie ist eine **Komponente**, nicht nur eine Top-Level-Ansicht. Das ist
analog zu BPMN-Call-Activities, die einen anderen Prozess als Subprozess
einbetten, oder zu Komponenten-Komposition in UI-Frameworks: das eingebettete
Artefakt bleibt eigenständig editierbar und versioniert.

## Entscheidung

### 1. Neuer Component-Typ: `formRef`

Ein Form kann ein anderes Form über einen neuen Komponenten-Typ einbetten:

```yaml
- type: formRef
  key: customer
  formRef: .nomos/views/Person_short.frm   # explizite Form-Datei
  source:
    action: get
    id: ${customerId}                       # FEEL-Ausdruck, ID des Zielobjekts
```

`formRef` ist — wie `selectRef` (ADR-0031) — eine Nomos-Erweiterung der
form-js-Komponente, registriert als Custom Component. Das engine-native Schema
bleibt gültig.

### 2. Form-Auswahl: explizit oder per Typ + Variante

Statt einer expliziten Datei kann auch über Typ + Variante referenziert werden:

```yaml
- type: formRef
  key: customer
  typeRef: .nomos/types/Person
  variant: short        # new | edit | list | short
  source:
    action: get
    id: ${customerId}
```

Der Renderer löst `(typeRef, variant)` zu genau einer `.frm`-Datei auf. Wenn
für eine Variante **mehrere** Forms existieren (z. B. kundenspezifische
Varianten), ist `formRef` Pflicht und der Form-Editor bietet eine Auswahlliste
beim Einfügen des Components an („Welche Person-Kurzansicht soll verwendet
werden?"). Damit bleibt die Modellrelation bevorzugter Weg, ohne flexible Sonderfälle zu verbieten.

### 3. Datenquellen: `get` (Default), `bind`, `inline`

Drei Modi, wie die Daten in das Subform kommen:

| Modus | Verhalten |
| --- | --- |
| `get` | REST-Aufruf an die Instance-API mit `id`; Ergebnis wird Datenkontext des Subforms. Read-only-Modus per Default (passend zu `_short`). |
| `bind` | Datenkontext ist ein **Feld** im Eltern-Form (verschachteltes Objekt). Edits propagieren in den Eltern-Submit. Passend zu `_edit`. |
| `inline` | Wert wird als Literal im Eltern-Form mitgegeben (z. B. seed-Daten). |

Der vom User beschriebene Hauptfall ist `get`: ID kommt aus einem
selectRef, Subform fetcht und rendert.

### 4. Rendering an Komponentenposition

Das Subform rendert **genau dort**, wo der `formRef`-Component im Eltern-Form
sitzt — eingerahmt durch ein optionales Label/Frame, ansonsten visuell wie ein
zusammenhängender Block. Layout-Kontext (Spaltenbreite, Theme) erbt vom
Eltern-Form. Eigene Header/Submit-Buttons des Subforms werden im eingebetteten
Kontext **unterdrückt**, sofern nicht `showSubmit: true` gesetzt ist.

### 5. Reaktivität: `dependsOn`

Wie bei `selectRef` (ADR-0031) kann ein `formRef` auf Änderungen anderer Felder
reagieren:

```yaml
- type: formRef
  key: customerView
  formRef: .nomos/views/Person_short.frm
  source:
    action: get
    id: ${customerId}
  dependsOn: [customerId]
```

Wechselt `customerId`, lädt der Renderer das Subform neu. Solange die ID leer
ist, zeigt der Slot einen Platzhalter („Keine Person ausgewählt").

### 6. Zyklenschutz und Tiefenlimit

Subforms können theoretisch andere Subforms enthalten. Zwei harte Regeln:

- **Zyklenerkennung** beim Laden: derselbe `.frm`-Pfad darf in einer
  Render-Kette nicht doppelt vorkommen (Form A → Form B → Form A bricht ab und
  zeigt eine sichtbare Fehlermeldung im Slot).
- **Max-Tiefe** (Default 4) ist konfigurierbar; tiefere Verschachtelung wird
  abgebrochen statt schweigend gerendert.

Beide Regeln werden bei `view`-Lint geprüft und blockieren das Speichern eines
Forms, das einen statischen Zyklus enthält.

### 7. Submit-Semantik

- `source.action: get` → Subform ist **read-only**, keine Submit-Beteiligung.
- `source.action: inline` → Subform-Werte werden Teil des Eltern-Submit unter
  `key`.
- `source.action: bind` → Subform-Werte werden Teil des Eltern-Submit; die
  **Validität** des Subforms ist Voraussetzung für die Eltern-Validität.

Ein editierbares Subform speichert **nicht eigenständig**: es trägt zum
Eltern-Submit bei. Wer „eigene Submit-Action" braucht, verwendet ein
freistehendes Form, nicht ein eingebettetes.

### 8. Föderierung

`formRef`-Pfade folgen denselben Referenzregeln wie alle Nomos-Artefakte
([ADR-0028](ADR-0028-DRAFT-id-based-references.md)). Ein Subform kann auf einem
**entfernten Server** liegen; sowohl die Spec als auch der `get`-Aufruf laufen
über den Mount-Proxy (ADR-0022/0023). Damit ist „Person aus fremder Domäne im
eigenen Bestellformular anzeigen" derselbe Mechanismus wie eine lokale
Einbettung.

## Konsequenzen

### Positiv

- **Komposition statt Duplikation**: dieselbe `_short`-Form überall, an einer
  Stelle gepflegt.
- Natürliches Master-Detail-Pattern, ohne Sonderkomponente — `selectRef`
  (ADR-0031) liefert die ID, `formRef` rendert das Detail.
- Federierung kostenlos: Subforms von entfernten Servern funktionieren über
  vorhandene Mount-Mechanik.
- Git-first reviewbar: Welches Form welches andere einbettet, steht im
  Artefakt.

### Negativ / Risiken

- Nested REST-Calls: jeder `formRef` mit `get` ist ein Request. Bei Listen
  („Bestellung zeigt 50 Positionen mit je eingebettetem Produkt-Short") wird
  das spürbar — Bulk-/Batch-Endpoint bzw. Caching nötig.
- Zyklenschutz muss konsequent durchgesetzt werden, sonst hängt das UI.
- Validierungs-/Submit-Semantik für `bind` ist nicht trivial (Teilvalidität,
  Fehlerpropagation in den Eltern-Renderer).
- Form-Editor wird komplexer: Auswahl des Sub-Forms beim Einfügen, Preview
  ohne Backend nur eingeschränkt möglich.
- Permissions: ein Subform kann Daten anzeigen, die der aktuelle Nutzer im
  Standalone-Kontext nicht sehen würde — Autorisierung **muss** beim
  `get`-Aufruf greifen, nicht erst im Renderer.

## Offene Punkte

- Batch-/Bulk-Fetch, wenn `formRef` in einer Liste oder dynamiclist sitzt
  (N+1-Problem).
- Verhalten bei dangling references (Zielobjekt gelöscht): Platzhalter,
  Fehlerblock, oder konfigurierbar?
- Caching-Strategie über mehrere Subform-Instanzen mit gleichem `(formRef, id)`.
- Wie äußert sich ein editierbares Subform im Eltern-Submit konkret —
  geschachteltes Objekt im JSON, separater Endpoint, transactional?
- Verhältnis zu BPMN-Call-Activity: gemeinsame Spec-Sprache oder bewusst
  unterschiedlich?
- UI-Konvention für „Frame an/aus", Titel, Collapsibility eines eingebetteten
  Blocks — Renderer-Detail, aber konsistenzrelevant.
- Wenn das Subform selbst `dependsOn`-Felder hat, deren Werte aus dem
  Eltern-Form kommen sollen: einheitliche Variable-Scoping-Regel (FEEL-Kontext
  des Eltern-Forms erbt nach unten? Default isoliert?).

## Referenz

- [ADR-0019 — DMN 1.5 / FEEL Evaluation Engine](ADR-0019-dmn-feel-engine-scope.md)
- [ADR-0022 (DRAFT) — Cosmos Explorer Server- und Repository-Mounts](ADR-0022-DRAFT-cosmos-explorer-server-and-repository-mounts.md)
- [ADR-0023 (DRAFT) — Inter-Server-Authentifizierung](ADR-0023-DRAFT-inter-server-authentication.md)
- [ADR-0024 (DRAFT) — Deklarative View-/Form-Artefakte](ADR-0024-DRAFT-declarative-view-artifacts.md)
- [ADR-0028 (DRAFT) — ID-basierte Referenzen](ADR-0028-DRAFT-id-based-references.md)
- [ADR-0031 (DRAFT) — Referenz-basierte Form-Komponenten](ADR-0031-DRAFT-reference-form-components.md)
