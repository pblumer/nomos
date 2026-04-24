#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
DEPLOY_DIR="$(dirname -- "$SCRIPT_DIR")"

cd "$DEPLOY_DIR"

if [ ! -f .env ]; then
  echo "[rollback] Missing deploy/.env."
  exit 1
fi

BACKEND_IMAGE="${1:-${PREVIOUS_BACKEND_IMAGE:-}}"
FRONTEND_IMAGE="${2:-${PREVIOUS_FRONTEND_IMAGE:-}}"

if [ -z "$BACKEND_IMAGE" ] || [ -z "$FRONTEND_IMAGE" ]; then
  echo "Usage: ./deploy/scripts/rollback.sh <backend-image> <frontend-image>"
  echo "Or set PREVIOUS_BACKEND_IMAGE and PREVIOUS_FRONTEND_IMAGE"
  exit 1
fi

echo "[rollback] Backend image: $BACKEND_IMAGE"
echo "[rollback] Frontend image: $FRONTEND_IMAGE"

NOMOS_BACKEND_IMAGE="$BACKEND_IMAGE" \
NOMOS_FRONTEND_IMAGE="$FRONTEND_IMAGE" \
  docker compose --env-file .env -f docker-compose.prod.yml pull

NOMOS_BACKEND_IMAGE="$BACKEND_IMAGE" \
NOMOS_FRONTEND_IMAGE="$FRONTEND_IMAGE" \
  docker compose --env-file .env -f docker-compose.prod.yml up -d --remove-orphans

echo "[rollback] Current status:"
NOMOS_BACKEND_IMAGE="$BACKEND_IMAGE" \
NOMOS_FRONTEND_IMAGE="$FRONTEND_IMAGE" \
  docker compose --env-file .env -f docker-compose.prod.yml ps

echo "[rollback] Done."
