# 014 - Katalog-Repository und Release-Modell

## Zweck
Dieses Dokument beschreibt den Betrieb mit getrenntem Katalog-Repository (Variante B) fuer wiederverwendbare Requirements und Rules.

## Zielbild
- Repo 1: `pblumer/nomos` fuer Anwendungscode (Backend, Frontend, Deploy).
- Repo 2: `pblumer/nomos-catalog` fuer fachliche Artefakte.

Das Katalog-Repository enthaelt:
- `products/`
- `requirements/`
- `rules/`

Produkte referenzieren nur IDs:
- `requirement_ids`
- `rule_ids`

## Warum diese Trennung
- Fachliche Aenderungen sind unabhaengig vom Anwendungsrelease.
- Requirements und Rules koennen ueber mehrere Produkte wiederverwendet werden.
- Governance und Review fuer Fachinhalte sind klar getrennt.
- Deployments koennen auf einen festen Tag oder Commit des Katalogs gepinnt werden.

## Lokales Setup
1. Repos nebeneinander auschecken:
   - `~/nomos`
   - `~/nomos-catalog`
2. In `nomos/deploy/.env` die Pfade setzen:
   - `NOMOS_PRODUCTS_DIR=../nomos-catalog/products`
   - `NOMOS_REQUIREMENTS_DIR=../nomos-catalog/requirements`
   - `NOMOS_RULES_DIR=../nomos-catalog/rules`

## Release per Tag (empfohlen)
Im Repo `nomos-catalog`:
1. Aenderungen mergen.
2. Tag erstellen, z. B. `catalog-v0.1.0`.
3. Tag pushen.

Auf Runtime-Seite:
1. Gewuenschten Tag auschecken.
2. Nomos mit den Katalogpfaden starten.

Vorteil: reproduzierbare, fachlich eindeutig markierte Katalogversion.

## Release per Commit
Alternativ kann ein fester Commit-Hash verwendet werden.

Vorteil: maximal praezise Reproduzierbarkeit.
Nachteil: weniger lesbar als ein Versionstag.

## Betriebsregeln
- Keine fachlichen Artefakte im Code-Repo pflegen, ausser minimalen Beispielen.
- Katalog-Aenderungen immer per Pull Request reviewen.
- Produktionsbetrieb nur mit gepinntem Tag oder Commit.
- Bei Breaking Changes zuerst Katalog mergen, dann Nomos-Code mit passender Version ausrollen.
