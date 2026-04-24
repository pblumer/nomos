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
                "rules": [
                    "rule-mailbox-quotas",
                    "rule-password-policy",
                    "rule-account-status",
                ],
            },
            sort_keys=False,
            allow_unicode=True,
        ),
        encoding="utf-8",
    )


def test_rules_endpoint_returns_rules_for_product(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()
    _write_product(products_dir)

    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))
    client = TestClient(app)

    response = client.get("/api/v1/products/produkt-1/rules")

    assert response.status_code == 200
    assert response.json() == {
        "product_id": "produkt-1",
        "items": [
            "rule-mailbox-quotas",
            "rule-password-policy",
            "rule-account-status",
        ],
        "count": 3,
    }


def test_rules_endpoint_returns_404_for_unknown_product(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()
    _write_product(products_dir)

    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))
    client = TestClient(app)

    response = client.get("/api/v1/products/unknown/rules")

    assert response.status_code == 404
    assert response.json() == {"detail": "Produkt nicht gefunden"}
