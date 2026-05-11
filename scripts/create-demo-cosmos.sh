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

"$NOMOS_BIN" domain add cloud.blumer.identity \
  --path "$COSMOS_PATH" \
  --owner "Identity Domain Team"

"$NOMOS_BIN" domain add cloud.blumer.collaboration \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Domain Team"

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
  --domain cloud.blumer.identity \
  --path "$COSMOS_PATH" \
  --owner "Identity Domain Team"

"$NOMOS_BIN" service add mailbox \
  --domain cloud.blumer.collaboration \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Domain Team"

"$NOMOS_BIN" service add license-assignment \
  --domain cloud.blumer.collaboration \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Domain Team"

echo ""
echo "==> Kopiere Blueprint- und Instance-Beispielkatalog"
mkdir -p "$COSMOS_PATH/.nomos/catalog"
cp -R examples/demo-cosmos/.nomos/catalog/blueprints "$COSMOS_PATH/.nomos/catalog/"
cp -R examples/demo-cosmos/.nomos/catalog/instances "$COSMOS_PATH/.nomos/catalog/"

echo "==> Ergänze Service-Ownership-Metadaten"
cat > "$COSMOS_PATH/.nomos/domains/identity/blumer/cloud/services/user-account/service.yaml" <<'YAML'
id: service-user-account
type: service
name: user-account
version: 0.1.0
status: draft
owner: Identity Domain Team
owned_by: cloud.blumer.identity
operated_by:
  - cloud.blumer.identity
capabilities:
  - user-account-management
supported_products:
  - PROD-ACC-MBX-001
summary: Domain-owned service capability for identity account management.
YAML
cat > "$COSMOS_PATH/.nomos/domains/collaboration/blumer/cloud/services/mailbox/service.yaml" <<'YAML'
id: service-mailbox
type: service
name: mailbox
version: 0.1.0
status: draft
owner: Collaboration Domain Team
owned_by: cloud.blumer.collaboration
operated_by:
  - cloud.blumer.collaboration
capabilities:
  - mailbox-provisioning
supported_products:
  - PROD-ACC-MBX-001
summary: Domain-owned service capability for mailbox provisioning.
YAML
cat > "$COSMOS_PATH/.nomos/domains/collaboration/blumer/cloud/services/license-assignment/service.yaml" <<'YAML'
id: service-license-assignment
type: service
name: license-assignment
version: 0.1.0
status: draft
owner: Collaboration Domain Team
owned_by: cloud.blumer.collaboration
operated_by:
  - cloud.blumer.collaboration
capabilities:
  - license-assignment
supported_products:
  - PROD-ACC-MBX-001
summary: Domain-owned service capability for license assignment.
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
