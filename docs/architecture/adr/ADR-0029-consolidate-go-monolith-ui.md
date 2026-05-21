# ADR-0029 - Konsolidierung auf den Go-Monolith mit eingebetteter Web-UI

## Status

Accepted

## Datum

2026-05-21

## Kontext

Das Repository enthielt parallel mehrere UI- bzw. API-Implementierungen:

1. **Go-Monolith** (`cmd/nomos` + `internal/server`): `nomos serve` liefert REST-API, MCP-Endpoint und eine server-gerenderte Web-UI aus. Die UI-Templates und statischen Assets liegen unter `internal/server/web/` und werden per `//go:embed` in die Binary eingebettet (`html/template` + `embed`). Dieser Stack wurde aktiv weiterentwickelt und ist der von `Dockerfile`/`docker-compose.yml`/`deploy/ansible` deployte Pfad.
2. **Python/FastAPI-Backend + React/Vite-Frontend** (`backend/`, `frontend/`): ein zweiter, paralleler Stack, der eine Teilmenge derselben Domäne (Products, Rules, Requirements, Auth) über dieselben YAML-Artefakte erneut implementierte und über `deploy/docker-compose.{dev,prod}.yml` hinter einem Caddy-Reverse-Proxy betrieben wurde. Festgehalten in den (numbered) Dokumenten [012](../012-technology-stack-mvp-0-1.md) und [013](../013-dev-und-prod-umgebung-mvp-0-1.md).
3. **Verwaistes `web/`** im Repo-Root: ein veraltetes Duplikat einer früheren Version der Go-Templates, das von keinem Code referenziert wurde (der `//go:embed`-Pfad ist relativ zu `internal/server/` und bettet `internal/server/web/` ein, nicht das Root-`web/`).

Diese Mehrgleisigkeit war eine Quelle von Verwirrung und Doppelpflege. Zwei divergierende Implementierungen derselben API über demselben Datenmodell widersprechen dem Determinismus- und Wartbarkeitsanspruch von Nomos. [ADR-0015](ADR-0015-default-http-port-7373.md) hatte die Runtime-Konsolidierung bereits als offenen Punkt benannt.

## Entscheidung

Nomos wird auf den **Go-Monolithen als einzige Runtime und einzige Web-UI** konsolidiert.

- Die Web-UI ist ausschließlich die in `internal/server/web/` eingebettete, server-gerenderte UI, ausgeliefert von `nomos serve`.
- Das Python/FastAPI-Backend (`backend/`) und das React/Vite-Frontend (`frontend/`) werden entfernt.
- Der zugehörige Zwei-Container-Deployment-Pfad wird entfernt: `deploy/docker-compose.{dev,prod}.yml`, `deploy/reverse-proxy/`, `deploy/scripts/` und die backend-spezifischen Env-/Auth-Beispiele.
- Das verwaiste Root-`web/` wird gelöscht.
- Das Go-Ansible-Deployment (`deploy/ansible/`, systemd + `nomos serve`) bleibt der unterstützte Betriebspfad.

## Begruendung

- **Single Source of Truth für die UI/API**: nur noch eine Implementierung über dem YAML-/Cosmos-Modell, keine Drift zwischen zwei API-Oberflächen.
- **Kein separater Build-Chain zur Laufzeit**: keine Node/React/Vite-Abhängigkeit; die Binary ist selbstständig deploybar.
- **Konsistenz mit bestehenden ADRs**: ADR-0003 (keine primäre DB), ADR-0006 (Go/Cobra-CLI) und ADR-0015 (Port 7373, offener Konsolidierungspunkt) zeigen bereits in Richtung Go-zentrierter, dateibasierter Runtime.
- **Weniger Betriebs- und Reviewaufwand**: ein Artefakt, ein CI-Pfad, ein Deployment.

## Konsequenzen

- Entfernt: `backend/`, `frontend/`, Root-`web/`, `deploy/docker-compose.{dev,prod}.yml`, `deploy/reverse-proxy/`, `deploy/scripts/`, `deploy/.env.example`, `deploy/catalog.env.example`, `deploy/users.example.yaml`, `pytest.ini`.
- CI (`.github/workflows/ci.yml`) verliert die Jobs `python` (FastAPI-Tests) und `frontend` (React-Build); es bleiben die Go-Build/Vet/Test- und Demo-Validierungs-Jobs.
- Die in `internal/server/web/static/vendor/` eingecheckten Browser-Bundles (bpmn-js/dmn-js/form-js) bleiben Teil der Binary. Ihr Refresh erfolgt nicht mehr über `frontend/`, sondern über eine temporäre, eigenständige npm-Installation (siehe `docs/product-process-modeling.md`).
- Die numbered Docs 012 und 013 sind als *Superseded* markiert und verweisen hierher.
- Authentifizierung: Der Go-Server nutzt API-Key-Auth (`X-API-Key`, vgl. ADR-0023); das users-file-basierte Multi-User-Auth-Modell des FastAPI-Backends entfällt.

## Offene Punkte

- **MCP-/Agent-Skills-Defaults**: Die in [ADR-0015](ADR-0015-default-http-port-7373.md) genannten Defaults in `mcp/**` und `.agent-skills/**` (historisch `localhost:8080`) sollten nun eindeutig auf den Go-Server (`7373`) gerichtet werden. Wird separat nachgezogen.

## Related

- [ADR-0006 - Go und Cobra als CLI-Technologie](ADR-0006-go-cobra-cli.md)
- [ADR-0015 - Default-Port 7373 für `nomos serve`](ADR-0015-default-http-port-7373.md)
- [ADR-0023 - Inter-Server-Authentifizierung](ADR-0023-DRAFT-inter-server-authentication.md)
- [012 - Technologie-Stack für MVP 0.1](../012-technology-stack-mvp-0-1.md) (superseded)
- [013 - Dev- und Prod-Umgebung für MVP 0.1](../013-dev-und-prod-umgebung-mvp-0-1.md) (superseded)
