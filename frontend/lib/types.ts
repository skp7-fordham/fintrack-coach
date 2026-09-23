export type AccountType =
  | "checking"
  | "savings"
  | "credit_card"
  | "cash"
  | "investment"
  | "loan";

export type CategoryType = "income" | "expense";
export type TransactionType = "income" | "expense" | "transfer";
export type TransactionStatus = "pending" | "completed" | "failed";

export type ImportStatus =
  | "queued"
  | "processing"
  | "completed"
  | "completed_with_errors"
  | "failed";

export interface Pagination {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
}

export interface AuthUser {
  id: string;
  email: string;
  is_demo: boolean;
  created_at: string;
}

export interface AuthResult {
  user: AuthUser;
  access_token: string;
  token_type: string;
  expires_in: number;
}

export interface Account {
  id: string;
  name: string;
  account_type: AccountType | string;
  currency: string;
  current_balance: string;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  name: string;
  category_type: CategoryType | string;
  color: string | null;
  icon: string | null;
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  id: string;
  user_id: string;
  account_id: string;
  category_id: string | null;
  description: string;
  merchant: string | null;
  amount: string;
  transaction_type: TransactionType | string;
  transaction_status: TransactionStatus | string;
  transaction_date: string;
  notes: string | null;
  created_at: string;
  updated_at: string;
}

export interface DashboardSummary {
  month: string;
  total_balance: string;
  monthly_income: string;
  monthly_expense: string;
  net_cash_flow: string;
  transaction_count: number;
}

export interface CategorySpendingItem {
  category_id: string | null;
  category_name: string;
  category_color: string | null;
  category_icon: string | null;
  amount: string;
  transaction_count: number;
  percentage: string;
}

export interface CategorySpendingResult {
  data: CategorySpendingItem[];
  meta: {
    month: string;
    total_expense: string;
  };
}

export interface MonthlyTrendItem {
  month: string;
  income: string;
  expense: string;
  net_cash_flow: string;
  transaction_count: number;
}

export interface MonthlyTrendsResult {
  data: MonthlyTrendItem[];
  meta: {
    months: number;
    from_month: string;
    to_month: string;
  };
}

export interface RecentTransaction {
  id: string;
  account_id: string;
  account_name: string;
  category_id: string | null;
  category_name: string;
  description: string;
  merchant: string | null;
  amount: string;
  transaction_type: string;
  transaction_status: string;
  transaction_date: string;
  created_at: string;
}

export interface ImportJob {
  id: string;
  account_id: string;
  original_filename: string;
  status: ImportStatus | string;
  total_rows: number;
  processed_rows: number;
  successful_rows: number;
  failed_rows: number;
  error_message: string | null;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface ImportError {
  row_number: number;
  error_message: string;
  raw_row: Record<string, string>;
}

export interface CoachMessage {
  id: string;
  role: "user" | "assistant" | string;
  content: string;
  created_at: string;
}

export interface CoachConversation {
  id: string;
  title: string | null;
  created_at: string;
  updated_at: string;
}

export interface CoachConversationDetail {
  id: string;
  title: string | null;
  messages: CoachMessage[];
  created_at: string;
  updated_at: string;
}

export interface CoachChatResult {
  conversation_id: string;
  message: CoachMessage;
  tools_used: string[];
}

export const ACCOUNT_TYPES: { value: AccountType; label: string }[] = [
  { value: "checking", label: "Checking" },
  { value: "savings", label: "Savings" },
  { value: "credit_card", label: "Credit card" },
  { value: "cash", label: "Cash" },
  { value: "investment", label: "Investment" },
  { value: "loan", label: "Loan" },
];
