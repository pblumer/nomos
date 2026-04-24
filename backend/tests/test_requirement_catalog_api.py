import os
from pathlib import Path

import pytest
from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


@pytest.fixture(autouse=True)
def clean_catalog(tmp_path, monkeypatch):
    products_dir = tmp_path / "products"
    products_dir.mkdir(parents=True, exist_ok=True)
    requirements_dir = tmp_path / "requirements"
    requirements_dir.mkdir(parents=True, exist_ok=True)
    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(products_dir))
    monkeypatch.setenv("NOMOS_REQUIREMENTS_DIR", str(requirements_dir))
    yield


def test_list_requirements_empty():
    response = client.get("/api/v1/requirements")
    assert response.status_code == 200
    data = response.json()
    assert data["items"] == []
    assert data["count"] == 0


def test_get_requirement_by_id():
    req_dir = Path(os.environ["NOMOS_REQUIREMENTS_DIR"])
    req_path = req_dir / "req-mailbox-enabled.yml"
    req_path.write_text("id: req-mailbox-enabled\nname: Mailbox enabled\ndescription: User must have a mailbox.\n")

    response = client.get("/api/v1/requirements/req-mailbox-enabled")
    assert response.status_code == 200
    data = response.json()
    assert data["id"] == "req-mailbox-enabled"
    assert data["name"] == "Mailbox enabled"


def test_get_requirement_not_found():
    response = client.get("/api/v1/requirements/missing")
    assert response.status_code == 404


def test_update_requirement():
    req_dir = Path(os.environ["NOMOS_REQUIREMENTS_DIR"])
    req_path = req_dir / "req-login-required.yml"
    req_path.write_text("id: req-login-required\nname: Login required\n")

    response = client.put(
        "/api/v1/requirements/req-login-required",
        json={"name": "Login mandatory", "description": "User needs login"},
    )
    assert response.status_code == 200
    data = response.json()
    assert data["name"] == "Login mandatory"
    assert data["description"] == "User needs login"

    response = client.get("/api/v1/requirements/req-login-required")
    assert response.json()["name"] == "Login mandatory"


def test_update_requirement_rejects_empty_name():
    req_dir = Path(os.environ["NOMOS_REQUIREMENTS_DIR"])
    req_path = req_dir / "req-a.yml"
    req_path.write_text("id: req-a\nname: Req A\n")

    response = client.put("/api/v1/requirements/req-a", json={"name": ""})
    assert response.status_code == 400
    assert "name" in response.json()["detail"].lower()


def test_update_requirement_rejects_invalid_priority():
    req_dir = Path(os.environ["NOMOS_REQUIREMENTS_DIR"])
    req_path = req_dir / "req-a.yml"
    req_path.write_text("id: req-a\nname: Req A\n")

    response = client.put("/api/v1/requirements/req-a", json={"name": "Req A", "priority": "urgent"})
    assert response.status_code == 400
    assert response.json()["detail"] == "Invalid priority: urgent"


def test_update_requirement_id_mismatch():
    req_dir = Path(os.environ["NOMOS_REQUIREMENTS_DIR"])
    req_path = req_dir / "req-a.yml"
    req_path.write_text("id: req-a\nname: Req A\n")

    response = client.put("/api/v1/requirements/req-a", json={"id": "req-b"})
    assert response.status_code == 400
    assert "match" in response.json()["detail"].lower()
