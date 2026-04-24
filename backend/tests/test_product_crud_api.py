from pathlib import Path

import yaml
from fastapi.testclient import TestClient

from backend.app.main import app


def _write_product(products_dir: Path, product_id: str) -> None:
    (products_dir / f"{product_id}.yml").write_text(
        yaml.safe_dump(
            {
                "id": product_id,
                "name": "Existing Product",
                "version": "0.1.0",
                "description": "Initial",
                "requirement_ids": ["req-login-required"],
                "rule_ids": ["rule-password-policy"],
            },
            sort_keys=False,
            allow_unicode=True,
        ),
        encoding="utf-8",
    )


def test_create_product_persists_yaml_and_returns_created_resource(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()
    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))

    client = TestClient(app)

    response = client.post(
        "/api/v1/products",
        json={
            "id": "produkt-2",
            "name": "Neues Produkt",
            "version": "1.0.0",
            "description": "Created in API",
        },
    )

    assert response.status_code == 201
    payload = response.json()
    assert payload["id"] == "produkt-2"
    assert payload["name"] == "Neues Produkt"

    created_file = products_dir / "produkt-2.yml"
    assert created_file.exists()
    created_data = yaml.safe_load(created_file.read_text(encoding="utf-8"))
    assert created_data["id"] == "produkt-2"
    assert created_data["version"] == "1.0.0"


def test_create_product_rejects_duplicate_id(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()
    _write_product(products_dir, "produkt-1")
    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))

    client = TestClient(app)

    response = client.post(
        "/api/v1/products",
        json={"id": "produkt-1", "name": "Duplicate", "version": "1.0.0"},
    )

    assert response.status_code == 409
    assert response.json()["detail"] == "product already exists"


def test_update_product_changes_fields_but_keeps_existing_relations(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()
    _write_product(products_dir, "produkt-1")
    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))

    client = TestClient(app)

    response = client.put(
        "/api/v1/products/produkt-1",
        json={"name": "Renamed Product", "version": "0.2.0", "description": "Updated"},
    )

    assert response.status_code == 200
    payload = response.json()
    assert payload["name"] == "Renamed Product"
    assert payload["version"] == "0.2.0"
    assert payload["requirements"] == ["req-login-required"]
    assert payload["rules"] == ["rule-password-policy"]

    updated_data = yaml.safe_load((products_dir / "produkt-1.yml").read_text(encoding="utf-8"))
    assert updated_data["name"] == "Renamed Product"
    assert updated_data["requirement_ids"] == ["req-login-required"]


def test_delete_product_removes_yaml_file(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()
    _write_product(products_dir, "produkt-1")
    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))

    client = TestClient(app)

    response = client.delete("/api/v1/products/produkt-1")

    assert response.status_code == 200
    assert response.json() == {"id": "produkt-1", "removed": True}
    assert not (products_dir / "produkt-1.yml").exists()


def test_update_product_rejects_invalid_payload(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()
    _write_product(products_dir, "produkt-1")
    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))

    client = TestClient(app)

    response = client.put("/api/v1/products/produkt-1", json={"name": "No Version", "version": ""})

    assert response.status_code == 400
    assert response.json()["detail"] == "Missing required field: version"


def test_get_product_resolves_id_even_if_filename_differs(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()

    (products_dir / "produkt-benutzerkonto.yml").write_text(
        yaml.safe_dump(
            {
                "id": "produkt-1",
                "name": "Benutzerkonto mit Mailbox",
                "version": "0.1.0",
            },
            sort_keys=False,
            allow_unicode=True,
        ),
        encoding="utf-8",
    )

    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))
    client = TestClient(app)

    response = client.get("/api/v1/products/produkt-1")

    assert response.status_code == 200
    assert response.json()["name"] == "Benutzerkonto mit Mailbox"
