"""Scope-basiertes Berechtigungsmodell für den Nomos MCP Server."""

import functools
import os
from enum import Enum


class AgentScope(str, Enum):
    READ = "read"
    DRAFT = "draft"
    WRITE = "write"
    # MERGE intentionally excluded — humans only


_SCOPE_LEVEL = {
    AgentScope.READ: 0,
    AgentScope.DRAFT: 1,
    AgentScope.WRITE: 2,
}


def get_scope() -> AgentScope:
    """Liest NOMOS_AGENT_SCOPE aus der Umgebung. Standardwert: read."""
    raw = os.getenv("NOMOS_AGENT_SCOPE", "read").lower().strip()
    try:
        return AgentScope(raw)
    except ValueError:
        allowed = ", ".join(s.value for s in AgentScope)
        raise ValueError(
            f"Ungültiger NOMOS_AGENT_SCOPE '{raw}'. Erlaubte Werte: {allowed}"
        )


def check_scope(needed: AgentScope) -> None:
    """Prüft ob der aktuelle Scope ausreicht. Wirft PermissionError wenn nicht."""
    current = get_scope()
    if _SCOPE_LEVEL[current] < _SCOPE_LEVEL[needed]:
        raise PermissionError(
            f"Scope '{current.value}' reicht nicht aus. "
            f"Dieses Tool erfordert mindestens scope '{needed.value}'."
        )


def require_scope(needed: AgentScope):
    """Decorator: wirft PermissionError wenn der aktuelle Scope nicht ausreicht."""
    def decorator(fn):
        @functools.wraps(fn)
        async def wrapper(*args, **kwargs):
            check_scope(needed)
            return await fn(*args, **kwargs)
        return wrapper
    return decorator
