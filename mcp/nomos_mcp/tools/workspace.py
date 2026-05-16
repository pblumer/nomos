"""MCP Tools für Branch/Workspace-Management.

Das Git-First-Prinzip von Nomos fordert, dass alle Änderungen auf isolierten
Branches erfolgen. Dieser Modul stellt Tools bereit um Workspaces (Git-Branches)
zu erstellen und zu verwalten.

TODO: Sobald /api/v1/workspaces im Backend implementiert ist, die direkten
git-Aufrufe durch REST-API-Aufrufe ersetzen.
"""

import json
import os
import subprocess
from datetime import datetime

from ..auth import AgentScope, require_scope

WORKSPACE_TOOLS = [
    {
        "name": "create_workspace",
        "description": (
            "Erstellt einen neuen Git-Branch als isolierten Arbeitsbereich für den Agenten. "
            "Alle Änderungen sollen auf diesem Branch erfolgen. "
            "Merge in main erfolgt ausschliesslich durch Menschen via Pull Request. "
            "Erfordert DRAFT-Scope. "
            "[EN] Creates a new Git branch as an isolated workspace. "
            "Merges to main are humans-only via PR. Requires DRAFT scope."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "branch_name": {
                    "type": "string",
                    "description": (
                        "Name des Branches, z.B. 'agent/add-gdpr-rules'. "
                        "Darf nicht 'main' sein."
                    ),
                },
                "description": {
                    "type": "string",
                    "description": "Kurzbeschreibung was auf diesem Branch geändert wird",
                },
            },
            "required": ["branch_name"],
        },
    },
    {
        "name": "get_current_workspace",
        "description": (
            "Gibt den aktuellen Git-Branch (Workspace) zurück. "
            "[EN] Returns the current Git branch (workspace)."
        ),
        "inputSchema": {"type": "object", "properties": {}, "required": []},
    },
    {
        "name": "list_workspaces",
        "description": (
            "Listet alle lokalen Git-Branches auf. "
            "[EN] Lists all local Git branches."
        ),
        "inputSchema": {"type": "object", "properties": {}, "required": []},
    },
]


def _repo_root() -> str:
    return os.getenv("NOMOS_REPO_PATH", os.getcwd())


def _run_git(*args: str) -> tuple[int, str, str]:
    result = subprocess.run(
        ["git", *args],
        cwd=_repo_root(),
        capture_output=True,
        text=True,
    )
    return result.returncode, result.stdout.strip(), result.stderr.strip()


@require_scope(AgentScope.DRAFT)
async def create_workspace_handler(branch_name: str, description: str = "") -> str:
    if branch_name in ("main", "master"):
        raise PermissionError(
            f"Direktes Schreiben auf '{branch_name}' ist nicht erlaubt. "
            "Wähle einen anderen Branch-Namen für deinen Workspace."
        )

    # Validate branch name characters
    forbidden = [" ", "..", "~", "^", ":", "?", "*", "[", "\\", "@{"]
    for char in forbidden:
        if char in branch_name:
            return (
                f"Ungültiger Branch-Name '{branch_name}': "
                f"Das Zeichen '{char}' ist nicht erlaubt."
            )

    code, out, err = _run_git("checkout", "-b", branch_name)
    if code != 0:
        if "already exists" in err:
            # Branch exists — switch to it
            code2, out2, err2 = _run_git("checkout", branch_name)
            if code2 != 0:
                return f"Fehler beim Wechseln auf Branch '{branch_name}': {err2}"
            return json.dumps({
                "branch": branch_name,
                "action": "switched",
                "description": description,
                "created_at": datetime.utcnow().isoformat() + "Z",
            }, ensure_ascii=False, indent=2)
        return f"Fehler beim Erstellen des Branches '{branch_name}': {err}"

    return json.dumps({
        "branch": branch_name,
        "action": "created",
        "description": description,
        "created_at": datetime.utcnow().isoformat() + "Z",
        "note": (
            "Merge in main erfolgt ausschliesslich durch Menschen via Pull Request. "
            "Nutze diesen Branch für alle deine Änderungen."
        ),
    }, ensure_ascii=False, indent=2)


async def get_current_workspace_handler() -> str:
    code, out, err = _run_git("rev-parse", "--abbrev-ref", "HEAD")
    if code != 0:
        return f"Fehler beim Lesen des aktuellen Branches: {err}"
    branch = out
    _, commit_out, _ = _run_git("rev-parse", "--short", "HEAD")
    return json.dumps({
        "branch": branch,
        "commit": commit_out,
        "is_main": branch in ("main", "master"),
    }, ensure_ascii=False, indent=2)


async def list_workspaces_handler() -> str:
    code, out, err = _run_git("branch", "--list")
    if code != 0:
        return f"Fehler beim Auflisten der Branches: {err}"
    branches = [b.strip().lstrip("* ") for b in out.splitlines() if b.strip()]
    return json.dumps({"branches": branches, "count": len(branches)}, ensure_ascii=False, indent=2)
