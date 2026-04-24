from pathlib import Path

import yaml
from fastapi.testclient import TestClient

from backend.app.main import app


def _write_product(products_dir: Path, product_id: str, requirement_ids: list[str], rule_ids: list[str]) -> None:
    product_file = products_dir / f"{product_id}.yml"
    product_file.write_text(
        yaml.safe_dump(
            {
                "id": product_id,
                "name": f"Product {product_id}",
                "version": "0.1.0",
                "requirement_ids": requirement_ids,
                "rule_ids": rule_ids,
            },
            sort_keys=False,
            allow_unicode=True,
        ),
        encoding="utf-8",
    )


def test_shared_requirement_and_rule_can_be_reused_by_multiple_products(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    requirements_dir = tmp_path / "requirements"
    rules_dir = tmp_path / "rules"
    products_dir.mkdir()
    requirements_dir.mkdir()
    rules_dir.mkdir()

    (requirements_dir / "req-shared.yml").write_text("id: req-shared\nname: Shared Requirement\n", encoding="utf-8")
    (rules_dir / "rule-shared.yml").write_text("id: rule-shared\nname: Shared Rule\n", encoding="utf-8")

    _write_product(products_dir, "produkt-1", ["req-shared"], ["rule-shared"])
    _write_product(products_dir, "produkt-2", ["req-shared"], ["rule-shared"])

    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))
    monkeypatch.setenv("NOMOS_REQUIREMENTS_DIR", str(requirements_dir))
    monkeypatch.setenv("NOMOS_RULES_DIR", str(rules_dir))

    client = TestClient(app)

    r1 = client.get("/api/v1/products/produkt-1/requirements")
    r2 = client.get("/api/v1/products/produkt-2/requirements")
    s1 = client.get("/api/v1/products/produkt-1/rules")
    s2 = client.get("/api/v1/products/produkt-2/rules")

    assert r1.status_code == 200
    assert r2.status_code == 200
    assert s1.status_code == 200
    assert s2.status_code == 200

    assert r1.json()["items"] == ["req-shared"]
    assert r2.json()["items"] == ["req-shared"]
    assert s1.json()["items"] == ["rule-shared"]
    assert s2.json()["items"] == ["rule-shared"]


def test_add_requirement_and_rule_create_catalog_artifact_files(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    requirements_dir = tmp_path / "requirements"
    rules_dir = tmp_path / "rules"
    products_dir.mkdir()
    requirements_dir.mkdir()
    rules_dir.mkdir()

    _write_product(products_dir, "produkt-1", [], [])

    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))
    monkeypatch.setenv("NOMOS_REQUIREMENTS_DIR", str(requirements_dir))
    monkeypatch.setenv("NOMOS_RULES_DIR", str(rules_dir))

    client = TestClient(app)

    add_req = client.post("/api/v1/products/produkt-1/requirements", json={"value": "req-new"})
    add_rule = client.post("/api/v1/products/produkt-1/rules", json={"value": "rule-new"})

    assert add_req.status_code == 201
    assert add_rule.status_code == 201
    assert (requirements_dir / "req-new.yml").exists()
    assert (rules_dir / "rule-new.yml").exists()
