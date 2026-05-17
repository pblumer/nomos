"""MCP Tools für Nomos Decisions (Go-Subsystem).

Eine Decision ist ein eigenständiges Artefakt in einer Domain.
Sie kann als DMN-Datei hinterlegt werden und wird von BPMN-Prozessen
über businessRuleTask-Schritte aufgerufen. Nach einem businessRuleTask
folgt typischerweise ein ExclusiveGateway, das auf den Output-Variablen
der Decision basiert.

Alle Operationen laufen über den nomos Go-Server (NOMOS_GO_API_BASE).
"""

import json
import os

import httpx

from ..auth import AgentScope, require_scope

DECISION_TOOLS = [
    {
        "name": "list_decisions",
        "description": (
            "Listet alle Decisions einer Domain auf (ID, Name, Status, ob DMN vorhanden). "
            "[EN] Lists all decision artifacts of a domain."
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
        "name": "get_decision",
        "description": (
            "Gibt Details einer Decision zurück: Inputs, Outputs, DMN-Status. "
            "[EN] Returns details of a decision including inputs and outputs."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "domain": {"type": "string", "description": "DNS-Name der Domain"},
                "decision_id": {"type": "string", "description": "Decision-ID, z.B. 'DEC-001'"},
            },
            "required": ["domain", "decision_id"],
        },
    },
    {
        "name": "create_decision",
        "description": (
            "Erstellt eine neue Decision in einer Domain. "
            "Eine Decision kapselt Entscheidungslogik (optional als DMN-Datei). "
            "Sie wird von BPMN-Prozessen über businessRuleTask-Schritte aufgerufen. "
            "Outputs der Decision steuern anschliessende ExclusiveGateways. "
            "Erfordert DRAFT-Scope. "
            "[EN] Creates a new decision artifact in a domain. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "domain": {"type": "string", "description": "DNS-Name der Domain"},
                "name": {"type": "string", "description": "Name der Decision"},
                "id": {
                    "type": "string",
                    "description": "Optionale ID (z.B. 'DEC-001'); wird sonst automatisch vergeben",
                },
                "number": {
                    "type": "string",
                    "description": "Fachliche Nummer (z.B. 'DEC-001')",
                },
                "owner": {"type": "string", "description": "Verantwortliches Team"},
                "summary": {"type": "string", "description": "Kurzbeschreibung"},
                "context": {
                    "type": "string",
                    "description": "Fachlicher Kontext / Problemstellung",
                },
                "status": {
                    "type": "string",
                    "enum": ["draft", "active", "deprecated"],
                    "description": "Status der Decision",
                },
                "inputs": {
                    "type": "array",
                    "description": "Input-Variablen der Decision",
                    "items": {
                        "type": "object",
                        "properties": {
                            "name": {"type": "string"},
                            "type": {
                                "type": "string",
                                "enum": ["string", "number", "boolean", "date"],
                            },
                            "description": {"type": "string"},
                        },
                        "required": ["name", "type"],
                    },
                },
                "outputs": {
                    "type": "array",
                    "description": "Output-Variablen der Decision (für Gateway-Routing)",
                    "items": {
                        "type": "object",
                        "properties": {
                            "name": {"type": "string"},
                            "type": {
                                "type": "string",
                                "enum": ["string", "number", "boolean", "date"],
                            },
                            "description": {"type": "string"},
                        },
                        "required": ["name", "type"],
                    },
                },
            },
            "required": ["domain", "name"],
        },
    },
    {
        "name": "update_decision_dmn",
        "description": (
            "Lädt eine DMN-XML-Datei für eine bestehende Decision hoch. "
            "DMN (Decision Model and Notation) ist der Standard für Entscheidungstabellen. "
            "Erfordert DRAFT-Scope. "
            "[EN] Uploads a DMN XML file for an existing decision. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "domain": {"type": "string", "description": "DNS-Name der Domain"},
                "decision_id": {"type": "string", "description": "Decision-ID"},
                "dmn_xml": {
                    "type": "string",
                    "description": "DMN XML-Inhalt (vollständiges DMN 1.3 XML-Dokument)",
                },
            },
            "required": ["domain", "decision_id", "dmn_xml"],
        },
    },
    {
        "name": "get_decision_dmn",
        "description": (
            "Gibt den DMN XML-Inhalt einer Decision zurück, falls vorhanden. "
            "[EN] Returns the DMN XML content of a decision if present."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "domain": {"type": "string", "description": "DNS-Name der Domain"},
                "decision_id": {"type": "string", "description": "Decision-ID"},
            },
            "required": ["domain", "decision_id"],
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
async def list_decisions_handler(domain: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_go_api_base()}/api/v1/domains/{domain}/decisions")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_decision_handler(domain: str, decision_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(
            f"{_go_api_base()}/api/v1/domains/{domain}/decisions/{decision_id}"
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def create_decision_handler(
    domain: str,
    name: str,
    id: str = "",
    number: str = "",
    owner: str = "",
    summary: str = "",
    context: str = "",
    status: str = "draft",
    inputs: list[dict] | None = None,
    outputs: list[dict] | None = None,
) -> str:
    payload: dict = {
        "name": name,
        "owner": owner,
        "summary": summary,
        "context": context,
        "status": status,
    }
    if id:
        payload["id"] = id
    if number:
        payload["number"] = number
    if inputs:
        payload["inputs"] = inputs
    if outputs:
        payload["outputs"] = outputs
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{_go_api_base()}/api/v1/domains/{domain}/decisions",
            json=payload,
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def update_decision_dmn_handler(domain: str, decision_id: str, dmn_xml: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.put(
            f"{_go_api_base()}/api/v1/domains/{domain}/decisions/{decision_id}/dmn",
            content=dmn_xml.encode("utf-8"),
            headers={"Content-Type": "application/xml; charset=utf-8"},
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_decision_dmn_handler(domain: str, decision_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(
            f"{_go_api_base()}/api/v1/domains/{domain}/decisions/{decision_id}/dmn"
        )
        if not resp.is_success:
            return _format_error(resp)
        return resp.text
