"use client";

import Link from "next/link";
import { FormEvent, useCallback, useState } from "react";
import { AppHeader } from "@/components/app-shell";
import { getErrorMessage, useAuth } from "@/components/auth-provider";
import {
  DemoReadOnlyNote,
  EmptyState,
  ErrorState,
  FormField,
  LoadingSkeleton,
  Modal,
  MoneyDisplay,
  Pagination,
  StatusBadge,
  inputClassName,
  primaryButtonClassName,
  secondaryButtonClassName,
} from "@/components/ui";
import { api } from "@/lib/api";
import { useDeferredLoad } from "@/lib/use-deferred-load";
import { formatDateLabel } from "@/lib/format";
import type { Account, Category, Transaction } from "@/lib/types";

export default function TransactionsPage() {
  const { user } = useAuth();
  const isDemo = Boolean(user?.is_demo);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [filters, setFilters] = useState({
    account_id: "",
    category_id: "",
    transaction_type: "",
    from: "",
    to: "",
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const [formError, setFormError] = useState("");
  const [form, setForm] = useState({
    account_id: "",
    category_id: "",
    description: "",
    merchant: "",
    amount: "",
    transaction_type: "expense",
    transaction_status: "completed",
    transaction_date: "",
    notes: "",
  });

  const loadMeta = useCallback(async () => {
    const [accountsRes, categoriesRes] = await Promise.all([
      api.listAccounts(),
      api.listCategories(),
    ]);
    setAccounts(accountsRes.data ?? []);
    setCategories(categoriesRes.data ?? []);
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await api.listTransactions({
        ...filters,
        page,
        page_size: 20,
      });
      setTransactions(res.data ?? []);
      setTotalPages(res.pagination.total_pages || 1);
    } catch (err) {
      setError(getErrorMessage(err, "We couldn’t load your transactions."));
    } finally {
      setLoading(false);
    }
  }, [filters, page]);

  useDeferredLoad(() => loadMeta().catch((err) => setError(getErrorMessage(err))), [loadMeta]);
  useDeferredLoad(load, [load]);

  const accountCurrency = (accountId: string) =>
    accounts.find((a) => a.id === accountId)?.currency;
  const accountName = (accountId: string) =>
    accounts.find((a) => a.id === accountId)?.name ?? "Unknown account";
  const categoryName = (categoryId: string | null) =>
    categories.find((category) => category.id === categoryId)?.name ?? "Uncategorized";

  async function onCreate(e: FormEvent) {
    e.preventDefault();
    if (isDemo) {
      setFormError("Demo account is read-only");
      return;
    }
    setPending(true);
    setFormError("");
    try {
      await api.createTransaction({
        account_id: form.account_id,
        category_id: form.category_id || null,
        description: form.description.trim(),
        merchant: form.merchant.trim() || null,
        amount: form.amount,
        transaction_type: form.transaction_type,
        transaction_status: form.transaction_status,
        transaction_date: form.transaction_date,
        notes: form.notes.trim() || null,
      });
      setCreateOpen(false);
      setForm((prev) => ({
        ...prev,
        description: "",
        merchant: "",
        amount: "",
        notes: "",
      }));
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
          title="Transactions"
          subtitle="Review and record income and expenses across your accounts."
        />
        <div className="flex flex-col items-start gap-1.5 sm:items-end">
          <button
            type="button"
            className={primaryButtonClassName()}
            disabled={isDemo}
            title={isDemo ? "Demo account is read-only" : undefined}
            onClick={() => {
              setFormError("");
              setForm((prev) => ({
                ...prev,
                account_id: accounts[0]?.id ?? "",
                transaction_date: new Date().toISOString().slice(0, 10),
              }));
              setCreateOpen(true);
            }}
          >
            Add transaction
          </button>
          {isDemo ? <DemoReadOnlyNote /> : null}
        </div>
      </div>

      <div className="mb-4 grid gap-3 rounded-lg border border-border bg-card p-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
        <select
          className={inputClassName()}
          value={filters.account_id}
          onChange={(e) => {
            setPage(1);
            setFilters({ ...filters, account_id: e.target.value });
          }}
          aria-label="Filter by account"
        >
          <option value="">All accounts</option>
          {accounts.map((account) => (
            <option key={account.id} value={account.id}>{account.name}</option>
          ))}
        </select>
        <select
          className={inputClassName()}
          value={filters.category_id}
          onChange={(e) => {
            setPage(1);
            setFilters({ ...filters, category_id: e.target.value });
          }}
          aria-label="Filter by category"
        >
          <option value="">All categories</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>{category.name}</option>
          ))}
        </select>
        <select
          className={inputClassName()}
          value={filters.transaction_type}
          onChange={(e) => {
            setPage(1);
            setFilters({ ...filters, transaction_type: e.target.value });
          }}
          aria-label="Filter by type"
        >
          <option value="">All types</option>
          <option value="income">Income</option>
          <option value="expense">Expense</option>
          <option value="transfer">Transfer</option>
        </select>
        <input
          type="date"
          className={inputClassName()}
          value={filters.from}
          onChange={(e) => {
            setPage(1);
            setFilters({ ...filters, from: e.target.value });
          }}
          aria-label="From date"
        />
        <input
          type="date"
          className={inputClassName()}
          value={filters.to}
          onChange={(e) => {
            setPage(1);
            setFilters({ ...filters, to: e.target.value });
          }}
          aria-label="To date"
        />
      </div>

      {error ? <ErrorState message={error} onRetry={load} /> : null}
      {loading ? <LoadingSkeleton rows={5} /> : null}

      {!loading && !error && transactions.length === 0 ? (
        <EmptyState
          title="No transactions yet"
          description="Add your first transaction or import a CSV to start tracking activity."
          action={
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                className={primaryButtonClassName()}
                disabled={isDemo}
                title={isDemo ? "Demo account is read-only" : undefined}
                onClick={() => setCreateOpen(true)}
              >
                Add transaction
              </button>
              {isDemo ? (
                <span
                  className={`${secondaryButtonClassName()} cursor-not-allowed opacity-60`}
                  title="Demo account is read-only"
                  aria-disabled="true"
                >
                  Import CSV
                </span>
              ) : (
                <Link href="/imports" className={secondaryButtonClassName()}>Import CSV</Link>
              )}
            </div>
          }
        />
      ) : null}

      {!loading && transactions.length > 0 ? (
        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <div className="overflow-x-auto">
            <table className="min-w-full text-left text-sm">
              <thead className="border-b border-border bg-slate-50/70 text-xs text-muted">
                <tr>
                  <th scope="col" className="px-4 py-2.5 font-medium">Date</th>
                  <th scope="col" className="px-4 py-2.5 font-medium">Transaction</th>
                  <th scope="col" className="px-4 py-2.5 font-medium">Type</th>
                  <th scope="col" className="px-4 py-2.5 font-medium">Status</th>
                  <th scope="col" className="px-4 py-2.5 font-medium text-right">Amount</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((txn) => (
                  <tr key={txn.id} className="border-b border-border/70 transition-colors last:border-0 hover:bg-slate-50/60">
                    <td className="px-4 py-3 whitespace-nowrap text-muted">{formatDateLabel(txn.transaction_date)}</td>
                    <td className="px-4 py-3">
                      <div className="font-medium">{txn.description}</div>
                      <div className="mt-0.5 text-xs text-muted">
                        {txn.merchant ? `${txn.merchant} · ` : ""}
                        {accountName(txn.account_id)} · {categoryName(txn.category_id)}
                      </div>
                    </td>
                    <td className="px-4 py-3"><StatusBadge status={txn.transaction_type} /></td>
                    <td className="px-4 py-3"><StatusBadge status={txn.transaction_status} /></td>
                    <td className="px-4 py-3 text-right">
                      <MoneyDisplay
                        amount={txn.transaction_type === "expense" ? `-${txn.amount}` : txn.amount}
                        currency={accountCurrency(txn.account_id)}
                        signed={txn.transaction_type !== "expense"}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="border-t border-border px-4 pb-4">
            <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
          </div>
        </div>
      ) : null}

      <Modal open={createOpen} title="Add transaction" onClose={() => setCreateOpen(false)}>
        <form onSubmit={onCreate} className="space-y-4">
          <FormField label="Account" htmlFor="txn_account">
            <select id="txn_account" required className={inputClassName()} value={form.account_id} onChange={(e) => setForm({ ...form, account_id: e.target.value })}>
              <option value="">Select account</option>
              {accounts.map((account) => (
                <option key={account.id} value={account.id}>{account.name}</option>
              ))}
            </select>
          </FormField>
          <FormField label="Category" htmlFor="txn_category">
            <select id="txn_category" className={inputClassName()} value={form.category_id} onChange={(e) => setForm({ ...form, category_id: e.target.value })}>
              <option value="">Uncategorized</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>{category.name}</option>
              ))}
            </select>
          </FormField>
          <FormField label="Description" htmlFor="txn_description">
            <input id="txn_description" required className={inputClassName()} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          </FormField>
          <FormField label="Merchant" htmlFor="txn_merchant">
            <input id="txn_merchant" className={inputClassName()} value={form.merchant} onChange={(e) => setForm({ ...form, merchant: e.target.value })} />
          </FormField>
          <FormField label="Amount" htmlFor="txn_amount">
            <input id="txn_amount" required inputMode="decimal" className={inputClassName()} value={form.amount} onChange={(e) => setForm({ ...form, amount: e.target.value })} />
          </FormField>
          <FormField label="Type" htmlFor="txn_type">
            <select id="txn_type" className={inputClassName()} value={form.transaction_type} onChange={(e) => setForm({ ...form, transaction_type: e.target.value })}>
              <option value="expense">Expense</option>
              <option value="income">Income</option>
              <option value="transfer">Transfer</option>
            </select>
          </FormField>
          <FormField label="Date" htmlFor="txn_date">
            <input id="txn_date" type="date" required className={inputClassName()} value={form.transaction_date} onChange={(e) => setForm({ ...form, transaction_date: e.target.value })} />
          </FormField>
          <FormField label="Notes" htmlFor="txn_notes">
            <textarea id="txn_notes" className={inputClassName()} rows={3} value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} />
          </FormField>
          {formError ? <p className="text-sm text-danger">{formError}</p> : null}
          <button type="submit" className={primaryButtonClassName()} disabled={pending}>
            {pending ? "Saving…" : "Add transaction"}
          </button>
        </form>
      </Modal>
    </div>
  );
}
