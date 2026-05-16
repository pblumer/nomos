"""MCP Tools für Nomos Requirements (Anforderungen)."""

import json
import os

import httpx

from ..auth import AgentScope, require_scope

REQUIREMENTS_TOOLS = [
    {
        "name": "list_requirements",
        "description": (
            "Listet alle Anforderungen im Nomos-Katalog auf. "
            "Gibt IDs, Namen, Kategorie, Priorität und Quelle zurück. "
            "[EN] Lists all requirements in the Nomos catalog."
        ),
        "inputSchema": {"type": "object", "properties": {}, "required": []},
    },
    {
        "name": "get_requirement",
        "description": (
            "Gibt vollständige Details einer Anforderung zurück. "
            "Erfordert die Anforderungs-ID. "
            "[EN] Returns full details of a requirement."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "requirement_id": {
                    "type": "string",
                    "description": "Die Anforderungs-ID, z.B. 'REQ-001'",
                }
            },
            "required": ["requirement_id"],
        },
    },
    {
        "name": "update_requirement",
        "description": (
            "Aktualisiert eine bestehende Anforderung. Erfordert DRAFT-Scope. "
            "Erlaubte Prioritäten: low, medium, high, critical. "
            "[EN] Updates an existing requirement. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "requirement_id": {"type": "string"},
                "updates": {
                    "type": "object",
                    "description": "Die zu ändernden Felder als Key-Value-Objekt",
                },
            },
            "required": ["requirement_id", "updates"],
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
async def list_requirements_handler() -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_api_base()}/api/v1/requirements")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_requirement_handler(requirement_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_api_base()}/api/v1/requirements/{requirement_id}")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def update_requirement_handler(requirement_id: str, updates: dict) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.put(
            f"{_api_base()}/api/v1/requirements/{requirement_id}", json=updates
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)
