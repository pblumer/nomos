import os
from pathlib import Path

import yaml
from fastapi import FastAPI, HTTPException

app = FastAPI(title="Nomos API", version="0.1.0")


def _products_dir() -> Path:
    return Path(os.getenv("NOMOS_PRODUCTS_DIR", "catalog/products"))


def _read_yaml_file(file_path: Path) -> dict[str, object]:
    return yaml.safe_load(file_path.read_text(encoding="utf-8")) or {}


def _validate_product(data: dict[str, object]) -> dict[str, object]:
    required_fields = ["id", "name", "version"]
    errors: list[str] = []

    for field_name in required_fields:
        value = data.get(field_name)
        if value is None or str(value).strip() == "":
            errors.append(f"Missing required field: {field_name}")

    return {
        "is_valid": len(errors) == 0,
        "errors": errors,
    }


def _load_product_or_404(product_id: str) -> dict[str, object]:
    yml_path = _products_dir() / f"{product_id}.yml"
    yaml_path = _products_dir() / f"{product_id}.yaml"

    if yml_path.exists():
        product = _read_yaml_file(yml_path)
        product["validation"] = _validate_product(product)
        return product

    if yaml_path.exists():
        product = _read_yaml_file(yaml_path)
        product["validation"] = _validate_product(product)
        return product

    raise HTTPException(status_code=404, detail="Produkt nicht gefunden")


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/api/v1/products")
def list_products() -> dict[str, object]:
    product_files = sorted([
        *_products_dir().glob("*.yml"),
        *_products_dir().glob("*.yaml"),
    ])

    items: list[dict[str, str]] = []
    for file_path in product_files:
        data = _read_yaml_file(file_path)
        items.append(
            {
                "id": str(data.get("id", "")),
                "name": str(data.get("name", "")),
                "version": str(data.get("version", "")),
            }
        )

    return {"items": items, "count": len(items)}


@app.get("/api/v1/products/{product_id}")
def get_product(product_id: str) -> dict[str, object]:
    return _load_product_or_404(product_id)


@app.get("/api/v1/products/{product_id}/requirements")
def get_product_requirements(product_id: str) -> dict[str, object]:
    product = _load_product_or_404(product_id)
    requirements = product.get("requirements", [])

    if not isinstance(requirements, list):
        requirements = []

    requirement_items = [str(item) for item in requirements]

    return {
        "product_id": product_id,
        "items": requirement_items,
        "count": len(requirement_items),
    }
