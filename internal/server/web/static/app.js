document.addEventListener('click', (e) => {
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
