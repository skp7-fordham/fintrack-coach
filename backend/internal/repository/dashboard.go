package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

type DashboardRepository struct {
	pool *pgxpool.Pool
}

func NewDashboardRepository(pool *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{pool: pool}
}

func (r *DashboardRepository) GetDashboardSummary(
	ctx context.Context,
	filter domain.DashboardSummaryFilter,
) (*domain.DashboardSummary, error) {
	totalBalance, err := r.sumAccountBalances(ctx, filter.UserID)
	if err != nil {
		return nil, err
	}

	income, expense, netCashFlow, count, err := r.aggregateMonthlyTransactions(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &domain.DashboardSummary{
		Month:            filter.Month,
		TotalBalance:     totalBalance,
		MonthlyIncome:    income,
		MonthlyExpense:   expense,
		NetCashFlow:      netCashFlow,
		TransactionCount: count,
	}, nil
}

func (r *DashboardRepository) sumAccountBalances(ctx context.Context, userID string) (string, error) {
	const query = `
		SELECT COALESCE(SUM(current_balance), 0)::numeric(14, 2)::text
		FROM accounts
		WHERE user_id = $1
	`

	var totalBalance string
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&totalBalance); err != nil {
		return "", fmt.Errorf("sum account balances: %w", err)
	}
	return totalBalance, nil
}

func (r *DashboardRepository) aggregateMonthlyTransactions(
	ctx context.Context,
	filter domain.DashboardSummaryFilter,
) (income, expense, netCashFlow string, count int64, err error) {
	const query = `
		SELECT
			COALESCE(
				SUM(CASE WHEN transaction_type = 'income' THEN amount ELSE 0 END),
				0
			)::numeric(14, 2)::text AS monthly_income,
			COALESCE(
				SUM(CASE WHEN transaction_type = 'expense' THEN amount ELSE 0 END),
				0
			)::numeric(14, 2)::text AS monthly_expense,
			COALESCE(
				SUM(CASE WHEN transaction_type = 'income' THEN amount ELSE 0 END)
				- SUM(CASE WHEN transaction_type = 'expense' THEN amount ELSE 0 END),
				0
			)::numeric(14, 2)::text AS net_cash_flow,
			COUNT(*)::bigint AS transaction_count
		FROM transactions
		WHERE user_id = $1
		  AND transaction_status = 'completed'
		  AND transaction_type IN ('income', 'expense')
		  AND transaction_date >= $2
		  AND transaction_date < $3
	`

	err = r.pool.QueryRow(
		ctx,
		query,
		filter.UserID,
		filter.MonthStart,
		filter.MonthEndExclusive,
	).Scan(&income, &expense, &netCashFlow, &count)
	if err != nil {
		return "", "", "", 0, fmt.Errorf("aggregate monthly transactions: %w", err)
	}

	return income, expense, netCashFlow, count, nil
}

func (r *DashboardRepository) GetCategorySpending(
	ctx context.Context,
	filter domain.CategorySpendingFilter,
) (*domain.CategorySpendingResult, error) {
	const query = `
		WITH filtered_expenses AS (
			SELECT
				t.category_id,
				t.amount,
				c.name AS category_name,
				c.color AS category_color,
				c.icon AS category_icon
			FROM transactions t
			LEFT JOIN categories c ON c.id = t.category_id
			WHERE t.user_id = $1
			  AND t.transaction_status = 'completed'
			  AND t.transaction_type = 'expense'
			  AND t.transaction_date >= $2
			  AND t.transaction_date < $3
		),
		grouped AS (
			SELECT
				category_id,
				category_name,
				category_color,
				category_icon,
				SUM(amount) AS category_amount,
				COUNT(*)::bigint AS transaction_count
			FROM filtered_expenses
			GROUP BY category_id, category_name, category_color, category_icon
		)
		SELECT
			category_id::text,
			COALESCE(category_name, 'Uncategorized') AS category_name,
			category_color,
			category_icon,
			category_amount::numeric(14, 2)::text AS amount,
			transaction_count,
			(
				category_amount
				/ SUM(category_amount) OVER ()
				* 100
			)::numeric(14, 2)::text AS percentage,
			(SUM(category_amount) OVER ())::numeric(14, 2)::text AS total_expense
		FROM grouped
		ORDER BY category_amount DESC, COALESCE(category_name, 'Uncategorized') ASC
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		filter.UserID,
		filter.MonthStart,
		filter.MonthEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf("get category spending: %w", err)
	}
	defer rows.Close()

	items := make([]domain.CategorySpendingItem, 0)
	totalExpense := "0.00"

	for rows.Next() {
		var item domain.CategorySpendingItem
		var rowTotalExpense string
		if err := rows.Scan(
			&item.CategoryID,
			&item.CategoryName,
			&item.CategoryColor,
			&item.CategoryIcon,
			&item.Amount,
			&item.TransactionCount,
			&item.Percentage,
			&rowTotalExpense,
		); err != nil {
			return nil, fmt.Errorf("scan category spending: %w", err)
		}
		totalExpense = rowTotalExpense
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate category spending: %w", err)
	}

	return &domain.CategorySpendingResult{
		Month:        filter.Month,
		TotalExpense: totalExpense,
		Items:        items,
	}, nil
}

func (r *DashboardRepository) GetMonthlyTrends(
	ctx context.Context,
	filter domain.MonthlyTrendFilter,
) (*domain.MonthlyTrendResult, error) {
	const query = `
		WITH months AS (
			SELECT generate_series(
				$2::date,
				$3::date,
				interval '1 month'
			)::date AS month_start
		),
		aggregated AS (
			SELECT
				date_trunc('month', transaction_date)::date AS month_start,
				COALESCE(
					SUM(CASE WHEN transaction_type = 'income' THEN amount ELSE 0 END),
					0
				) AS income,
				COALESCE(
					SUM(CASE WHEN transaction_type = 'expense' THEN amount ELSE 0 END),
					0
				) AS expense,
				COUNT(*)::bigint AS transaction_count
			FROM transactions
			WHERE user_id = $1
			  AND transaction_status = 'completed'
			  AND transaction_type IN ('income', 'expense')
			  AND transaction_date >= $2
			  AND transaction_date < $4
			GROUP BY 1
		)
		SELECT
			to_char(m.month_start, 'YYYY-MM') AS month,
			COALESCE(a.income, 0)::numeric(14, 2)::text AS income,
			COALESCE(a.expense, 0)::numeric(14, 2)::text AS expense,
			(
				COALESCE(a.income, 0) - COALESCE(a.expense, 0)
			)::numeric(14, 2)::text AS net_cash_flow,
			COALESCE(a.transaction_count, 0)::bigint AS transaction_count
		FROM months m
		LEFT JOIN aggregated a USING (month_start)
		ORDER BY m.month_start ASC
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		filter.UserID,
		filter.RangeStart,
		filter.CurrentMonthStart,
		filter.RangeEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf("get monthly trends: %w", err)
	}
	defer rows.Close()

	items := make([]domain.MonthlyTrendItem, 0, filter.Months)
	for rows.Next() {
		var item domain.MonthlyTrendItem
		if err := rows.Scan(
			&item.Month,
			&item.Income,
			&item.Expense,
			&item.NetCashFlow,
			&item.TransactionCount,
		); err != nil {
			return nil, fmt.Errorf("scan monthly trend: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate monthly trends: %w", err)
	}

	return &domain.MonthlyTrendResult{
		Months:    filter.Months,
		FromMonth: filter.FromMonth,
		ToMonth:   filter.ToMonth,
		Items:     items,
	}, nil
}

func (r *DashboardRepository) GetRecentTransactions(
	ctx context.Context,
	filter domain.RecentTransactionsFilter,
) (*domain.RecentTransactionsResult, error) {
	const query = `
		SELECT
			t.id::text,
			t.account_id::text,
			a.name AS account_name,
			t.category_id::text,
			COALESCE(c.name, 'Uncategorized') AS category_name,
			t.description,
			t.merchant,
			t.amount::numeric(14, 2)::text AS amount,
			t.transaction_type,
			t.transaction_status,
			t.transaction_date,
			t.created_at
		FROM transactions t
		INNER JOIN accounts a ON a.id = t.account_id
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1
		ORDER BY t.transaction_date DESC, t.created_at DESC, t.id DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, filter.UserID, filter.Limit)
	if err != nil {
		return nil, fmt.Errorf("get recent transactions: %w", err)
	}
	defer rows.Close()

	items := make([]domain.RecentTransactionItem, 0, filter.Limit)
	for rows.Next() {
		var item domain.RecentTransactionItem
		if err := rows.Scan(
			&item.ID,
			&item.AccountID,
			&item.AccountName,
			&item.CategoryID,
			&item.CategoryName,
			&item.Description,
			&item.Merchant,
			&item.Amount,
			&item.TransactionType,
			&item.TransactionStatus,
			&item.TransactionDate,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan recent transaction: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent transactions: %w", err)
	}

	return &domain.RecentTransactionsResult{
		Limit: filter.Limit,
		Items: items,
	}, nil
}
