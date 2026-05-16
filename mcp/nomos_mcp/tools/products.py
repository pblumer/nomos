"""MCP Tools für Nomos Products (CRUD)."""

import json
import os

import httpx

from ..auth import AgentScope, require_scope

NOMOS_API_BASE = os.getenv("NOMOS_API_BASE", "http://localhost:8080")

PRODUCT_TOOLS = [
    {
        "name": "list_products",
        "description": (
            "Listet alle Produkte im Nomos-Katalog auf. "
            "Gibt IDs, Namen und Versionen zurück. "
            "Verwende dieses Tool um einen Überblick zu bekommen, bevor du ein "
            "spezifisches Produkt abrufst oder änderst. "
            "[EN] Lists all products in the Nomos catalog."
        ),
        "inputSchema": {"type": "object", "properties": {}, "required": []},
    },
    {
        "name": "get_product",
        "description": (
            "Gibt vollständige Details eines Produkts zurück inkl. Anforderungen, "
            "Business Rules und Varianten. Erfordert die Produkt-ID. "
            "[EN] Returns full details of a product including requirements, rules, and variants."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "product_id": {
                    "type": "string",
                    "description": "Die Produkt-ID, z.B. 'PROD-001'",
                }
            },
            "required": ["product_id"],
        },
    },
    {
        "name": "create_product",
        "description": (
            "Erstellt ein neues Produkt-Artefakt im Nomos-Katalog. "
            "Erfordert DRAFT-Scope. Das Produkt wird auf dem aktuellen Branch erstellt. "
            "Nach der Erstellung sollte validate_catalog aufgerufen werden. "
            "[EN] Creates a new product artifact. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "id": {"type": "string", "description": "Eindeutige ID, z.B. 'PROD-042'"},
                "name": {"type": "string", "description": "Produktname"},
                "version": {"type": "string", "description": "Versionsnummer, z.B. '0.1.0'"},
                "status": {
                    "type": "string",
                    "enum": ["draft", "active", "deprecated"],
                    "description": "Status des Produkts",
                },
                "owner": {"type": "string", "description": "Verantwortlicher"},
                "summary": {"type": "string", "description": "Kurzbeschreibung"},
                "requirement_ids": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "Liste der verknüpften Anforderungs-IDs",
                },
                "rule_ids": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "Liste der verknüpften Regel-IDs",
                },
            },
            "required": ["id", "name", "version"],
        },
    },
    {
        "name": "update_product",
        "description": (
            "Aktualisiert ein bestehendes Produkt. Erfordert DRAFT-Scope. "
            "Nur auf Branches erlaubt, niemals auf main. "
            "[EN] Updates an existing product. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "product_id": {"type": "string", "description": "Die Produkt-ID"},
                "updates": {
                    "type": "object",
                    "description": "Die zu ändernden Felder als Key-Value-Objekt",
                },
            },
            "required": ["product_id", "updates"],
        },
    },
    {
        "name": "delete_product",
        "description": (
            "Löscht ein Produkt aus dem Katalog. Erfordert DRAFT-Scope. "
            "Nur auf Branches erlaubt, niemals auf main. "
            "[EN] Deletes a product from the catalog. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "product_id": {"type": "string", "description": "Die Produkt-ID"},
            },
            "required": ["product_id"],
        },
    },
    {
        "name": "add_product_requirement",
        "description": (
            "Verknüpft eine Anforderungs-ID mit einem Produkt. Erfordert DRAFT-Scope. "
            "[EN] Links a requirement ID to a product. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "product_id": {"type": "string"},
                "requirement_id": {"type": "string", "description": "Die Anforderungs-ID"},
            },
            "required": ["product_id", "requirement_id"],
        },
    },
    {
        "name": "remove_product_requirement",
        "description": (
            "Entfernt eine Anforderungs-ID von einem Produkt. Erfordert DRAFT-Scope. "
            "[EN] Removes a requirement ID from a product. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "product_id": {"type": "string"},
                "requirement_id": {"type": "string"},
            },
            "required": ["product_id", "requirement_id"],
        },
    },
    {
        "name": "add_product_rule",
        "description": (
            "Verknüpft eine Regel-ID mit einem Produkt. Erfordert DRAFT-Scope. "
            "[EN] Links a rule ID to a product. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "product_id": {"type": "string"},
                "rule_id": {"type": "string", "description": "Die Regel-ID"},
            },
            "required": ["product_id", "rule_id"],
        },
    },
    {
        "name": "remove_product_rule",
        "description": (
            "Entfernt eine Regel-ID von einem Produkt. Erfordert DRAFT-Scope. "
            "[EN] Removes a rule ID from a product. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "product_id": {"type": "string"},
                "rule_id": {"type": "string"},
            },
            "required": ["product_id", "rule_id"],
        },
    },
    {
        "name": "get_product_variants",
        "description": (
            "Gibt alle Varianten eines Produkts zurück. "
            "[EN] Returns all variants of a product."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "product_id": {"type": "string"},
            },
            "required": ["product_id"],
        },
    },
]


def _api_base() -> str:
    return os.getenv("NOMOS_API_BASE", NOMOS_API_BASE)


def _format_response(resp: httpx.Response) -> str:
    try:
        return json.dumps(resp.json(), ensure_ascii=False, indent=2)
    except Exception:
        return resp.text


def _format_error(resp: httpx.Response) -> str:
    return (
        f"HTTP {resp.status_code}: {resp.text}"
    )


@require_scope(AgentScope.READ)
async def list_products_handler() -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_api_base()}/api/v1/products")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_product_handler(product_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{_api_base()}/api/v1/products/{product_id}")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def create_product_handler(payload: dict) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.post(f"{_api_base()}/api/v1/products", json=payload)
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def update_product_handler(product_id: str, updates: dict) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.put(
            f"{_api_base()}/api/v1/products/{product_id}", json=updates
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def delete_product_handler(product_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.delete(f"{_api_base()}/api/v1/products/{product_id}")
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def add_product_requirement_handler(product_id: str, requirement_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{_api_base()}/api/v1/products/{product_id}/requirements",
            json={"value": requirement_id},
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def remove_product_requirement_handler(product_id: str, requirement_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.delete(
            f"{_api_base()}/api/v1/products/{product_id}/requirements/{requirement_id}"
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def add_product_rule_handler(product_id: str, rule_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{_api_base()}/api/v1/products/{product_id}/rules",
            json={"value": rule_id},
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.DRAFT)
async def remove_product_rule_handler(product_id: str, rule_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.delete(
            f"{_api_base()}/api/v1/products/{product_id}/rules/{rule_id}"
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)


@require_scope(AgentScope.READ)
async def get_product_variants_handler(product_id: str) -> str:
    async with httpx.AsyncClient() as client:
        resp = await client.get(
            f"{_api_base()}/api/v1/products/{product_id}/variants"
        )
        if not resp.is_success:
            return _format_error(resp)
        return _format_response(resp)
