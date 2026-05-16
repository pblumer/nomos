# Architecture Decision Records

This directory contains the canonical ADRs for the Nomos platform.

ADRs record significant architectural and technology decisions. Each ADR captures context, the decision taken, and its consequences. Once accepted, ADRs are not edited — superseded decisions get a new ADR.

## Index

| ADR | Title | Status |
| --- | --- | --- |
| [ADR-0001](ADR-0001-git-first-source-of-truth.md) | Git-first als Quelle der Wahrheit | Accepted |
| [ADR-0002](ADR-0002-yaml-artefacts-for-mvp.md) | YAML als primäres Artefaktformat in MVP 0.1 | Accepted |
| [ADR-0003](ADR-0003-no-primary-database-in-mvp.md) | Keine primäre Datenbank für autoritative Artefakte in MVP 0.1 | Accepted |
| [ADR-0004](ADR-0004-ai-assistance-only-no-autonomous-decisions.md) | KI-Assistenz nur für Drafts, keine autonomen Entscheidungen | Accepted |
| [ADR-0005](ADR-0005-blueprint-instance-assurance-model.md) | Blueprint-/Instance-/Assurance-Modell als Kernmodell | Accepted |
| [ADR-0006](ADR-0006-go-cobra-cli.md) | Go und Cobra als CLI-Technologie | Accepted |
| [ADR-0007](ADR-0007-domain-owned-product-offerings.md) | Domain-owned product offerings and service-based fulfillment | Accepted |
| [ADR-0008](ADR-0008-mcp-server-ai-agent-interface.md) | MCP Server als KI-Agenten-Schnittstelle für Nomos-Artefakte | Accepted |

## Superseded / legacy stubs

Earlier stubs in `docs/adr/` are redirect files pointing to the canonical ADRs above.

## Guidelines

- New ADRs go in this directory with the next sequential number.
- Use the filename pattern `ADR-XXXX-short-title.md`.
- Status values: `Proposed` → `Accepted` → `Superseded` (with a reference to the superseding ADR).
