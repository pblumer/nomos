"""MCP Tools für Nomos Instances (read-only)."""

import json
import os

import httpx

from ..auth import AgentScope, require_scope

INSTANCES_TOOLS = [
    {
        "name": "list_instances",
        "description": (
            "Listet alle Product-Instances im Nomos-Katalog auf. "
            "Gibt ID, Typ, Name, Version und Status zurück. Read-only. "
            "[EN] Lists all product instances in the Nomos catalog. Read-only."
        ),
        "inputSchema": {"type": "object", "properties": {}, "required": []},
    },
    {
        "name": "get_instance",
        "description": (
            "Gibt vollständige Details einer Product-Instance zurück inkl. "
            "blueprint_ref, inputs, observed_state und evidence. Read-only. "
            "[EN] Returns full details of a product instance. Read-only."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "instance_id": {
                    "type": "string",
                    "description": "Die Instance-ID, z.B. 'PI-ACC-MBX-001'",
                }
            },
            "required": ["instance_id"],
        },
    },
    {
        "name": "get_instance_compliance",
        "description": (
            "Gibt den Compliance-Status einer Product-Instance zurück inkl. "
            "Evidence und Findings. Read-only. "
            "[EN] Returns compliance status of a product instance including evidence and findings. Read-only."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "instance_id": {
                    "type": "string",
                    "description": "Die Instance-ID",
                }
            },
            "required": ["instance_id"],
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
async def list_instances_handler() -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_api_base()}/api/v1/instances")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_instance_handler(instance_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_api_base()}/api/v1/instances/{instance_id}")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_instance_compliance_handler(instance_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(
            f"{_api_base()}/api/v1/instances/{instance_id}/compliance"
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)
