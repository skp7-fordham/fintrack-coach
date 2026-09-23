"use client";

import { moneyTone, formatMoney, formatSignedMoney } from "@/lib/format";
import { useEffect, useId, type ReactNode } from "react";

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

  return <span className={`font-semibold tabular-nums tracking-tight ${color} ${className}`}>{text}</span>;
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
    <div className="rounded-lg border border-border bg-card p-4">
      <p className="text-xs font-medium text-muted">{label}</p>
      <div className="mt-2 text-2xl font-semibold tracking-[-0.025em]">{value}</div>
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
  const labels: Record<string, string> = {
    completed: "Completed",
    pending: "Pending",
    failed: "Failed",
    queued: "Queued",
    processing: "Processing",
    completed_with_errors: "Completed with errors",
    income: "Income",
    expense: "Expense",
    transfer: "Transfer",
  };

  return (
    <span
      className={`inline-flex whitespace-nowrap rounded-full px-2.5 py-1 text-[11px] font-medium leading-none ${
        styles[status] ?? "bg-slate-100 text-slate-700"
      }`}
    >
      {labels[status] ?? status.replaceAll("_", " ")}
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
    <div className="flex flex-col items-start rounded-lg border border-dashed border-slate-300 bg-slate-50/50 p-6 sm:p-8">
      <h3 className="text-base font-semibold text-foreground">{title}</h3>
      <p className="mt-1.5 max-w-lg text-sm leading-6 text-muted">{description}</p>
      {action ? <div className="mt-4">{action}</div> : null}
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
    <div className="rounded-lg border border-danger/20 bg-danger-soft p-4 text-danger">
      <p className="text-sm font-medium">{message}</p>
      {onRetry ? (
        <button
          type="button"
          onClick={onRetry}
          className="mt-3 rounded-md border border-danger/15 bg-card px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-white"
        >
          Try again
        </button>
      ) : null}
    </div>
  );
}

export function DemoReadOnlyNote({ className = "" }: { className?: string }) {
  return (
    <p className={`text-xs font-medium text-muted ${className}`}>
      Demo account is read-only
    </p>
  );
}

export function LoadingSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="space-y-3" aria-busy="true" aria-label="Loading">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="h-16 animate-pulse rounded-lg bg-slate-200/70" />
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
  const titleId = useId();

  useEffect(() => {
    if (!open) return;
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [open, onClose]);

  if (!open) return null;
  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-4 sm:items-center"
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      onClick={onClose}
    >
      <div
        className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-xl border border-border bg-card p-5 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-start justify-between gap-3">
          <h2 id={titleId} className="text-lg font-semibold">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md px-2 py-1 text-sm text-muted transition-colors hover:bg-slate-100 hover:text-foreground"
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
          className="rounded-md border border-border px-3 py-2 text-sm transition-colors hover:bg-slate-50"
          disabled={pending}
        >
          Cancel
        </button>
        <button
          type="button"
          onClick={onConfirm}
          disabled={pending}
          className="rounded-md bg-danger px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {pending ? "Deleting…" : confirmLabel}
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
  return "w-full rounded-md border border-border bg-card px-3 py-2 text-sm text-foreground transition-colors placeholder:text-slate-400 hover:border-slate-300 disabled:cursor-not-allowed disabled:bg-slate-50 disabled:text-muted";
}

export function primaryButtonClassName() {
  return "inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-60";
}

export function secondaryButtonClassName() {
  return "inline-flex items-center justify-center rounded-md border border-border bg-card px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-slate-300 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60";
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

function renderInlineMarkdown(text: string): ReactNode[] {
  return text
    .split(/(\*\*[^*\n]+?\*\*)/g)
    .filter(Boolean)
    .map((part, index) =>
      part.startsWith("**") && part.endsWith("**") ? (
        <strong key={index} className="font-semibold text-foreground">
          {part.slice(2, -2)}
        </strong>
      ) : (
        <span key={index}>{part}</span>
      ),
    );
}

function AssistantMessageContent({ content }: { content: string }) {
  const lines = content.replaceAll("\r\n", "\n").split("\n");

  return (
    <div className="space-y-1.5">
      {lines.map((line, index) => {
        if (!line.trim()) {
          return <div key={index} className="h-1" aria-hidden="true" />;
        }

        const bullet = line.match(/^\s*[-*]\s+(.+)$/);
        if (bullet) {
          return (
            <div key={index} className="flex gap-2">
              <span className="text-primary" aria-hidden="true">•</span>
              <span>{renderInlineMarkdown(bullet[1])}</span>
            </div>
          );
        }

        const numbered = line.match(/^\s*(\d+\.)\s+(.+)$/);
        if (numbered) {
          return (
            <div key={index} className="flex gap-2">
              <span className="shrink-0 font-medium text-muted">{numbered[1]}</span>
              <span>{renderInlineMarkdown(numbered[2])}</span>
            </div>
          );
        }

        return <p key={index}>{renderInlineMarkdown(line)}</p>;
      })}
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
        className={`max-w-[88%] px-4 py-3 text-sm leading-relaxed ${
          isUser
            ? "rounded-[16px_16px_4px_16px] bg-primary text-white whitespace-pre-wrap"
            : "rounded-[16px_16px_16px_4px] border border-border bg-slate-50/60 text-foreground"
        }`}
      >
        {!isUser ? (
          <p className="mb-2 text-xs font-medium text-muted">
            Based on your FinTrack data
          </p>
        ) : null}
        {isUser ? content : <AssistantMessageContent content={content} />}
      </div>
    </div>
  );
}
