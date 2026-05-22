document.addEventListener('click', (e) => {
  const toggle = e.target.closest('[data-sidebar-toggle]');
  if (toggle) {
    const order = ['expanded', 'collapsed', 'hidden'];
    const root = document.documentElement;
    const current = root.classList.contains('sidebar-hidden')
      ? 'hidden'
      : root.classList.contains('sidebar-collapsed')
        ? 'collapsed'
        : 'expanded';
    const next = order[(order.indexOf(current) + 1) % order.length];
    root.classList.toggle('sidebar-collapsed', next === 'collapsed');
    root.classList.toggle('sidebar-hidden', next === 'hidden');
    try {
      localStorage.setItem('nomos-sidebar', next);
    } catch (err) {}
    return;
  }
  const btn = e.target.closest('[data-copy]');
  if (!btn) return;
  const el = document.querySelector(btn.getAttribute('data-copy'));
  if (!el) return;
  navigator.clipboard?.writeText(el.textContent || '');
  btn.textContent = 'Copied';
  setTimeout(() => (btn.textContent = 'Copy'), 1200);
});
document.addEventListener('input', (e) => {
  const input = e.target.closest('[data-filter-target]');
  if (!input) return;
  const q = input.value.toLowerCase();
  document.querySelectorAll(input.dataset.filterTarget).forEach((el) => {
    el.style.display = el.textContent.toLowerCase().includes(q) ? '' : 'none';
  });
});

document.addEventListener('input', (e) => {
  const form = e.target.closest('[data-child-domain-form]');
  if (form) updateChildDomainPreview(form);
  const serviceForm = e.target.closest('[data-service-form]');
  if (serviceForm) updateServicePreview(serviceForm);
});
document.addEventListener('DOMContentLoaded', () => {
  document.querySelectorAll('[data-child-domain-form]').forEach(updateChildDomainPreview);
  document.querySelectorAll('[data-service-form]').forEach(updateServicePreview);
});
function updateChildDomainPreview(form) {
  const parent = form.querySelector('[data-parent-domain]')?.value.trim() || '';
  const segmentInput = form.querySelector('[data-segment-input]');
  const preview = form.querySelector('[data-domain-preview]');
  const warning = form.querySelector('[data-segment-warning]');
  const segment = segmentInput?.value.trim() || '';
  const messages = [];
  if (!segment) messages.push('Use only the new segment, for example: test2.');
  if (segment.includes('.')) messages.push('You entered a full domain name. In this form, enter only the new segment. Use Advanced mode if you want to create a full canonical name.');
  if (/\s/.test(segment)) messages.push('Spaces are not allowed.');
  if (segment.includes('/')) messages.push('Slashes are not allowed.');
  if (segment.startsWith('.') || segment.endsWith('.')) messages.push('Do not use a leading or trailing dot.');
  if (segment && !/^[a-z0-9-]+$/.test(segment)) messages.push('Lowercase letters, numbers, and hyphens are preferred.');
  if (warning) warning.textContent = messages.join(' ');
  if (segmentInput) segmentInput.setCustomValidity(messages.some((m) => !m.includes('preferred')) ? messages[0] : '');
  if (preview) preview.value = segment && parent ? `${segment}.${parent}` : `new-segment.${parent || 'parent.example'}`;
}
function updateServicePreview(form) {
  const parent = form.querySelector('[data-service-parent]')?.value.trim() || 'parent.example';
  const name = form.querySelector('[data-service-name]')?.value.trim() || 'service-name';
  const preview = form.querySelector('[data-service-preview]');
  if (preview) preview.value = `${parent} / services / ${name}`;
}

// Inline source viewer/editor for YAML, JSON and Markdown artifact files.
// Each [data-source-editor] element carries data-path and data-lang and is
// backed by the vendored CodeMirror 6 bundle plus the /api/v1/source endpoints.
function initSourceEditors() {
  const editors = document.querySelectorAll('[data-source-editor]');
  if (!editors.length) return;
  loadCodeMirror().then(() => editors.forEach(setupSourceEditor));
}

let cmLoader = null;
function loadCodeMirror() {
  if (window.CM6) return Promise.resolve();
  if (cmLoader) return cmLoader;
  cmLoader = new Promise((resolve, reject) => {
    const s = document.createElement('script');
    s.src = '/static/vendor/codemirror/codemirror.js';
    s.onload = () => resolve();
    s.onerror = () => reject(new Error('failed to load editor bundle'));
    document.head.appendChild(s);
  });
  return cmLoader;
}

// Like fetch(), but reports the /api/v1 request+response to the Cosmos REST
// inspector when present (window.__nomosRecordApiCall). On pages without the
// inspector it behaves exactly like a plain fetch.
async function trackedFetch(url, init) {
  const record = window.__nomosRecordApiCall;
  if (!record) return fetch(url, init);
  const method = ((init && init.method) || 'GET').toUpperCase();
  const reqBody = init && init.body;
  const started = performance.now();
  let res;
  try {
    res = await fetch(url, init);
  } catch (err) {
    record({ method, url, reqBody, error: err.message, durationMs: performance.now() - started });
    throw err;
  }
  let text = '';
  try {
    text = await res.clone().text();
  } catch (_) {}
  record({ method, url, reqBody, status: res.status, ok: res.ok, body: text, durationMs: performance.now() - started });
  return res;
}

function setupSourceEditor(root) {
  if (root.dataset.sourceReady) return;
  root.dataset.sourceReady = '1';
  const path = root.dataset.path;
  const language = root.dataset.lang || 'text';
  if (!path) return;

  const host = root.querySelector('[data-editor-host]');
  const preview = root.querySelector('[data-editor-preview]');
  const status = root.querySelector('[data-editor-status]');
  const findings = root.querySelector('[data-editor-findings]');
  const saveBtn = root.querySelector('[data-editor-save]');
  const previewBtn = root.querySelector('[data-editor-preview-toggle]');
  let editor = null;
  let dirty = false;

  const setStatus = (msg, kind) => {
    if (!status) return;
    status.textContent = msg || '';
    status.className = 'editor-status' + (kind ? ' editor-status-' + kind : '');
  };

  trackedFetch(`/api/v1/source?path=${encodeURIComponent(path)}`)
    .then((r) => r.json().then((b) => ({ ok: r.ok, body: b })))
    .then(({ ok, body }) => {
      if (!ok) throw new Error(body.error || 'load failed');
      editor = window.CM6.create(host, {
        doc: body.content || '',
        language,
        onChange: () => {
          dirty = true;
          setStatus('Unsaved changes', 'warn');
        },
      });
      setStatus('Loaded', 'ok');
    })
    .catch((err) => setStatus('Could not load source: ' + err.message, 'error'));

  if (saveBtn) {
    saveBtn.addEventListener('click', () => {
      if (!editor) return;
      setStatus('Saving…', '');
      if (findings) findings.innerHTML = '';
      trackedFetch(`/api/v1/source?path=${encodeURIComponent(path)}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: editor.getValue() }),
      })
        .then((r) => r.json().then((b) => ({ status: r.status, body: b })))
        .then(({ status: code, body }) => {
          if (code === 422) {
            setStatus('Syntax error — not saved', 'error');
            renderFindings(findings, [{ severity: 'error', message: body.syntaxError }]);
            return;
          }
          if (!body.ok) {
            setStatus(body.error || 'Save failed', 'error');
            return;
          }
          dirty = false;
          // Scope findings to the file just saved: a per-file editor should not
          // surface unrelated cosmos-wide findings (e.g. a product blueprint's
          // errors while editing cosmos.yaml).
          const all = body.validation && body.validation.findings ? body.validation.findings : [];
          const list = all.filter((f) => findingRelevantToPath(f.path, path));
          setStatus('Saved' + (list.length ? ` · ${list.length} finding(s)` : ''), list.length ? 'warn' : 'ok');
          renderFindings(findings, list);
        })
        .catch((err) => setStatus('Save failed: ' + err.message, 'error'));
    });
  }

  if (previewBtn && preview) {
    previewBtn.addEventListener('click', () => {
      if (!editor) return;
      const showing = preview.hidden === false;
      if (showing) {
        preview.hidden = true;
        host.hidden = false;
        previewBtn.textContent = 'Preview';
        return;
      }
      trackedFetch('/api/v1/render/markdown', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: editor.getValue() }),
      })
        .then((r) => r.json())
        .then((b) => {
          preview.innerHTML = b.html || '';
          preview.hidden = false;
          host.hidden = true;
          previewBtn.textContent = 'Edit';
        })
        .catch((err) => setStatus('Preview failed: ' + err.message, 'error'));
    });
  }

  window.addEventListener('beforeunload', (e) => {
    if (dirty) {
      e.preventDefault();
      e.returnValue = '';
    }
  });
}

// findingRelevantToPath decides whether a validation finding belongs to the
// file currently open in the editor. A finding is relevant when its path equals
// the edited file, sits in the same artifact directory, or one path nests under
// the other. Findings without a path (cosmos-global) are not shown per-file.
function findingRelevantToPath(findingPath, editPath) {
  if (!findingPath || !editPath) return false;
  const norm = (p) => p.replace(/\\/g, '/').replace(/^\.?\//, '').replace(/\/+$/, '');
  const a = norm(findingPath);
  const b = norm(editPath);
  if (a === b) return true;
  if (b.startsWith(a + '/') || a.startsWith(b + '/')) return true;
  const dir = (p) => (p.includes('/') ? p.slice(0, p.lastIndexOf('/')) : '');
  return dir(a) !== '' && dir(a) === dir(b);
}

function renderFindings(container, findings) {
  if (!container) return;
  container.innerHTML = '';
  if (!findings || !findings.length) return;
  const list = document.createElement('ul');
  list.className = 'editor-finding-list';
  findings.forEach((f) => {
    const li = document.createElement('li');
    li.className = 'editor-finding editor-finding-' + (f.severity || 'info');
    const sev = document.createElement('span');
    sev.className = 'editor-finding-severity';
    sev.textContent = (f.severity || 'info').toUpperCase();
    li.appendChild(sev);
    li.appendChild(document.createTextNode(' ' + (f.message || '')));
    list.appendChild(li);
  });
  container.appendChild(list);
}

document.addEventListener('DOMContentLoaded', initSourceEditors);
