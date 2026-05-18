# ADR-0013 - User Contact Interfaces (UCI) als erstklassige Nomos-Artefakte

## Status

Proposed

## Datum

2026-05-18

## Kontext

Nomos verwaltet heute strukturierte Artefakte für Produkte, Anforderungen, Business Rules,
Validierungsszenarien und Findings. Prozessabläufe werden konzeptionell mit BPMN modelliert.

In BPMN existiert der Artefakttyp **User Task** – eine Aufgabe, die explizit eine menschliche
Interaktion erfordert. User Tasks sind nicht vollständig automatisierbar; sie benötigen ein
Interface, über das ein Benutzer Eingaben macht, Informationen bestätigt oder Entscheidungen
trifft.

Diese Interfaces werden in Nomos als **User Contact Interfaces (UCI)** bezeichnet. Ein UCI ist
das konkrete Gegenstück zu einer BPMN User Task: Es definiert, *wie* eine menschliche
Interaktion technisch realisiert wird – unabhängig vom verwendeten UI-Framework.

UCIs können in sehr unterschiedlichen Technologien implementiert sein, z. B.:

- **HTML/JS/CSS** – einfachste, plattformunabhängige Form (Referenzimplementierung)
- **React / Vue / Angular** – komponentenbasierte SPA-Frameworks
- **Go-Templates** – server-rendered (wie in Nomos selbst bereits genutzt)
- **Native Mobile** – iOS / Android
- **Formular-Engines** – z. B. JSON Forms, XForms
- **Low-Code-Plattformen** – z. B. Microsoft Power Apps, Retool
- **CLI-Prompts** – für terminalbasierte Interaktionen

Bisher fehlt in Nomos ein Mechanismus, um UCIs zu versionieren, mit BPMN-Prozessen zu
verknüpfen und denselben Git-basierten Review- und Governance-Prozess zu durchlaufen wie
alle anderen Artefakte.

## Entscheidung

UCIs werden als eigener Artefakttyp eingeführt und im `.nomos/`-Workspace verwaltet.
Sie sind **framework-agnostisch** definiert und werden mit BPMN User Tasks verknüpft.

### Artefaktstruktur im Cosmos-Workspace

```
.nomos/
  uci/
    <uci-id>/
      uci.yaml          # Metadaten, Verknüpfung, Deklaration
      implementation/   # Framework-spezifische Dateien (z.B. index.html, app.js, style.css)
      schema.yaml       # Optionales Ein-/Ausgabeschema (JSON Schema-kompatibel)
      README.md         # Fachliche Beschreibung des Interfaces
```

### `uci.yaml` – Metadatenformat

```yaml
id: uci-konto-erstellung-benutzer
version: "1.0.0"
title: Benutzerkonto erstellen – Eingabemaske
description: |
  Erfassung der Basisdaten für ein neues Benutzerkonto durch den Antragsteller.

bpmn_process_ref: PROC-ACC-001
bpmn_task_ref: UT-001

framework:
  type: html-js-css    # html-js-css | react | vue | angular | go-template | cli | native-mobile | other
  runtime: browser     # browser | server | mobile | terminal | other
  entry_point: implementation/index.html

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

Nomos **speichert, versioniert und validiert** UCIs. Es **rendert oder führt** sie nicht
selbst aus. Die Rendering-Verantwortung liegt bei der jeweiligen Prozess-Runtime
(BPMN-Engine, Applikations-Shell, Portal).

Nomos stellt sicher, dass:

1. Die `uci.yaml`-Metadaten wohlgeformt und vollständig sind (strukturelle Validierung).
2. Die Verknüpfung zu BPMN-Artefakten konsistent ist (referentielle Integrität).
3. Das deklarierte `entry_point` im `implementation/`-Verzeichnis vorhanden ist.
4. Alle Änderungen den Git-Workflow durchlaufen (Branch → PR → Review → Merge),
   konsistent mit ADR-0001.

### CLI-Erweiterung

```bash
nomos uci init <id> --task <task-ref> --process <proc-ref> --framework html-js-css
nomos uci list
nomos uci get <id>
nomos validate --scope uci
```

### REST-API-Erweiterung

| Methode | Endpunkt | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/uci` | Alle UCIs auflisten |
| `GET` | `/api/v1/uci/{id}` | UCI-Metadaten abrufen |
| `GET` | `/api/v1/uci/{id}/implementation/{file}` | Implementierungsdatei abrufen |
| `GET` | `/api/v1/processes/{id}/user-tasks/{taskId}/uci` | UCI zu einer User Task |

### Web-UI

Die Go-Template-basierte Web-UI erhält einen UCI-Explorer analog zum bestehenden
Domain/Service-Explorer: Auflistung, Detailansicht mit Metadaten/Schema/Verknüpfungen
sowie einen Vorschau-Link auf den `entry_point` (wo technisch möglich).

## Verworfene Alternativen

**UCI nur als Attribut in BPMN User Tasks**: UCIs wären keine eigenständigen Artefakte,
sondern nur ein `implementation_url`-Feld. Abgelehnt, weil UCIs eigene Lebenszyklen,
eigene Reviews und Wiederverwendbarkeit über mehrere Prozesse hinweg erfordern.

**Ein einziges UI-Framework vorschreiben (z. B. nur HTML/JS/CSS)**: Würde den
Einsatz in heterogenen Organisationsumgebungen einschränken. Abgelehnt — Nomos soll
framework-agnostisch bleiben.

**UCIs in einem separaten Repository**: Referentielle Integrität wäre schwerer
sicherstellbar und der Git-first-Ansatz (ADR-0001) würde aufgeweicht. Externe
Referenzen können als optionale Erweiterung in einer späteren Phase ergänzt werden.

## Konsequenzen

### Positiv

- UCIs sind vollständig versioniert und auditierbar (Git-History).
- Verknüpfung zu BPMN User Tasks ist explizit und maschinell validierbar.
- Framework-Agnostizität erlaubt den Einsatz in unterschiedlichen Technologiekontexten.
- Governance greift automatisch durch den bestehenden Branch/PR-Workflow (ADR-0001).
- Keine neue technische Laufzeitabhängigkeit — Nomos rendert UCIs nicht selbst.

### Negativ / Risiken

- Nomos kann UCIs nicht selbst ausführen oder in einer BPMN-Runtime einbetten
  (bewusst Out of Scope, konsistent mit MVP-Prinzipien).
- Validierung bleibt auf Struktur und referentielle Integrität beschränkt;
  semantische Korrektheit des UI-Codes wird nicht geprüft.
- Teams müssen das `uci.yaml`-Format kennen und aktiv pflegen.

### Offene Punkte

- Synchronisation des `schema.yaml` mit der BPMN-Datenmodellierung → separates ADR.
- Wiederverwendung von UCIs über Cosmos-Grenzen hinweg (UCI-Bibliothek) → Backlog.
- Iframe-Sandbox-Preview für HTML-UCIs in der Web-UI → Backlog.

## Referenz

- ADR-0001: Git-first als Quelle der Wahrheit
- ADR-0004: KI-Assistenz nur für Drafts, keine autonomen Entscheidungen
- ADR-0008: MCP Server als KI-Agenten-Schnittstelle (UCIs als lesbare Artefakte)
