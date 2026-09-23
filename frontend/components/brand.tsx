type BrandMarkProps = {
  className?: string;
  inverse?: boolean;
};

export function BrandMark({ className = "h-9 w-9", inverse = false }: BrandMarkProps) {
  const frameColor = inverse ? "#cbd5e1" : "#ccfbf1";
  const signalColor = inverse ? "#5eead4" : "#ffffff";

  return (
    <span
      className={`inline-flex shrink-0 items-center justify-center rounded-lg ${
        inverse ? "bg-teal-400/10 ring-1 ring-inset ring-teal-300/20" : "bg-primary"
      } ${className}`}
      aria-hidden="true"
    >
      <svg viewBox="0 0 32 32" className="h-[72%] w-[72%]" fill="none">
        <path
          d="M8.75 8.5h14.5a2.25 2.25 0 0 1 2.25 2.25v10.5a2.25 2.25 0 0 1-2.25 2.25H8.75a2.25 2.25 0 0 1-2.25-2.25v-10.5A2.25 2.25 0 0 1 8.75 8.5Z"
          stroke={frameColor}
          strokeWidth="1.75"
          opacity={inverse ? 0.72 : 0.8}
        />
        <path
          d="m10 20 4.2-4.1 3.1 2.45 4.7-5.85"
          stroke={signalColor}
          strokeWidth="2.35"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M18.9 12.5H22v3.1"
          stroke={signalColor}
          strokeWidth="2.35"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </span>
  );
}

export function BrandLockup({
  inverse = false,
  compact = false,
}: {
  inverse?: boolean;
  compact?: boolean;
}) {
  return (
    <div className="flex min-w-0 items-center gap-3">
      <BrandMark className={compact ? "h-8 w-8" : "h-9 w-9"} inverse={inverse} />
      <div className="min-w-0 leading-tight">
        <p className={`text-sm font-semibold tracking-tight ${inverse ? "text-white" : "text-foreground"}`}>
          FinTrack Coach
        </p>
        {!compact ? (
          <p className={`mt-1 text-xs leading-4 ${inverse ? "text-sidebar-muted" : "text-muted"}`}>
            Spending insights and AI guidance
          </p>
        ) : null}
      </div>
    </div>
  );
}
