#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
DEPLOY_DIR="$(dirname -- "$SCRIPT_DIR")"
ROOT_DIR="$(dirname -- "$DEPLOY_DIR")"

CATALOG_REPO="${NOMOS_CATALOG_REPO:-https://github.com/pblumer/nomos-catalog.git}"
CATALOG_DIR="${NOMOS_CATALOG_DIR:-$ROOT_DIR/../nomos-catalog}"
CATALOG_REF="${CATALOG_REF:-catalog-v0.1.0}"

mkdir -p "$(dirname -- "$CATALOG_DIR")"

if [ ! -d "$CATALOG_DIR/.git" ]; then
  echo "[catalog] Cloning $CATALOG_REPO into $CATALOG_DIR"
  git clone "$CATALOG_REPO" "$CATALOG_DIR"
fi

echo "[catalog] Fetching remote state..."
git -C "$CATALOG_DIR" fetch --tags origin

echo "[catalog] Checking out ref: $CATALOG_REF"
if git -C "$CATALOG_DIR" rev-parse --verify --quiet "$CATALOG_REF^{commit}" >/dev/null; then
  git -C "$CATALOG_DIR" checkout --detach "$CATALOG_REF"
else
  echo "[catalog] ERROR: ref '$CATALOG_REF' not found in $CATALOG_DIR"
  exit 1
fi

PINNED_COMMIT="$(git -C "$CATALOG_DIR" rev-parse --short HEAD)"
echo "[catalog] Pinned catalog at $PINNED_COMMIT"
