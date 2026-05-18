#!/bin/sh
set -e

# If no cosmos.yaml exists yet, bootstrap a demo cosmos
if [ ! -f "$COSMOS_PATH/.nomos/cosmos.yaml" ]; then
  echo "==> Bootstrapping demo cosmos at $COSMOS_PATH"
  NOMOS_BIN="$NOMOS_BIN" COSMOS_PATH="$COSMOS_PATH" sh /usr/local/bin/create-demo-cosmos.sh
fi

echo "==> Starting Nomos server on $NOMOS_LISTEN"
exec nomos serve --path "$COSMOS_PATH" --listen "$NOMOS_LISTEN"
