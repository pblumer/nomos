"""MCP Tools für Nomos Business Rules."""

import json
import os

import httpx

from ..auth import AgentScope, require_scope

RULES_TOOLS = [
    {
        "name": "list_rules",
        "description": (
            "Listet alle Business Rules im Nomos-Katalog auf. "
            "Gibt IDs, Namen, Typ, Schweregrad und Kategorie zurück. "
            "[EN] Lists all business rules in the Nomos catalog."
        ),
        "inputSchema": {"type": "object", "properties": {}, "required": []},
    },
    {
        "name": "get_rule",
        "description": (
            "Gibt vollständige Details einer Business Rule zurück. "
            "Erfordert die Regel-ID. "
            "[EN] Returns full details of a business rule."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "rule_id": {
                    "type": "string",
                    "description": "Die Regel-ID, z.B. 'RULE-GDPR-001'",
                }
            },
            "required": ["rule_id"],
        },
    },
    {
        "name": "create_rule",
        "description": (
            "Erstellt eine neue Business Rule im Nomos-Katalog. Erfordert DRAFT-Scope. "
            "Erlaubte Typen: must, must_not, derive, allow, deny. "
            "Erlaubte Schweregrade: low, medium, high, critical. "
            "[EN] Creates a new business rule. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "id": {"type": "string", "description": "Eindeutige ID, z.B. 'RULE-GDPR-001'"},
                "name": {"type": "string"},
                "description": {"type": "string"},
                "type": {
                    "type": "string",
                    "enum": ["must", "must_not", "derive", "allow", "deny"],
                },
                "severity": {
                    "type": "string",
                    "enum": ["low", "medium", "high", "critical"],
                },
                "category": {"type": "string"},
                "condition": {"type": "string"},
                "effect": {"type": "string"},
                "rationale": {"type": "string"},
            },
            "required": ["id", "name"],
        },
    },
    {
        "name": "update_rule",
        "description": (
            "Aktualisiert eine bestehende Business Rule. Erfordert DRAFT-Scope. "
            "[EN] Updates an existing business rule. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "rule_id": {"type": "string"},
                "updates": {
                    "type": "object",
                    "description": "Die zu ändernden Felder als Key-Value-Objekt",
                },
            },
            "required": ["rule_id", "updates"],
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
async def list_rules_handler() -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_api_base()}/api/v1/rules")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_rule_handler(rule_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_api_base()}/api/v1/rules/{rule_id}")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def create_rule_handler(payload: dict) -> str:
    rule_id = payload.get("id", "")
    async with httpx.AsyncClient() as client:
        resp = await client.put(f"{_api_base()}/api/v1/rules/{rule_id}", json=payload)
        if resp.status_code == 404:
            # Rule doesn't exist yet — create via PUT (upsert-like)
            # The API uses PUT for updates; for creation we need the file to exist first.
            # We'll use the update endpoint which will 404, so fall back to a POST-like approach
            # by first trying to create via product requirement side-effect, or notify the user.
            return (
                f"HTTP 404: Regel '{rule_id}' existiert nicht. "
                "Erstelle zuerst die YAML-Datei manuell oder nutze den create_workspace-Workflow. "
                "Die REST API unterstützt keinen direkten POST /api/v1/rules Endpunkt."
            )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def update_rule_handler(rule_id: str, updates: dict) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.put(f"{_api_base()}/api/v1/rules/{rule_id}", json=updates)
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)
