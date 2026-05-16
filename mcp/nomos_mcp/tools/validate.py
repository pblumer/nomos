"""MCP Tool für Nomos Katalog-Validierung."""

import json
import os
import subprocess

from ..auth import AgentScope, require_scope

VALIDATE_TOOLS = [
    {
        "name": "validate_catalog",
        "description": (
            "Führt die deterministische Nomos-Validierung aus und gibt alle Findings zurück. "
            "Rufe dieses Tool nach jeder Änderung auf um sicherzustellen, dass keine Errors "
            "entstanden sind. Exit 0 = keine Errors (Warnings sind OK). "
            "Erfordert READ-Scope. "
            "[EN] Runs Nomos catalog validation and returns all findings. "
            "Call after every change to check for errors. Exit 0 = no errors (warnings OK)."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "path": {
                    "type": "string",
                    "description": (
                        "Pfad zum Cosmos-Repository (Standard: NOMOS_REPO_PATH oder '.')"
                    ),
                },
                "format": {
                    "type": "string",
                    "enum": ["text", "json"],
                    "description": "Ausgabeformat (Standard: json)",
                },
            },
            "required": [],
        },
    },
    {
        "name": "get_health",
        "description": (
            "Prüft ob der Nomos API-Server erreichbar ist. "
            "[EN] Checks if the Nomos API server is reachable."
        ),
        "inputSchema": {"type": "object", "properties": {}, "required": []},
    },
]


def _repo_root() -> str:
    return os.getenv("NOMOS_REPO_PATH", os.getcwd())


def _nomos_binary() -> str:
    return os.getenv("NOMOS_BINARY", "nomos")


@require_scope(AgentScope.READ)
async def validate_catalog_handler(path: str = "", fmt: str = "json") -> str:
    repo_path = path or _repo_root()
    binary = _nomos_binary()
    cmd = [binary, "validate", "--path", repo_path]
    if fmt == "json":
        cmd += ["--format", "json"]

    result = subprocess.run(cmd, capture_output=True, text=True)
    output = result.stdout or result.stderr or "(kein Output)"

    if fmt == "json":
        try:
            parsed = json.loads(output)
            findings = parsed if isinstance(parsed, list) else parsed.get("findings", parsed)
            errors = [f for f in (findings if isinstance(findings, list) else []) if f.get("severity") == "error"]
            return json.dumps({
                "exit_code": result.returncode,
                "has_errors": result.returncode != 0,
                "error_count": len(errors),
                "findings": findings,
            }, ensure_ascii=False, indent=2)
        except json.JSONDecodeError:
            pass

    status = "FEHLER" if result.returncode != 0 else "OK"
    return f"Validierung {status} (Exit {result.returncode}):\n{output}"


@require_scope(AgentScope.READ)
async def get_health_handler() -> str:
    import httpx
    api_base = os.getenv("NOMOS_API_BASE", "http://localhost:8080")
    try:
        async with httpx.AsyncClient(timeout=5.0) as client:
            resp = await client.get(f"{api_base}/health")
            return json.dumps({
                "status": "reachable",
                "http_status": resp.status_code,
                "body": resp.json() if resp.is_success else resp.text,
            }, ensure_ascii=False, indent=2)
    except Exception as exc:
        return json.dumps({
            "status": "unreachable",
            "error": str(exc),
            "api_base": api_base,
        }, ensure_ascii=False, indent=2)
