document.addEventListener('click', (e) => {
  const btn = e.target.closest('[data-copy]');
  if (!btn) return;
  const el = document.querySelector(btn.getAttribute('data-copy'));
  if (!el) return;
  navigator.clipboard?.writeText(el.textContent || '');
});
