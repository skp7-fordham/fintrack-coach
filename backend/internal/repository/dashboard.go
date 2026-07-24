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
