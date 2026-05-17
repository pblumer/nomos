"""Tutorial-Szenario: commerce.blumer.cloud via Nomos MCP aufbauen.

Dieses Test-Szenario ist gleichzeitig Tutorial und Integrationstest.
Es zeigt den vollständigen Ablauf:
  1. Cosmos initialisieren
  2. Domain anlegen
  3. Services anlegen
  4. Produkt-Offering (Blueprint) anlegen und Services verknüpfen
  5. Prozess anlegen
  6. Decision anlegen (mit DMN-Logik)
  7. Prozessschritt verknüpfen (businessRuleTask → Decision → Gateway)

Voraussetzungen:
  - `nomos` Binary im PATH
  - `nomos serve` läuft auf NOMOS_GO_API_BASE (default: http://localhost:9090)
  - NOMOS_AGENT_SCOPE=draft

Die Tests mocken alle HTTP-Aufrufe mit respx und simulieren das Go-Backend.
Damit läuft der Test ohne laufenden Server.
"""

import json
import os
import subprocess
import sys
import tempfile
from unittest.mock import AsyncMock, patch

import pytest
import respx
from httpx import Response

# Scope auf DRAFT setzen damit schreibende Operationen erlaubt sind
os.environ.setdefault("NOMOS_AGENT_SCOPE", "draft")
os.environ.setdefault("NOMOS_GO_API_BASE", "http://localhost:9090")

from nomos_mcp.tools.cosmos import cosmos_init_handler, cosmos_info_handler
from nomos_mcp.tools.decisions import (
    create_decision_handler,
    get_decision_handler,
    list_decisions_handler,
    update_decision_dmn_handler,
)
from nomos_mcp.tools.domains import (
    create_domain_handler,
    get_domain_handler,
    list_domains_handler,
)
from nomos_mcp.tools.processes import (
    create_process_handler,
    get_process_handler,
    list_processes_handler,
)
from nomos_mcp.tools.services import (
    create_service_handler,
    get_service_handler,
    list_services_handler,
)

BASE = "http://localhost:9090"

# ---------------------------------------------------------------------------
# Schritt 1: Cosmos initialisieren
# ---------------------------------------------------------------------------

class TestCosmosInit:
    """Verifiziert dass cosmos_init den nomos CLI aufruft und ein sinnvolles Ergebnis liefert."""

    def test_cosmos_init_calls_nomos_cli(self, tmp_path):
        with patch("nomos_mcp.tools.cosmos._run_nomos", return_value=(0, "ok", "")) as mock_run:
            import asyncio
            result = asyncio.run(
                cosmos_init_handler(str(tmp_path), "commerce-cosmos", "Commerce Cosmos")
            )
        mock_run.assert_called_once_with("cosmos", "init", str(tmp_path))
        data = json.loads(result)
        assert data["id"] == "commerce-cosmos"
        assert data["action"] == "initialized"

    def test_cosmos_init_propagates_cli_error(self, tmp_path):
        with patch("nomos_mcp.tools.cosmos._run_nomos", return_value=(1, "", "permission denied")):
            import asyncio
            result = asyncio.run(
                cosmos_init_handler(str(tmp_path), "x", "X")
            )
        assert "Fehler" in result or "permission denied" in result


# ---------------------------------------------------------------------------
# Schritt 2: Domain anlegen
# ---------------------------------------------------------------------------

class TestDomainTools:
    """Erstellt die Domain commerce.blumer.cloud."""

    @respx.mock
    @pytest.mark.asyncio
    async def test_create_domain(self):
        respx.post(f"{BASE}/api/v1/domains").mock(
            return_value=Response(
                201,
                json={
                    "canonical": "commerce.blumer.cloud",
                    "name": "Commerce",
                    "owner": "commerce-team",
                    "serviceCount": 0,
                },
            )
        )
        result = await create_domain_handler("commerce.blumer.cloud", "commerce-team")
        data = json.loads(result)
        assert data["canonical"] == "commerce.blumer.cloud"

    @respx.mock
    @pytest.mark.asyncio
    async def test_list_domains(self):
        respx.get(f"{BASE}/api/v1/domains").mock(
            return_value=Response(
                200,
                json={"domains": [{"canonical": "commerce.blumer.cloud"}]},
            )
        )
        result = await list_domains_handler()
        data = json.loads(result)
        assert len(data["domains"]) == 1

    @respx.mock
    @pytest.mark.asyncio
    async def test_get_domain(self):
        respx.get(f"{BASE}/api/v1/domains/commerce.blumer.cloud").mock(
            return_value=Response(
                200,
                json={
                    "canonical": "commerce.blumer.cloud",
                    "serviceCount": 3,
                    "productCount": 1,
                },
            )
        )
        result = await get_domain_handler("commerce.blumer.cloud")
        data = json.loads(result)
        assert data["serviceCount"] == 3


# ---------------------------------------------------------------------------
# Schritt 3: Services anlegen (Checkout, Inventory, Payment)
# ---------------------------------------------------------------------------

class TestServiceTools:
    """Erstellt die drei Services des Commerce-Domains."""

    @respx.mock
    @pytest.mark.asyncio
    async def test_create_checkout_service(self):
        respx.post(f"{BASE}/api/v1/domains/commerce.blumer.cloud/services").mock(
            return_value=Response(
                201,
                json={
                    "name": "checkout",
                    "domain": "commerce.blumer.cloud",
                    "owner": "commerce-team",
                },
            )
        )
        result = await create_service_handler(
            domain="commerce.blumer.cloud",
            name="checkout",
            owner="commerce-team",
            summary="Verarbeitet Warenkörbe und leitet Bestellungen ein",
            capabilities=["cart-management", "order-initiation"],
        )
        data = json.loads(result)
        assert data["name"] == "checkout"

    @respx.mock
    @pytest.mark.asyncio
    async def test_create_payment_service(self):
        respx.post(f"{BASE}/api/v1/domains/commerce.blumer.cloud/services").mock(
            return_value=Response(
                201,
                json={"name": "payment", "domain": "commerce.blumer.cloud"},
            )
        )
        result = await create_service_handler(
            domain="commerce.blumer.cloud",
            name="payment",
            owner="finance-team",
            summary="Abwicklung von Zahlungen und Rückerstattungen",
            capabilities=["payment-processing", "refund"],
        )
        data = json.loads(result)
        assert data["name"] == "payment"

    @respx.mock
    @pytest.mark.asyncio
    async def test_list_services(self):
        respx.get(f"{BASE}/api/v1/domains/commerce.blumer.cloud/services").mock(
            return_value=Response(
                200,
                json={
                    "domain": "commerce.blumer.cloud",
                    "services": [
                        {"name": "checkout"},
                        {"name": "inventory"},
                        {"name": "payment"},
                    ],
                },
            )
        )
        result = await list_services_handler("commerce.blumer.cloud")
        data = json.loads(result)
        assert len(data["services"]) == 3


# ---------------------------------------------------------------------------
# Schritt 4: Prozess anlegen (Bestellabwicklung)
# ---------------------------------------------------------------------------

class TestProcessTools:
    """Erstellt den Bestellabwicklungs-Prozess für das Commerce-Produkt."""

    @respx.mock
    @pytest.mark.asyncio
    async def test_create_process(self):
        respx.post(f"{BASE}/api/v1/products/bp-order-fulfillment/processes").mock(
            return_value=Response(
                201,
                json={
                    "id": "proc-001",
                    "name": "Bestellabwicklung",
                    "status": "draft",
                    "related_product": "bp-order-fulfillment",
                },
            )
        )
        result = await create_process_handler(
            product_id="bp-order-fulfillment",
            name="Bestellabwicklung",
            summary="End-to-End Bestellprozess vom Warenkorb bis zur Lieferung",
        )
        data = json.loads(result)
        assert data["id"] == "proc-001"
        assert data["related_product"] == "bp-order-fulfillment"

    @respx.mock
    @pytest.mark.asyncio
    async def test_list_processes(self):
        respx.get(f"{BASE}/api/v1/products/bp-order-fulfillment/processes").mock(
            return_value=Response(
                200,
                json={"items": [{"id": "proc-001", "name": "Bestellabwicklung"}], "count": 1},
            )
        )
        result = await list_processes_handler("bp-order-fulfillment")
        data = json.loads(result)
        assert data["count"] == 1


# ---------------------------------------------------------------------------
# Schritt 5: Decision anlegen (Bestellung genehmigen)
# ---------------------------------------------------------------------------

class TestDecisionTools:
    """Erstellt die Decision 'Bestellung genehmigen' mit DMN-Logik."""

    @respx.mock
    @pytest.mark.asyncio
    async def test_create_decision(self):
        respx.post(f"{BASE}/api/v1/domains/commerce.blumer.cloud/decisions").mock(
            return_value=Response(
                201,
                json={
                    "id": "DEC-001",
                    "name": "Bestellung genehmigen",
                    "number": "DEC-001",
                    "status": "draft",
                    "has_dmn": False,
                    "inputs": [
                        {"name": "orderAmount", "type": "number"},
                        {"name": "customerTier", "type": "string"},
                    ],
                    "outputs": [
                        {"name": "approved", "type": "boolean"},
                        {"name": "approvalReason", "type": "string"},
                    ],
                },
            )
        )
        result = await create_decision_handler(
            domain="commerce.blumer.cloud",
            name="Bestellung genehmigen",
            id="DEC-001",
            number="DEC-001",
            owner="commerce-team",
            summary="Entscheidet ob eine Bestellung automatisch genehmigt wird",
            context=(
                "Bestellungen über einem Schwellenwert oder von unbekannten Kunden "
                "müssen manuell freigegeben werden."
            ),
            inputs=[
                {"name": "orderAmount", "type": "number", "description": "Bestellwert in CHF"},
                {"name": "customerTier", "type": "string", "description": "Kundenkategorie: standard|premium|vip"},
            ],
            outputs=[
                {"name": "approved", "type": "boolean", "description": "True = automatisch genehmigt"},
                {"name": "approvalReason", "type": "string", "description": "Begründung der Entscheidung"},
            ],
        )
        data = json.loads(result)
        assert data["id"] == "DEC-001"
        assert len(data["inputs"]) == 2
        assert len(data["outputs"]) == 2

    @respx.mock
    @pytest.mark.asyncio
    async def test_upload_dmn(self):
        """Lädt eine DMN-Entscheidungstabelle für die Decision hoch."""
        dmn_xml = _minimal_dmn("DEC-001", "Bestellung genehmigen")
        respx.put(
            f"{BASE}/api/v1/domains/commerce.blumer.cloud/decisions/DEC-001/dmn"
        ).mock(
            return_value=Response(
                200,
                json={"id": "DEC-001", "has_dmn": True, "dmn_file": "decision.dmn"},
            )
        )
        result = await update_decision_dmn_handler(
            domain="commerce.blumer.cloud",
            decision_id="DEC-001",
            dmn_xml=dmn_xml,
        )
        data = json.loads(result)
        assert data["has_dmn"] is True

    @respx.mock
    @pytest.mark.asyncio
    async def test_list_decisions(self):
        respx.get(f"{BASE}/api/v1/domains/commerce.blumer.cloud/decisions").mock(
            return_value=Response(
                200,
                json={
                    "domain": "commerce.blumer.cloud",
                    "items": [{"id": "DEC-001", "name": "Bestellung genehmigen", "has_dmn": True}],
                    "count": 1,
                },
            )
        )
        result = await list_decisions_handler("commerce.blumer.cloud")
        data = json.loads(result)
        assert data["count"] == 1
        assert data["items"][0]["has_dmn"] is True

    @respx.mock
    @pytest.mark.asyncio
    async def test_get_decision(self):
        respx.get(
            f"{BASE}/api/v1/domains/commerce.blumer.cloud/decisions/DEC-001"
        ).mock(
            return_value=Response(
                200,
                json={
                    "id": "DEC-001",
                    "name": "Bestellung genehmigen",
                    "has_dmn": True,
                    "inputs": [{"name": "orderAmount", "type": "number"}],
                    "outputs": [{"name": "approved", "type": "boolean"}],
                },
            )
        )
        result = await get_decision_handler("commerce.blumer.cloud", "DEC-001")
        data = json.loads(result)
        assert data["id"] == "DEC-001"
        assert data["has_dmn"] is True


# ---------------------------------------------------------------------------
# Vollständiges Tutorial-Szenario (end-to-end Sequenz dokumentiert)
# ---------------------------------------------------------------------------

class TestTutorialScenario:
    """
    Dokumentiert den vollständigen Ablauf: Cosmos → Domain → Services
    → Produkt → Prozess → Decision (mit DMN) → Verknüpfung.

    Diese Klasse zeigt wie ein KI-Agent das nomos MCP nutzt um einen
    vollständigen Commerce-Cosmos aufzubauen.
    """

    @respx.mock
    @pytest.mark.asyncio
    async def test_full_commerce_cosmos_flow(self):
        """
        Vollständiger Aufbau eines Commerce-Cosmos via MCP.

        Szenario: commerce.blumer.cloud
        - Domain: commerce.blumer.cloud
        - Services: checkout, inventory, payment
        - Produkt: Bestellabwicklung (bp-order-fulfillment)
        - Prozess: Bestellabwicklung (proc-001)
        - Decision: Bestellung genehmigen (DEC-001) mit DMN
        - Prozessschritte:
            1. serviceTask: checkout → initiateOrder
            2. serviceTask: inventory → reserveItems
            3. businessRuleTask → DEC-001 (approved?)
            4. [Gateway] → approved=true → serviceTask: payment → charge
                        → approved=false → userTask: manuelle Freigabe
        """

        # 1. Domain anlegen
        respx.post(f"{BASE}/api/v1/domains").mock(
            return_value=Response(201, json={"canonical": "commerce.blumer.cloud"})
        )
        domain_result = await create_domain_handler("commerce.blumer.cloud", "commerce-team")
        assert "commerce.blumer.cloud" in domain_result

        # 2. Services anlegen
        respx.post(f"{BASE}/api/v1/domains/commerce.blumer.cloud/services").mock(
            return_value=Response(201, json={"name": "checkout", "domain": "commerce.blumer.cloud"})
        )
        svc_result = await create_service_handler("commerce.blumer.cloud", "checkout", "commerce-team")
        assert "checkout" in svc_result

        # 3. Decision erstellen
        respx.post(f"{BASE}/api/v1/domains/commerce.blumer.cloud/decisions").mock(
            return_value=Response(
                201,
                json={
                    "id": "DEC-001",
                    "name": "Bestellung genehmigen",
                    "has_dmn": False,
                    "outputs": [{"name": "approved", "type": "boolean"}],
                },
            )
        )
        dec_result = await create_decision_handler(
            domain="commerce.blumer.cloud",
            name="Bestellung genehmigen",
            outputs=[{"name": "approved", "type": "boolean"}],
        )
        assert "DEC-001" in dec_result

        # 4. DMN hochladen
        respx.put(
            f"{BASE}/api/v1/domains/commerce.blumer.cloud/decisions/DEC-001/dmn"
        ).mock(return_value=Response(200, json={"id": "DEC-001", "has_dmn": True}))
        dmn_result = await update_decision_dmn_handler(
            "commerce.blumer.cloud", "DEC-001", _minimal_dmn("DEC-001", "Bestellung genehmigen")
        )
        assert "has_dmn" in dmn_result

        # 5. Prozess erstellen
        respx.post(f"{BASE}/api/v1/products/bp-order-fulfillment/processes").mock(
            return_value=Response(201, json={"id": "proc-001", "name": "Bestellabwicklung"})
        )
        proc_result = await create_process_handler(
            "bp-order-fulfillment", "Bestellabwicklung", "End-to-End Bestellprozess"
        )
        assert "proc-001" in proc_result

        # Szenario vollständig — dokumentiert den Aufbau-Flow


# ---------------------------------------------------------------------------
# Hilfsfunktionen
# ---------------------------------------------------------------------------

def _minimal_dmn(decision_id: str, decision_name: str) -> str:
    """Erzeugt ein minimales DMN 1.3 XML für Tests."""
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/"
             xmlns:dmndi="https://www.omg.org/spec/DMN/20191111/DMNDI/"
             id="definitions_{decision_id}"
             name="{decision_name}"
             namespace="http://camunda.org/schema/1.0/dmn">
  <decision id="{decision_id}" name="{decision_name}">
    <decisionTable id="decisionTable_{decision_id}" hitPolicy="FIRST">
      <input id="input_orderAmount" label="Bestellwert">
        <inputExpression id="inputExpression_orderAmount" typeRef="number">
          <text>orderAmount</text>
        </inputExpression>
      </input>
      <input id="input_customerTier" label="Kundenkategorie">
        <inputExpression id="inputExpression_customerTier" typeRef="string">
          <text>customerTier</text>
        </inputExpression>
      </input>
      <output id="output_approved" label="Genehmigt" name="approved" typeRef="boolean"/>
      <output id="output_reason" label="Begründung" name="approvalReason" typeRef="string"/>
      <rule id="rule_vip_always">
        <inputEntry id="ie1"><text></text></inputEntry>
        <inputEntry id="ie2"><text>"vip"</text></inputEntry>
        <outputEntry id="oe1"><text>true</text></outputEntry>
        <outputEntry id="oe2"><text>"VIP-Kunde: automatisch genehmigt"</text></outputEntry>
      </rule>
      <rule id="rule_small_order">
        <inputEntry id="ie3"><text>&lt; 1000</text></inputEntry>
        <inputEntry id="ie4"><text></text></inputEntry>
        <outputEntry id="oe3"><text>true</text></outputEntry>
        <outputEntry id="oe4"><text>"Kleinstbestellung: automatisch genehmigt"</text></outputEntry>
      </rule>
      <rule id="rule_large_order">
        <inputEntry id="ie5"><text>&gt;= 1000</text></inputEntry>
        <inputEntry id="ie6"><text></text></inputEntry>
        <outputEntry id="oe5"><text>false</text></outputEntry>
        <outputEntry id="oe6"><text>"Grossbestellung: manuelle Freigabe erforderlich"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>"""
