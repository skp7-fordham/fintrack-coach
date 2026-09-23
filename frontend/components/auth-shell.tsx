import type { ReactNode } from "react";
import { BrandLockup } from "@/components/brand";

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
    <div className="flex min-h-screen items-center justify-center bg-background px-4 py-10">
      <div className="w-full max-w-md overflow-hidden rounded-xl border border-border bg-card shadow-[0_18px_50px_-30px_rgba(15,23,42,0.45)]">
        <div className="border-b border-border bg-slate-50/70 px-7 py-5">
          <BrandLockup />
        </div>
        <div className="px-7 py-7 sm:px-8">
          <h1 className="text-2xl font-semibold tracking-[-0.025em]">{title}</h1>
          <p className="mt-1.5 text-sm leading-6 text-muted">{subtitle}</p>
          <div className="mt-6">{children}</div>
        </div>
      </div>
    </div>
  );
}
