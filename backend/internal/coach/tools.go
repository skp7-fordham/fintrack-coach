package coach

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/aiprovider"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/service"
)

const (
	toolGetDashboardSummary   = "get_dashboard_summary"
	toolGetCategorySpending   = "get_category_spending"
	toolGetMonthlyTrends      = "get_monthly_trends"
	toolGetRecentTransactions = "get_recent_transactions"
	toolListAccounts          = "list_accounts"
	toolGetImportStatus       = "get_import_status"
)

type ToolRegistry struct {
	dashboard *service.DashboardService
	accounts  *service.AccountService
	imports   *service.ImportService
	logger    *slog.Logger
}

func NewToolRegistry(
	dashboard *service.DashboardService,
	accounts *service.AccountService,
	imports *service.ImportService,
	logger *slog.Logger,
) *ToolRegistry {
	return &ToolRegistry{
		dashboard: dashboard,
		accounts:  accounts,
		imports:   imports,
		logger:    logger,
	}
}

func (r *ToolRegistry) Definitions() []aiprovider.ToolDefinition {
	return []aiprovider.ToolDefinition{
		{
			Type: "function",
			Function: aiprovider.ToolFunctionSchema{
				Name:        toolGetDashboardSummary,
				Description: "Get transaction activity for one calendar month (income, expense, net cash flow, transaction count) plus the user's current total balance snapshot. total_balance is present-tense only and is not a historical month-end balance.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"month": map[string]any{
							"type":        "string",
							"description": "Month in YYYY-MM format for transaction activity. Defaults to the current UTC calendar month when omitted. Do not invent unrelated years.",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.ToolFunctionSchema{
				Name:        toolGetCategorySpending,
				Description: "Get expense spending grouped by category for one calendar month of transaction activity.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"month": map[string]any{
							"type":        "string",
							"description": "Month in YYYY-MM format. Defaults to the current UTC calendar month when omitted. Do not invent unrelated years.",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.ToolFunctionSchema{
				Name:        toolGetMonthlyTrends,
				Description: "Compare monthly income, expense, and net cash flow over recent calendar months. Use this for month-over-month activity comparisons. It does not provide historical account balances.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"months": map[string]any{
							"type":        "integer",
							"description": "Number of months to include ending at the current UTC month, from 1 to 12. Defaults to 6.",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.ToolFunctionSchema{
				Name:        toolGetRecentTransactions,
				Description: "Get recent transactions with date, description, merchant, amount, type, category, and account.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"limit": map[string]any{
							"type":        "integer",
							"description": "Number of transactions to return, from 1 to 20. Defaults to 10.",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.ToolFunctionSchema{
				Name:        toolListAccounts,
				Description: "List the authenticated user's accounts with name, type, currency, and current balances only. Balances are present snapshots, not historical values.",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.ToolFunctionSchema{
				Name:        toolGetImportStatus,
				Description: "Get recent CSV import job statuses including filename, status, successful rows, failed rows, and created_at.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"limit": map[string]any{
							"type":        "integer",
							"description": "Number of recent import jobs to return, from 1 to 20. Defaults to 5.",
						},
					},
				},
			},
		},
	}
}

func (r *ToolRegistry) Execute(ctx context.Context, userID, toolName, rawArgs string) (string, error) {
	start := time.Now()
	result, err := r.execute(ctx, userID, toolName, rawArgs)
	r.logger.Info(
		"coach tool executed",
		"user_id", userID,
		"tool_name", toolName,
		"duration_ms", time.Since(start).Milliseconds(),
		"ok", err == nil,
	)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (r *ToolRegistry) execute(ctx context.Context, userID, toolName, rawArgs string) (string, error) {
	args, err := parseToolArgs(rawArgs)
	if err != nil {
		return "", err
	}
	delete(args, "user_id")

	switch toolName {
	case toolGetDashboardSummary:
		return r.getDashboardSummary(ctx, userID, args)
	case toolGetCategorySpending:
		return r.getCategorySpending(ctx, userID, args)
	case toolGetMonthlyTrends:
		return r.getMonthlyTrends(ctx, userID, args)
	case toolGetRecentTransactions:
		return r.getRecentTransactions(ctx, userID, args)
	case toolListAccounts:
		return r.listAccounts(ctx, userID)
	case toolGetImportStatus:
		return r.getImportStatus(ctx, userID, args)
	default:
		return "", errUnknownTool
	}
}

func (r *ToolRegistry) getDashboardSummary(ctx context.Context, userID string, args map[string]any) (string, error) {
	month, err := optionalStringArg(args, "month")
	if err != nil {
		return "", err
	}
	summary, err := r.dashboard.GetSummary(ctx, userID, dto.DashboardSummaryQuery{Month: month})
	if err != nil {
		return "", err
	}
	return marshalCompact(map[string]any{
		"month":             summary.Month,
		"total_balance":     summary.TotalBalance,
		"monthly_income":    summary.MonthlyIncome,
		"monthly_expense":   summary.MonthlyExpense,
		"net_cash_flow":     summary.NetCashFlow,
		"transaction_count": summary.TransactionCount,
	})
}

func (r *ToolRegistry) getCategorySpending(ctx context.Context, userID string, args map[string]any) (string, error) {
	month, err := optionalStringArg(args, "month")
	if err != nil {
		return "", err
	}
	result, err := r.dashboard.GetCategorySpending(ctx, userID, dto.CategorySpendingQuery{Month: month})
	if err != nil {
		return "", err
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, map[string]any{
			"category_name":     item.CategoryName,
			"amount":            item.Amount,
			"transaction_count": item.TransactionCount,
			"percentage":        item.Percentage,
		})
	}
	return marshalCompact(map[string]any{
		"month":         result.Month,
		"total_expense": result.TotalExpense,
		"categories":    items,
	})
}

func (r *ToolRegistry) getMonthlyTrends(ctx context.Context, userID string, args map[string]any) (string, error) {
	months, err := optionalIntArg(args, "months", 6, 1, 12)
	if err != nil {
		return "", err
	}
	result, err := r.dashboard.GetMonthlyTrends(ctx, userID, dto.MonthlyTrendsQuery{
		Months: strconv.Itoa(months),
	})
	if err != nil {
		return "", err
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, map[string]any{
			"month":             item.Month,
			"income":            item.Income,
			"expense":           item.Expense,
			"net_cash_flow":     item.NetCashFlow,
			"transaction_count": item.TransactionCount,
		})
	}
	return marshalCompact(map[string]any{
		"months":     result.Months,
		"from_month": result.FromMonth,
		"to_month":   result.ToMonth,
		"trends":     items,
	})
}

func (r *ToolRegistry) getRecentTransactions(ctx context.Context, userID string, args map[string]any) (string, error) {
	limit, err := optionalIntArg(args, "limit", 10, 1, 20)
	if err != nil {
		return "", err
	}
	result, err := r.dashboard.GetRecentTransactions(ctx, userID, dto.RecentTransactionsQuery{
		Limit: strconv.Itoa(limit),
	})
	if err != nil {
		return "", err
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, item := range result.Items {
		row := map[string]any{
			"date":          item.TransactionDate.Format("2006-01-02"),
			"description":   item.Description,
			"amount":        item.Amount,
			"type":          item.TransactionType,
			"category_name": item.CategoryName,
			"account_name":  item.AccountName,
		}
		if item.Merchant != nil {
			row["merchant"] = *item.Merchant
		} else {
			row["merchant"] = nil
		}
		items = append(items, row)
	}
	return marshalCompact(map[string]any{
		"limit":        result.Limit,
		"transactions": items,
	})
}

func (r *ToolRegistry) listAccounts(ctx context.Context, userID string) (string, error) {
	accounts, err := r.accounts.ListAccounts(ctx, userID)
	if err != nil {
		return "", err
	}
	items := make([]map[string]any, 0, len(accounts))
	for _, account := range accounts {
		items = append(items, map[string]any{
			"name":            account.Name,
			"account_type":    account.AccountType,
			"currency":        account.Currency,
			"current_balance": account.CurrentBalance,
		})
	}
	return marshalCompact(map[string]any{"accounts": items})
}

func (r *ToolRegistry) getImportStatus(ctx context.Context, userID string, args map[string]any) (string, error) {
	limit, err := optionalIntArg(args, "limit", 5, 1, 20)
	if err != nil {
		return "", err
	}
	result, err := r.imports.ListImports(ctx, userID, "1", strconv.Itoa(limit))
	if err != nil {
		return "", err
	}

	items := make([]map[string]any, 0, len(result.Jobs))
	for _, job := range result.Jobs {
		items = append(items, map[string]any{
			"original_filename": job.OriginalFilename,
			"status":            job.Status,
			"successful_rows":   job.SuccessfulRows,
			"failed_rows":       job.FailedRows,
			"created_at":        job.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return marshalCompact(map[string]any{"imports": items})
}

func parseToolArgs(raw string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]any{}, nil
	}
	var args map[string]any
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		return nil, fmt.Errorf("%w: %v", errMalformedToolArgs, err)
	}
	if args == nil {
		return map[string]any{}, nil
	}
	return args, nil
}

func optionalStringArg(args map[string]any, key string) (string, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return "", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s must be a string", errMalformedToolArgs, key)
	}
	return strings.TrimSpace(value), nil
}

func optionalIntArg(args map[string]any, key string, defaultValue, minValue, maxValue int) (int, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return defaultValue, nil
	}

	var value int
	switch typed := raw.(type) {
	case float64:
		if typed != float64(int(typed)) {
			return 0, fmt.Errorf("%w: %s must be an integer", errMalformedToolArgs, key)
		}
		value = int(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, fmt.Errorf("%w: %s must be an integer", errMalformedToolArgs, key)
		}
		value = int(parsed)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0, fmt.Errorf("%w: %s must be an integer", errMalformedToolArgs, key)
		}
		value = parsed
	default:
		return 0, fmt.Errorf("%w: %s must be an integer", errMalformedToolArgs, key)
	}

	if value < minValue || value > maxValue {
		return 0, &domain.ValidationError{
			Message: fmt.Sprintf("%s must be an integer between %d and %d", key, minValue, maxValue),
		}
	}
	return value, nil
}

func marshalCompact(payload any) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal tool result: %w", err)
	}
	return string(data), nil
}
