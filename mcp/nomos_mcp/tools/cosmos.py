"""MCP Tools für Nomos Cosmos-Verwaltung (Go-Subsystem).

Cosmos ist der oberste Container eines Nomos-Repositories.
Diese Tools initialisieren und lesen Cosmos-Artefakte über den
nomos CLI (cosmos init) bzw. den nomos Go-Server (cosmos info).
"""

import json
import os
import subprocess

import httpx

from ..auth import AgentScope, require_scope

COSMOS_TOOLS = [
    {
        "name": "cosmos_init",
        "description": (
            "Initialisiert ein neues Nomos-Cosmos-Repository an einem lokalen Pfad. "
            "Erstellt cosmos.yaml, domains/ und catalog/ Verzeichnisse. "
            "Erfordert DRAFT-Scope. "
            "[EN] Initialises a new Nomos cosmos at a local path. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "path": {
                    "type": "string",
                    "description": "Absoluter Pfad wo der Cosmos angelegt werden soll, z.B. '/tmp/my-cosmos'",
                },
                "id": {
                    "type": "string",
                    "description": "Cosmos-ID, z.B. 'my-cosmos'",
                },
                "name": {
                    "type": "string",
                    "description": "Anzeigename des Cosmos",
                },
                "owner": {
                    "type": "string",
                    "description": "Verantwortlicher/Team",
                },
            },
            "required": ["path", "id", "name"],
        },
    },
    {
        "name": "cosmos_info",
        "description": (
            "Gibt Metadaten des aktuell konfigurierten Cosmos zurück: "
            "ID, Name, Anzahl Domains und Services. "
            "[EN] Returns metadata of the currently configured cosmos."
        ),
        "inputSchema": {"type": "object", "properties": {}, "required": []},
    },
]


def _go_api_base() -> str:
    return os.getenv("NOMOS_GO_API_BASE", "http://localhost:9090")


def _nomos_binary() -> str:
    return os.getenv("NOMOS_BINARY", "nomos")


def _run_nomos(*args: str, cwd: str | None = None) -> tuple[int, str, str]:
    result = subprocess.run(
        [_nomos_binary(), *args],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    return result.returncode, result.stdout.strip(), result.stderr.strip()


def _format_response(resp: httpx.Response) -> str:
    try:
        return json.dumps(resp.json(), ensure_ascii=False, indent=2)
    except Exception:
        return resp.text


@require_scope(AgentScope.DRAFT)
async def cosmos_init_handler(path: str, id: str, name: str, owner: str = "") -> str:
    code, out, err = _run_nomos("cosmos", "init", path)
    if code != 0:
        return f"Fehler beim Initialisieren des Cosmos: {err or out}"

    result = {
        "path": path,
        "id": id,
        "name": name,
        "owner": owner,
        "action": "initialized",
        "note": (
            "Cosmos wurde angelegt. Starte 'nomos serve --path <path>' "
            "und setze NOMOS_GO_API_BASE auf die Serveradresse."
        ),
    }
    return json.dumps(result, ensure_ascii=False, indent=2)


@require_scope(AgentScope.READ)
async def cosmos_info_handler() -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_go_api_base()}/api/v1/cosmos")
        if not resp.is_success:
            return f"HTTP {resp.status_code}: {resp.text}"
        return _format_response(resp)
