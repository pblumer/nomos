# 012 - Technologie-Stack fuer MVP 0.1

## Zweck
Dieses Dokument fixiert den initial gewaehlten Technologie-Stack fuer die Umsetzung von MVP 0.1. Es ergaenzt die Architekturgrundlagen um konkrete Framework-Entscheidungen fuer Backend, Frontend, Validierung, Git-Integration und Delivery.

## Rahmenbedingungen aus ADRs
Die Stack-Entscheidung folgt den bereits beschlossenen Leitplanken:
- Git-first als Quelle der Wahrheit (ADR-0001)
- YAML als primaeres Artefaktformat (ADR-0002)
- keine primaere Artefaktdatenbank im MVP 0.1 (ADR-0003)
- KI nur assistiv, nie autonom entscheidend (ADR-0004)

## Entscheidungsueberblick

### Backend
- Sprache: Python 3.12
- API Framework: FastAPI
- Datenmodelle und Contracts: Pydantic v2
- YAML Verarbeitung: ruamel.yaml
- Strukturvalidierung: JSON Schema
- Linting fuer YAML: yamllint
- Git-Anbindung: Git CLI (subprocess) oder alternativ GitPython
- Testframework: pytest

### Frontend
- Sprache: TypeScript
- Framework: React
- Build Tool: Vite
- Server-State und API Caching: TanStack Query
- Formular- und Editor-Validierung: React Hook Form + Zod
- Artefakt-Editor und Diff-Naehe: Monaco Editor
- UI Komponenten: shadcn/ui (Radix-basiert)

### Delivery und Qualitaet
- CI: GitHub Actions
- Pflichtchecks in PRs:
  - yamllint
  - Schema-Validierung
  - Backend Tests (pytest)
  - Frontend Typcheck und Build
- Lokale Qualitaet: pre-commit Hooks fuer Linting und Basiskontrollen

## Begruendung der Wahl

### Warum Python + FastAPI
- Sehr gute Passung fuer parse-, validierungs- und regelorientierte Backend-Logik.
- Schnelle Bereitstellung klarer REST-Endpunkte mit OpenAPI.
- Pydantic v2 erlaubt strikte, gut testbare Domain- und API-Modelle.
- Geringe Reibung bei YAML- und Schema-getriebener Verarbeitung.

### Warum React + TypeScript + Vite
- Schnelle Entwicklungszyklen fuer ein iteratives MVP.
- TypeScript reduziert Fehler in komplexen Editor- und Review-Flows.
- Vite liefert schnelles lokales Feedback und schlanke Build-Pipeline.
- Monaco unterstuetzt den Git-first Ansatz durch gute Diff- und Editor-Ergonomie.

### Warum ohne primaere Datenbank im MVP
- Entspricht ADR-0003.
- Reduziert Architekturkomplexitaet in frueher Phase.
- Fokus bleibt auf Git-gestuetzter Nachvollziehbarkeit und Governance.
- Optional spaeter: abgeleitetes Read Model oder Suchindex fuer Performance.

## Zuordnung zur Roadmap
- Phase 1: YAML-Schemas, Beispielartefakte, Linting-Basis
- Phase 2: FastAPI Service-Rahmen, Read-Endpunkte, Parser
- Phase 3: Workspace/Branch/Commit/Review-Endpunkte
- Phase 4: React UI fuer Produkt-, Requirement- und Rule-Bearbeitung
- Phase 5: Deterministische Validierungsengine in Python
- Phase 6: KI-Assistance als optionaler Draft-Flow

## Abgrenzung
Dieser Stack ist fuer MVP 0.1 festgelegt. Aenderungen an Grundprinzipien (z. B. neue primaere Persistenz oder Abkehr von Git-first) benoetigen zuerst eine ADR-Anpassung.

## Offene Punkte fuer die naechsten Schritte
- Endgueltige OpenAPI Spezifikation fuer MVP-Endpunkte
- Konkrete Repo-Struktur fuer backend/ und frontend/
- Definition der ersten JSON Schemas fuer Kernartefakte
- Auswahl der CI Job-Matrix und Mindestqualitaetsgrenzen
