from pathlib import Path

from fastapi.testclient import TestClient

from backend.app.main import app


def _write_yaml(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")


def test_blueprints_and_instances_api(monkeypatch, tmp_path):
    blueprints_dir = tmp_path / "blueprints"
    instances_dir = tmp_path / "instances"
    monkeypatch.setenv("NOMOS_BLUEPRINTS_DIR", str(blueprints_dir))
    monkeypatch.setenv("NOMOS_INSTANCES_DIR", str(instances_dir))

    _write_yaml(
        blueprints_dir / "products" / "account.yaml",
        """
id: PB-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
""",
    )
    _write_yaml(
        instances_dir / "products" / "account-instance.yaml",
        """
id: PI-ACC-MBX-EXAMPLE-001
type: product_instance
name: Beispielinstanz Benutzerkonto mit Mailbox
blueprint_ref: PB-ACC-MBX-001
blueprint_version: 0.1.0
compliance_status: compliant
evidence: []
findings: []
""",
    )

    client = TestClient(app)

    blueprints = client.get("/api/v1/blueprints")
    assert blueprints.status_code == 200
    assert blueprints.json()["count"] == 1
    assert blueprints.json()["items"][0]["id"] == "PB-ACC-MBX-001"

    blueprint = client.get("/api/v1/blueprints/PB-ACC-MBX-001")
    assert blueprint.status_code == 200
    assert blueprint.json()["type"] == "product_blueprint"

    instances = client.get("/api/v1/instances")
    assert instances.status_code == 200
    assert instances.json()["count"] == 1

    compliance = client.get("/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/compliance")
    assert compliance.status_code == 200
    assert compliance.json()["status"] == "compliant"
