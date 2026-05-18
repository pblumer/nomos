import base64
import hashlib
import hmac
import os
import time

from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.responses import JSONResponse

DEFAULT_USERNAME = "nomos"
DEFAULT_TOKEN_TTL_SECONDS = 12 * 60 * 60

PUBLIC_PATHS = {
    "/health",
    "/api/v1/auth/login",
    "/api/v1/auth/status",
    "/docs",
    "/redoc",
    "/openapi.json",
}


def _password() -> str:
    return os.getenv("NOMOS_AUTH_PASSWORD", "").strip()


def _username() -> str:
    return os.getenv("NOMOS_AUTH_USERNAME", "").strip() or DEFAULT_USERNAME


def _signing_key() -> bytes:
    explicit = os.getenv("NOMOS_AUTH_SECRET", "").strip()
    if explicit:
        return explicit.encode("utf-8")
    password = _password()
    if password:
        return hashlib.sha256(password.encode("utf-8")).digest()
    return b""


def auth_required() -> bool:
    return _password() != ""


def _b64u_encode(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).decode("ascii").rstrip("=")


def _b64u_decode(value: str) -> bytes:
    padding = "=" * (-len(value) % 4)
    return base64.urlsafe_b64decode(value + padding)


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
    return hmac.compare_digest(username, _username())


def verify_credentials(username: str, password: str) -> bool:
    expected_password = _password()
    if not expected_password:
        return False
    user_ok = hmac.compare_digest(username.encode("utf-8"), _username().encode("utf-8"))
    pwd_ok = hmac.compare_digest(password.encode("utf-8"), expected_password.encode("utf-8"))
    return user_ok and pwd_ok


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
