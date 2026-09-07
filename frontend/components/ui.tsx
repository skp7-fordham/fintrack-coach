"use client";

import { moneyTone, formatMoney, formatSignedMoney } from "@/lib/format";
import type { ReactNode } from "react";

export function MoneyDisplay({
  amount,
  currency,
  signed = false,
  className = "",
}: {
  amount: string | number;
  currency?: string | null;
  signed?: boolean;
  className?: string;
}) {
  const tone = moneyTone(amount);
  const color =
    tone === "positive"
      ? "text-success"
      : tone === "negative"
        ? "text-danger"
        : "text-foreground";
  const text = signed
    ? formatSignedMoney(amount, currency)
    : formatMoney(amount, currency);

  return <span className={`font-medium tabular-nums ${color} ${className}`}>{text}</span>;
}

export function SummaryCard({
  label,
  value,
  hint,
}: {
  label: string;
  value: ReactNode;
  hint?: string;
}) {
  return (
    <div className="rounded-xl border border-border bg-card p-4 shadow-sm">
      <p className="text-sm text-muted">{label}</p>
      <div className="mt-2 text-2xl font-semibold tracking-tight">{value}</div>
      {hint ? <p className="mt-1 text-xs text-muted">{hint}</p> : null}
    </div>
  );
}

export function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    completed: "bg-success-soft text-success",
    pending: "bg-warning-soft text-warning",
    failed: "bg-danger-soft text-danger",
    queued: "bg-slate-100 text-slate-700",
    processing: "bg-primary-soft text-primary",
    completed_with_errors: "bg-warning-soft text-warning",
    income: "bg-success-soft text-success",
    expense: "bg-danger-soft text-danger",
    transfer: "bg-slate-100 text-slate-700",
  };

  return (
    <span
      className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium capitalize ${
        styles[status] ?? "bg-slate-100 text-slate-700"
      }`}
    >
      {status.replaceAll("_", " ")}
    </span>
  );
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-start gap-3 rounded-xl border border-dashed border-border bg-card p-8">
      <h3 className="text-lg font-medium text-foreground">{title}</h3>
      <p className="max-w-lg text-sm text-muted">{description}</p>
      {action}
    </div>
  );
}

export function ErrorState({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div className="rounded-xl border border-danger/20 bg-danger-soft p-4 text-danger">
      <p className="text-sm font-medium">{message}</p>
      {onRetry ? (
        <button
          type="button"
          onClick={onRetry}
          className="mt-3 rounded-lg bg-card px-3 py-1.5 text-sm font-medium text-foreground shadow-sm"
        >
          Try again
        </button>
      ) : null}
    </div>
  );
}

export function LoadingSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="space-y-3" aria-busy="true" aria-label="Loading">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="h-16 animate-pulse rounded-xl bg-slate-200/70" />
      ))}
    </div>
  );
}

export function Modal({
  open,
  title,
  children,
  onClose,
}: {
  open: boolean;
  title: string;
  children: ReactNode;
  onClose: () => void;
}) {
  if (!open) return null;
  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-4 sm:items-center"
      role="dialog"
      aria-modal="true"
      aria-label={title}
      onClick={onClose}
    >
      <div
        className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-card p-5 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-start justify-between gap-3">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg px-2 py-1 text-sm text-muted hover:bg-slate-100"
          >
            Close
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = "Confirm",
  pending,
  onConfirm,
  onClose,
}: {
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  pending?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}) {
  return (
    <Modal open={open} title={title} onClose={onClose}>
      <p className="text-sm text-muted">{message}</p>
      <div className="mt-5 flex justify-end gap-2">
        <button
          type="button"
          onClick={onClose}
          className="rounded-lg border border-border px-3 py-2 text-sm"
          disabled={pending}
        >
          Cancel
        </button>
        <button
          type="button"
          onClick={onConfirm}
          disabled={pending}
          className="rounded-lg bg-danger px-3 py-2 text-sm font-medium text-white disabled:opacity-60"
        >
          {pending ? "Working…" : confirmLabel}
        </button>
      </div>
    </Modal>
  );
}

export function FormField({
  label,
  htmlFor,
  children,
  hint,
}: {
  label: string;
  htmlFor: string;
  children: ReactNode;
  hint?: string;
}) {
  return (
    <label className="block space-y-1.5" htmlFor={htmlFor}>
      <span className="text-sm font-medium text-foreground">{label}</span>
      {children}
      {hint ? <span className="block text-xs text-muted">{hint}</span> : null}
    </label>
  );
}

export function inputClassName() {
  return "w-full rounded-lg border border-border bg-white px-3 py-2 text-sm text-foreground shadow-sm placeholder:text-slate-400";
}

export function primaryButtonClassName() {
  return "inline-flex items-center justify-center rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary-hover disabled:opacity-60";
}

export function secondaryButtonClassName() {
  return "inline-flex items-center justify-center rounded-lg border border-border bg-card px-4 py-2 text-sm font-medium text-foreground hover:bg-slate-50 disabled:opacity-60";
}

export function Pagination({
  page,
  totalPages,
  onPageChange,
}: {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}) {
  if (totalPages <= 1) return null;
  return (
    <div className="flex items-center justify-between gap-3 pt-4">
      <p className="text-sm text-muted">
        Page {page} of {totalPages}
      </p>
      <div className="flex gap-2">
        <button
          type="button"
          className={secondaryButtonClassName()}
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          Previous
        </button>
        <button
          type="button"
          className={secondaryButtonClassName()}
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
        >
          Next
        </button>
      </div>
    </div>
  );
}

export function CoachMessageBubble({
  role,
  content,
}: {
  role: string;
  content: string;
}) {
  const isUser = role === "user";
  return (
    <div className={`flex ${isUser ? "justify-end" : "justify-start"}`}>
      <div
        className={`max-w-[85%] rounded-2xl px-4 py-3 text-sm leading-relaxed whitespace-pre-wrap ${
          isUser
            ? "bg-primary text-white"
            : "border border-border bg-card text-foreground shadow-sm"
        }`}
      >
        {!isUser ? (
          <p className="mb-2 text-[11px] font-medium uppercase tracking-wide text-muted">
            Based on your FinTrack data
          </p>
        ) : null}
        {content}
      </div>
    </div>
  );
}
