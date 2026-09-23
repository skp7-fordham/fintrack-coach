"use client";

import Link from "next/link";
import { useCallback, useMemo, useState } from "react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { AppHeader } from "@/components/app-shell";
import { getErrorMessage } from "@/components/auth-provider";
import {
  EmptyState,
  ErrorState,
  LoadingSkeleton,
  MoneyDisplay,
  StatusBadge,
  SummaryCard,
  inputClassName,
  primaryButtonClassName,
} from "@/components/ui";
import { api } from "@/lib/api";
import { useDeferredLoad } from "@/lib/use-deferred-load";
import { currentYearMonthUTC, formatDateLabel, formatMoney } from "@/lib/format";
import type {
  Account,
  CategorySpendingItem,
  DashboardSummary,
  MonthlyTrendItem,
  RecentTransaction,
} from "@/lib/types";

export default function DashboardPage() {
  const [month, setMonth] = useState(currentYearMonthUTC());
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [spending, setSpending] = useState<CategorySpendingItem[]>([]);
  const [trends, setTrends] = useState<MonthlyTrendItem[]>([]);
  const [recent, setRecent] = useState<RecentTransaction[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [summaryRes, spendingRes, trendsRes, recentRes, accountsRes] =
        await Promise.all([
          api.getDashboardSummary(month),
          api.getCategorySpending(month),
          api.getMonthlyTrends(6),
          api.getRecentTransactions(8),
          api.listAccounts(),
        ]);
      setSummary(summaryRes.data);
      setSpending(spendingRes.data ?? []);
      setTrends(trendsRes.data ?? []);
      setRecent(recentRes.data ?? []);
      setAccounts(accountsRes.data ?? []);
    } catch (err) {
      setError(getErrorMessage(err, "We couldn’t load your dashboard."));
    } finally {
      setLoading(false);
    }
  }, [month]);

  useDeferredLoad(load, [load]);

  const currencyHint = useMemo(() => {
    const codes = Array.from(new Set(accounts.map((a) => a.currency)));
    if (codes.length === 1) return codes[0];
    return null;
  }, [accounts]);

  const spendingChart = spending.map((item) => ({
    name: item.category_name,
    amount: Number.parseFloat(item.amount) || 0,
    fill: item.category_color || "#0f766e",
  }));

  const trendsChart = trends.map((item) => ({
    month: item.month,
    income: Number.parseFloat(item.income) || 0,
    expense: Number.parseFloat(item.expense) || 0,
  }));

  return (
    <div>
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between [&>header]:mb-0">
        <AppHeader
          title="Dashboard"
          subtitle="Monitor balances, spending patterns, and recent activity."
        />
        <label className="block text-sm">
          <span className="mb-1 block text-muted">Month</span>
          <input
            type="month"
            className={inputClassName()}
            value={month}
            onChange={(e) => setMonth(e.target.value)}
          />
        </label>
      </div>

      {error ? <ErrorState message={error} onRetry={load} /> : null}
      {loading ? <LoadingSkeleton rows={5} /> : null}

      {!loading && !error && summary ? (
        <div className="space-y-6">
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
            <SummaryCard
              label="Total balance"
              value={<MoneyDisplay amount={summary.total_balance} currency={currencyHint} />}
              hint={
                currencyHint
                  ? undefined
                  : accounts.length > 1
                    ? "Totals combined across currencies (not converted)"
                    : undefined
              }
            />
            <SummaryCard
              label="Monthly income"
              value={<MoneyDisplay amount={summary.monthly_income} currency={currencyHint} signed />}
            />
            <SummaryCard
              label="Monthly expense"
              value={<MoneyDisplay amount={summary.monthly_expense} currency={currencyHint} />}
            />
            <SummaryCard
              label="Net cash flow"
              value={<MoneyDisplay amount={summary.net_cash_flow} currency={currencyHint} signed />}
            />
            <SummaryCard label="Transactions" value={summary.transaction_count} />
          </div>

          <div className="grid gap-6 xl:grid-cols-2">
            <section className="rounded-lg border border-border bg-card p-5">
              <h2 className="mb-4 text-base font-semibold">Category spending</h2>
              {spendingChart.length === 0 ? (
                <EmptyState
                  title="No spending recorded this month"
                  description="Add transactions or import a CSV to see how your spending is distributed."
                  action={
                    <Link href="/transactions" className={primaryButtonClassName()}>
                      Add transaction
                    </Link>
                  }
                />
              ) : (
                <div className="h-72">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={spendingChart}>
                      <CartesianGrid strokeDasharray="3 3" vertical={false} />
                      <XAxis dataKey="name" tick={{ fontSize: 12 }} interval={0} angle={-20} textAnchor="end" height={60} />
                      <YAxis tick={{ fontSize: 12 }} />
                      <Tooltip formatter={(value) => formatMoney(Number(value), currencyHint)} />
                      <Bar dataKey="amount" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              )}
            </section>

            <section className="rounded-lg border border-border bg-card p-5">
              <h2 className="mb-4 text-base font-semibold">Monthly trends</h2>
              {trendsChart.length === 0 ? (
                <EmptyState
                  title="No financial trends yet"
                  description="Record income and expenses to compare your activity over time."
                />
              ) : (
                <div className="h-72">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={trendsChart}>
                      <CartesianGrid strokeDasharray="3 3" vertical={false} />
                      <XAxis dataKey="month" tick={{ fontSize: 12 }} />
                      <YAxis tick={{ fontSize: 12 }} />
                      <Tooltip formatter={(value) => formatMoney(Number(value), currencyHint)} />
                      <Legend />
                      <Bar dataKey="income" fill="var(--success)" radius={[4, 4, 0, 0]} />
                      <Bar dataKey="expense" fill="var(--danger)" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              )}
            </section>
          </div>

          <section className="rounded-lg border border-border bg-card p-5">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h2 className="text-base font-semibold">Recent transactions</h2>
              <Link href="/transactions" className="text-sm font-medium text-primary hover:underline">
                View all
              </Link>
            </div>
            {recent.length === 0 ? (
              <EmptyState
                title="No recent transactions"
                description="Add an account and record transactions to start building your financial overview."
                action={
                  <div className="flex flex-wrap gap-2">
                    <Link href="/accounts" className={primaryButtonClassName()}>
                      Add account
                    </Link>
                    <Link href="/imports" className="rounded-md border border-border px-4 py-2 text-sm font-medium transition-colors hover:bg-slate-50">
                      Import CSV
                    </Link>
                  </div>
                }
              />
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full text-left text-sm">
                  <thead className="border-b border-border bg-slate-50/70 text-xs text-muted">
                    <tr>
                      <th scope="col" className="px-2 py-2 font-medium">Date</th>
                      <th scope="col" className="px-2 py-2 font-medium">Description</th>
                      <th scope="col" className="px-2 py-2 font-medium">Account</th>
                      <th scope="col" className="px-2 py-2 font-medium">Category</th>
                      <th scope="col" className="px-2 py-2 font-medium">Type</th>
                      <th scope="col" className="px-2 py-2 font-medium text-right">Amount</th>
                    </tr>
                  </thead>
                  <tbody>
                    {recent.map((txn) => (
                      <tr key={txn.id} className="border-b border-border/70 transition-colors last:border-0 hover:bg-slate-50/60">
                        <td className="px-2 py-3 whitespace-nowrap text-muted">{formatDateLabel(txn.transaction_date)}</td>
                        <td className="px-2 py-3">
                          <div className="font-medium">{txn.description}</div>
                          {txn.merchant ? (
                            <div className="text-xs text-muted">{txn.merchant}</div>
                          ) : null}
                        </td>
                        <td className="px-2 py-3">{txn.account_name}</td>
                        <td className="px-2 py-3">{txn.category_name || "Uncategorized"}</td>
                        <td className="px-2 py-3">
                          <StatusBadge status={txn.transaction_type} />
                        </td>
                        <td className="px-2 py-3 text-right">
                          <MoneyDisplay
                            amount={
                              txn.transaction_type === "expense"
                                ? `-${txn.amount}`
                                : txn.amount
                            }
                            signed={txn.transaction_type !== "expense"}
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        </div>
      ) : null}
    </div>
  );
}
