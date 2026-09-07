"use client";

import { FormEvent, useCallback, useMemo, useState } from "react";
import { AppHeader } from "@/components/app-shell";
import { getErrorMessage } from "@/components/auth-provider";
import {
  ConfirmDialog,
  EmptyState,
  ErrorState,
  FormField,
  LoadingSkeleton,
  Modal,
  StatusBadge,
  inputClassName,
  primaryButtonClassName,
  secondaryButtonClassName,
} from "@/components/ui";
import { api } from "@/lib/api";
import { useDeferredLoad } from "@/lib/use-deferred-load";
import type { Category } from "@/lib/types";

const emptyForm = {
  name: "",
  category_type: "expense",
  color: "#0f766e",
  icon: "",
};

function CategoryList({
  items,
  title,
  onEdit,
  onDelete,
}: {
  items: Category[];
  title: string;
  onEdit: (category: Category) => void;
  onDelete: (category: Category) => void;
}) {
  if (items.length === 0) return null;
  return (
    <section className="space-y-3">
      <h2 className="text-base font-semibold">{title}</h2>
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {items.map((category) => (
          <article key={category.id} className="rounded-xl border border-border bg-card p-4 shadow-sm">
            <div className="flex items-center gap-3">
              <span
                className="inline-flex h-9 w-9 items-center justify-center rounded-full text-sm"
                style={{ backgroundColor: category.color || "#e2e8f0" }}
                aria-hidden
              >
                {category.icon || "•"}
              </span>
              <div>
                <h3 className="font-medium">{category.name}</h3>
                <StatusBadge status={category.category_type} />
              </div>
            </div>
            <div className="mt-4 flex gap-2">
              <button type="button" className={secondaryButtonClassName()} onClick={() => onEdit(category)}>
                Edit
              </button>
              <button
                type="button"
                className="rounded-lg border border-danger/30 px-4 py-2 text-sm text-danger"
                onClick={() => onDelete(category)}
              >
                Delete
              </button>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}

function CategoryForm({
  form,
  setForm,
  error,
  pending,
  onSubmit,
}: {
  form: typeof emptyForm;
  setForm: (value: typeof emptyForm) => void;
  error: string;
  pending: boolean;
  onSubmit: (e: FormEvent) => void;
}) {
  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <FormField label="Name" htmlFor="cat_name">
        <input id="cat_name" required className={inputClassName()} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
      </FormField>
      <FormField label="Type" htmlFor="cat_type">
        <select id="cat_type" className={inputClassName()} value={form.category_type} onChange={(e) => setForm({ ...form, category_type: e.target.value })}>
          <option value="expense">Expense</option>
          <option value="income">Income</option>
        </select>
      </FormField>
      <FormField label="Colour" htmlFor="cat_color">
        <div className="flex gap-2">
          <input id="cat_color" type="color" className="h-10 w-14 rounded border border-border" value={form.color} onChange={(e) => setForm({ ...form, color: e.target.value })} />
          <input className={inputClassName()} value={form.color} onChange={(e) => setForm({ ...form, color: e.target.value })} />
        </div>
      </FormField>
      <FormField label="Icon" htmlFor="cat_icon" hint="Optional short label or emoji">
        <input id="cat_icon" className={inputClassName()} value={form.icon} onChange={(e) => setForm({ ...form, icon: e.target.value })} />
      </FormField>
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      <button type="submit" className={primaryButtonClassName()} disabled={pending}>
        {pending ? "Saving…" : "Save category"}
      </button>
    </form>
  );
}

export default function CategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [filter, setFilter] = useState<"all" | "income" | "expense">("all");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [editCategory, setEditCategory] = useState<Category | null>(null);
  const [deleteCategory, setDeleteCategory] = useState<Category | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [pending, setPending] = useState(false);
  const [formError, setFormError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await api.listCategories(filter === "all" ? undefined : filter);
      setCategories(res.data ?? []);
    } catch (err) {
      setError(getErrorMessage(err, "Failed to load categories"));
    } finally {
      setLoading(false);
    }
  }, [filter]);

  useDeferredLoad(load, [load]);

  const grouped = useMemo(
    () => ({
      income: categories.filter((c) => c.category_type === "income"),
      expense: categories.filter((c) => c.category_type === "expense"),
    }),
    [categories],
  );

  async function onCreate(e: FormEvent) {
    e.preventDefault();
    setPending(true);
    setFormError("");
    try {
      await api.createCategory({
        name: form.name.trim(),
        category_type: form.category_type,
        color: form.color || null,
        icon: form.icon.trim() || null,
      });
      setCreateOpen(false);
      setForm(emptyForm);
      await load();
    } catch (err) {
      setFormError(getErrorMessage(err));
    } finally {
      setPending(false);
    }
  }

  async function onEdit(e: FormEvent) {
    e.preventDefault();
    if (!editCategory) return;
    setPending(true);
    setFormError("");
    try {
      await api.updateCategory(editCategory.id, {
        name: form.name.trim(),
        category_type: form.category_type,
        color: form.color || null,
        icon: form.icon.trim() || null,
      });
      setEditCategory(null);
      await load();
    } catch (err) {
      setFormError(getErrorMessage(err));
    } finally {
      setPending(false);
    }
  }

  async function onDelete() {
    if (!deleteCategory) return;
    setPending(true);
    try {
      await api.deleteCategory(deleteCategory.id);
      setDeleteCategory(null);
      await load();
    } catch (err) {
      setFormError(getErrorMessage(err));
    } finally {
      setPending(false);
    }
  }

  function beginEdit(category: Category) {
    setFormError("");
    setEditCategory(category);
    setForm({
      name: category.name,
      category_type: category.category_type,
      color: category.color || "#0f766e",
      icon: category.icon || "",
    });
  }

  return (
    <div>
      <div className="mb-6 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <AppHeader title="Categories" subtitle="Organise income and expense labels" />
        <div className="flex flex-wrap gap-2">
          {(["all", "income", "expense"] as const).map((value) => (
            <button
              key={value}
              type="button"
              onClick={() => setFilter(value)}
              className={`rounded-lg px-3 py-2 text-sm capitalize ${
                filter === value ? "bg-primary text-white" : "border border-border bg-card"
              }`}
            >
              {value}
            </button>
          ))}
          <button
            type="button"
            className={primaryButtonClassName()}
            onClick={() => {
              setForm(emptyForm);
              setFormError("");
              setCreateOpen(true);
            }}
          >
            Add category
          </button>
        </div>
      </div>

      {error ? <ErrorState message={error} onRetry={load} /> : null}
      {loading ? <LoadingSkeleton rows={4} /> : null}

      {!loading && !error && categories.length === 0 ? (
        <EmptyState
          title="Add a category"
          description="Categories help you understand where money goes. Create income and expense labels before importing CSV files."
          action={
            <button type="button" className={primaryButtonClassName()} onClick={() => setCreateOpen(true)}>
              Add category
            </button>
          }
        />
      ) : null}

      {!loading && categories.length > 0 ? (
        <div className="space-y-8">
          <CategoryList items={grouped.expense} title="Expense" onEdit={beginEdit} onDelete={setDeleteCategory} />
          <CategoryList items={grouped.income} title="Income" onEdit={beginEdit} onDelete={setDeleteCategory} />
        </div>
      ) : null}

      <Modal open={createOpen} title="Create category" onClose={() => setCreateOpen(false)}>
        <CategoryForm form={form} setForm={setForm} error={formError} pending={pending} onSubmit={onCreate} />
      </Modal>

      <Modal open={Boolean(editCategory)} title="Edit category" onClose={() => setEditCategory(null)}>
        <CategoryForm form={form} setForm={setForm} error={formError} pending={pending} onSubmit={onEdit} />
      </Modal>

      <ConfirmDialog
        open={Boolean(deleteCategory)}
        title="Delete category"
        message={`Delete “${deleteCategory?.name ?? ""}”? Existing transactions keep their history.`}
        confirmLabel="Delete"
        pending={pending}
        onConfirm={onDelete}
        onClose={() => setDeleteCategory(null)}
      />
    </div>
  );
}
