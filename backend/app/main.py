import os
from pathlib import Path

import yaml
from fastapi import Body, FastAPI, HTTPException

app = FastAPI(title="Nomos API", version="0.1.0")


def _products_dir() -> Path:
    return Path(os.getenv("NOMOS_PRODUCTS_DIR", "catalog/products"))


def _read_yaml_file(file_path: Path) -> dict[str, object]:
    return yaml.safe_load(file_path.read_text(encoding="utf-8")) or {}


def _write_yaml_file(file_path: Path, data: dict[str, object]) -> None:
    file_path.write_text(
        yaml.safe_dump(data, sort_keys=False, allow_unicode=True),
        encoding="utf-8",
    )


def _product_files() -> list[Path]:
    return sorted([
        *_products_dir().glob("*.yml"),
        *_products_dir().glob("*.yaml"),
    ])


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


def _load_product_with_path_or_404(product_id: str) -> tuple[Path, dict[str, object]]:
    yml_path = _products_dir() / f"{product_id}.yml"
    yaml_path = _products_dir() / f"{product_id}.yaml"

    if yml_path.exists():
        product = _read_yaml_file(yml_path)
        product["validation"] = _validate_product(product)
        return yml_path, product

    if yaml_path.exists():
        product = _read_yaml_file(yaml_path)
        product["validation"] = _validate_product(product)
        return yaml_path, product

    raise HTTPException(status_code=404, detail="Produkt nicht gefunden")


def _load_product_or_404(product_id: str) -> dict[str, object]:
    _, product = _load_product_with_path_or_404(product_id)
    return product


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/api/v1/products")
def list_products() -> dict[str, object]:
    product_files = _product_files()

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


def _list_field(product: dict[str, object], field_name: str) -> list[str]:
    value = product.get(field_name, [])
    if not isinstance(value, list):
        return []
    return [str(item) for item in value]


@app.get("/api/v1/products/{product_id}/requirements")
def get_product_requirements(product_id: str) -> dict[str, object]:
    product = _load_product_or_404(product_id)
    requirement_items = _list_field(product, "requirements")

    return {
        "product_id": product_id,
        "items": requirement_items,
        "count": len(requirement_items),
    }


@app.get("/api/v1/products/{product_id}/rules")
def get_product_rules(product_id: str) -> dict[str, object]:
    product = _load_product_or_404(product_id)
    rule_items = _list_field(product, "rules")

    return {
        "product_id": product_id,
        "items": rule_items,
        "count": len(rule_items),
    }


@app.get("/api/v1/products/{product_id}/summary")
def get_product_summary(product_id: str) -> dict[str, object]:
    product = _load_product_or_404(product_id)
    validation = product.get("validation", {})

    if not isinstance(validation, dict):
        validation = {"is_valid": False, "errors": []}

    errors = validation.get("errors", [])
    if not isinstance(errors, list):
        errors = []

    return {
        "product_id": str(product.get("id", product_id)),
        "name": str(product.get("name", "")),
        "version": str(product.get("version", "")),
        "requirements_count": len(_list_field(product, "requirements")),
        "rules_count": len(_list_field(product, "rules")),
        "validation_is_valid": bool(validation.get("is_valid", False)),
        "validation_error_count": len(errors),
    }


@app.get("/api/v1/overview")
def get_overview() -> dict[str, int]:
    product_count = 0
    requirements_count = 0
    rules_count = 0
    invalid_products_count = 0

    for file_path in _product_files():
        product_count += 1
        product = _read_yaml_file(file_path)
        validation = _validate_product(product)
        requirements_count += len(_list_field(product, "requirements"))
        rules_count += len(_list_field(product, "rules"))
        if not validation["is_valid"]:
            invalid_products_count += 1

    return {
        "product_count": product_count,
        "requirements_count": requirements_count,
        "rules_count": rules_count,
        "invalid_products_count": invalid_products_count,
    }


@app.post("/api/v1/products/{product_id}/requirements", status_code=201)
def add_product_requirement(product_id: str, payload: dict[str, str] = Body(...)) -> dict[str, str]:
    file_path, product = _load_product_with_path_or_404(product_id)
    item = str(payload.get("value", "")).strip()
    if item == "":
        raise HTTPException(status_code=400, detail="value is required")

    items = _list_field(product, "requirements")
    if item not in items:
        items.append(item)

    product["requirements"] = items
    product["validation"] = _validate_product(product)
    _write_yaml_file(file_path, product)

    return {"item": item}


@app.delete("/api/v1/products/{product_id}/requirements/{item}")
def delete_product_requirement(product_id: str, item: str) -> dict[str, object]:
    file_path, product = _load_product_with_path_or_404(product_id)
    items = _list_field(product, "requirements")

    if item not in items:
        raise HTTPException(status_code=404, detail="item not found")

    items = [current for current in items if current != item]
    product["requirements"] = items
    product["validation"] = _validate_product(product)
    _write_yaml_file(file_path, product)

    return {"item": item, "removed": True}


@app.post("/api/v1/products/{product_id}/rules", status_code=201)
def add_product_rule(product_id: str, payload: dict[str, str] = Body(...)) -> dict[str, str]:
    file_path, product = _load_product_with_path_or_404(product_id)
    item = str(payload.get("value", "")).strip()
    if item == "":
        raise HTTPException(status_code=400, detail="value is required")

    items = _list_field(product, "rules")
    if item not in items:
        items.append(item)

    product["rules"] = items
    product["validation"] = _validate_product(product)
    _write_yaml_file(file_path, product)

    return {"item": item}


@app.delete("/api/v1/products/{product_id}/rules/{item}")
def delete_product_rule(product_id: str, item: str) -> dict[str, object]:
    file_path, product = _load_product_with_path_or_404(product_id)
    items = _list_field(product, "rules")

    if item not in items:
        raise HTTPException(status_code=404, detail="item not found")

    items = [current for current in items if current != item]
    product["rules"] = items
    product["validation"] = _validate_product(product)
    _write_yaml_file(file_path, product)

    return {"item": item, "removed": True}
