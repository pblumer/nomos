# development-lifecycle

## Go Web UI in MVP

The MVP web interface remains read-only and is served by the Go binary (`nomos serve`) with embedded templates/assets. This keeps delivery simple while preserving a consistent Nomos UI style without introducing a second runtime build pipeline.


## Source repo vs. Cosmos workspace

Development happens in the Nomos source repository (`cmd/`, `internal/`, `docs/`, `examples/`, `scripts/`). Concrete Cosmos data must not be committed at the source repository root. Use `nomos cosmos init <path>` to create a separate workspace whose mutable state is stored under `<path>/.nomos/`.

Recommended ignores:

- Dedicated Cosmos repository: `.nomos/cache/` and `.nomos/index/`.
- Local-only Nomos data in another repository: `.nomos/`.
