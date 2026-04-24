import os
from pathlib import Path

import yaml
from fastapi import FastAPI

app = FastAPI(title="Nomos API", version="0.1.0")


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/api/v1/products")
def list_products() -> dict[str, object]:
    products_dir = Path(os.getenv("NOMOS_PRODUCTS_DIR", "catalog/products"))
    product_files = sorted([
        *products_dir.glob("*.yml"),
        *products_dir.glob("*.yaml"),
    ])

    items: list[dict[str, str]] = []
    for file_path in product_files:
        data = yaml.safe_load(file_path.read_text(encoding="utf-8")) or {}
        items.append(
            {
                "id": str(data.get("id", "")),
                "name": str(data.get("name", "")),
                "version": str(data.get("version", "")),
            }
        )

    return {"items": items, "count": len(items)}
