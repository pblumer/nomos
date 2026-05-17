"""MCP Tools für Nomos Processes (Go-Subsystem).

Prozesse sind BPMN-Prozess-Artefakte die einem Produkt (Blueprint) zugeordnet sind.
Alle Operationen laufen über den nomos Go-Server (NOMOS_GO_API_BASE).
"""

import json
import os

import httpx

from ..auth import AgentScope, require_scope

PROCESS_TOOLS = [
    {
        "name": "list_processes",
        "description": (
            "Listet alle Prozesse eines Produkts auf (ID, Name, Status, Schritte). "
            "[EN] Lists all processes of a product."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "product_id": {
                    "type": "string",
                    "description": "Blueprint-ID des Produkts, z.B. 'bp-order-fulfillment'",
                }
            },
            "required": ["product_id"],
        },
    },
    {
        "name": "get_process",
        "description": (
            "Gibt Details eines Prozesses zurück: Schritte, BPMN-Referenz, Task-Mappings. "
            "[EN] Returns process details including steps and BPMN reference."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "process_id": {
                    "type": "string",
                    "description": "Prozess-ID, z.B. 'proc-001'",
                }
            },
            "required": ["process_id"],
        },
    },
    {
        "name": "create_process",
        "description": (
            "Erstellt einen neuen BPMN-Prozess und verknüpft ihn mit einem Produkt. "
            "Der Prozess erhält einen leeren BPMN-Start/End-Rahmen und kann dann "
            "mit Schritten und Business Rule Tasks befüllt werden. "
            "Erfordert DRAFT-Scope. "
            "[EN] Creates a new BPMN process linked to a product. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "product_id": {
                    "type": "string",
                    "description": "Blueprint-ID des Produkts",
                },
                "name": {
                    "type": "string",
                    "description": "Name des Prozesses, z.B. 'Bestellabwicklung'",
                },
                "summary": {
                    "type": "string",
                    "description": "Kurzbeschreibung des Prozesses",
                },
                "id": {
                    "type": "string",
                    "description": "Optionale Prozess-ID (wird sonst automatisch vergeben)",
                },
            },
            "required": ["product_id", "name"],
        },
    },
    {
        "name": "add_process_step",
        "description": (
            "Fügt einen Schritt zu einem bestehenden Prozess hinzu. "
            "Für einen Business Rule Task (task_type='businessRuleTask') muss ein "
            "decision_ref angegeben werden — der Schritt erhält automatisch ein "
            "Gateway nach dem Task. "
            "Erfordert DRAFT-Scope. "
            "[EN] Adds a step to a process. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "process_id": {"type": "string", "description": "Prozess-ID"},
                "name": {"type": "string", "description": "Schrittname"},
                "task_type": {
                    "type": "string",
                    "enum": [
                        "task",
                        "userTask",
                        "serviceTask",
                        "businessRuleTask",
                        "manualTask",
                        "scriptTask",
                    ],
                    "description": "BPMN Task-Typ",
                },
                "service_ref": {
                    "type": "string",
                    "description": "Service-Referenz, z.B. 'commerce.blumer.cloud/checkout'",
                },
                "method": {
                    "type": "string",
                    "description": "Aufgerufene Methode des Services",
                },
                "decision_ref": {
                    "type": "string",
                    "description": "Decision-ID für businessRuleTask, z.B. 'DEC-001'",
                },
                "required": {"type": "boolean", "description": "Pflichtschritt?"},
                "depends_on": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "IDs von Schritten die zuerst abgeschlossen sein müssen",
                },
            },
            "required": ["process_id", "name", "task_type"],
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
async def list_processes_handler(product_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_go_api_base()}/api/v1/products/{product_id}/processes")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_process_handler(process_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_go_api_base()}/api/v1/processes/{process_id}")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def create_process_handler(product_id: str, name: str, summary: str = "", id: str = "") -> str:
    payload: dict = {"name": name, "summary": summary}
    if id:
        payload["id"] = id
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{_go_api_base()}/api/v1/products/{product_id}/processes",
            json=payload,
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def add_process_step_handler(
    process_id: str,
    name: str,
    task_type: str,
    service_ref: str = "",
    method: str = "",
    decision_ref: str = "",
    required: bool = True,
    depends_on: list[str] | None = None,
) -> str:
    payload: dict = {
        "name": name,
        "taskType": task_type,
        "serviceRef": service_ref,
        "method": method,
        "required": required,
    }
    if depends_on:
        payload["dependsOn"] = depends_on
    if decision_ref:
        payload["decisionRef"] = decision_ref
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{_go_api_base()}/api/v1/processes/{process_id}/steps",
            json=payload,
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)
