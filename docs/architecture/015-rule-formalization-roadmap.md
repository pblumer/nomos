# 015 Rule Formalization Roadmap

## Status

Draft

## Kontext

Aktuell nutzen Rules im MVP nur die Pflichtfelder:

- `id`
- `name`
- `description`

Das reicht für Anzeige und einfache Katalogpflege, ist aber für automatische Validierung und Governance zu dünn.

## Entscheidung

Rules werden schrittweise fachlich und technisch ausformuliert.

### Phase 1 (catalog only, rückwärtskompatibel)

Zusätzliche optionale Felder im Catalog erlauben:

- `category`
- `severity`
- `scope`
- `condition`
- `enforcement`
- `validation` (strukturiert)
- `owner`
- `references`
- `version`

Backend bleibt kompatibel und akzeptiert weiterhin mindestens `id`, `name`, `description`.

### Phase 2 (backend contract)

API-Schema für Rules erweitern:

- Neue Felder in Response-Modellen
- Validierung der `severity` Enum
- Validierung von `validation.type` + Operatoren

### Phase 3 (frontend UX)

Rule-Detailansicht erweitern:

- Severity Badge
- Kategorie/Scope Filter
- Validation-Block lesbar darstellen

### Phase 4 (policy checks)

Optionaler Policy-Check Endpoint, der Rule-Validation maschinenlesbar auswertet.

## Konsequenzen

- Bessere Fachlichkeit und Prüfbarkeit
- Höhere Qualität für spätere Automatisierung
- Etwas höherer Pflegeaufwand beim Rule-Authoring

## Referenzen

- `nomos-catalog/docs/rule-schema.md`
- `docs/architecture/014-katalog-repository-und-release-modell.md`
