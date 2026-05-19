import base64
import hashlib
import importlib
import secrets

import pytest
from fastapi.testclient import TestClient


def _hash(password: str) -> str:
    salt = secrets.token_bytes(16)
    iterations = 600_000
    derived = hashlib.pbkdf2_hmac("sha256", password.encode("utf-8"), salt, iterations)
    return (
        f"pbkdf2_sha256${iterations}$"
        f"{base64.b64encode(salt).decode('ascii')}$"
        f"{base64.b64encode(derived).decode('ascii')}"
    )


def _reload_app():
    from backend.app import auth as auth_module
    from backend.app import main as main_module

    importlib.reload(auth_module)
    importlib.reload(main_module)
    return main_module.app


@pytest.fixture
def multi_user_client(monkeypatch, tmp_path):
    monkeypatch.delenv("NOMOS_AUTH_PASSWORD", raising=False)
    monkeypatch.delenv("NOMOS_AUTH_USERNAME", raising=False)
    users_file = tmp_path / "users.yaml"
    users_file.write_text(
        "users:\n"
        f"  - username: nomos\n    password_hash: \"{_hash('nomos-pw')}\"\n"
        f"  - username: sven\n    password_hash: \"{_hash('sven-pw')}\"\n",
        encoding="utf-8",
    )
    monkeypatch.setenv("NOMOS_AUTH_USERS_FILE", str(users_file))
    monkeypatch.setenv("NOMOS_AUTH_SECRET", "test-secret-multi")
    monkeypatch.setenv("NOMOS_PRODUCTS_DIR", str(tmp_path / "products"))
    (tmp_path / "products").mkdir()
    yield TestClient(_reload_app())


def test_status_requires_auth_when_users_file_present(multi_user_client):
    response = multi_user_client.get("/api/v1/auth/status")
    assert response.status_code == 200
    assert response.json() == {"auth_required": True}


def test_each_listed_user_can_log_in(multi_user_client):
    for username, password in [("nomos", "nomos-pw"), ("sven", "sven-pw")]:
        response = multi_user_client.post(
            "/api/v1/auth/login",
            json={"username": username, "password": password},
        )
        assert response.status_code == 200, (username, response.text)
        body = response.json()
        assert body["username"] == username
        assert body["token"]


def test_token_from_listed_user_unlocks_protected_route(multi_user_client):
    login = multi_user_client.post(
        "/api/v1/auth/login",
        json={"username": "sven", "password": "sven-pw"},
    )
    token = login.json()["token"]
    response = multi_user_client.get(
        "/api/v1/products",
        headers={"Authorization": f"Bearer {token}"},
    )
    assert response.status_code == 200


def test_unknown_user_is_rejected(multi_user_client):
    response = multi_user_client.post(
        "/api/v1/auth/login",
        json={"username": "ghost", "password": "whatever"},
    )
    assert response.status_code == 401


def test_wrong_password_for_known_user_is_rejected(multi_user_client):
    response = multi_user_client.post(
        "/api/v1/auth/login",
        json={"username": "sven", "password": "nomos-pw"},
    )
    assert response.status_code == 401
