# ADR-0008 - MCP Server als KI-Agenten-Schnittstelle für Nomos-Artefakte

## Status

Accepted

## Kontext

KI-Agenten (z.B. Claude, andere LLM-basierte Systeme) benötigen eine strukturierte
Schnittstelle um Nomos-Artefakte zu lesen und kontrolliert zu verändern. Die bestehende
REST API des Python-Backends ist für direkte HTTP-Aufrufe konzipiert, nicht für den
Einsatz als Tool-Schnittstelle in Agenten-Frameworks.

Das Model Context Protocol (MCP) von Anthropic hat sich als Standard für die Integration
von Werkzeugen in LLM-Agenten etabliert. Es ermöglicht strukturierte Tool-Definitionen
mit Eingabe-Schemata und maschinenlesbaren Beschreibungen.

Gleichzeitig muss das Governance-Modell von Nomos gewahrt bleiben: KI-Agenten dürfen
keine Änderungen direkt auf `main` vornehmen, und Freigaben bleiben Menschen vorbehalten
(vgl. ADR-0004).

## Entscheidung

Wir führen einen eigenständigen MCP Server (`mcp/`) ein, der:

1. **Intern die bestehende REST API nutzt** — kein direkter Dateisystemzugriff,
   keine neuen Backend-Abhängigkeiten.

2. **Ein scope-basiertes Berechtigungsmodell implementiert** via `NOMOS_AGENT_SCOPE`:
   - `read` — nur lesende Operationen
   - `draft` — lesen + schreiben auf isolierten Branches
   - `write` — wie draft (für zukünftige Commit/PR-Operationen vorgesehen)
   - `merge` existiert intentionally nicht — Merges in main bleiben Menschen vorbehalten.

3. **Branch-Isolierung erzwingt**: Das `create_workspace`-Tool erstellt Git-Branches
   für Agent-Läufe und verweigert `main`/`master` explizit mit einem PermissionError.

4. **Als eigenständiges Python-Package** (`mcp/`) lebt — nicht als Teil des FastAPI-Backends.

## Verworfene Alternativen

**Direkter Dateisystemzugriff im MCP Server**: Würde Backend-Logik duplizieren und
Validierungsregeln umgehen. Abgelehnt.

**Erweiterung des FastAPI-Backends um MCP-Endpunkte**: Vermischt zwei verschiedene
Protokoll-Paradigmen im gleichen Service. Abgelehnt.

**Keine Scope-Kontrolle (voller Zugriff für alle Agenten)**: Verletzt ADR-0004 und
das Governance-Prinzip. Abgelehnt.

## Konsequenzen

### Positiv
- KI-Agenten können Nomos-Artefakte strukturiert lesen und Änderungsvorschläge auf
  Branches erstellen — im Einklang mit dem Git-First-Prinzip (ADR-0001).
- Alle Scope-Checks erfolgen vor dem API-Aufruf, nicht danach.
- Tool-Beschreibungen sind maschinenlesbar und bilingual — Agenten können eigenständig
  das richtige Tool wählen.
- Keine Regression im Backend: bestehende Tests bleiben unverändert grün.

### Negativ / Risiken
- Der MCP Server ist eine weitere Laufzeitkomponente die konfiguriert und gestartet
  werden muss (NOMOS_API_BASE, NOMOS_AGENT_SCOPE, NOMOS_REPO_PATH).
- `create_rule` hat keine echte POST-Semantik in der REST API (nur PUT/update) —
  als TODO dokumentiert, bis ein `/api/v1/rules` POST-Endpunkt existiert.
- Workspace-Tools nutzen `subprocess`/`git` direkt — sobald `/api/v1/workspaces`
  im Backend implementiert ist, soll auf REST umgestellt werden.

## Referenz

- Umsetzung: `mcp/` (Package), `.agent-skills/nomos-mcp/SKILL.md`
- Basis: ADR-0001 (Git-first), ADR-0004 (KI nur assistiv)
- Konzept: `docs/architecture/007-ai-assistance-concept.md`
