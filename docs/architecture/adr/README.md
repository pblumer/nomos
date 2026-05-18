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
| [ADR-0009](ADR-0009-nomos-cosmos-network-and-core-engine.md) | Nomos Cosmos Netzwerkarchitektur und Core Engine | Proposed |
| [ADR-0010](ADR-0010-DRAFT-well-known-endpoint-and-domain-proof.md) | Well-known-Endpoint-Spezifikation und Domain-Proof-Format | Draft |
| [ADR-0011](ADR-0011-DRAFT-service-plugin-model-and-hot-reload.md) | Service-Plugin-Modell und Hot-Reload-Isolation | Draft |
| [ADR-0012](ADR-0012-DRAFT-global-cosmos-index-service.md) | Globaler Cosmos-Index-Service (opt-in Crawler) | Draft |
| [ADR-0013](ADR-0013-user-contact-interfaces.md) | User Contact Interfaces (UCI) als erstklassige Nomos-Artefakte | Draft |
| [ADR-0014](ADR-0014-DRAFT-nomos-self-model-bundle.md) | Nomos-Selbstmodellierung als versioniertes Cosmos-Bundle (Nomos-as-Nomos) | Draft |
| [ADR-0015](ADR-0015-default-http-port-7373.md) | Default-Port `7373` für `nomos serve` | Accepted |
| [ADR-0016](ADR-0016-DRAFT-graph-storage-and-query-model.md) | Graph-Speicherung und Abfragemodell des Cosmos (Git-projizierter openCypher-Graph) | Draft |
| [ADR-0017](ADR-0017-DRAFT-decision-traceability.md) | Decision Traceability via inhaltsadressierte, hash-verkettete Trace-Artefakte | Draft |
| [ADR-0018](ADR-0018-DRAFT-process-trigger-events.md) | Process Trigger Events: Modell-Layer für Timer- und Message-Starts | Draft |

## Superseded / legacy stubs

Earlier stubs in `docs/adr/` are redirect files pointing to the canonical ADRs above.

## Guidelines

- New ADRs go in this directory with the next sequential number.
- Use the filename pattern `ADR-XXXX-short-title.md`.
- Status values: `Proposed` → `Accepted` → `Superseded` (with a reference to the superseding ADR).
