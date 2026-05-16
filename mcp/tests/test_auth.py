"""Tests für das Nomos MCP Scope/Berechtigungsmodell."""

import os
import pytest

from nomos_mcp.auth import AgentScope, check_scope, get_scope, require_scope


def _set_scope(scope: str):
    os.environ["NOMOS_AGENT_SCOPE"] = scope


def test_get_scope_read(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    assert get_scope() == AgentScope.READ


def test_get_scope_draft(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "draft")
    assert get_scope() == AgentScope.DRAFT


def test_get_scope_write(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "write")
    assert get_scope() == AgentScope.WRITE


def test_get_scope_default(monkeypatch):
    monkeypatch.delenv("NOMOS_AGENT_SCOPE", raising=False)
    assert get_scope() == AgentScope.READ


def test_get_scope_invalid(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "superadmin")
    with pytest.raises(ValueError, match="Ungültiger NOMOS_AGENT_SCOPE"):
        get_scope()


def test_get_scope_case_insensitive(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "DRAFT")
    assert get_scope() == AgentScope.DRAFT


def test_check_scope_read_allows_read(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    check_scope(AgentScope.READ)  # should not raise


def test_check_scope_read_blocks_draft(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    with pytest.raises(PermissionError, match="draft"):
        check_scope(AgentScope.DRAFT)


def test_check_scope_read_blocks_write(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")
    with pytest.raises(PermissionError, match="write"):
        check_scope(AgentScope.WRITE)


def test_check_scope_draft_allows_read(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "draft")
    check_scope(AgentScope.READ)  # should not raise


def test_check_scope_draft_allows_draft(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "draft")
    check_scope(AgentScope.DRAFT)  # should not raise


def test_check_scope_draft_blocks_write(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "draft")
    with pytest.raises(PermissionError, match="write"):
        check_scope(AgentScope.WRITE)


def test_check_scope_write_allows_all(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "write")
    check_scope(AgentScope.READ)
    check_scope(AgentScope.DRAFT)
    check_scope(AgentScope.WRITE)


@pytest.mark.asyncio
async def test_require_scope_decorator_blocks(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "read")

    @require_scope(AgentScope.DRAFT)
    async def protected_fn():
        return "should not reach"

    with pytest.raises(PermissionError):
        await protected_fn()


@pytest.mark.asyncio
async def test_require_scope_decorator_allows(monkeypatch):
    monkeypatch.setenv("NOMOS_AGENT_SCOPE", "draft")

    @require_scope(AgentScope.DRAFT)
    async def protected_fn():
        return "allowed"

    result = await protected_fn()
    assert result == "allowed"
