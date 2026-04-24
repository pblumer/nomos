# Nomos Runbook: Catalog upgrade and pinning

## Ziel

Nomos soll reproduzierbar auf einem fixen Stand des `nomos-catalog` laufen.
Dafür wird ein Tag oder Commit über `CATALOG_REF` gepinnt.

## Voraussetzungen

- `deploy/.env` ist vorhanden.
- `deploy/catalog.env` ist vorhanden (aus `deploy/catalog.env.example` kopiert).
- Der Ziel-Tag existiert im Repo `pblumer/nomos-catalog`.

## Initiales Setup

```bash
cd deploy
cp catalog.env.example catalog.env
```

Beispiel in `deploy/catalog.env`:

```env
NOMOS_CATALOG_REPO=https://github.com/pblumer/nomos-catalog.git
NOMOS_CATALOG_DIR=../nomos-catalog
CATALOG_REF=catalog-v0.1.0
```

## Upgrade auf neuen Catalog-Stand

1) Ziel-Tag im Catalog erstellen (im Repo `nomos-catalog`), z. B. `catalog-v0.2.0`.

2) In `nomos/deploy/catalog.env` den Pin anpassen:

```env
CATALOG_REF=catalog-v0.2.0
```

3) Deployment ausführen:

```bash
./deploy/scripts/deploy.sh
```

`deploy.sh` ruft automatisch `deploy/scripts/sync-catalog.sh` auf, wenn `catalog.env` gesetzt ist.

## Verifikation

- Ausgabe enthält z. B.:
  - `[catalog] Checking out ref: catalog-v0.2.0`
  - `[catalog] Pinned catalog at <commit>`
- Healthcheck ist grün.

## Rollback

- `CATALOG_REF` auf vorherigen Tag zurücksetzen.
- Erneut deployen:

```bash
./deploy/scripts/deploy.sh
```

## Hinweise

- In Produktion nie auf `main` pinnen.
- Immer immutable refs verwenden (Tag oder Commit SHA).
