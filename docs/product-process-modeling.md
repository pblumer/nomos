# Product process modeling with BPMN

Nomos product offerings can reference product-level process artifacts under `.nomos/catalog/blueprints/processes`.
The metadata stays in YAML and the BPMN diagram stays in a sibling `.bpmn` XML file so reviews remain Git-friendly.

```text
.nomos/catalog/blueprints/products/PROD-0000001.yaml
.nomos/catalog/blueprints/processes/PRC-0000001.yaml
.nomos/catalog/blueprints/processes/PRC-0000001.bpmn
```

The Cosmos Explorer product showcase exposes a **Process** tab. If `bpmn-js` browser assets are available at
`/static/vendor/bpmn-js/bpmn-viewer.production.min.js`, the tab renders the XML with bpmn.io's `bpmn-js` viewer and adds
small task mapping overlays. In constrained/offline development environments where the npm package cannot be fetched, the
same API and mapping UI remain usable and the tab falls back to showing the BPMN XML text instead of a custom renderer.

The vendored browser bundles are committed under `internal/server/web/static/vendor/` and embedded into the
`nomos` binary, so no Node toolchain is required at build or runtime. To refresh them from upstream in a connected
environment, fetch the package into a throwaway directory and copy the dist files in:

```bash
tmp="$(mktemp -d)"
npm --prefix "$tmp" install bpmn-js@^18
mkdir -p internal/server/web/static/vendor/bpmn-js
cp "$tmp"/node_modules/bpmn-js/dist/bpmn-viewer.production.min.js internal/server/web/static/vendor/bpmn-js/
cp "$tmp"/node_modules/bpmn-js/dist/bpmn-modeler.production.min.js internal/server/web/static/vendor/bpmn-js/
rm -rf "$tmp"
```

## API quick reference

- `GET /api/v1/products/{productId}/processes`
- `POST /api/v1/products/{productId}/processes`
- `GET /api/v1/processes/{processId}`
- `GET /api/v1/processes/{processId}/bpmn`
- `PUT /api/v1/processes/{processId}/bpmn`
- `GET /api/v1/processes/{processId}/tasks`
- `PUT /api/v1/processes/{processId}/task-mappings`

Task mappings connect BPMN task IDs to domain-owned fulfillment service references. Nomos stores this as descriptive
product/process knowledge only; it is not a workflow execution engine.
