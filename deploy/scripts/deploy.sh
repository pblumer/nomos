#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
DEPLOY_DIR="$(dirname -- "$SCRIPT_DIR")"

cd "$DEPLOY_DIR"

if [ ! -f .env ]; then
  echo "[deploy] Missing deploy/.env. Copy deploy/.env.example to deploy/.env and fill secrets."
  exit 1
fi

if [ -f ./catalog.env ]; then
  # shellcheck disable=SC1091
  . ./catalog.env
fi

if [ -n "${CATALOG_REF:-}" ] || [ -n "${NOMOS_CATALOG_REPO:-}" ] || [ -n "${NOMOS_CATALOG_DIR:-}" ]; then
  echo "[deploy] Syncing external catalog..."
  "$SCRIPT_DIR/sync-catalog.sh"
fi

TAG="${1:-}"
if [ -n "$TAG" ]; then
  IMAGE_REPO="${NOMOS_IMAGE_REPO:-ghcr.io/pblumer/nomos}"
  export NOMOS_BACKEND_IMAGE="${IMAGE_REPO}-backend:${TAG}"
  export NOMOS_FRONTEND_IMAGE="${IMAGE_REPO}-frontend:${TAG}"
  echo "[deploy] Using tag $TAG"
  echo "[deploy] Backend image: $NOMOS_BACKEND_IMAGE"
  echo "[deploy] Frontend image: $NOMOS_FRONTEND_IMAGE"
fi

echo "[deploy] Pulling images..."
docker compose --env-file .env -f docker-compose.prod.yml pull

echo "[deploy] Starting services..."
docker compose --env-file .env -f docker-compose.prod.yml up -d --remove-orphans

echo "[deploy] Current status:"
docker compose --env-file .env -f docker-compose.prod.yml ps

if command -v curl >/dev/null 2>&1; then
  HEALTH_URL="${NOMOS_HEALTHCHECK_URL:-http://localhost/health}"
  echo "[deploy] Healthcheck: $HEALTH_URL"
  curl -fsS "$HEALTH_URL" >/dev/null && echo "[deploy] Healthcheck OK" || echo "[deploy] Healthcheck failed (check logs)"
fi

echo "[deploy] Done."
