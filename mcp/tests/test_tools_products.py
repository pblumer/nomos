"""Tests für Nomos MCP Product-Tools mit gemockten HTTP-Responses."""

import json
import os

import httpx
import pytest
import respx

from nomos_mcp.tools.products import (
    create_product_handler,
    delete_product_handler,
    get_product_handler,
    list_products_handler,
    update_product_handler,
    add_product_requirement_handler,
    remove_product_requirement_handler,
)

API_BASE = "http://localhost:8080"

SAMPLE_PRODUCT = {
    "id": "PROD-001",
    "name": "Test Produkt",
    "version": "0.1.0",
    "status": "draft",
    "owner": "Test Team",
    "requirements": [],
    "rules": [],
    "variants": [],
    "validation": {"is_valid": True, "errors": []},
}

PRODUCTS_LIST = {"items": [{"id": "PROD-001", "name": "Test Produkt", "version": "0.1.0"}], "count": 1}


@pytest.fixture(autouse=True)
def set_read_scope(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    monkeypatch.setenv("NOMOS_API_BASE", API_BASE)


@respx.mock
@pytest.mark.asyncio
async def test_list_products(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    respx.get(f"{API_BASE}/api/v1/products").mock(
        return_value=httpx.Response(200, json=PRODUCTS_LIST)
    )
    result = await list_products_handler()
    data = json.loads(result)
    assert data["count"] == 1
    assert data["items"][0]["id"] == "PROD-001"


@respx.mock
@pytest.mark.asyncio
async def test_get_product(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    respx.get(f"{API_BASE}/api/v1/products/PROD-001").mock(
        return_value=httpx.Response(200, json=SAMPLE_PRODUCT)
    )
    result = await get_product_handler("PROD-001")
    data = json.loads(result)
    assert data["id"] == "PROD-001"
    assert data["name"] == "Test Produkt"


@respx.mock
@pytest.mark.asyncio
async def test_get_product_not_found(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    respx.get(f"{API_BASE}/api/v1/products/PROD-999").mock(
        return_value=httpx.Response(404, json={"detail": "Produkt nicht gefunden"})
    )
    result = await get_product_handler("PROD-999")
    assert "HTTP 404" in result


@respx.mock
@pytest.mark.asyncio
async def test_create_product_requires_draft_scope(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    with pytest.raises(PermissionError):
        await create_product_handler({"id": "PROD-002", "name": "Neu", "version": "0.1.0"})


@respx.mock
@pytest.mark.asyncio
async def test_create_product_with_draft_scope(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "draft")
    new_product = {**SAMPLE_PRODUCT, "id": "PROD-002", "name": "Neues Produkt"}
    respx.post(f"{API_BASE}/api/v1/products").mock(
        return_value=httpx.Response(201, json=new_product)
    )
    result = await create_product_handler({"id": "PROD-002", "name": "Neues Produkt", "version": "0.1.0"})
    data = json.loads(result)
    assert data["id"] == "PROD-002"


@respx.mock
@pytest.mark.asyncio
async def test_update_product_requires_draft_scope(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    with pytest.raises(PermissionError):
        await update_product_handler("PROD-001", {"name": "Geändert"})


@respx.mock
@pytest.mark.asyncio
async def test_update_product_with_draft_scope(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "draft")
    updated = {**SAMPLE_PRODUCT, "name": "Geändert"}
    respx.put(f"{API_BASE}/api/v1/products/PROD-001").mock(
        return_value=httpx.Response(200, json=updated)
    )
    result = await update_product_handler("PROD-001", {"name": "Geändert"})
    data = json.loads(result)
    assert data["name"] == "Geändert"


@respx.mock
@pytest.mark.asyncio
async def test_delete_product_requires_draft_scope(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    with pytest.raises(PermissionError):
        await delete_product_handler("PROD-001")


@respx.mock
@pytest.mark.asyncio
async def test_delete_product_with_draft_scope(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "draft")
    respx.delete(f"{API_BASE}/api/v1/products/PROD-001").mock(
        return_value=httpx.Response(200, json={"id": "PROD-001", "removed": True})
    )
    result = await delete_product_handler("PROD-001")
    data = json.loads(result)
    assert data["removed"] is True


@respx.mock
@pytest.mark.asyncio
async def test_add_product_requirement_with_draft_scope(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "draft")
    respx.post(f"{API_BASE}/api/v1/products/PROD-001/requirements").mock(
        return_value=httpx.Response(201, json={"item": "REQ-001"})
    )
    result = await add_product_requirement_handler("PROD-001", "REQ-001")
    data = json.loads(result)
    assert data["item"] == "REQ-001"


@respx.mock
@pytest.mark.asyncio
async def test_remove_product_requirement_with_draft_scope(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "draft")
    respx.delete(f"{API_BASE}/api/v1/products/PROD-001/requirements/REQ-001").mock(
        return_value=httpx.Response(200, json={"item": "REQ-001", "removed": True})
    )
    result = await remove_product_requirement_handler("PROD-001", "REQ-001")
    data = json.loads(result)
    assert data["removed"] is True
