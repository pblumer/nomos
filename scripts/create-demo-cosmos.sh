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
echo "==> Erstelle Domänen"

"$NOMOS_BIN" domain add identity.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Identity Team"

"$NOMOS_BIN" domain add collaboration.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Team"

"$NOMOS_BIN" domain add governance.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Governance Team"

"$NOMOS_BIN" domain add platform.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Platform Team"

"$NOMOS_BIN" domain add assurance.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Assurance Team"

echo ""
echo "==> Erstelle Services"

# Identity Domain
"$NOMOS_BIN" service add user-account \
  --domain identity.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Identity Team"

"$NOMOS_BIN" service add privileged-account \
  --domain identity.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Identity Team"

"$NOMOS_BIN" service add external-user-account \
  --domain identity.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Identity Team"

# Collaboration Domain
"$NOMOS_BIN" service add mailbox \
  --domain collaboration.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Team"

"$NOMOS_BIN" service add license-assignment \
  --domain collaboration.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Team"

"$NOMOS_BIN" service add teams-workspace \
  --domain collaboration.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Team"

"$NOMOS_BIN" service add shared-mailbox \
  --domain collaboration.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Collaboration Team"

# Governance Domain
"$NOMOS_BIN" service add provisioning-rules \
  --domain governance.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Governance Team"

"$NOMOS_BIN" service add approval-policy \
  --domain governance.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Governance Team"

"$NOMOS_BIN" service add naming-policy \
  --domain governance.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Governance Team"

# Platform Domain
"$NOMOS_BIN" service add rule-validation-api \
  --domain platform.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Platform Team"

"$NOMOS_BIN" service add decision-api \
  --domain platform.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Platform Team"

"$NOMOS_BIN" service add skill-registry \
  --domain platform.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Platform Team"

# Assurance Domain
"$NOMOS_BIN" service add findings \
  --domain assurance.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Assurance Team"

"$NOMOS_BIN" service add evidence \
  --domain assurance.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Assurance Team"

"$NOMOS_BIN" service add audit-log \
  --domain assurance.blumer.cloud \
  --path "$COSMOS_PATH" \
  --owner "Assurance Team"

echo ""
echo "==> Kopiere Blueprint- und Instance-Beispielkatalog"
mkdir -p "$COSMOS_PATH/catalog"
cp -R catalog/blueprints "$COSMOS_PATH/catalog/"
cp -R catalog/instances "$COSMOS_PATH/catalog/"

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
