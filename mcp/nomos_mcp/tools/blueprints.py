"""MCP Tools für Nomos Blueprints (read-only)."""

import json
import os

import httpx

from ..auth import AgentScope, require_scope

BLUEPRINTS_TOOLS = [
    {
        "name": "list_blueprints",
        "description": (
            "Listet alle Blueprints im Nomos-Katalog auf (Product- und Service-Blueprints). "
            "Gibt ID, Typ, Name, Version und Status zurück. Read-only. "
            "[EN] Lists all blueprints in the Nomos catalog. Read-only."
        ),
        "inputSchema": {"type": "object", "properties": {}, "required": []},
    },
    {
        "name": "get_blueprint",
        "description": (
            "Gibt vollständige Details eines Blueprints zurück inkl. required_inputs, "
            "required_service_blueprints und required_services. Read-only. "
            "[EN] Returns full details of a blueprint. Read-only."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "blueprint_id": {
                    "type": "string",
                    "description": "Die Blueprint-ID, z.B. 'PB-ACC-MBX-001'",
                }
            },
            "required": ["blueprint_id"],
        },
    },
]


def _api_base() -> str:
    return os.getenv("NOMOS_API_BASE", "http://localhost:8080")


def _format_response(resp: httpx.Response) -> str:
    try:
        return json.dumps(resp.json(), ensure_ascii=False, indent=2)
    except Exception:
        return resp.text


def _format_error(resp: httpx.Response) -> str:
    return f"HTTP {resp.status_code}: {resp.text}"


@require_scope(AgentScope.READ)
async def list_blueprints_handler() -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_api_base()}/api/v1/blueprints")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_blueprint_handler(blueprint_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_api_base()}/api/v1/blueprints/{blueprint_id}")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)
