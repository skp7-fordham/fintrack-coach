import { clearSession, getToken } from "@/lib/auth";
import type {
  Account,
  AuthResult,
  Category,
  CategorySpendingResult,
  CoachChatResult,
  CoachConversation,
  CoachConversationDetail,
  DashboardSummary,
  ImportError,
  ImportJob,
  MonthlyTrendsResult,
  Pagination,
  RecentTransaction,
  Transaction,
} from "@/lib/types";

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

type DataResponse<T> = { data: T };
type ListResponse<T> = { data: T[]; pagination: Pagination };

function apiBaseUrl(): string {
  const raw =
    process.env.NEXT_PUBLIC_API_BASE_URL ||
    (process.env.NODE_ENV !== "production" ? "http://localhost:8080" : "");
  const base = raw.trim().replace(/\/+$/, "");
  if (!base) {
    throw new ApiError("NEXT_PUBLIC_API_BASE_URL is not configured", 500);
  }
  return base;
}

function apiURL(path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl()}${normalizedPath}`;
}

function handleUnauthorized(): void {
  clearSession();
  if (typeof window !== "undefined") {
    const path = window.location.pathname;
    if (!path.startsWith("/login") && !path.startsWith("/register")) {
      window.location.assign("/login");
    }
  }
}

async function parseError(response: Response): Promise<ApiError> {
  try {
    const body = (await response.json()) as { error?: string };
    if (body.error) {
      return new ApiError(body.error, response.status);
    }
  } catch {
    // ignore
  }
  return new ApiError(`Request failed (${response.status})`, response.status);
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  auth = true,
): Promise<T> {
  const headers = new Headers(options.headers);
  if (!(options.body instanceof FormData) && !headers.has("Content-Type") && options.body) {
    headers.set("Content-Type", "application/json");
  }

  if (auth) {
    const token = getToken();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
  }

  const response = await fetch(apiURL(path), {
    ...options,
    headers,
  });

  if (response.status === 401) {
    handleUnauthorized();
    throw await parseError(response);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  if (!response.ok) {
    throw await parseError(response);
  }

  if (response.status === 204 || response.headers.get("content-length") === "0") {
    return undefined as T;
  }

  return (await response.json()) as T;
}

function toQuery(params: Record<string, string | number | undefined | null>): string {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === null || value === "") return;
    search.set(key, String(value));
  });
  const qs = search.toString();
  return qs ? `?${qs}` : "";
}

export const api = {
  register(email: string, password: string) {
    return request<DataResponse<AuthResult>>(
      "/auth/register",
      {
        method: "POST",
        body: JSON.stringify({ email, password }),
      },
      false,
    );
  },

  login(email: string, password: string) {
    return request<DataResponse<AuthResult>>(
      "/auth/login",
      {
        method: "POST",
        body: JSON.stringify({ email, password }),
      },
      false,
    );
  },

  demoLogin() {
    return request<DataResponse<AuthResult>>(
      "/auth/demo",
      { method: "POST" },
      false,
    );
  },

  listAccounts() {
    return request<DataResponse<Account[]>>("/accounts");
  },

  createAccount(body: {
    name: string;
    account_type: string;
    currency: string;
    initial_balance?: string;
  }) {
    return request<DataResponse<Account>>("/accounts", {
      method: "POST",
      body: JSON.stringify(body),
    });
  },

  updateAccount(
    id: string,
    body: { name?: string; account_type?: string; currency?: string },
  ) {
    return request<DataResponse<Account>>(`/accounts/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body),
    });
  },

  deleteAccount(id: string) {
    return request<void>(`/accounts/${id}`, { method: "DELETE" });
  },

  listCategories(type?: string) {
    return request<DataResponse<Category[]>>(
      `/categories${toQuery({ type })}`,
    );
  },

  createCategory(body: {
    name: string;
    category_type: string;
    color?: string | null;
    icon?: string | null;
  }) {
    return request<DataResponse<Category>>("/categories", {
      method: "POST",
      body: JSON.stringify(body),
    });
  },

  updateCategory(
    id: string,
    body: {
      name?: string;
      category_type?: string;
      color?: string | null;
      icon?: string | null;
    },
  ) {
    return request<DataResponse<Category>>(`/categories/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body),
    });
  },

  deleteCategory(id: string) {
    return request<void>(`/categories/${id}`, { method: "DELETE" });
  },

  listTransactions(params: {
    account_id?: string;
    category_id?: string;
    transaction_type?: string;
    from?: string;
    to?: string;
    search?: string;
    page?: number;
    page_size?: number;
  }) {
    return request<ListResponse<Transaction>>(
      `/transactions${toQuery(params)}`,
    );
  },

  createTransaction(body: {
    account_id: string;
    category_id?: string | null;
    description: string;
    merchant?: string | null;
    amount: string;
    transaction_type: string;
    transaction_status: string;
    transaction_date: string;
    notes?: string | null;
  }) {
    return request<DataResponse<Transaction>>("/transactions", {
      method: "POST",
      body: JSON.stringify(body),
    });
  },

  getDashboardSummary(month?: string) {
    return request<DataResponse<DashboardSummary>>(
      `/dashboard/summary${toQuery({ month })}`,
    );
  },

  getCategorySpending(month?: string) {
    return request<CategorySpendingResult>(
      `/dashboard/category-spending${toQuery({ month })}`,
    );
  },

  getMonthlyTrends(months = 6) {
    return request<MonthlyTrendsResult>(
      `/dashboard/monthly-trends${toQuery({ months })}`,
    );
  },

  getRecentTransactions(limit = 5) {
    return request<DataResponse<RecentTransaction[]> & { meta: { limit: number } }>(
      `/dashboard/recent-transactions${toQuery({ limit })}`,
    );
  },

  listImports(page = 1, pageSize = 20) {
    return request<ListResponse<ImportJob>>(
      `/imports${toQuery({ page, page_size: pageSize })}`,
    );
  },

  getImport(id: string) {
    return request<DataResponse<ImportJob>>(`/imports/${id}`);
  },

  listImportErrors(id: string, page = 1, pageSize = 50) {
    return request<ListResponse<ImportError>>(
      `/imports/${id}/errors${toQuery({ page, page_size: pageSize })}`,
    );
  },

  uploadImport(accountId: string, file: File) {
    const form = new FormData();
    form.append("account_id", accountId);
    form.append("file", file);
    return request<DataResponse<ImportJob>>("/imports/transactions", {
      method: "POST",
      body: form,
    });
  },

  coachChat(message: string, conversationId?: string) {
    return request<DataResponse<CoachChatResult>>("/coach/chat", {
      method: "POST",
      body: JSON.stringify({
        message,
        conversation_id: conversationId || undefined,
      }),
    });
  },

  listConversations(page = 1, pageSize = 50) {
    return request<ListResponse<CoachConversation>>(
      `/coach/conversations${toQuery({ page, page_size: pageSize })}`,
    );
  },

  getConversation(id: string) {
    return request<DataResponse<CoachConversationDetail>>(
      `/coach/conversations/${id}`,
    );
  },

  deleteConversation(id: string) {
    return request<void>(`/coach/conversations/${id}`, { method: "DELETE" });
  },
};
