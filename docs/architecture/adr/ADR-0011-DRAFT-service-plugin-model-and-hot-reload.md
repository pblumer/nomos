# ADR-0011 (DRAFT): Service-Plugin-Modell und Hot-Reload-Isolation

## Status

Draft

## Datum

2026-05-16

## Kontext

ADR-0009 legt fest, dass fachliche Faehigkeiten als registrierte Services implementiert werden,
die hot-reload-faehig sind. Dieses ADR entscheidet das genaue Isolations- und Lademodell.

## Offene Fragen

- Wie werden Services isoliert: in-process (Go-Plugins), subprocess (eigener Prozess) oder WASM?
- Wie wird das Service-Manifest (YAML) valide gegen Race Conditions reloaded?
- Wie wird der Service Router waehrend eines Reloads weiter bedient (zero-downtime)?
- Wie werden abgestuerzte Services detektiert und neu gestartet?
- Wie wird das Capability-Contract-System (welcher Service kann welche Operationen) durchgesetzt?

## Kandidaten-Optionen (noch nicht entschieden)

**Option A — In-process Go-Plugins (`plugin.so`):**
Go-Plugin-Mechanismus. Vorteil: Niedrige Latenz, kein IPC. Nachteil: Fragil bei Typinkompatibilitaet,
kein echter Crash-Isolation.

**Option B — Subprocess mit gRPC/stdio:**
Jeder Service laeuft als eigener Prozess, kommuniziert via gRPC oder stdio. Vorteil: Volle
Isolation, crash-safe, sprachagnostisch. Nachteil: IPC-Overhead, Prozessmanagement.

**Option C — WASM-Module:**
Services werden als WASM-Binaries geliefert und in einer WASM-Runtime ausgefuehrt. Vorteil:
Portabel, sandbox-sicher, hot-swappable. Nachteil: WASM-Toolchain-Komplexitaet, I/O-Einschraenkungen.

## Naechste Schritte

- Evaluierung Option B (Subprocess) als pragmatische Baseline.
- Spike: WASM-Modul fuer einen einfachen Read-only-Service.
- Entscheidung des Hot-Reload-Protokolls (inotify vs. Git-Webhook vs. API-Trigger).

## Referenz

- ADR-0009 — Nomos Cosmos Netzwerkarchitektur und Core Engine
