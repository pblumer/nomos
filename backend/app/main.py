import os
from pathlib import Path

import yaml
from fastapi import Body, FastAPI, HTTPException

app = FastAPI(title="Nomos API", version="0.1.0")


def _products_dir() -> Path:
    products_env = os.getenv("NOMOS_PRODUCTS_DIR")
    if products_env:
        return Path(products_env)
    return Path("catalog/products")


def _requirements_dir() -> Path:
    requirements_env = os.getenv("NOMOS_REQUIREMENTS_DIR")
    if requirements_env:
        return Path(requirements_env)
    return _products_dir().parent / "requirements"


def _rules_dir() -> Path:
    rules_env = os.getenv("NOMOS_RULES_DIR")
    if rules_env:
        return Path(rules_env)
    return _products_dir().parent / "rules"


def _read_yaml_file(file_path: Path) -> dict[str, object]:
    return yaml.safe_load(file_path.read_text(encoding="utf-8")) or {}


def _write_yaml_file(file_path: Path, data: dict[str, object]) -> None:
    file_path.parent.mkdir(parents=True, exist_ok=True)
    file_path.write_text(
        yaml.safe_dump(data, sort_keys=False, allow_unicode=True),
        encoding="utf-8",
    )


def _product_files() -> list[Path]:
    return sorted([
        *_products_dir().glob("*.yml"),
        *_products_dir().glob("*.yaml"),
    ])


def _artifact_file_path(base_dir: Path, item_id: str) -> Path:
    return base_dir / f"{item_id}.yml"


def _rule_files() -> list[Path]:
    return sorted([
        *_rules_dir().glob("*.yml"),
        *_rules_dir().glob("*.yaml"),
    ])


def _requirement_files() -> list[Path]:
    return sorted([
        *_requirements_dir().glob("*.yml"),
        *_requirements_dir().glob("*.yaml"),
    ])


def _load_rule_with_path_or_404(rule_id: str) -> tuple[Path, dict[str, object]]:
    yml_path = _rules_dir() / f"{rule_id}.yml"
    yaml_path = _rules_dir() / f"{rule_id}.yaml"

    if yml_path.exists():
        return yml_path, _read_yaml_file(yml_path)

    if yaml_path.exists():
        return yaml_path, _read_yaml_file(yaml_path)

    for file_path in _rule_files():
        data = _read_yaml_file(file_path)
        if str(data.get("id", "")).strip() == rule_id:
            return file_path, data

    raise HTTPException(status_code=404, detail="Rule not found")


def _load_rule_or_404(rule_id: str) -> dict[str, object]:
    _, rule = _load_rule_with_path_or_404(rule_id)
    return rule


def _load_requirement_with_path_or_404(req_id: str) -> tuple[Path, dict[str, object]]:
    yml_path = _requirements_dir() / f"{req_id}.yml"
    yaml_path = _requirements_dir() / f"{req_id}.yaml"

    if yml_path.exists():
        return yml_path, _read_yaml_file(yml_path)

    if yaml_path.exists():
        return yaml_path, _read_yaml_file(yaml_path)

    for file_path in _requirement_files():
        data = _read_yaml_file(file_path)
        if str(data.get("id", "")).strip() == req_id:
            return file_path, data

    raise HTTPException(status_code=404, detail="Requirement not found")


def _load_requirement_or_404(req_id: str) -> dict[str, object]:
    _, req = _load_requirement_with_path_or_404(req_id)
    return req


def _validate_requirement(data: dict[str, object]) -> dict[str, object]:
    errors: list[str] = []

    req_id = str(data.get("id", "")).strip()
    if req_id == "":
        errors.append("Missing required field: id")

    req_name = str(data.get("name", "")).strip()
    if req_name == "":
        errors.append("Missing required field: name")

    return {
        "is_valid": len(errors) == 0,
        "errors": errors,
    }


def _ensure_artifact_file(base_dir: Path, item_id: str, kind: str) -> None:
    file_path = _artifact_file_path(base_dir, item_id)
    if file_path.exists():
        return

    _write_yaml_file(
        file_path,
        {
            "id": item_id,
            "name": f"{kind} {item_id}",
        },
    )


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


def _validate_rule(data: dict[str, object]) -> dict[str, object]:
    errors: list[str] = []

    rule_id = str(data.get("id", "")).strip()
    if rule_id == "":
        errors.append("Missing required field: id")

    rule_name = str(data.get("name", "")).strip()
    if rule_name == "":
        errors.append("Missing required field: name")

    severity = str(data.get("severity", "")).strip()
    if severity != "":
        allowed_severity = {"low", "medium", "high", "critical"}
        if severity not in allowed_severity:
            errors.append(f"Invalid severity: {severity}")

    validation_value = data.get("validation")
    if validation_value is not None:
        if not isinstance(validation_value, dict):
            errors.append("validation must be an object")
        else:
            validation_type = str(validation_value.get("type", "")).strip()
            validation_field = str(validation_value.get("field", "")).strip()
            if validation_type == "":
                errors.append("Missing required field: validation.type")
            if validation_field == "":
                errors.append("Missing required field: validation.field")

            operator = validation_value.get("operator")
            if operator is not None and str(operator).strip() != "":
                allowed_operators = {"eq", "neq", "lt", "lte", "gt", "gte", "matches"}
                operator_str = str(operator).strip()
                if operator_str not in allowed_operators:
                    errors.append(f"Invalid validation.operator: {operator_str}")

    return {
        "is_valid": len(errors) == 0,
        "errors": errors,
    }


def _requirement_ids(product: dict[str, object]) -> list[str]:
    value = product.get("requirement_ids")
    if isinstance(value, list):
        return [str(item) for item in value]

    legacy = product.get("requirements")
    if isinstance(legacy, list):
        return [str(item) for item in legacy]

    return []


def _rule_ids(product: dict[str, object]) -> list[str]:
    value = product.get("rule_ids")
    if isinstance(value, list):
        return [str(item) for item in value]

    legacy = product.get("rules")
    if isinstance(legacy, list):
        return [str(item) for item in legacy]

    return []


def _enrich_product_for_response(product: dict[str, object]) -> dict[str, object]:
    enriched = dict(product)
    enriched["requirements"] = _requirement_ids(product)
    enriched["rules"] = _rule_ids(product)
    enriched["validation"] = _validate_product(product)
    return enriched


def _load_product_with_path_or_404(product_id: str) -> tuple[Path, dict[str, object]]:
    yml_path = _products_dir() / f"{product_id}.yml"
    yaml_path = _products_dir() / f"{product_id}.yaml"

    if yml_path.exists():
        return yml_path, _enrich_product_for_response(_read_yaml_file(yml_path))

    if yaml_path.exists():
        return yaml_path, _enrich_product_for_response(_read_yaml_file(yaml_path))

    for file_path in _product_files():
        data = _read_yaml_file(file_path)
        if str(data.get("id", "")).strip() == product_id:
            return file_path, _enrich_product_for_response(data)

    raise HTTPException(status_code=404, detail="Produkt nicht gefunden")


def _load_product_or_404(product_id: str) -> dict[str, object]:
    _, product = _load_product_with_path_or_404(product_id)
    return product


def _normalize_string_list(value: object) -> list[str]:
    if not isinstance(value, list):
        return []
    return [str(item).strip() for item in value if str(item).strip()]


def _prepare_product_for_write(product: dict[str, object]) -> dict[str, object]:
    prepared = dict(product)
    prepared.pop("requirements", None)
    prepared.pop("rules", None)
    prepared.pop("validation", None)

    requirement_ids = _normalize_string_list(prepared.get("requirement_ids", []))
    rule_ids = _normalize_string_list(prepared.get("rule_ids", []))

    prepared["requirement_ids"] = requirement_ids
    prepared["requirements"] = requirement_ids
    prepared["rule_ids"] = rule_ids
    prepared["rules"] = rule_ids
    return prepared


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/api/v1/products")
def list_products() -> dict[str, object]:
    items: list[dict[str, str]] = []
    for file_path in _product_files():
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
    requirement_items = _requirement_ids(product)
    return {
        "product_id": product_id,
        "items": requirement_items,
        "count": len(requirement_items),
    }


@app.get("/api/v1/products/{product_id}/rules")
def get_product_rules(product_id: str) -> dict[str, object]:
    product = _load_product_or_404(product_id)
    rule_items = _rule_ids(product)
    return {
        "product_id": product_id,
        "items": rule_items,
        "count": len(rule_items),
    }


@app.get("/api/v1/rules")
def list_rules() -> dict[str, object]:
    items: list[dict[str, object]] = []
    for file_path in _rule_files():
        data = _read_yaml_file(file_path)
        items.append(
            {
                "id": str(data.get("id", "")),
                "name": str(data.get("name", "")),
                "description": str(data.get("description", "")),
                "severity": str(data.get("severity", "")),
                "category": str(data.get("category", "")),
            }
        )

    return {"items": items, "count": len(items)}


@app.get("/api/v1/rules/{rule_id}")
def get_rule(rule_id: str) -> dict[str, object]:
    return _load_rule_or_404(rule_id)


@app.put("/api/v1/rules/{rule_id}")
def update_rule(rule_id: str, payload: dict[str, object] = Body(...)) -> dict[str, object]:
    file_path, current_rule = _load_rule_with_path_or_404(rule_id)

    if "id" in payload and str(payload.get("id", "")).strip() not in {"", rule_id}:
        raise HTTPException(status_code=400, detail="id in payload must match path")

    merged_rule = {**current_rule, **payload, "id": rule_id}
    validation = _validate_rule(merged_rule)
    if not validation["is_valid"]:
        raise HTTPException(status_code=400, detail=validation["errors"][0])

    _write_yaml_file(file_path, merged_rule)
    return merged_rule


@app.get("/api/v1/products/{product_id}/summary")
def get_product_summary(product_id: str) -> dict[str, object]:
    product = _load_product_or_404(product_id)
    validation = product.get("validation", {"is_valid": False, "errors": []})
    if not isinstance(validation, dict):
        validation = {"is_valid": False, "errors": []}

    errors = validation.get("errors", [])
    if not isinstance(errors, list):
        errors = []

    return {
        "product_id": str(product.get("id", product_id)),
        "name": str(product.get("name", "")),
        "version": str(product.get("version", "")),
        "requirements_count": len(_requirement_ids(product)),
        "rules_count": len(_rule_ids(product)),
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
        requirements_count += len(_requirement_ids(product))
        rules_count += len(_rule_ids(product))
        if not validation["is_valid"]:
            invalid_products_count += 1

    return {
        "product_count": product_count,
        "requirements_count": requirements_count,
        "rules_count": rules_count,
        "invalid_products_count": invalid_products_count,
    }


@app.post("/api/v1/products", status_code=201)
def create_product(payload: dict[str, object] = Body(...)) -> dict[str, object]:
    product = _prepare_product_for_write(payload)
    product_id = str(product.get("id", "")).strip()
    if product_id == "":
        raise HTTPException(status_code=400, detail="Missing required field: id")

    yml_path = _products_dir() / f"{product_id}.yml"
    yaml_path = _products_dir() / f"{product_id}.yaml"
    if yml_path.exists() or yaml_path.exists():
        raise HTTPException(status_code=409, detail="product already exists")

    product["id"] = product_id
    validation = _validate_product(product)
    if not validation["is_valid"]:
        raise HTTPException(status_code=400, detail=validation["errors"][0])

    _write_yaml_file(yml_path, product)
    return _enrich_product_for_response(product)


@app.put("/api/v1/products/{product_id}")
def update_product(product_id: str, payload: dict[str, object] = Body(...)) -> dict[str, object]:
    file_path, _ = _load_product_with_path_or_404(product_id)
    current_product = _read_yaml_file(file_path)

    if "id" in payload and str(payload.get("id", "")).strip() not in {"", product_id}:
        raise HTTPException(status_code=400, detail="id in payload must match path")

    merged_product = {**current_product, **payload, "id": product_id}
    merged_product = _prepare_product_for_write(merged_product)

    validation = _validate_product(merged_product)
    if not validation["is_valid"]:
        raise HTTPException(status_code=400, detail=validation["errors"][0])

    _write_yaml_file(file_path, merged_product)
    return _enrich_product_for_response(merged_product)


@app.delete("/api/v1/products/{product_id}")
def delete_product(product_id: str) -> dict[str, object]:
    file_path, _ = _load_product_with_path_or_404(product_id)
    file_path.unlink(missing_ok=False)
    return {"id": product_id, "removed": True}


@app.post("/api/v1/products/{product_id}/requirements", status_code=201)
def add_product_requirement(product_id: str, payload: dict[str, str] = Body(...)) -> dict[str, str]:
    file_path, product = _load_product_with_path_or_404(product_id)
    item = str(payload.get("value", "")).strip()
    if item == "":
        raise HTTPException(status_code=400, detail="value is required")

    items = _requirement_ids(product)
    if item in items:
        raise HTTPException(status_code=409, detail="item already exists")

    items.append(item)
    product["requirement_ids"] = items
    product["requirements"] = items
    product["validation"] = _validate_product(product)
    _write_yaml_file(file_path, product)
    _ensure_artifact_file(_requirements_dir(), item, "requirement")
    return {"item": item}


@app.delete("/api/v1/products/{product_id}/requirements/{item}")
def delete_product_requirement(product_id: str, item: str) -> dict[str, object]:
    file_path, product = _load_product_with_path_or_404(product_id)
    items = _requirement_ids(product)
    if item not in items:
        raise HTTPException(status_code=404, detail="item not found")

    items = [current for current in items if current != item]
    product["requirement_ids"] = items
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

    items = _rule_ids(product)
    if item in items:
        raise HTTPException(status_code=409, detail="item already exists")

    items.append(item)
    product["rule_ids"] = items
    product["rules"] = items
    product["validation"] = _validate_product(product)
    _write_yaml_file(file_path, product)
    _ensure_artifact_file(_rules_dir(), item, "rule")
    return {"item": item}


@app.delete("/api/v1/products/{product_id}/rules/{item}")
def delete_product_rule(product_id: str, item: str) -> dict[str, object]:
    file_path, product = _load_product_with_path_or_404(product_id)
    items = _rule_ids(product)
    if item not in items:
        raise HTTPException(status_code=404, detail="item not found")

    items = [current for current in items if current != item]
    product["rule_ids"] = items
    product["rules"] = items
    product["validation"] = _validate_product(product)
    _write_yaml_file(file_path, product)
    return {"item": item, "removed": True}


@app.get("/api/v1/requirements")
def list_requirements() -> dict[str, object]:
    items: list[dict[str, object]] = []
    for file_path in _requirement_files():
        data = _read_yaml_file(file_path)
        items.append(
            {
                "id": str(data.get("id", "")),
                "name": str(data.get("name", "")),
                "description": str(data.get("description", "")),
            }
        )

    return {"items": items, "count": len(items)}


@app.get("/api/v1/requirements/{req_id}")
def get_requirement(req_id: str) -> dict[str, object]:
    return _load_requirement_or_404(req_id)


@app.put("/api/v1/requirements/{req_id}")
def update_requirement(req_id: str, payload: dict[str, object] = Body(...)) -> dict[str, object]:
    file_path, current_req = _load_requirement_with_path_or_404(req_id)

    if "id" in payload and str(payload.get("id", "")).strip() not in {"", req_id}:
        raise HTTPException(status_code=400, detail="id in payload must match path")

    merged_req = {**current_req, **payload, "id": req_id}
    validation = _validate_requirement(merged_req)
    if not validation["is_valid"]:
        raise HTTPException(status_code=400, detail=validation["errors"][0])

    _write_yaml_file(file_path, merged_req)
    return merged_req
