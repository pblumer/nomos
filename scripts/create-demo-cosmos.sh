#!/usr/bin/env sh
set -eu

# Erstellt einen vollständigen Demo-Cosmos für Nomos.
#
# Aufruf:
#   sh scripts/create-demo-cosmos.sh
#   sh scripts/create-demo-cosmos.sh /tmp/nomos-demo
#
# Optional:
#   NOMOS_BIN=./bin/nomos COSMOS_PATH=./tmp/demo-cosmos sh scripts/create-demo-cosmos.sh

NOMOS_BIN="${NOMOS_BIN:-./bin/nomos}"

# Priorität: 1. Argument, 2. COSMOS_PATH Environment Variable, 3. Default.
COSMOS_PATH="${1:-${COSMOS_PATH:-./tmp/demo-cosmos}}"

echo "==> Prüfe Nomos CLI"
if [ ! -x "$NOMOS_BIN" ]; then
  echo "ERROR: Nomos CLI nicht gefunden oder nicht ausführbar: $NOMOS_BIN"
  echo "Baue den CLI zuerst mit:"
  echo "  make build"
  echo "oder:"
  echo "  go build -o ./bin/nomos ./cmd/nomos"
  exit 1
fi

echo "==> Nomos Version"
"$NOMOS_BIN" version

echo ""
echo "==> Entferne bestehenden Demo-Cosmos: $COSMOS_PATH"
rm -rf "$COSMOS_PATH"

echo ""
echo "==> Erstelle neuen Demo-Cosmos"
"$NOMOS_BIN" cosmos init "$COSMOS_PATH" --git

echo ""
echo "==> Erstelle DNS-ähnliche Domänen"

"$NOMOS_BIN" domain add blumer.com \
  --path "$COSMOS_PATH" \
  --owner "Blumer Web Team"

"$NOMOS_BIN" domain add blumer.net \
  --path "$COSMOS_PATH" \
  --owner "Blumer Network Team"

"$NOMOS_BIN" domain add identity.blumer.com \
  --path "$COSMOS_PATH" \
  --owner "Identity Team"

"$NOMOS_BIN" domain add governance.blumer.com \
  --path "$COSMOS_PATH" \
  --owner "Governance Team"

"$NOMOS_BIN" domain add blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Cloud Team"

"$NOMOS_BIN" domain add home.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Home Team"

"$NOMOS_BIN" domain add zytlog.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Zytlog Team"

"$NOMOS_BIN" domain add beispiel.ch \
  --path "$COSMOS_PATH" \
  --owner "Swiss Example Team"

echo ""

"$NOMOS_BIN" domain add identity.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Identity Domain Team"

"$NOMOS_BIN" domain add collaboration.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Domain Team"

"$NOMOS_BIN" domain add mailing.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Mailing Domain Team"

echo "==> Erstelle Services"

"$NOMOS_BIN" service add user-account \
  --domain identity.blumer.com \
  --path "$COSMOS_PATH" \
  --owner "Identity Team"

"$NOMOS_BIN" service add provisioning-rules \
  --domain governance.blumer.com \
  --path "$COSMOS_PATH" \
  --owner "Governance Team"

"$NOMOS_BIN" service add home-dashboard \
  --domain home.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Home Team"

"$NOMOS_BIN" service add zytlog-api \
  --domain zytlog.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Zytlog Team"


"$NOMOS_BIN" service add user-account \
  --domain identity.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Identity Domain Team"

"$NOMOS_BIN" service add mailbox \
  --domain collaboration.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Domain Team"

"$NOMOS_BIN" service add exchange \
  --domain mailing.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Mailing Domain Team"

"$NOMOS_BIN" service add license-assignment \
  --domain collaboration.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Domain Team"

echo ""
echo "==> Kopiere Blueprint- und Instance-Beispielkatalog"
mkdir -p "$COSMOS_PATH/.nomos/catalog"
cp -R examples/demo-cosmos/.nomos/catalog/blueprints "$COSMOS_PATH/.nomos/catalog/"
cp -R examples/demo-cosmos/.nomos/catalog/instances "$COSMOS_PATH/.nomos/catalog/"

echo "==> Ergänze Service-Ownership-Metadaten"
cat > "$COSMOS_PATH/.nomos/domains/cloud/blumer/identity/services/user-account/service.yaml" <<'YAML'
id: service-user-account
type: service
name: user-account
version: 0.1.0
status: draft
owner: Identity Domain Team
owned_by: identity.blumer.cloud
operated_by:
  - identity.blumer.cloud
capabilities:
  - user-account-management
supported_products:
  - PROD-ACC-MBX-001
  - PROD-CLOUD-MAILBOX-001
summary: Domain-owned service capability for identity account management.
ola:
  name: Identity Account OLA
  target: 4h
  availability: business-hours
YAML
cat > "$COSMOS_PATH/.nomos/domains/cloud/blumer/collaboration/services/mailbox/service.yaml" <<'YAML'
id: service-mailbox
type: service
name: mailbox
version: 0.1.0
status: draft
owner: Collaboration Domain Team
owned_by: collaboration.blumer.cloud
operated_by:
  - collaboration.blumer.cloud
capabilities:
  - mailbox-provisioning
supported_products:
  - PROD-ACC-MBX-001
  - PROD-CLOUD-MAILBOX-001
summary: Domain-owned service capability for mailbox provisioning.
sla:
  name: Mailbox Provisioning SLA
  target: 8h
  availability: business-hours
YAML
cat > "$COSMOS_PATH/.nomos/domains/cloud/blumer/mailing/services/exchange/service.yaml" <<'YAML'
id: service-exchange
type: service
name: exchange
version: 0.1.0
status: draft
owner: Mailing Domain Team
owned_by: mailing.blumer.cloud
operated_by:
  - mailing.blumer.cloud
capabilities:
  - exchange-mailbox-provisioning
supported_products:
  - PROD-ACC-MBX-001
summary: Domain-owned service capability for Exchange mailbox provisioning.
sla:
  name: Exchange Mailbox SLA
  target: 8h
  availability: business-hours
YAML
cat > "$COSMOS_PATH/.nomos/domains/cloud/blumer/collaboration/services/license-assignment/service.yaml" <<'YAML'
id: service-license-assignment
type: service
name: license-assignment
version: 0.1.0
status: draft
owner: Collaboration Domain Team
owned_by: collaboration.blumer.cloud
operated_by:
  - collaboration.blumer.cloud
capabilities:
  - license-assignment
supported_products:
  - PROD-ACC-MBX-001
summary: Domain-owned service capability for license assignment.
ola:
  name: License Assignment OLA
  target: 4h
  availability: business-hours
YAML

# Service alias used in the product process mapping example.
"$NOMOS_BIN" domain add account.blumer.cloud --path "$COSMOS_PATH" --owner "Account Domain Team"
"$NOMOS_BIN" service add license --domain account.blumer.cloud --path "$COSMOS_PATH" --owner "Account Domain Team"
cat > "$COSMOS_PATH/.nomos/domains/cloud/blumer/account/services/license/service.yaml" <<'YAML'
id: service-license
type: service
name: license
version: 0.1.0
status: draft
owner: Account Domain Team
owned_by: account.blumer.cloud
operated_by:
  - account.blumer.cloud
capabilities:
  - license-assignment
supported_products:
  - PROD-ACC-MBX-001
summary: Domain-owned service capability for account license assignment.
ola:
  name: Account License OLA
  target: 2h
  availability: business-hours
YAML

python3 - <<'PYDEMO'
from pathlib import Path
import os
product = Path(os.environ["COSMOS_PATH"]) / ".nomos/catalog/blueprints/products/benutzerkonto-mit-mailbox.yaml"
text = product.read_text()
if "processes:" not in text:
    text = text.replace("rules:\n", "processes:\n  - PRC-ACC-MBX-001\nrules:\n")
if "account.blumer.cloud/license" not in text:
    old = """    - service_ref: mailing.blumer.cloud/exchange
      role: primary
      required: true
      description: Provides the mailbox capability.
      sla_ref: SLA-MAILBOX-STANDARD
      ola_ref: OLA-MAILING-OPS-STANDARD
"""
    new = old + """    - service_ref: account.blumer.cloud/license
      role: supporting
      required: true
      description: Assigns the required product license.
      sla:
        target: 4h
        availability: 99.5%
        support_window: business_hours
      ola:
        owner: Account Operations
        target: 2h
"""
    text = text.replace(old, new)
if "purpose:" not in text:
    old = """summary: >
  Bereitstellung eines Benutzerkontos mit zugehoeriger Mailbox.
"""
    new = old + """purpose: Schneller, nachvollziehbarer Onboarding-Baustein fuer Mitarbeitende.
description: Das Angebot kombiniert Identitaet, Lizenz und Mailbox in einem kontrollierten Fulfillment-Prozess.
consumers:
  - Employees
  - Service desk
lifecycle_status: draft
tags:
  - identity
  - mailbox
  - provisioning
"""
    text = text.replace(old, new)
product.write_text(text)
PYDEMO
mkdir -p "$COSMOS_PATH/.nomos/catalog/blueprints/processes"
cat > "$COSMOS_PATH/.nomos/catalog/blueprints/processes/PRC-ACC-MBX-001.yaml" <<'YAML'
id: PRC-ACC-MBX-001
type: process
name: Provision Benutzeraccount mit Mailbox
version: 0.1.0
status: draft
owner: blumer.cloud
summary: End-to-end fulfillment process for the product offering.
tags:
  - provisioning
  - product-offering
related_product: PROD-ACC-MBX-001
bpmn:
  file: PRC-ACC-MBX-001.bpmn
  process_id: Process_UserAccountMailboxProvisioning
  primary: true
task_mappings:
  - bpmn_element_id: Task_CreateUserAccount
    task_name: Create user account
    bpmn_element_type: bpmn:ServiceTask
    service_ref: identity.blumer.cloud/user-account
    role: primary
    required: true
  - bpmn_element_id: Task_AssignLicense
    task_name: Assign license
    bpmn_element_type: bpmn:ServiceTask
    service_ref: account.blumer.cloud/license
    role: supporting
    required: true
  - bpmn_element_id: Task_CreateMailbox
    task_name: Create mailbox
    bpmn_element_type: bpmn:ServiceTask
    service_ref: mailing.blumer.cloud/exchange
    role: primary
    required: true
YAML
cat > "$COSMOS_PATH/.nomos/catalog/blueprints/processes/PRC-ACC-MBX-001.bpmn" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" id="Definitions_UserAccountMailbox" targetNamespace="https://nomos.local/bpmn">
  <bpmn:process id="Process_UserAccountMailboxProvisioning" name="Provision Benutzeraccount mit Mailbox" isExecutable="false">
    <bpmn:startEvent id="StartEvent_Request" name="Request received" />
    <bpmn:userTask id="Task_ValidateRequest" name="Validate request" />
    <bpmn:serviceTask id="Task_CreateUserAccount" name="Create user account" />
    <bpmn:serviceTask id="Task_AssignLicense" name="Assign license" />
    <bpmn:serviceTask id="Task_CreateMailbox" name="Create mailbox" />
    <bpmn:manualTask id="Task_QualityCheck" name="Quality check" />
    <bpmn:task id="Task_DocumentEvidence" name="Document evidence" />
    <bpmn:endEvent id="EndEvent_Done" name="Fulfilled" />
  </bpmn:process>
  <bpmndi:BPMNDiagram id="BPMNDiagram_UserAccountMailbox"><bpmndi:BPMNPlane id="BPMNPlane_UserAccountMailbox" bpmnElement="Process_UserAccountMailboxProvisioning" /></bpmndi:BPMNDiagram>
</bpmn:definitions>
XML


echo ""
echo "==> Erstelle Commerce-Domain mit Services und Methoden"
"$NOMOS_BIN" domain add commerce.blumer.cloud --path "$COSMOS_PATH" --owner "Commerce Team"
"$NOMOS_BIN" service add checkout  --domain commerce.blumer.cloud --path "$COSMOS_PATH" --owner "Commerce Team"
"$NOMOS_BIN" service add inventory --domain commerce.blumer.cloud --path "$COSMOS_PATH" --owner "Commerce Team"
"$NOMOS_BIN" service add payment   --domain commerce.blumer.cloud --path "$COSMOS_PATH" --owner "Commerce Team"

cat > "$COSMOS_PATH/.nomos/domains/cloud/blumer/commerce/services/checkout/service.yaml" <<'YAML'
id: service-checkout
type: service
name: checkout
version: 0.1.0
status: draft
owner: Commerce Team
owned_by: commerce.blumer.cloud
operated_by:
  - commerce.blumer.cloud
methods:
  - name: initiateOrder
  - name: confirmOrder
  - name: cancelOrder
  - name: getOrderStatus
summary: Handles order creation and lifecycle management.
YAML

cat > "$COSMOS_PATH/.nomos/domains/cloud/blumer/commerce/services/inventory/service.yaml" <<'YAML'
id: service-inventory
type: service
name: inventory
version: 0.1.0
status: draft
owner: Commerce Team
owned_by: commerce.blumer.cloud
operated_by:
  - commerce.blumer.cloud
methods:
  - name: checkStock
  - name: reserveItems
  - name: releaseItems
  - name: getInventoryLevel
summary: Manages stock levels and item reservations.
YAML

cat > "$COSMOS_PATH/.nomos/domains/cloud/blumer/commerce/services/payment/service.yaml" <<'YAML'
id: service-payment
type: service
name: payment
version: 0.1.0
status: draft
owner: Commerce Team
owned_by: commerce.blumer.cloud
operated_by:
  - commerce.blumer.cloud
methods:
  - name: processPayment
  - name: refundPayment
  - name: getPaymentStatus
  - name: authorizePayment
summary: Processes payments and refunds for orders.
YAML

echo ""
echo "==> Cosmos Info"
"$NOMOS_BIN" cosmos info --path "$COSMOS_PATH"

echo ""
echo "==> Cosmos Doctor"
"$NOMOS_BIN" cosmos doctor --path "$COSMOS_PATH"

echo ""
echo "==> Domains"
"$NOMOS_BIN" domain list --path "$COSMOS_PATH"

echo ""
echo "==> Validierung"
"$NOMOS_BIN" validate --path "$COSMOS_PATH"

echo ""
echo "==> Dateien"
find "$COSMOS_PATH" \
  -path "$COSMOS_PATH/.git" -prune -o \
  -type f -print | sort

echo ""
echo "==> Git Status"
(
  cd "$COSMOS_PATH"
  git status --short
)

echo ""
echo "Demo-Cosmos wurde erfolgreich erstellt: $COSMOS_PATH"
echo "Start Web UI with:"
echo "  ./bin/nomos serve --path $COSMOS_PATH --listen 127.0.0.1:8080"
