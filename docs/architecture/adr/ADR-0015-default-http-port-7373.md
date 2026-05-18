# ADR-0015 - Default-Port `7373` für `nomos serve`

## Status

Accepted

## Datum

2026-05-18

## Kontext

`nomos serve` startet einen HTTP-Server, der sowohl die REST-API als auch die eingebettete Web-UI ausliefert (siehe [ADR-0006](ADR-0006-go-cobra-cli.md)). Bislang war der Default des Flags `--listen` `127.0.0.1:8080`.

Port `8080` ist in der Praxis stark überladen: Reverse Proxies (Caddy/Traefik), Tomcat, alternative HTTP-Endpoints diverser Tools und CI-Runner belegen ihn regelmäßig. Im bestehenden `deploy/docker-compose.dev.yml` wird `8080` außerdem bereits für den Caddy-Proxy verwendet, was beim Wechsel auf das Go-Binary zu Port-Kollisionen geführt hätte.

Beim Aufbau einer containerisierten lokalen Entwicklungsumgebung und perspektivisch eines Helm-Charts mit optionalem Monitoring-Stack (Prometheus `9090`, Grafana `3000`, Loki `3100`) sowie optionalen Event-Brokern (NATS `4222`, RabbitMQ `5672`, Kafka `9092`) soll der Nomos-Default-Port nicht mit etablierten Ports dieser Nachbarsysteme kollidieren.

## Entscheidung

Der Default-Wert von `--listen` für `nomos serve` wechselt von `127.0.0.1:8080` auf `127.0.0.1:7373`.

`7373` wird der konventionelle Nomos-Port in CLI-Defaults, Doku-Beispielen, Compose-Files und im zukünftigen Helm-Chart (`service.port`).

## Begruendung

- **IANA-unassigned**: `7373` ist im IANA Service Name and Transport Protocol Port Number Registry zum Entscheidungszeitpunkt nicht zugewiesen.
- **Außerhalb der Dev-Hot-Zone**: Liegt nicht im stark belegten Bereich `3000–9000`, in dem die meisten Web-/API-Tools per Default lauschen.
- **Klare Trennung zu Nachbarsystemen**: Keine Kollision mit Prometheus (`9090`), Grafana (`3000`), Loki (`3100`), Tempo (`3200`), NATS (`4222`), RabbitMQ (`5672`), Kafka (`9092`), Postgres (`5432`), Redis (`6379`).
- **Merkbar**: Palindromisch, kurz, eindeutig einer Anwendung zuzuordnen.
- **Backward Compatibility unkritisch**: Es gibt zum Zeitpunkt der Entscheidung keinen produktiven Betrieb; alle Beispiele und Skripte sind im Repository selbst gepflegt und werden in einem Schritt mitgezogen.

## Konsequenzen

- Default in `internal/cli/root.go` ist `127.0.0.1:7373`.
- `README.md`, `docs/cli-usage.md` und `scripts/create-demo-cosmos.sh` zeigen `7373` in allen Beispielen.
- Zukünftige Container-Artefakte (Dockerfile `EXPOSE`, Compose, Helm `service.port`, Ingress-Beispiele) verwenden `7373` als Default.
- Nutzer können den Port jederzeit per `--listen host:port` überschreiben — der Wechsel ist nicht bindend.
- MCP-Tools (`mcp/nomos_mcp/tools/*.py`) und Agent-Skills (`.agent-skills/**`) zeigen weiterhin auf `localhost:8080`. Diese werden bewusst nicht in diesem ADR mitgezogen, weil unklar ist, ob sie das Go-Backend oder das Python-FastAPI-Backend adressieren. Die Anpassung erfolgt zusammen mit der Runtime-Konsolidierung (siehe „Offene Punkte").
- Bestehende lokale Skripte/Aliases von Nutzern, die hartkodiert `8080` ansprechen, müssen einmalig angepasst werden.

## Betrachtete Alternativen

- **`8080` beibehalten**: Vertraut, aber bereits durch Caddy-Proxy in `docker-compose.dev.yml` belegt und in vielen Tool-Stacks doppelt vergeben. Verworfen.
- **`8473`**: Ebenfalls IANA-unassigned und gedanklich nahe an `8443` (HTTPS-Alt). Funktional gleichwertig, aber weniger einprägsam als `7373`.
- **Port in privater Range (z. B. `47853`)**: Garantiert konfliktfrei, aber unhandlich zu merken und unüblich für UI-zugängliche Services.
- **`7474`**: Vergeben durch Neo4j HTTP — explizit verworfen.

## Offene Punkte

- **MCP-/Agent-Skills-Defaults**: Anpassung der `NOMOS_API_BASE`-Defaults in `mcp/**` und `.agent-skills/**` erfolgt im Zuge der Entscheidung, ob das Python-FastAPI-Backend zugunsten des Go-`nomos serve` konsolidiert wird.
- **Caddy-Mapping in Compose**: `deploy/docker-compose.dev.yml` mapped aktuell `8080:80` für den Proxy. Diese Mappings werden überarbeitet, sobald die Compose-Files auf das Go-Binary umgestellt werden.

## Related

- [ADR-0006 - Go und Cobra als CLI-Technologie](ADR-0006-go-cobra-cli.md)
