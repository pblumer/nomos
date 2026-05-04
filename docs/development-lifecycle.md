# development-lifecycle

## Go Web UI in MVP

The MVP web interface remains read-only and is served by the Go binary (`nomos serve`) with embedded templates/assets. This keeps delivery simple while preserving a consistent Nomos UI style without introducing a second runtime build pipeline.
