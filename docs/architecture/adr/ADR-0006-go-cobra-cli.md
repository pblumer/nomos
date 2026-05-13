# ADR-0006 - Go und Cobra als CLI-Technologie

## Status

Accepted

## Datum

2026-05-13

## Kontext

Nomos ist ein Git-first-Werkzeug für Produkt- und Anforderungsmanagement. Der primäre Interaktionsmodus für Fach- und Technikrollen ist ein portables Kommandozeilen-Interface. Das CLI muss plattformübergreifend funktionieren (Linux, macOS, Windows), eine kurze Startup-Latenz haben und ohne Runtime-Abhängigkeiten verteilt werden können.

Zusätzlich dient das CLI als Einsteigspunkt für die REST-API und die Web-UI: `nomos serve` startet den integrierten HTTP-Server und liefert die eingebettete Web-UI aus.

## Entscheidung

Nomos verwendet Go für das CLI und die eingebettete Web-UI. Cobra wird als Command-Router eingesetzt.

Die CLI-Befehle sind in `internal/cli/root.go` zentralisiert und rufen Use-Cases aus `internal/app/` auf. Der Go-Server bindet das Web-UI-Template-Verzeichnis `web/` zur Compile-Zeit via `//go:embed` ein, sodass kein separater Node-Build-Schritt zur Laufzeit nötig ist.

## Begruendung

- Go kompiliert zu einem einzelnen, statisch gelinkten Binary ohne externe Runtime-Abhängigkeiten.
- Kurze Startup-Zeit passend für interaktive CLI-Nutzung.
- Cobra bietet eine stabile, konventionsorientierte API für Subcommands, Flags und Hilfe-Generierung.
- `//go:embed` ermöglicht die Einbettung von Web-Templates und Static Assets ohne separaten Deploymentschritt.
- Cross-Compilation für Linux, macOS und Windows ist nativ in der Go-Toolchain enthalten.

## Konsequenzen

- Das Binary enthält CLI, REST-API und Web-UI in einem einzigen Artefakt.
- Versionsmetadaten werden über ldflags zur Build-Zeit injiziert (siehe `internal/version/`).
- Cobra-Subcommands in `internal/cli/root.go` sind die einzige Einstiegsstelle für CLI-Befehle.
- CLI-Lesebefehle rufen die lokale REST-API auf; Schreibbefehle gehen direkt durch `internal/app/`.
- Das Web-UI-Template-Verzeichnis `web/` hat keinen eigenen Build-Schritt und keine Node-Abhängigkeit zur Laufzeit.

## Betrachtete Alternativen

- **Python + Click/Typer**: Bessere Scripting-Integration, aber abhängig von installiertem Python-Interpreter. Nicht geeignet für Zero-Dependency-Deployment.
- **Rust + Clap**: Vergleichbare Performance und Binary-Portabilität, aber höhere Einstiegshürde im Team und schlechtere Go-Interoperabilität.
- **Node.js + Commander**: Bestehende Frontend-Tooling-Nähe, aber langsame Startup-Zeit und Node-Runtime-Abhängigkeit.
- **Shell-Skripte**: Nicht portabel genug für Windows; keine typsichere Geschäftslogik.

## Related

- [ADR-0001 - Git-first als Quelle der Wahrheit](ADR-0001-git-first-source-of-truth.md)
- [ADR-0002 - YAML als primäres Artefaktformat](ADR-0002-yaml-artefacts-for-mvp.md)
- [012 - Technology Stack MVP 0.1](../012-technology-stack-mvp-0-1.md)
