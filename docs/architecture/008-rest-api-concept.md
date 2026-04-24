# 008 - REST API Konzept (konzeptionell)

## Zweck
Konzeptioneller API-Rahmen für MVP 0.1 ohne Implementierungsvertrag.

## Wichtiger Hinweis
Die hier gezeigten Endpunkte sind Beispiele und keine finalen Contracts. OpenAPI Spezifikation ist nicht Teil dieses Schritts.

## API-Gruppen
- Product Catalogue API
- Requirement API
- Rule API
- Validation API
- Git Workspace API
- AI Assistance API
- Governance/Review API
- Metadata API

## Beispielendpunkte (konzeptionell)
- GET /api/v1/products
- GET /api/v1/products/{product_id}
- GET /api/v1/products/{product_id}/requirements
- GET /api/v1/products/{product_id}/rules
- POST /api/v1/products/{product_id}/validate
- POST /api/v1/workspaces
- POST /api/v1/workspaces/{workspace_id}/commit
- POST /api/v1/workspaces/{workspace_id}/review-request
- POST /api/v1/ai/rule-draft
- POST /api/v1/ai/requirement-draft

## Architekturprinzipien
- Trennung von Leseoperationen und Änderungsvorschlag-Operationen.
- Branch-Kontext ist zentral für reproduzierbare Ergebnisse.
- Validierung kann gegen einen ausgewählten Branch laufen.
- API darf Governance nicht umgehen.
- Hauptbranch-Schreibzugriffe nur über geregelten Reviewprozess.
