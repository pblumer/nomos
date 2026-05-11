# development-lifecycle

## Go Web UI in MVP

The MVP web interface remains read-only and is served by the Go binary (`nomos serve`) with embedded templates/assets. This keeps delivery simple while preserving a consistent Nomos UI style without introducing a second runtime build pipeline.


## Source repo vs. Cosmos workspace

Development happens in the Nomos source repository (`cmd/`, `internal/`, `docs/`, `examples/`, `scripts/`). Concrete Cosmos data must not be committed at the source repository root. Use `nomos cosmos init <path>` to create a separate workspace whose mutable state is stored under `<path>/.nomos/`.

Recommended ignores:

- Dedicated Cosmos repository: `.nomos/cache/` and `.nomos/index/`.
- Local-only Nomos data in another repository: `.nomos/`.

## Domain-owned product offering foundation

When changing catalog, validation, CLI or REST behavior, keep the ADR-0001 foundation intact: product blueprints load with explicit `offered_by`, services load with explicit `owned_by`, fulfillment references resolve to `<domain>/<service>`, and CLI/REST DTOs expose the same JSON fields. Existing legacy catalog files must continue to load and should produce validation findings rather than load errors.

## Domain-owned product offering workflow

The demo cosmos now demonstrates the ADR-0001 workflow with `identity.blumer.cloud` offering `PROD-ACC-MBX-001` and fulfillment services from both `identity.blumer.cloud` and `collaboration.blumer.cloud`. Regenerate it with:

```bash
sh scripts/create-demo-cosmos.sh /tmp/nomos-demo
```

The generated product remains under `.nomos/catalog/blueprints/products`, while the Cosmos Explorer shows it below the identity domain's `Products` pseudo-node and in the global Catalog Index.
