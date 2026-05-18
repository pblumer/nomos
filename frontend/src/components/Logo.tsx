type LogoProps = {
  size?: number;
  className?: string;
  title?: string;
};

export function Logo({ size = 32, className, title }: LogoProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      role={title ? "img" : undefined}
      aria-label={title}
      aria-hidden={title ? undefined : true}
    >
      {title ? <title>{title}</title> : null}
      <defs>
        <linearGradient id="nomos-logo-bg" x1="0" y1="0" x2="32" y2="32" gradientUnits="userSpaceOnUse">
          <stop offset="0" stopColor="#6366F1" />
          <stop offset="1" stopColor="#7C3AED" />
        </linearGradient>
        <linearGradient id="nomos-logo-shine" x1="16" y1="0" x2="16" y2="32" gradientUnits="userSpaceOnUse">
          <stop offset="0" stopColor="#FFFFFF" stopOpacity="0.28" />
          <stop offset="0.6" stopColor="#FFFFFF" stopOpacity="0" />
        </linearGradient>
        <linearGradient id="nomos-logo-node" x1="13.6" y1="13.6" x2="18.4" y2="18.4" gradientUnits="userSpaceOnUse">
          <stop offset="0" stopColor="#A5B4FC" />
          <stop offset="1" stopColor="#7C3AED" />
        </linearGradient>
      </defs>

      <rect x="0" y="0" width="32" height="32" rx="8" fill="url(#nomos-logo-bg)" />
      <rect x="0" y="0" width="32" height="32" rx="8" fill="url(#nomos-logo-shine)" />
      <rect x="0.5" y="0.5" width="31" height="31" rx="7.5" stroke="#FFFFFF" strokeOpacity="0.12" />

      <g fill="#FFFFFF">
        <rect x="8" y="7" width="3.2" height="18" rx="1.2" />
        <rect x="20.8" y="7" width="3.2" height="18" rx="1.2" />
        <path d="M8 7 L11.2 7 L24 25 L20.8 25 Z" />
      </g>

      <circle cx="16" cy="16" r="2.6" fill="url(#nomos-logo-node)" stroke="#FFFFFF" strokeWidth="1.2" />
    </svg>
  );
}
