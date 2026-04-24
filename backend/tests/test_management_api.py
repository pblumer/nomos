from pathlib import Path

import yaml
from fastapi.testclient import TestClient

from backend.app.main import app


def _write_product(products_dir: Path, product_id: str = "produkt-1") -> None:
    product_file = products_dir / f"{product_id}.yml"
    product_file.write_text(
        yaml.safe_dump(
            {
                "id": product_id,
                "name": "Benutzerkonto mit Mailbox",
                "version": "0.1.0",
                "requirements": ["mailbox-enabled", "login-required"],
                "rules": ["rule-password-policy", "rule-mailbox-quotas"],
            },
            sort_keys=False,
            allow_unicode=True,
        ),
        encoding="utf-8",
    )


def test_overview_endpoint_returns_catalog_aggregates(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()
    _write_product(products_dir, "produkt-1")
    _write_product(products_dir, "produkt-2")

    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))
    client = TestClient(app)

    response = client.get("/api/v1/overview")

    assert response.status_code == 200
    assert response.json() == {
        "product_count": 2,
        "requirements_count": 4,
        "rules_count": 4,
        "invalid_products_count": 0,
    }


def test_add_and_delete_requirement(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()
    _write_product(products_dir)

    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))
    client = TestClient(app)

    add_response = client.post("/api/v1/products/produkt-1/requirements", json={"value": "mfa-required"})
    assert add_response.status_code == 201
    assert add_response.json() == {"item": "mfa-required"}

    list_response = client.get("/api/v1/products/produkt-1/requirements")
    assert list_response.status_code == 200
    assert "mfa-required" in list_response.json()["items"]

    delete_response = client.delete("/api/v1/products/produkt-1/requirements/mfa-required")
    assert delete_response.status_code == 200
    assert delete_response.json() == {"item": "mfa-required", "removed": True}


def test_add_and_delete_rule(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()
    _write_product(products_dir)

    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))
    client = TestClient(app)

    add_response = client.post("/api/v1/products/produkt-1/rules", json={"value": "rule-mfa-enforcement"})
    assert add_response.status_code == 201
    assert add_response.json() == {"item": "rule-mfa-enforcement"}

    list_response = client.get("/api/v1/products/produkt-1/rules")
    assert list_response.status_code == 200
    assert "rule-mfa-enforcement" in list_response.json()["items"]

    delete_response = client.delete("/api/v1/products/produkt-1/rules/rule-mfa-enforcement")
    assert delete_response.status_code == 200
    assert delete_response.json() == {"item": "rule-mfa-enforcement", "removed": True}
