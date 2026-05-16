# Nomos MCP Server

Ein [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) Server der KI-Agenten ermöglicht,
Nomos-Artefakte mit entsprechender Berechtigung zu lesen, erstellen, verändern und löschen.

Der Server nutzt intern die bestehende REST API des Nomos Python-Backends und
respektiert das Git-First-Prinzip und das Berechtigungskonzept von Nomos.

---

## Voraussetzungen

| Variable           | Beschreibung                              | Standard                 |
|--------------------|-------------------------------------------|--------------------------|
| `NOMOS_API_BASE`   | URL des Nomos Python-Backends             | `http://localhost:8080`  |
| `NOMOS_AGENT_SCOPE`| Scope des Agenten (`read`/`draft`/`write`)| `read`                   |
| `NOMOS_REPO_PATH`  | Pfad zum Nomos Git-Repository             | Aktuelles Verzeichnis    |
| `NOMOS_BINARY`     | Pfad zur `nomos` CLI (für Validierung)    | `nomos`                  |

Das Nomos Python-Backend muss laufen bevor der MCP Server gestartet wird:

```bash
cd backend && uvicorn app.main:app --port 8080
```

---

## Installation

```bash
cd mcp
pip install -e .

# Für Entwicklung (inkl. Test-Abhängigkeiten):
pip install -e ".[dev]"
```

---

## Start

```bash
# Direkt als Python-Modul starten (für stdio-basierte MCP-Verbindung):
python -m nomos_mcp.server

# Oder via installiertem Skript:
nomos-mcp
```

---

## Scope-Modell

Agents erhalten einen von drei Scopes via `NOMOS_AGENT_SCOPE`:

| Scope   | Beschreibung                                              |
|---------|-----------------------------------------------------------|
| `read`  | Nur lesen: list, get, validate, compliance-Status         |
| `draft` | Lesen + auf Branch schreiben (create, update, delete)     |
| `write` | Wie draft + zusätzliche Commit/PR-Operationen (geplant)  |

**Merge in `main` ist intentionally kein Scope** — das bleibt Menschen vorbehalten.

### Scope-Matrix (alle verfügbaren Tools)

| Tool                        | read | draft | write |
|-----------------------------|:----:|:-----:|:-----:|
| `list_products`             | ✓    | ✓     | ✓     |
| `get_product`               | ✓    | ✓     | ✓     |
| `create_product`            | ✗    | ✓     | ✓     |
| `update_product`            | ✗    | ✓     | ✓     |
| `delete_product`            | ✗    | ✓     | ✓     |
| `add_product_requirement`   | ✗    | ✓     | ✓     |
| `remove_product_requirement`| ✗    | ✓     | ✓     |
| `add_product_rule`          | ✗    | ✓     | ✓     |
| `remove_product_rule`       | ✗    | ✓     | ✓     |
| `get_product_variants`      | ✓    | ✓     | ✓     |
| `list_rules`                | ✓    | ✓     | ✓     |
| `get_rule`                  | ✓    | ✓     | ✓     |
| `create_rule`               | ✗    | ✓     | ✓     |
| `update_rule`               | ✗    | ✓     | ✓     |
| `list_requirements`         | ✓    | ✓     | ✓     |
| `get_requirement`           | ✓    | ✓     | ✓     |
| `update_requirement`        | ✗    | ✓     | ✓     |
| `list_blueprints`           | ✓    | ✓     | ✓     |
| `get_blueprint`             | ✓    | ✓     | ✓     |
| `list_instances`            | ✓    | ✓     | ✓     |
| `get_instance`              | ✓    | ✓     | ✓     |
| `get_instance_compliance`   | ✓    | ✓     | ✓     |
| `create_workspace`          | ✗    | ✓     | ✓     |
| `get_current_workspace`     | ✓    | ✓     | ✓     |
| `list_workspaces`           | ✓    | ✓     | ✓     |
| `validate_catalog`          | ✓    | ✓     | ✓     |
| `get_health`                | ✓    | ✓     | ✓     |

---

## Claude Desktop Konfiguration

Füge folgendes in deine `claude_desktop_config.json` ein:

```json
{
  "mcpServers": {
    "nomos": {
      "command": "python",
      "args": ["-m", "nomos_mcp.server"],
      "cwd": "/path/to/nomos/mcp",
      "env": {
        "NOMOS_API_BASE": "http://localhost:8080",
        "NOMOS_AGENT_SCOPE": "draft",
        "NOMOS_REPO_PATH": "/path/to/nomos"
      }
    }
  }
}
```

**Speicherort der Konfigurationsdatei:**
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

---

## Tests ausführen

```bash
cd mcp

# Abhängigkeiten installieren
pip install -e ".[dev]"
pip install respx  # HTTP-Mock-Bibliothek

# Tests starten
python -m pytest

# Oder aus dem Repo-Root:
cd /path/to/nomos && python -m pytest mcp/tests/
```

---

## Architektur

```
mcp/
├── nomos_mcp/
│   ├── server.py        # Haupt-MCP-Server (Tool-Registrierung und Dispatch)
│   ├── auth.py          # Scope-basiertes Berechtigungsmodell
│   └── tools/
│       ├── products.py      # CRUD für Products
│       ├── rules.py         # CRUD für Business Rules
│       ├── requirements.py  # CRUD für Requirements
│       ├── blueprints.py    # Read-only für Blueprints
│       ├── instances.py     # Read-only für Instances
│       ├── workspace.py     # Git Branch-/Workspace-Management
│       └── validate.py      # Katalog-Validierung und Health-Check
└── tests/
    ├── test_auth.py            # Scope-Logik Tests
    └── test_tools_products.py  # Product-Tools mit HTTP-Mocks
```

---

## Sicherheitshinweis: Warum Merge auf `main` nicht möglich ist

Das Nomos Governance-Modell (ADR 006, ADR 007) schreibt vor:

> _KI ist optional und strikt assistiv. Autoritative Entscheidungen bleiben bei Menschen._

Daher:

1. **Kein Merge-Tool** — Es gibt keinen `merge_to_main`-Tool im MCP Server. Merges
   erfolgen ausschliesslich durch Menschen via Pull Requests.

2. **Branch-Schutz** — Das `create_workspace`-Tool verweigert explizit Branch-Namen
   wie `main` oder `master` mit einem klaren Fehler.

3. **Scope-Kontrolle** — Alle schreibenden Operationen prüfen den Scope vor dem
   Ausführen. Ein Agent mit `read`-Scope kann keine Änderungen vornehmen.

4. **Nachvollziehbarkeit** — Alle Agent-Änderungen landen auf einem isolierten Branch
   und werden durch normalen Git-History und PR-Review sichtbar.

---

## Referenz

- Architecture Decision 007: `docs/architecture/007-ai-assistance-concept.md`
- Architecture Decision 008: `docs/architecture/008-rest-api-concept.md`
- Agent Skill: `.agent-skills/nomos-mcp/SKILL.md`
