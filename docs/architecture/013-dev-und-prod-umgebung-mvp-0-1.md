# 013 - Dev- und Prod-Umgebung für MVP 0.1

## Zweck
Dieses Dokument beschreibt den einfachsten, kontrollierten Aufbau für Entwicklungs- und Produktionsumgebung in MVP 0.1. Ziel ist ein schneller Start mit wenig Betriebsaufwand und klarer Erweiterbarkeit.

## Leitprinzipien für die Umgebungen
- Einfachheit vor Vollausbau.
- Gleiche Runtime-Bausteine in Dev und Prod, damit Verhalten möglichst konsistent bleibt.
- Git-first bleibt unverändert: fachliche Artefakte liegen im Repository, nicht in einer primären Datenbank.
- Containerisierung für reproduzierbare Setups.
- Secrets nie im Repository speichern.

## Umgebungsmodell (MVP)

### 1) Dev-Umgebung (lokal)
Empfehlung: Docker Compose lokal auf Entwicklerrechnern.

Bausteine:
- frontend: React + Vite Dev Server
- backend: FastAPI App (uvicorn)
- optionaler reverse proxy: nur falls lokal gleiche Pfade wie Prod getestet werden sollen
- lokale Arbeitsverzeichnisse als Volumes:
  - artefact workspace
  - temporäre git workspaces

Vorteile:
- Ein Kommando für Start und Stop.
- Weniger lokale Abhängigkeitsprobleme.
- Gleiches Containerbild kann später in Prod genutzt werden.

### 2) Prod-Umgebung (initial)
Empfehlung für MVP: Single-VM Deployment mit Docker Compose.

Bausteine:
- reverse proxy mit TLS (z. B. Caddy oder Traefik)
- frontend als statische Assets (Nginx Container oder direkt über Proxy)
- backend als FastAPI Container
- persistente Volumes für:
  - temporäre Workspace-Daten
  - Logs (falls lokal gesammelt)
- Git Provider als externes System (GitHub/GitLab)

Warum Single-VM für MVP:
- Schnellster Weg in stabilen Betrieb.
- Geringe Betriebskomplexität.
- Ausreichend für erwartete Last in MVP 0.1.

## Nicht-Ziele im MVP-Betrieb
- Kein Kubernetes-Zwang.
- Kein verteiltes Multi-Region Setup.
- Kein primäres Datenbankcluster für Business-Artefakte.
- Kein vollständiges Enterprise-SSO als Blocker für den Start.

## Konfiguration pro Umgebung

### Gemeinsame Konfigurationsprinzipien
- Alle Umgebungswerte über ENV Variablen.
- Keine hardcodierten URLs, Tokens oder Branchnamen.
- Strikte Trennung zwischen:
  - APP Konfiguration
  - Git Integration
  - AI Integration (optional)

### Beispielhafte ENV Variablen
- NOMOS_ENV=dev|prod
- NOMOS_API_HOST
- NOMOS_API_PORT
- NOMOS_FRONTEND_ORIGIN
- NOMOS_GIT_PROVIDER
- NOMOS_GIT_REPO_URL
- NOMOS_GIT_DEFAULT_BRANCH=main
- NOMOS_GIT_TOKEN (Secret)
- NOMOS_WORKSPACE_ROOT
- NOMOS_LOG_LEVEL
- NOMOS_AI_ENABLED=true|false
- NOMOS_AI_PROVIDER (optional)
- NOMOS_AI_API_KEY (Secret, optional)

## Build- und Release-Ansatz

### Dev
- Frontend: hot reload via Vite
- Backend: autoreload via uvicorn --reload
- Linting und Tests lokal sowie in CI

### Prod
- Multi-stage Docker Build für frontend und backend
- Versionierte Container-Tags pro Release (z. B. git sha)
- Deployment über compose pull + compose up -d
- Rollback über vorherigen Image-Tag

## Minimaler Betriebsablauf für Prod
1. CI baut und testet Images.
2. CI pusht Images in Registry.
3. Deployment-Pipeline aktualisiert compose auf Ziel-VM.
4. Healthcheck prüft API und UI Erreichbarkeit.
5. Bei Fehler: automatischer oder manüller Rollback auf vorherigen Tag.

## Sicherheitsbaseline für MVP
- TLS nur über reverse proxy.
- Secrets ausschliesslich über Secret Store oder .env auf Server (nicht im Git).
- Netzwerk: nur 80/443 offen, Backend nicht direkt aus dem Internet erreichbar.
- Branchschutz und PR-Checks als technische Governance.
- Auditfähigkeit über Git Historie und CI Logs.

## Observability baseline
- Strukturierte Backend-Logs (json oder klar formatierte text logs).
- Reverse-Proxy Access Logs.
- Einfache Uptime-Checks für /health Endpunkt.
- Fehlertracking kann später ergänzt werden (kein MVP-Blocker).

## Verzeichnisvorschlag für Betriebsartefakte
- /deploy
  - docker-compose.dev.yml
  - docker-compose.prod.yml
  - .env.example
  - reverse-proxy/
    - Caddyfile oder traefik config
  - scripts/
    - deploy.sh
    - rollback.sh

## Entscheidung: warum dieser Weg der einfachste ist
- Nutzt den bereits gewählten Stack direkt ohne Zusatzplattform.
- Trennt klar zwischen lokalem Arbeiten und stabilem Betrieb.
- Erlaubt spätere Migration auf Kubernetes oder Managed Plattform ohne Neuschreiben der Anwendung.
- Bleibt konsistent mit ADR-0001 bis ADR-0004.

## Upgradepfad nach MVP
- Einführung einer dedizierten Staging-Umgebung zwischen Dev und Prod.
- Read Model/Suchindex als abgeleitete Komponente bei wachsendem Query-Bedarf.
- Zentralisiertes Logging und Monitoring.
- Erweiterte Authentifizierung und Rollenmodell.

## Konkrete Dateien im Repository
- deploy/.env.example
- deploy/docker-compose.dev.yml
- deploy/docker-compose.prod.yml
- deploy/reverse-proxy/Caddyfile.dev
- deploy/reverse-proxy/Caddyfile
- deploy/scripts/deploy.sh
- deploy/scripts/rollback.sh

## Schnellstart Dev
1. cp deploy/.env.example deploy/.env
2. Werte in deploy/.env anpassen.
3. docker compose --env-file deploy/.env -f deploy/docker-compose.dev.yml up --build
4. Optional mit Proxy-Profil: docker compose --env-file deploy/.env -f deploy/docker-compose.dev.yml --profile proxy up --build

## Schnellstart Prod
1. Frische VM mit Docker und Docker Compose Plugin vorbereiten.
2. Repository auschecken und deploy/.env erstellen.
3. Initial deploy: ./deploy/scripts/deploy.sh
4. Deploy mit Tag: ./deploy/scripts/deploy.sh <tag>
5. Rollback: ./deploy/scripts/rollback.sh <backend-image> <frontend-image>

## Akzeptanzkriterien für Phase-Start
- Entwickler können die Umgebung mit einem klaren Startkommando lokal hochfahren.
- Prod läuft reproduzierbar auf einer frischen VM anhand dokumentierter Schritte.
- Secrets sind getrennt von Quellcode verwaltet.
- Deployment und Rollback sind dokumentiert und testbar.
