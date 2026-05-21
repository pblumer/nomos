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
