# 009 - UI Konzept für Produktmanagement

## Zweck
Beschreibung einer businessfreundlichen UI, die Git-Komplexitaet verbirgt und Governance sichtbar macht.

## Leitprinzipien
- Klarheit vor Funktionsvielfalt.
- Business-Lesbarkeit.
- Sicheres Editieren.
- Sichtbares Validierungsfeedback.
- Sichtbarer Änderungs- und Reviewstatus.

## Git Begriffe in Businesssprache
- Branch = Änderungsvorschlag
- Commit = Speicherpunkt
- Pull Request = Review-Antrag
- Merge = Freigabe übernehmen
- Main branch = freigegebener Katalog
- Diff = Änderungsvergleich
- Tag = veroeffentlichte Version

## Schlüsselansichten
- Produktübersicht
- Produkteditor
- Anforderungseditor
- Regeleditor
- Validierung testen
- Findings anzeigen
- Änderungsvorschlag verwalten
- Review-Übersicht
- KI-Assistenz

## Nutzerinteraktion
Businessnutzer arbeiten mit Produkten, Varianten, Anforderungen, Regeln und Szenarien in fachlicher Sprache. Die UI fuehrt intern Branch, Commit und Pull Request durch, ohne Git Know-how vorauszusetzen.

## MVP Prioritaeten
1. Produktliste und Produktdetail.
2. Strukturierte Bearbeitung von Anforderungen und Regeln.
3. Testvalidierung mit klaren Findings.
4. Transparenter Änderungsvergleich.
5. Sichtbarer Review- und Freigabestatus.
6. Optionales KI-Panel für Draft-Unterstuetzung.

## Implemented Go Web UI extension

The implemented Nomos web UI is a Go-served, embedded-template interface for a local Cosmos repository. It intentionally avoids a primary database and avoids a separate Node/Vite frontend for this serving path. The UI uses `internal/app` as the application layer so CLI, API, and web pages share DTOs and filesystem behavior.

### Navigation and pages

The global navigation contains Dashboard, Cosmos, Domains, Services, Blueprints, Instances, Validation, Graph, Namespace Tree, Verification, and API. Each page shows the active section, the selected Cosmos path, and a Cosmos status badge.

### Capability mapping

| CLI command | API endpoint | Web page |
|---|---|---|
| `nomos cosmos info` | `/api/v1/cosmos` | `/cosmos` |
| `nomos cosmos doctor` | TBD/app doctor endpoint | `/cosmos` |
| `nomos domain list` | `/api/v1/domains` | `/domains` |
| `nomos domain get` | `/api/v1/domains/{domain}` | `/domains/{domain}` |
| `nomos domain add` | `POST /api/v1/domains` | `/domains` form |
| `nomos service list` | `/api/v1/domains/{domain}/services` | `/services` |
| `nomos service get` | `/api/v1/domains/{domain}/services/{service}` | `/services` detail |
| `nomos service add` | `POST /api/v1/domains/{domain}/services` | `/services` form |
| `nomos validate` | `/api/v1/validate` | `/validate` |
| `nomos graph` | `/api/v1/graph` | `/graph` |
| `nomos namespace tree` | `/api/v1/namespaces` | `/namespaces` |
| `nomos verify domain` | TBD | `/verify` |
| `nomos blueprint list` | `/api/v1/blueprints` | `/blueprints` |
| `nomos blueprint show` | `/api/v1/blueprints/{id}` | `/blueprints/{id}` |
| `nomos instance list` | `/api/v1/instances` | `/instances` |
| `nomos instance show` | `/api/v1/instances/{id}` | `/instances/{id}` |
| `nomos instance compliance` | `/api/v1/instances/{id}/compliance` | `/instances/{id}` |

### Current limitations

- Doctor checks are application-layer page data and do not yet have a dedicated `/api/v1/doctor` endpoint.
- Verification evidence is visible on `/verify`; a stable read API for all evidence can be added later.
- The UI remains deliberately lightweight and uses only small progressive-enhancement JavaScript for copy and filter interactions.
