import type { ReactNode } from "react";

export function AuthShell({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle: string;
  children: ReactNode;
}) {
  return (
    <div className="relative flex min-h-screen items-center justify-center px-4 py-10">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,#99f6e4_0%,transparent_40%),radial-gradient(circle_at_bottom_right,#bae6fd_0%,transparent_35%)]" />
      <div className="relative w-full max-w-md rounded-2xl border border-border bg-card p-8 shadow-lg">
        <div className="mb-8 flex items-center gap-3">
          <span className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-sm font-bold text-white">
            F
          </span>
          <div>
            <p className="text-sm font-semibold text-foreground">FinTrack Coach</p>
            <p className="text-xs text-muted">Money clarity with an AI coach</p>
          </div>
        </div>
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        <p className="mt-1 text-sm text-muted">{subtitle}</p>
        <div className="mt-6">{children}</div>
      </div>
    </div>
  );
}
