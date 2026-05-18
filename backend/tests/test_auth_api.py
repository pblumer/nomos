import importlib

import pytest
from fastapi.testclient import TestClient


def _reload_app():
    from backend.app import auth as auth_module
    from backend.app import main as main_module

    importlib.reload(auth_module)
    importlib.reload(main_module)
    return main_module.app


@pytest.fixture
def client_with_auth(monkeypatch):
    monkeypatch.setenv("NOMOS_AUTH_PASSWORD", "geheim")
    monkeypatch.setenv("NOMOS_AUTH_USERNAME", "kollege")
    monkeypatch.setenv("NOMOS_AUTH_SECRET", "test-secret-key")
    app = _reload_app()
    yield TestClient(app)


@pytest.fixture
def client_without_auth(monkeypatch):
    monkeypatch.delenv("NOMOS_AUTH_PASSWORD", raising=False)
    monkeypatch.delenv("NOMOS_AUTH_USERNAME", raising=False)
    monkeypatch.delenv("NOMOS_AUTH_SECRET", raising=False)
    app = _reload_app()
    yield TestClient(app)


def test_status_reports_required_when_password_set(client_with_auth):
    response = client_with_auth.get("/api/v1/auth/status")
    assert response.status_code == 200
    assert response.json() == {"auth_required": True}


def test_status_reports_not_required_when_password_unset(client_without_auth):
    response = client_without_auth.get("/api/v1/auth/status")
    assert response.status_code == 200
    assert response.json() == {"auth_required": False}


def test_health_is_public_even_with_auth_enabled(client_with_auth):
    response = client_with_auth.get("/health")
    assert response.status_code == 200


def test_protected_route_rejects_missing_token(client_with_auth):
    response = client_with_auth.get("/api/v1/products")
    assert response.status_code == 401


def test_protected_route_rejects_bad_token(client_with_auth):
    response = client_with_auth.get(
        "/api/v1/products",
        headers={"Authorization": "Bearer not-a-real-token"},
    )
    assert response.status_code == 401


def test_login_with_wrong_credentials_returns_401(client_with_auth):
    response = client_with_auth.post(
        "/api/v1/auth/login",
        json={"username": "kollege", "password": "falsch"},
    )
    assert response.status_code == 401


def test_login_returns_token_and_unlocks_protected_route(client_with_auth, tmp_path, monkeypatch):
    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(tmp_path))
    (tmp_path).mkdir(parents=True, exist_ok=True)

    login = client_with_auth.post(
        "/api/v1/auth/login",
        json={"username": "kollege", "password": "geheim"},
    )
    assert login.status_code == 200
    body = login.json()
    assert "token" in body and body["token"]
    assert body["username"] == "kollege"

    response = client_with_auth.get(
        "/api/v1/products",
        headers={"Authorization": f"Bearer {body['token']}"},
    )
    assert response.status_code == 200


def test_login_endpoint_disabled_without_password(client_without_auth):
    response = client_without_auth.post(
        "/api/v1/auth/login",
        json={"username": "x", "password": "y"},
    )
    assert response.status_code == 400


def test_existing_endpoints_unaffected_when_auth_disabled(client_without_auth):
    response = client_without_auth.get("/api/v1/products")
    assert response.status_code == 200
