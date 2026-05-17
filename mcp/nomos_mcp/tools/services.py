"""MCP Tools für Nomos Services (Go-Subsystem).

Services sind konkrete logische Dienste unter einer Domain.
Alle Operationen laufen über den nomos Go-Server (NOMOS_GO_API_BASE).
"""

import json
import os

import httpx

from ..auth import AgentScope, require_scope

SERVICE_TOOLS = [
    {
        "name": "list_services",
        "description": (
            "Listet alle Services einer Domain auf. "
            "[EN] Lists all services of a domain."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "domain": {
                    "type": "string",
                    "description": "DNS-Name der Domain, z.B. 'commerce.blumer.cloud'",
                }
            },
            "required": ["domain"],
        },
    },
    {
        "name": "get_service",
        "description": (
            "Gibt Details eines Services zurück: Owner, Capabilities, SLA, Methoden. "
            "[EN] Returns details of a service."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "domain": {
                    "type": "string",
                    "description": "DNS-Name der Domain",
                },
                "service": {
                    "type": "string",
                    "description": "Service-Name, z.B. 'checkout'",
                },
            },
            "required": ["domain", "service"],
        },
    },
    {
        "name": "create_service",
        "description": (
            "Erstellt einen neuen Service unter einer Domain. "
            "Erfordert DRAFT-Scope. "
            "[EN] Creates a new service under a domain. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "domain": {
                    "type": "string",
                    "description": "DNS-Name der Domain",
                },
                "name": {
                    "type": "string",
                    "description": "Service-Name (lowercase, Bindestriche erlaubt), z.B. 'checkout'",
                },
                "owner": {
                    "type": "string",
                    "description": "Verantwortliches Team",
                },
                "summary": {
                    "type": "string",
                    "description": "Kurzbeschreibung des Services",
                },
                "capabilities": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "Liste von Capabilities, z.B. ['payment', 'refund']",
                },
            },
            "required": ["domain", "name"],
        },
    },
]


def _go_api_base() -> str:
    return os.getenv("NOMOS_GO_API_BASE", "http://localhost:9090")


def _format_response(resp: httpx.Response) -> str:
    try:
        return json.dumps(resp.json(), ensure_ascii=False, indent=2)
    except Exception:
        return resp.text


def _format_error(resp: httpx.Response) -> str:
    return f"HTTP {resp.status_code}: {resp.text}"


@require_scope(AgentScope.READ)
async def list_services_handler(domain: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_go_api_base()}/api/v1/domains/{domain}/services")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_service_handler(domain: str, service: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_go_api_base()}/api/v1/domains/{domain}/services/{service}")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def create_service_handler(
    domain: str,
    name: str,
    owner: str = "",
    summary: str = "",
    capabilities: list[str] | None = None,
) -> str:
    payload: dict = {"name": name, "owner": owner, "summary": summary}
    if capabilities:
        payload["capabilities"] = capabilities
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{_go_api_base()}/api/v1/domains/{domain}/services",
            json=payload,
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)
