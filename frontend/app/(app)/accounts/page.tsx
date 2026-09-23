"use client";

import { FormEvent, useCallback, useState } from "react";
import { Banknote, Landmark, PiggyBank, WalletCards } from "lucide-react";
import { AppHeader } from "@/components/app-shell";
import { getErrorMessage } from "@/components/auth-provider";
import {
  ConfirmDialog,
  EmptyState,
  ErrorState,
  FormField,
  LoadingSkeleton,
  Modal,
  MoneyDisplay,
  inputClassName,
  primaryButtonClassName,
  secondaryButtonClassName,
} from "@/components/ui";
import { api } from "@/lib/api";
import { useDeferredLoad } from "@/lib/use-deferred-load";
import { formatAccountType } from "@/lib/format";
import { ACCOUNT_TYPES, type Account } from "@/lib/types";

const emptyCreate = {
  name: "",
  account_type: "checking",
  currency: "USD",
  initial_balance: "0.00",
};

function AccountIcon({ type }: { type: string }) {
  const iconClass = "h-5 w-5";
  switch (type) {
    case "savings":
      return <PiggyBank className={iconClass} strokeWidth={1.8} aria-hidden />;
    case "cash":
      return <Banknote className={iconClass} strokeWidth={1.8} aria-hidden />;
    case "checking":
      return <Landmark className={iconClass} strokeWidth={1.8} aria-hidden />;
    default:
      return <WalletCards className={iconClass} strokeWidth={1.8} aria-hidden />;
  }
}

export default function AccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [editAccount, setEditAccount] = useState<Account | null>(null);
  const [deleteAccount, setDeleteAccount] = useState<Account | null>(null);
  const [form, setForm] = useState(emptyCreate);
  const [editForm, setEditForm] = useState({ name: "", account_type: "", currency: "" });
  const [pending, setPending] = useState(false);
  const [formError, setFormError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await api.listAccounts();
      setAccounts(res.data ?? []);
    } catch (err) {
      setError(getErrorMessage(err, "We couldn’t load your accounts."));
    } finally {
      setLoading(false);
    }
  }, []);

  useDeferredLoad(load, [load]);

  async function onCreate(e: FormEvent) {
    e.preventDefault();
    setFormError("");
    setPending(true);
    try {
      await api.createAccount({
        name: form.name.trim(),
        account_type: form.account_type,
        currency: form.currency.trim().toUpperCase(),
        initial_balance: form.initial_balance || "0.00",
      });
      setCreateOpen(false);
      setForm(emptyCreate);
      await load();
    } catch (err) {
      setFormError(getErrorMessage(err));
    } finally {
      setPending(false);
    }
  }

  async function onEdit(e: FormEvent) {
    e.preventDefault();
    if (!editAccount) return;
    setFormError("");
    setPending(true);
    try {
      await api.updateAccount(editAccount.id, {
        name: editForm.name.trim(),
        account_type: editForm.account_type,
        currency: editForm.currency.trim().toUpperCase(),
      });
      setEditAccount(null);
      await load();
    } catch (err) {
      setFormError(getErrorMessage(err));
    } finally {
      setPending(false);
    }
  }

  async function onDelete() {
    if (!deleteAccount) return;
    setPending(true);
    setFormError("");
    try {
      await api.deleteAccount(deleteAccount.id);
      setDeleteAccount(null);
      await load();
    } catch (err) {
      setFormError(getErrorMessage(err));
    } finally {
      setPending(false);
    }
  }

  return (
    <div>
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between [&>header]:mb-0">
        <AppHeader
          title="Accounts"
          subtitle="Manage the accounts that make up your financial picture."
        />
        <button
          type="button"
          className={primaryButtonClassName()}
          onClick={() => {
            setFormError("");
            setForm(emptyCreate);
            setCreateOpen(true);
          }}
        >
          Add account
        </button>
      </div>

      {error ? <ErrorState message={error} onRetry={load} /> : null}
      {loading ? <LoadingSkeleton rows={4} /> : null}

      {!loading && !error && accounts.length === 0 ? (
        <EmptyState
          title="No accounts yet"
          description="Add an account to start tracking balances and transactions."
          action={
            <button type="button" className={primaryButtonClassName()} onClick={() => setCreateOpen(true)}>
              Add account
            </button>
          }
        />
      ) : null}

      {!loading && accounts.length > 0 ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {accounts.map((account) => (
            <article key={account.id} className="rounded-lg border border-border bg-card p-5">
              <div className="flex items-start justify-between gap-3">
                <div className="flex min-w-0 items-center gap-3">
                  <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-primary-soft text-primary">
                    <AccountIcon type={account.account_type} />
                  </span>
                  <div className="min-w-0">
                  <h2 className="text-lg font-semibold">{account.name}</h2>
                  <p className="text-sm text-muted">
                    {formatAccountType(account.account_type)} · {account.currency}
                  </p>
                  </div>
                </div>
              </div>
              <p className="mt-5 text-xs font-medium text-muted">Current balance</p>
              <p className="mt-1 text-2xl font-semibold">
                <MoneyDisplay amount={account.current_balance} currency={account.currency} />
              </p>
              <div className="mt-5 flex gap-2">
                <button
                  type="button"
                  className={secondaryButtonClassName()}
                  onClick={() => {
                    setFormError("");
                    setEditAccount(account);
                    setEditForm({
                      name: account.name,
                      account_type: account.account_type,
                      currency: account.currency,
                    });
                  }}
                >
                  Edit
                </button>
                <button
                  type="button"
                  className="rounded-md border border-danger/30 px-4 py-2 text-sm font-medium text-danger transition-colors hover:bg-danger-soft"
                  onClick={() => {
                    setFormError("");
                    setDeleteAccount(account);
                  }}
                >
                  Delete
                </button>
              </div>
            </article>
          ))}
        </div>
      ) : null}

      <Modal open={createOpen} title="Add account" onClose={() => setCreateOpen(false)}>
        <form onSubmit={onCreate} className="space-y-4">
          <FormField label="Name" htmlFor="name">
            <input id="name" required className={inputClassName()} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          </FormField>
          <FormField label="Account type" htmlFor="account_type">
            <select id="account_type" className={inputClassName()} value={form.account_type} onChange={(e) => setForm({ ...form, account_type: e.target.value })}>
              {ACCOUNT_TYPES.map((type) => (
                <option key={type.value} value={type.value}>{type.label}</option>
              ))}
            </select>
          </FormField>
          <FormField label="Currency" htmlFor="currency" hint="3-letter ISO code">
            <input id="currency" required maxLength={3} className={inputClassName()} value={form.currency} onChange={(e) => setForm({ ...form, currency: e.target.value })} />
          </FormField>
          <FormField label="Initial balance" htmlFor="initial_balance">
            <input id="initial_balance" required className={inputClassName()} value={form.initial_balance} onChange={(e) => setForm({ ...form, initial_balance: e.target.value })} />
          </FormField>
          {formError ? <p className="text-sm text-danger">{formError}</p> : null}
          <button type="submit" className={primaryButtonClassName()} disabled={pending}>
            {pending ? "Saving…" : "Add account"}
          </button>
        </form>
      </Modal>

      <Modal open={Boolean(editAccount)} title="Edit account" onClose={() => setEditAccount(null)}>
        <form onSubmit={onEdit} className="space-y-4">
          <FormField label="Name" htmlFor="edit_name">
            <input id="edit_name" required className={inputClassName()} value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} />
          </FormField>
          <FormField label="Account type" htmlFor="edit_type">
            <select id="edit_type" className={inputClassName()} value={editForm.account_type} onChange={(e) => setEditForm({ ...editForm, account_type: e.target.value })}>
              {ACCOUNT_TYPES.map((type) => (
                <option key={type.value} value={type.value}>{type.label}</option>
              ))}
            </select>
          </FormField>
          <FormField label="Currency" htmlFor="edit_currency">
            <input id="edit_currency" required maxLength={3} className={inputClassName()} value={editForm.currency} onChange={(e) => setEditForm({ ...editForm, currency: e.target.value })} />
          </FormField>
          {formError ? <p className="text-sm text-danger">{formError}</p> : null}
          <button type="submit" className={primaryButtonClassName()} disabled={pending}>
            {pending ? "Saving…" : "Save changes"}
          </button>
        </form>
      </Modal>

      <ConfirmDialog
        open={Boolean(deleteAccount)}
        title="Delete account"
        message={
          formError ||
          `Delete “${deleteAccount?.name ?? ""}”? Accounts with transactions cannot be deleted.`
        }
        confirmLabel="Delete"
        pending={pending}
        onConfirm={onDelete}
        onClose={() => setDeleteAccount(null)}
      />
    </div>
  );
}
