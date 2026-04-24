import pytest
from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


@pytest.fixture(autouse=True)
def clean_servers(tmp_path, monkeypatch):
    servers_dir = tmp_path / "servers"
    servers_dir.mkdir(parents=True, exist_ok=True)
    monkeypatch.setenv("NOMOS_SERVERS_DIR", str(servers_dir))
    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(tmp_path / "products"))
    (tmp_path / "products").mkdir(parents=True, exist_ok=True)
    yield


def test_list_servers_empty():
    response = client.get("/api/v1/servers")
    assert response.status_code == 200
    data = response.json()
    assert data["items"] == []
    assert data["count"] == 0


def test_create_and_get_server():
    payload = {"id": "srv-web-01", "name": "Web Server 01", "description": "Main web server"}
    response = client.post("/api/v1/servers", json=payload)
    assert response.status_code == 201
    data = response.json()
    assert data["id"] == "srv-web-01"
    assert data["name"] == "Web Server 01"

    response = client.get("/api/v1/servers/srv-web-01")
    assert response.status_code == 200
    data = response.json()
    assert data["name"] == "Web Server 01"


def test_create_server_missing_id():
    response = client.post("/api/v1/servers", json={"name": "No ID"})
    assert response.status_code == 400
    assert "id" in response.json()["detail"].lower()


def test_create_server_missing_name():
    response = client.post("/api/v1/servers", json={"id": "srv-01"})
    assert response.status_code == 400
    assert "name" in response.json()["detail"].lower()


def test_create_server_conflict():
    client.post("/api/v1/servers", json={"id": "srv-01", "name": "Server 1"})
    response = client.post("/api/v1/servers", json={"id": "srv-01", "name": "Server 1 again"})
    assert response.status_code == 409


def test_update_server():
    client.post("/api/v1/servers", json={"id": "srv-01", "name": "Old name"})
    response = client.put("/api/v1/servers/srv-01", json={"name": "New name"})
    assert response.status_code == 200
    assert response.json()["name"] == "New name"

    response = client.get("/api/v1/servers/srv-01")
    assert response.json()["name"] == "New name"


def test_update_server_id_mismatch():
    client.post("/api/v1/servers", json={"id": "srv-01", "name": "Server 1"})
    response = client.put("/api/v1/servers/srv-01", json={"id": "srv-02"})
    assert response.status_code == 400
    assert "match" in response.json()["detail"].lower()


def test_update_server_not_found():
    response = client.put("/api/v1/servers/missing", json={"name": "x"})
    assert response.status_code == 404


def test_delete_server():
    client.post("/api/v1/servers", json={"id": "srv-01", "name": "Server 1"})
    response = client.delete("/api/v1/servers/srv-01")
    assert response.status_code == 200
    assert response.json()["removed"] is True

    response = client.get("/api/v1/servers/srv-01")
    assert response.status_code == 404


def test_delete_server_not_found():
    response = client.delete("/api/v1/servers/missing")
    assert response.status_code == 404


def test_list_servers():
    client.post("/api/v1/servers", json={"id": "srv-a", "name": "A"})
    client.post("/api/v1/servers", json={"id": "srv-b", "name": "B"})
    response = client.get("/api/v1/servers")
    assert response.status_code == 200
    data = response.json()
    assert data["count"] == 2
    ids = [item["id"] for item in data["items"]]
    assert "srv-a" in ids
    assert "srv-b" in ids
