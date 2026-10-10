export function BrandMark({ className = 'size-8' }: { className?: string }) {
  return (
    <span className={`grid shrink-0 place-items-center rounded-field bg-primary text-primary-content ${className}`}>
      <svg viewBox="0 0 32 32" className="size-[70%]" aria-hidden="true">
        <path d="M6.5 17h4.5l2.75-6.5 4 11 2.75-4.5h5" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    </span>
  )
}

export function Brand() {
  return (
    <div className="flex items-center gap-2.5">
      <BrandMark />
      <span className="text-[15px] font-semibold tracking-tight">VPSentinel</span>
    </div>
  )
}
