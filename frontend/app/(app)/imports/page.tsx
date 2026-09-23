"use client";

import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AppHeader } from "@/components/app-shell";
import { getErrorMessage } from "@/components/auth-provider";
import {
  EmptyState,
  ErrorState,
  FormField,
  LoadingSkeleton,
  Modal,
  Pagination,
  StatusBadge,
  inputClassName,
  primaryButtonClassName,
  secondaryButtonClassName,
} from "@/components/ui";
import { api } from "@/lib/api";
import { useDeferredLoad } from "@/lib/use-deferred-load";
import type { Account, ImportError, ImportJob } from "@/lib/types";

const MAX_FILE_SIZE = 5 * 1024 * 1024;
const TERMINAL = new Set(["completed", "completed_with_errors", "failed"]);

export default function ImportsPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [jobs, setJobs] = useState<ImportJob[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [accountId, setAccountId] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [uploadError, setUploadError] = useState("");
  const [pending, setPending] = useState(false);
  const [errorsJob, setErrorsJob] = useState<ImportJob | null>(null);
  const [rowErrors, setRowErrors] = useState<ImportError[]>([]);
  const mounted = useRef(true);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [accountsRes, importsRes] = await Promise.all([
        api.listAccounts(),
        api.listImports(page, 20),
      ]);
      if (!mounted.current) return;
      setAccounts(accountsRes.data ?? []);
      setJobs(importsRes.data ?? []);
      setTotalPages(importsRes.pagination.total_pages || 1);
      if (!accountId && accountsRes.data?.[0]) {
        setAccountId(accountsRes.data[0].id);
      }
    } catch (err) {
      if (mounted.current) setError(getErrorMessage(err, "We couldn’t load your CSV imports."));
    } finally {
      if (mounted.current) setLoading(false);
    }
  }, [accountId, page]);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  useDeferredLoad(load, [load]);

  const activeJobIds = useMemo(
    () => jobs.filter((job) => !TERMINAL.has(job.status)).map((job) => job.id),
    [jobs],
  );

  useEffect(() => {
    if (activeJobIds.length === 0) return;
    const timer = window.setInterval(async () => {
      try {
        const updates = await Promise.all(activeJobIds.map((id) => api.getImport(id)));
        if (!mounted.current) return;
        setJobs((current) =>
          current.map((job) => {
            const updated = updates.find((item) => item.data.id === job.id);
            return updated ? updated.data : job;
          }),
        );
      } catch {
        // keep previous state; next poll may recover
      }
    }, 2000);
    return () => window.clearInterval(timer);
  }, [activeJobIds]);

  async function onUpload(e: FormEvent) {
    e.preventDefault();
    setUploadError("");
    if (!accountId) {
      setUploadError("Select an account");
      return;
    }
    if (!file) {
      setUploadError("Choose a .csv file");
      return;
    }
    if (!file.name.toLowerCase().endsWith(".csv")) {
      setUploadError("File must be a .csv");
      return;
    }
    if (file.size > MAX_FILE_SIZE) {
      setUploadError("File exceeds the 5 MB limit");
      return;
    }

    setPending(true);
    try {
      await api.uploadImport(accountId, file);
      setFile(null);
      await load();
    } catch (err) {
      setUploadError(getErrorMessage(err));
    } finally {
      setPending(false);
    }
  }

  async function openErrors(job: ImportJob) {
    setErrorsJob(job);
    try {
      const res = await api.listImportErrors(job.id, 1, 100);
      setRowErrors(res.data ?? []);
    } catch (err) {
      setRowErrors([]);
      setUploadError(getErrorMessage(err));
    }
  }

  return (
    <div>
      <AppHeader
        title="CSV Imports"
        subtitle="Upload transaction files and review processing results."
      />

      <section className="mb-6 rounded-lg border border-border bg-card p-5">
        <h2 className="text-base font-semibold">Import transactions</h2>
        <p className="mt-1 text-sm text-muted">
          Select an account and upload a CSV with these columns: <code>date, description, merchant, amount, type, category, notes</code>.
        </p>
        <pre className="mt-3 overflow-x-auto rounded-lg bg-slate-950 p-3 font-mono text-xs text-slate-100">
{`date,description,merchant,amount,type,category,notes
2026-07-01,Grocery run,Fresh Mart,84.20,expense,Groceries,
2026-07-02,Payday,,3200.00,income,Salary,Biweekly pay`}
        </pre>
        <form onSubmit={onUpload} className="mt-4 grid gap-3 md:grid-cols-[1fr_1fr_auto]">
          <FormField label="Account" htmlFor="import_account">
            <select id="import_account" className={inputClassName()} value={accountId} onChange={(e) => setAccountId(e.target.value)}>
              <option value="">Select account</option>
              {accounts.map((account) => (
                <option key={account.id} value={account.id}>{account.name}</option>
              ))}
            </select>
          </FormField>
          <FormField label="CSV file" htmlFor="import_file">
            <input
              id="import_file"
              type="file"
              accept=".csv,text/csv"
              className={inputClassName()}
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            />
          </FormField>
          <div className="flex items-end">
            <button type="submit" className={primaryButtonClassName()} disabled={pending}>
              {pending ? "Importing…" : "Import CSV"}
            </button>
          </div>
        </form>
        {uploadError ? <p className="mt-3 text-sm text-danger">{uploadError}</p> : null}
      </section>

      {error ? <ErrorState message={error} onRetry={load} /> : null}
      {loading ? <LoadingSkeleton rows={4} /> : null}

      {!loading && !error && jobs.length === 0 ? (
        <EmptyState
          title="No imports yet"
          description="Upload a CSV file to add transactions in bulk."
        />
      ) : null}

      {!loading && jobs.length > 0 ? (
        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <div className="overflow-x-auto">
            <table className="min-w-full text-left text-sm">
              <thead className="border-b border-border bg-slate-50/70 text-xs text-muted">
                <tr>
                  <th scope="col" className="px-4 py-2.5 font-medium">File</th>
                  <th scope="col" className="px-4 py-2.5 font-medium">Status</th>
                  <th scope="col" className="px-4 py-2.5 font-medium">Results</th>
                  <th scope="col" className="px-4 py-2.5 font-medium">Uploaded</th>
                  <th scope="col" className="px-4 py-2.5 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {jobs.map((job) => (
                  <tr key={job.id} className="border-b border-border/70 transition-colors last:border-0 hover:bg-slate-50/60">
                    <td className="px-4 py-3 font-medium">{job.original_filename}</td>
                    <td className="px-4 py-3"><StatusBadge status={job.status} /></td>
                    <td className="px-4 py-3 text-muted">
                      {job.successful_rows} imported · {job.failed_rows} issues · {job.total_rows} rows
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-muted">{new Date(job.created_at).toLocaleString()}</td>
                    <td className="px-4 py-3">
                      {job.failed_rows > 0 ? (
                        <button type="button" className={secondaryButtonClassName()} onClick={() => void openErrors(job)}>
                          View issues
                        </button>
                      ) : (
                        <span className="text-xs text-muted">—</span>
                      )}
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

      <Modal open={Boolean(errorsJob)} title="Import issues" onClose={() => setErrorsJob(null)}>
        {rowErrors.length === 0 ? (
          <p className="text-sm text-muted">No issues were found for this import.</p>
        ) : (
          <div className="max-h-80 space-y-3 overflow-y-auto">
            {rowErrors.map((err) => (
              <div key={`${err.row_number}-${err.error_message}`} className="rounded-lg border border-border p-3 text-sm">
                <p className="font-medium">Row {err.row_number}</p>
                <p className="text-danger">{err.error_message}</p>
              </div>
            ))}
          </div>
        )}
      </Modal>
    </div>
  );
}
