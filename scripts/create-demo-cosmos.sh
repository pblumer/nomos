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

echo ""
echo "==> Kopiere Blueprint- und Instance-Beispielkatalog"
mkdir -p "$COSMOS_PATH/.nomos/catalog"
cp -R examples/demo-cosmos/.nomos/catalog/blueprints "$COSMOS_PATH/.nomos/catalog/"
cp -R examples/demo-cosmos/.nomos/catalog/instances "$COSMOS_PATH/.nomos/catalog/"

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
