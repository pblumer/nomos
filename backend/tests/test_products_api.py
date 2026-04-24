from pathlib import Path

import yaml
from fastapi.testclient import TestClient

from backend.app.main import app


def test_products_endpoint_reads_yaml_files(monkeypatch, tmp_path: Path) -> None:
    products_dir = tmp_path / "products"
    products_dir.mkdir()

    product_file = products_dir / "produkt-1.yml"
    product_file.write_text(
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

    response = client.get("/api/v1/products")

    assert response.status_code == 200
    assert response.json() == {
        "items": [
            {
                "id": "produkt-1",
                "name": "Benutzerkonto mit Mailbox",
                "version": "0.1.0",
            }
        ],
        "count": 1,
    }
