# Vendored @bpmn-io/form-js (ADR-0024)

The Cosmos Explorer renders `engine: form-js` view artifacts with
[`@bpmn-io/form-js`](https://github.com/bpmn-io/form-js), vendored here like
`bpmn-js`/`dmn-js`. Until the bundle is present, the Explorer falls back to
showing the form schema as JSON.

## Expected files

Place these two files in this directory (names matter — they are referenced by
`internal/server/web/templates/cosmos.html` in `loadFormJs()`):

- `form-viewer.umd.js` — the UMD build of the form-js **viewer**
- `form-js.css` — the form-js stylesheet

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

From an environment with npm/CDN access, e.g.:

```sh
# pin a version, then copy the UMD viewer build + CSS into this folder
npm pack @bpmn-io/form-js-viewer@1   # or @bpmn-io/form-js
# extract and copy:
#   dist/<umd build>.js  -> form-viewer.umd.js
#   dist/assets/form-js.css (or dist/*.css) -> form-js.css
```

Or download directly:

```sh
curl -L -o form-viewer.umd.js "https://unpkg.com/@bpmn-io/form-js-viewer@1/dist/form-viewer.umd.js"
curl -L -o form-js.css        "https://unpkg.com/@bpmn-io/form-js@1/dist/assets/form-js.css"
```

## Notes

- The loader probes a few global names
  (`@bpmn-io/form-js-viewer`, `@bpmn-io/form-js`, `FormViewer`, `FormJS`) and
  supports both the `createForm({container, schema, data})` and
  `new Form({container}).importSchema(schema, data)` APIs. If your build exposes
  a different global, adjust `loadFormJs()` in `cosmos.html`.
- The editor (`@bpmn-io/form-js-editor`) can be vendored later as a second file
  to enable in-place form editing; the viewer is enough for rendering.
- Keep the version pinned; bump deliberately (see ADR-0024 `engine_version`).
