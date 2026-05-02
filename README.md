# Nomos Cosmos Client MVP

Nomos ist ein portables CLI Werkzeug fuer Git-first Cosmos Repositories.

## Was der MVP kann
- Cosmos lokal initialisieren
- Domaenen und Services anlegen
- Struktur validieren
- Graph ausgeben
- DNS TXT Verifikation pruefen
- lokale Read-only API starten

## Was noch nicht enthalten ist
- BPMN oder DMN Runtime
- Web UI
- Datenbank
- autonome KI Entscheidungen

## Quickstart
```bash
go test ./...
go run ./cmd/nomos --help
go run ./cmd/nomos cosmos init ./tmp/demo-cosmos --git
go run ./cmd/nomos domain add identity.blumer.cloud --path ./tmp/demo-cosmos --owner "Identity Team"
go run ./cmd/nomos service add user-account --domain identity.blumer.cloud --path ./tmp/demo-cosmos --owner "Identity Team"
go run ./cmd/nomos validate --path ./tmp/demo-cosmos
go run ./cmd/nomos graph --path ./tmp/demo-cosmos --format mermaid
go run ./cmd/nomos serve --path ./tmp/demo-cosmos --listen 127.0.0.1:8080
```
