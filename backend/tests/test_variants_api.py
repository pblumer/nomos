import pytest
from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


@pytest.fixture(autouse=True)
def clean_catalog(tmp_path, monkeypatch):
    products_dir = tmp_path / "products"
    products_dir.mkdir(parents=True, exist_ok=True)
    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))
    yield


def test_list_variants_empty():
    client.post("/api/v1/products", json={"id": "prod-1", "name": "Product 1", "version": "1.0.0"})
    response = client.get("/api/v1/products/prod-1/variants")
    assert response.status_code == 200
    data = response.json()
    assert data["items"] == []
    assert data["count"] == 0


def test_add_and_list_variants():
    client.post("/api/v1/products", json={"id": "prod-1", "name": "Product 1", "version": "1.0.0"})
    response = client.post("/api/v1/products/prod-1/variants", json={"id": "intern", "name": "Intern", "context": "internal"})
    assert response.status_code == 201
    assert response.json()["item"] == "intern"

    response = client.get("/api/v1/products/prod-1/variants")
    assert response.status_code == 200
    data = response.json()
    assert data["count"] == 1
    assert data["items"][0]["id"] == "intern"
    assert data["items"][0]["name"] == "Intern"


def test_add_variant_rejects_empty_id():
    client.post("/api/v1/products", json={"id": "prod-1", "name": "Product 1", "version": "1.0.0"})
    response = client.post("/api/v1/products/prod-1/variants", json={"name": "x"})
    assert response.status_code == 400
    assert "id" in response.json()["detail"].lower()


def test_add_variant_rejects_duplicate():
    client.post("/api/v1/products", json={"id": "prod-1", "name": "Product 1", "version": "1.0.0"})
    client.post("/api/v1/products/prod-1/variants", json={"id": "intern"})
    response = client.post("/api/v1/products/prod-1/variants", json={"id": "intern"})
    assert response.status_code == 409


def test_delete_variant():
    client.post("/api/v1/products", json={"id": "prod-1", "name": "Product 1", "version": "1.0.0"})
    client.post("/api/v1/products/prod-1/variants", json={"id": "intern"})
    response = client.delete("/api/v1/products/prod-1/variants/intern")
    assert response.status_code == 200
    assert response.json()["removed"] is True

    response = client.get("/api/v1/products/prod-1/variants")
    assert response.json()["count"] == 0


def test_delete_variant_not_found():
    client.post("/api/v1/products", json={"id": "prod-1", "name": "Product 1", "version": "1.0.0"})
    response = client.delete("/api/v1/products/prod-1/variants/missing")
    assert response.status_code == 404
