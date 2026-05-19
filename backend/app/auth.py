import base64
import binascii
import hashlib
import hmac
import os
import time
from pathlib import Path

import yaml
from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.responses import JSONResponse

DEFAULT_USERNAME = "nomos"
DEFAULT_TOKEN_TTL_SECONDS = 12 * 60 * 60
PBKDF2_ITERATIONS = 600_000

PUBLIC_PATHS = {
    "/health",
    "/api/v1/auth/login",
    "/api/v1/auth/status",
    "/docs",
    "/redoc",
    "/openapi.json",
}


def _users_file() -> Path | None:
    path = os.getenv("NOMOS_AUTH_USERS_FILE", "").strip()
    return Path(path) if path else None


def _load_users() -> dict[str, str]:
    """Return a mapping of username to stored password representation.

    File entries use ``pbkdf2_sha256$<iter>$<salt_b64>$<hash_b64>``.
    The env-var fallback stores the literal password under the ``plain:`` prefix.
    """
    path = _users_file()
    if path and path.exists():
        raw = yaml.safe_load(path.read_text(encoding="utf-8")) or {}
        users: dict[str, str] = {}
        for entry in raw.get("users", []) or []:
            name = str(entry.get("username", "")).strip()
            stored = str(entry.get("password_hash", "")).strip()
            if name and stored:
                users[name] = stored
        return users

    password = os.getenv("NOMOS_AUTH_PASSWORD", "").strip()
    if not password:
        return {}
    name = os.getenv("NOMOS_AUTH_USERNAME", "").strip() or DEFAULT_USERNAME
    return {name: f"plain:{password}"}


def auth_required() -> bool:
    return bool(_load_users())


def _b64u_encode(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).decode("ascii").rstrip("=")


def _b64u_decode(value: str) -> bytes:
    padding = "=" * (-len(value) % 4)
    return base64.urlsafe_b64decode(value + padding)


def _signing_key() -> bytes:
    explicit = os.getenv("NOMOS_AUTH_SECRET", "").strip()
    if explicit:
        return explicit.encode("utf-8")
    users = _load_users()
    if not users:
        return b""
    digest = hashlib.sha256()
    for name in sorted(users):
        digest.update(name.encode("utf-8"))
        digest.update(b":")
        digest.update(users[name].encode("utf-8"))
        digest.update(b"\n")
    return digest.digest()


def _sign(payload: str) -> str:
    digest = hmac.new(_signing_key(), payload.encode("utf-8"), hashlib.sha256).digest()
    return _b64u_encode(digest)


def issue_token(username: str, ttl_seconds: int = DEFAULT_TOKEN_TTL_SECONDS) -> tuple[str, int]:
    expires_at = int(time.time()) + ttl_seconds
    payload = f"{username}:{expires_at}"
    encoded = _b64u_encode(payload.encode("utf-8"))
    return f"{encoded}.{_sign(payload)}", expires_at


def verify_token(token: str) -> bool:
    if not token or "." not in token:
        return False
    encoded_payload, signature = token.split(".", 1)
    try:
        payload = _b64u_decode(encoded_payload).decode("utf-8")
        username, expires_str = payload.rsplit(":", 1)
        expires_at = int(expires_str)
    except (ValueError, UnicodeDecodeError):
        return False
    if not hmac.compare_digest(_sign(payload), signature):
        return False
    if expires_at < int(time.time()):
        return False
    return username in _load_users()


def _verify_password(password: str, stored: str) -> bool:
    if stored.startswith("plain:"):
        expected = stored[len("plain:"):]
        return hmac.compare_digest(password.encode("utf-8"), expected.encode("utf-8"))
    if stored.startswith("pbkdf2_sha256$"):
        try:
            _, iter_str, salt_b64, hash_b64 = stored.split("$", 3)
            iterations = int(iter_str)
            salt = base64.b64decode(salt_b64)
            expected_hash = base64.b64decode(hash_b64)
        except (ValueError, binascii.Error):
            return False
        derived = hashlib.pbkdf2_hmac(
            "sha256", password.encode("utf-8"), salt, iterations
        )
        return hmac.compare_digest(derived, expected_hash)
    return False


def verify_credentials(username: str, password: str) -> bool:
    users = _load_users()
    stored = users.get(username)
    if stored is None:
        return False
    return _verify_password(password, stored)


class AuthMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        if not auth_required():
            return await call_next(request)
        if request.method == "OPTIONS":
            return await call_next(request)
        if request.url.path in PUBLIC_PATHS:
            return await call_next(request)
        header = request.headers.get("authorization", "")
        if not header.lower().startswith("bearer "):
            return JSONResponse({"detail": "Not authenticated"}, status_code=401)
        if not verify_token(header[len("bearer "):].strip()):
            return JSONResponse({"detail": "Invalid or expired token"}, status_code=401)
        return await call_next(request)
