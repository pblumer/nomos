# Vendored @bpmn-io/form-js (ADR-0024)

The Cosmos Explorer renders `engine: form-js` view artifacts with
[`@bpmn-io/form-js`](https://github.com/bpmn-io/form-js), vendored here like
`bpmn-js`/`dmn-js`. Until the viewer bundle is present, the Explorer falls back
to showing the form schema as JSON.

## Expected files

Place these four files in this directory (names matter — they are referenced by
`internal/server/web/templates/cosmos.html`):

- `form-viewer.umd.js` — the UMD build of the form-js **viewer**
- `form-js.css` — the viewer stylesheet
- `form-editor.umd.js` — the UMD build of the form-js **editor**
- `form-js-editor.css` — the editor stylesheet

For in-place editing (ADR-0024 step 1), also add the **editor** build:

- `form-editor.umd.js` — the UMD build of the form-js **editor**
- `form-js-editor.css` — the form-js editor stylesheet

```sh
curl -L -o form-editor.umd.js "https://unpkg.com/@bpmn-io/form-js@1/dist/form-editor.umd.js"
curl -L -o form-js-editor.css "https://unpkg.com/@bpmn-io/form-js@1/dist/assets/form-js-editor.css"
```

The editor loader (`loadFormJsEditor()` in `cosmos.html`) probes the globals
`@bpmn-io/form-js-editor` / `@bpmn-io/form-js` / `FormEditor` / `FormJSEditor`
and supports both `createFormEditor({container, schema})` and
`new FormEditor({container}).importSchema(schema)`; on save it reads
`getSchema()`/`saveSchema()`. Adjust if your build differs.

## How to vendor

From an environment with npm/CDN access (all four files come from the same
`@bpmn-io/form-js` umbrella package):

```sh
cd internal/server/web/static/vendor/form-js

curl -L -o form-viewer.umd.js "https://unpkg.com/@bpmn-io/form-js@1/dist/form-viewer.umd.js"
curl -L -o form-js.css        "https://unpkg.com/@bpmn-io/form-js@1/dist/assets/form-js.css"
curl -L -o form-editor.umd.js "https://unpkg.com/@bpmn-io/form-js@1/dist/form-editor.umd.js"
curl -L -o form-js-editor.css "https://unpkg.com/@bpmn-io/form-js@1/dist/assets/form-js-editor.css"
```

Then `make build` (the embed picks up the files at build time).

## Notes

- **Viewer** global: `window.FormViewer`; exports `createForm` + `Form`.
  The loader calls `createForm({container, schema, data})`.
- **Editor** global: `window.FormEditor`; exports `createFormEditor` + `FormEditor`.
  The loader calls `createFormEditor({container, schema})`, then `editor.getSchema()`
  on save.
- The editor (`form-editor.umd.js`) is only loaded on demand when the user clicks
  "Schema bearbeiten" in the user-interface detail panel — the viewer is always
  loaded for rendering.
- Keep the version pinned; bump deliberately (see ADR-0024 `engine_version`).
