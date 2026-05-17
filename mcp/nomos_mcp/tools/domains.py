"""MCP Tools für Nomos Domains (Go-Subsystem).

Domains sind DNS-benannte Namespaces im Cosmos-Tree.
Alle Operationen laufen über den nomos Go-Server (NOMOS_GO_API_BASE).
"""

import json
import os

import httpx

from ..auth import AgentScope, require_scope

DOMAIN_TOOLS = [
    {
        "name": "list_domains",
        "description": (
            "Listet alle Domains im Cosmos auf (DNS-Namen, Owners, Service-Anzahl). "
            "[EN] Lists all domains in the cosmos."
        ),
        "inputSchema": {"type": "object", "properties": {}, "required": []},
    },
    {
        "name": "get_domain",
        "description": (
            "Gibt Details einer Domain zurück: DNS-Name, Services, Produkte, Status. "
            "[EN] Returns details of a domain including services and products."
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
        "name": "create_domain",
        "description": (
            "Erstellt eine neue Top-Level-Domain im Cosmos. "
            "Der DNS-Name muss der DNS-Namespacing-Konvention folgen (z.B. 'identity.blumer.cloud'). "
            "Erfordert DRAFT-Scope. "
            "[EN] Creates a new top-level domain. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "dns_name": {
                    "type": "string",
                    "description": "DNS-Name der neuen Domain, z.B. 'commerce.blumer.cloud'",
                },
                "owner": {
                    "type": "string",
                    "description": "Verantwortliches Team oder Person",
                },
            },
            "required": ["dns_name"],
        },
    },
    {
        "name": "create_child_domain",
        "description": (
            "Erstellt eine Kind-Domain unter einer bestehenden Domain. "
            "Erfordert DRAFT-Scope. "
            "[EN] Creates a child domain under an existing domain. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "parent_domain": {
                    "type": "string",
                    "description": "DNS-Name der Eltern-Domain, z.B. 'blumer.cloud'",
                },
                "segment": {
                    "type": "string",
                    "description": "Neues Segment (nur lowercase, z.B. 'commerce')",
                },
                "owner": {
                    "type": "string",
                    "description": "Verantwortliches Team",
                },
            },
            "required": ["parent_domain", "segment"],
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
async def list_domains_handler() -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_go_api_base()}/api/v1/domains")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_domain_handler(domain: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_go_api_base()}/api/v1/domains/{domain}")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def create_domain_handler(dns_name: str, owner: str = "") -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{_go_api_base()}/api/v1/domains",
            json={"dns": dns_name, "owner": owner},
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def create_child_domain_handler(parent_domain: str, segment: str, owner: str = "") -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{_go_api_base()}/api/v1/domains/{parent_domain}/children",
            json={"segment": segment, "owner": owner},
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)
