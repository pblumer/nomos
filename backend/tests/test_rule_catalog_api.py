from pathlib import Path

import yaml
from fastapi.testclient import TestClient

from backend.app.main import app


def test_get_rule_by_id_returns_extended_fields(monkeypatch, tmp_path: Path) -> None:
    rules_dir = tmp_path / "rules"
    rules_dir.mkdir()

    (rules_dir / "rule-password-policy.yml").write_text(
        yaml.safe_dump(
            {
                "id": "rule-password-policy",
                "name": "Password policy",
                "description": "Password policy assignment is mandatory.",
                "category": "security",
                "severity": "high",
                "scope": "identity",
                "enforcement": "Account must be linked to standard-password-policy.",
            },
            sort_keys=False,
            allow_unicode=True,
        ),
        encoding="utf-8",
    )

    monkeypatch.setenv("NOMOS_RULES_DIR", str(rules_dir))
    client = TestClient(app)

    response = client.get("/api/v1/rules/rule-password-policy")

    assert response.status_code == 200
    payload = response.json()
    assert payload["id"] == "rule-password-policy"
    assert payload["name"] == "Password policy"
    assert payload["severity"] == "high"
    assert payload["category"] == "security"


def test_list_rules_returns_count_and_basic_metadata(monkeypatch, tmp_path: Path) -> None:
    rules_dir = tmp_path / "rules"
    rules_dir.mkdir()

    (rules_dir / "rule-1.yml").write_text("id: rule-1\nname: Rule One\nseverity: low\n", encoding="utf-8")
    (rules_dir / "rule-2.yml").write_text("id: rule-2\nname: Rule Two\nseverity: critical\n", encoding="utf-8")

    monkeypatch.setenv("NOMOS_RULES_DIR", str(rules_dir))
    client = TestClient(app)

    response = client.get("/api/v1/rules")

    assert response.status_code == 200
    payload = response.json()
    assert payload["count"] == 2
    ids = [item["id"] for item in payload["items"]]
    assert ids == ["rule-1", "rule-2"]
