// Entry point bundled into a single IIFE exposed as window.CM6.
// Rebuild with `make assets-codemirror` after changing dependencies.
import { EditorView, basicSetup } from "codemirror";
import { EditorState, Compartment } from "@codemirror/state";
import { keymap } from "@codemirror/view";
import { indentWithTab } from "@codemirror/commands";
import { json } from "@codemirror/lang-json";
import { yaml } from "@codemirror/lang-yaml";
import { markdown } from "@codemirror/lang-markdown";

const languages = { json, yaml, markdown };

function languageExtension(lang) {
  const factory = languages[lang];
  return factory ? [factory()] : [];
}

// create mounts an editor in `parent`. opts: { doc, language, readOnly, onChange }.
function create(parent, opts = {}) {
  const { doc = "", language = "text", readOnly = false, onChange } = opts;
  const extensions = [
    basicSetup,
    keymap.of([indentWithTab]),
    ...languageExtension(language),
    EditorView.lineWrapping,
    EditorState.readOnly.of(!!readOnly),
  ];
  if (typeof onChange === "function") {
    extensions.push(
      EditorView.updateListener.of((u) => {
        if (u.docChanged) onChange(u.state.doc.toString());
      }),
    );
  }
  const view = new EditorView({
    state: EditorState.create({ doc, extensions }),
    parent,
  });
  return {
    view,
    getValue: () => view.state.doc.toString(),
    setValue: (text) =>
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: text },
      }),
    destroy: () => view.destroy(),
  };
}

window.CM6 = { create, EditorView, EditorState, Compartment };
