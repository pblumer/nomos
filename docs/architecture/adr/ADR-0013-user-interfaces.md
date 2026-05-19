# ADR-0013 - User Interfaces (UI) als erstklassige Nomos-Artefakte

## Status

Proposed (revidiert 2026-05-19: Begriff "User Contact Interface (UCI)" durch
"User Interface (UI)" ersetzt; lightweight UI-Deklaration unter Service
ergänzt.)

## Datum

2026-05-18 (Erstfassung) · 2026-05-19 (Revision: Terminologie + Service-UI)

## Kontext

Nomos verwaltet heute strukturierte Artefakte für Produkte, Anforderungen, Business Rules,
Validierungsszenarien und Findings. Prozessabläufe werden konzeptionell mit BPMN modelliert.

In BPMN existiert der Artefakttyp **User Task** – eine Aufgabe, die explizit eine menschliche
Interaktion erfordert. User Tasks sind nicht vollständig automatisierbar; sie benötigen ein
Interface, über das ein Benutzer Eingaben macht, Informationen bestätigt oder Entscheidungen
trifft.

Diese Interfaces werden in Nomos als **User Interfaces (UI)** bezeichnet. Ein UI ist das
konkrete Gegenstück zu einer BPMN User Task: Es definiert, *wie* eine menschliche
Interaktion technisch realisiert wird – unabhängig vom verwendeten UI-Framework.

> **Terminologie-Historie:** Frühere Entwürfe nutzten den Begriff *User Contact Interface
> (UCI)*. Der Begriff wurde zugunsten von *User Interface (UI)* aufgegeben, weil "UI" der
> branchenübliche Begriff ist und das frühere "Contact" keinen zusätzlichen Erkenntnisgewinn
> brachte. Wo "UCI" historisch in Code, YAML oder anderen ADRs noch auftauchte, ist der
> Begriff sinngemäß durch "UI" zu ersetzen.

UIs können in sehr unterschiedlichen Technologien implementiert sein, z. B.:

- **bpmn-io-form** – schema-getriebene Formulare via `form-js` (Referenzimplementierung
  für formularbasierte User Tasks, siehe Abschnitt *Referenz-Implementierungstyp*)
- **HTML/JS/CSS** – einfachste, plattformunabhängige Form
- **React / Vue / Angular** – komponentenbasierte SPA-Frameworks
- **Go-Templates** – server-rendered (wie in Nomos selbst bereits genutzt)
- **Native Mobile** – iOS / Android
- **Formular-Engines** – z. B. JSON Forms, XForms
- **Low-Code-Plattformen** – z. B. Microsoft Power Apps, Retool
- **CLI-Prompts** – für terminalbasierte Interaktionen

Bisher fehlte in Nomos ein Mechanismus, um UIs zu versionieren, mit BPMN-Prozessen zu
verknüpfen und denselben Git-basierten Review- und Governance-Prozess zu durchlaufen wie
alle anderen Artefakte.

## Entscheidung

UIs werden auf zwei komplementären Ebenen geführt:

1. **Lightweight UI-Deklaration auf Service-Ebene** – jeder Service kann in seiner
   `service.yaml` unter `user_interfaces:` eine Liste der von ihm angebotenen UIs
   deklarieren (Name, ID, Channel, URL, Stabilität, Summary). Diese Form dient der
   Architektur-Sicht: Welche Services bieten welche UIs an, ähnlich zu Capabilities und
   Data Objects. Sie ist absichtlich schlank gehalten und erfordert keine
   Implementierungs-Dateien.

2. **Vollständiges UI-Artefakt** – wenn ein UI als eigenständiges, versionierbares
   Artefakt mit Implementierungs-Dateien und Schemas geführt werden soll (z. B. um es
   gemeinsam mit einer BPMN User Task zu reviewen), wird es unter `.nomos/ui/<ui-id>/`
   abgelegt. Diese Form ist **framework-agnostisch** und wird mit BPMN User Tasks
   verknüpft.

Beide Formen koexistieren: Die Service-Deklaration kann via `entry_point` oder `url` auf
ein vollständiges UI-Artefakt verweisen; umgekehrt kann ein UI-Artefakt über seine
Service-Referenz auf den anbietenden Service zeigen.

### Service-Ebene: `user_interfaces:` in `service.yaml`

```yaml
# .nomos/domains/<domain>/services/<service>/service.yaml
user_interfaces:
  - id: ui-account-portal
    name: Self-Service Portal
    summary: Web-Portal für Endkunden zur Konto-Self-Service-Verwaltung.
    channel: web                # web | mobile | cli | desktop | voice | api
    url: https://portal.example
    stability: stable
  - id: ui-onboarding-form
    name: Konto-Erstellung – Eingabemaske
    channel: web
    stability: draft
```

Diese Einträge werden in der Web-UI unter jedem Service als eigene Gruppe "User
Interfaces" neben Capabilities und Data Objects angezeigt und sind über
`POST /api/v1/services/{domain}/{service}/user-interfaces` editierbar.

### Vollständiges UI-Artefakt im Cosmos-Workspace

```
.nomos/
  ui/
    <ui-id>/
      ui.yaml           # Metadaten, Verknüpfung, Deklaration
      implementation/   # Framework-spezifische Dateien (z.B. index.html, app.js, style.css)
      schema.yaml       # Optionales Ein-/Ausgabeschema (JSON Schema-kompatibel)
      README.md         # Fachliche Beschreibung des Interfaces
```

### `ui.yaml` – Metadatenformat

```yaml
id: ui-konto-erstellung-benutzer
version: "1.0.0"
title: Benutzerkonto erstellen – Eingabemaske
description: |
  Erfassung der Basisdaten für ein neues Benutzerkonto durch den Antragsteller.

bpmn_process_ref: PROC-ACC-001
bpmn_task_ref: UT-001

# Optionaler Rückverweis auf den anbietenden Service
service_ref: identity.blumer.cloud/user-account

framework:
  type: bpmn-io-form   # bpmn-io-form | html-js-css | react | vue | angular | go-template | cli | native-mobile | other
  runtime: browser     # browser | server | mobile | terminal | other
  entry_point: implementation/form.json

input_schema: schema.yaml#input
output_schema: schema.yaml#output

status: draft          # draft | review | approved | deprecated
owner_domain: identity.blumer.cloud
review_required: true

tags:
  - onboarding
  - konto
created: 2026-05-18
```

### Verantwortung von Nomos

Nomos **speichert, versioniert und validiert** UIs. Es **rendert oder führt** sie nicht
selbst aus. Die Rendering-Verantwortung liegt bei der jeweiligen Prozess-Runtime
(BPMN-Engine, Applikations-Shell, Portal).

Nomos stellt sicher, dass:

1. Die `ui.yaml`-Metadaten wohlgeformt und vollständig sind (strukturelle Validierung).
2. Die Verknüpfung zu BPMN-Artefakten konsistent ist (referentielle Integrität).
3. Das deklarierte `entry_point` im `implementation/`-Verzeichnis vorhanden ist.
4. Service-Referenzen (`service_ref` bzw. die `user_interfaces:`-Liste der Services)
   konsistent zueinander sind.
5. Alle Änderungen den Git-Workflow durchlaufen (Branch → PR → Review → Merge),
   konsistent mit ADR-0001.

### Referenz-Implementierungstyp: `bpmn-io-form`

Für formularbasierte User Tasks wird `bpmn-io-form` ([`form-js`](https://github.com/bpmn-io/form-js))
als bevorzugter Implementierungstyp empfohlen. Begründung:

- Die Form-Definition ist reines JSON → versionierbar, diff-freundlich, reviewfähig
  (konsistent zu ADR-0001).
- Native semantische Kopplung an BPMN User Tasks (Camunda Forms).
- Vendor-Konsistenz mit dem bereits eingesetzten `bpmn-js` Viewer/Modeler
  (siehe `docs/product-process-modeling.md`).
- Schema-Determinismus: Die Form-Definition kann gegen `schema.yaml` (JSON Schema)
  abgeglichen oder daraus abgeleitet werden.

Konventionen bei `framework.type: bpmn-io-form`:

- `entry_point` zeigt auf eine `form.json`-Datei (form-js Schema, Version ≥ 1).
- `runtime` ist typischerweise `browser`; server-seitiges Rendering bleibt möglich,
  solange die Form-Definition unverändert bleibt.
- Browser-Assets werden – wie `bpmn-js` – unter `internal/server/web/static/vendor/`
  ausgeliefert (Air-Gap-fähig). Kein Online-CDN.

`bpmn-io-form` ist *eine* Referenzimplementierung, nicht der einzig zulässige Typ.
Komplexere Interaktionen (Wizards, Dashboards, CLI-Prompts, Native Mobile) bleiben
über die anderen `framework.type`-Werte abbildbar; ADR-0013 bleibt damit
framework-agnostisch.

### CLI-Erweiterung

```bash
nomos ui init <id> --task <task-ref> --process <proc-ref> --framework bpmn-io-form
nomos ui list
nomos ui get <id>
nomos validate --scope ui
```

### REST-API-Erweiterung

| Methode | Endpunkt | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/ui` | Alle vollständigen UI-Artefakte auflisten |
| `GET` | `/api/v1/ui/{id}` | UI-Metadaten abrufen |
| `GET` | `/api/v1/ui/{id}/implementation/{file}` | Implementierungsdatei abrufen |
| `GET` | `/api/v1/processes/{id}/user-tasks/{taskId}/ui` | UI zu einer User Task |
| `POST` | `/api/v1/services/{domain}/{service}/user-interfaces` | Service-UI-Deklaration anlegen |

### Web-UI

Die Go-Template-basierte Web-UI zeigt UIs an zwei Stellen:

- **Im Service-Detail** wird die `user_interfaces:`-Liste eines Services als eigene
  Gruppe neben Capabilities und Data Objects angezeigt. Die Gruppe ist filterbar; ein
  Add-Button öffnet ein Modal zur Erfassung.
- **Als UI-Explorer** (analog Domain/Service-Explorer): Auflistung aller vollständigen
  UI-Artefakte mit Detailansicht (Metadaten/Schema/Verknüpfungen) und – wo technisch
  möglich – einem Vorschau-Link auf den `entry_point`.

## Verworfene Alternativen

**UI nur als Attribut in BPMN User Tasks**: UIs wären keine eigenständigen Artefakte,
sondern nur ein `implementation_url`-Feld. Abgelehnt, weil vollständige UIs eigene
Lebenszyklen, eigene Reviews und Wiederverwendbarkeit über mehrere Prozesse hinweg
erfordern. Die schlanke Variante existiert dennoch — als `user_interfaces:`-Liste am
Service.

**Nur Service-Deklaration ohne vollständiges Artefakt**: Die Service-Liste allein reicht
für Architektur-Sichten, nicht aber für versionierte Forms, Schemas oder Implementations-
Dateien. Beide Ebenen koexistieren bewusst.

**Ein einziges UI-Framework vorschreiben (z. B. nur HTML/JS/CSS oder nur
`bpmn-io-form`)**: Würde den Einsatz in heterogenen Organisationsumgebungen
einschränken (z. B. CLI-Prompts, Native Mobile, Low-Code-Plattformen). Abgelehnt –
Nomos soll framework-agnostisch bleiben. `bpmn-io-form` wird stattdessen als
*empfohlene Referenzimplementierung* für formularbasierte User Tasks geführt.

**UIs in einem separaten Repository**: Referentielle Integrität wäre schwerer
sicherstellbar und der Git-first-Ansatz (ADR-0001) würde aufgeweicht. Externe
Referenzen können als optionale Erweiterung in einer späteren Phase ergänzt werden.

**Begriff "User Contact Interface (UCI)" beibehalten**: Der Begriff war wenig
gebräuchlich und führte zu Verwirrung; "User Interface (UI)" ist der branchenübliche
Begriff und passt zum gleichnamigen Element auf Service-Ebene.

## Konsequenzen

### Positiv

- UIs sind als Architektur-Element (Service-Liste) und als versioniertes Artefakt
  (UI-Bundle) erfassbar — eine Stufe Detail je nach Bedarf.
- Verknüpfung zu BPMN User Tasks ist explizit und maschinell validierbar.
- Framework-Agnostizität erlaubt den Einsatz in unterschiedlichen Technologiekontexten.
- Governance greift automatisch durch den bestehenden Branch/PR-Workflow (ADR-0001).
- Keine neue technische Laufzeitabhängigkeit — Nomos rendert UIs nicht selbst.
- Konsistente Terminologie ("UI") quer durch Code, YAML, Docs und Web-UI.

### Negativ / Risiken

- Nomos kann UIs nicht selbst ausführen oder in einer BPMN-Runtime einbetten
  (bewusst Out of Scope, konsistent mit MVP-Prinzipien).
- Validierung bleibt auf Struktur und referentielle Integrität beschränkt;
  semantische Korrektheit des UI-Codes wird nicht geprüft.
- Teams müssen sowohl das `ui.yaml`-Format als auch die `user_interfaces:`-Liste am
  Service kennen und konsistent pflegen.
- Bestehender Service `core.nomos/uci-manager` trägt den alten Begriff im Namen.
  Eine Umbenennung in `ui-manager` ist als Follow-up vorgesehen; bis dahin gilt der
  Name als historisches Artefakt.

### Offene Punkte

- Synchronisation des `schema.yaml` mit der BPMN-Datenmodellierung → separates ADR.
- Wiederverwendung von UIs über Cosmos-Grenzen hinweg (UI-Bibliothek) → Backlog.
- Iframe-Sandbox-Preview für HTML-UIs in der Web-UI → Backlog.
- Umbenennung `uci-manager` → `ui-manager` inkl. Anpassung der Prozess-Blueprints
  und der `bundle.yaml` → Backlog.

## Referenz

- ADR-0001: Git-first als Quelle der Wahrheit
- ADR-0004: KI-Assistenz nur für Drafts, keine autonomen Entscheidungen
- ADR-0008: MCP Server als KI-Agenten-Schnittstelle (UIs als lesbare Artefakte)
- ADR-0014: Nomos Self-Model Bundle (bündelt Self-Model-UIs)
- ADR-0016: Graph Storage and Query Model (UI als Knotentyp)
