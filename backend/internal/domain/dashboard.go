package domain

import "time"

// DashboardSummaryFilter holds the scoped inputs for dashboard aggregation.
type DashboardSummaryFilter struct {
	UserID            string
	Month             string
	MonthStart        time.Time
	MonthEndExclusive time.Time
}

// DashboardSummary is the aggregated financial snapshot for one user and month.
type DashboardSummary struct {
	Month            string
	TotalBalance     string
	MonthlyIncome    string
	MonthlyExpense   string
	NetCashFlow      string
	TransactionCount int64
}
