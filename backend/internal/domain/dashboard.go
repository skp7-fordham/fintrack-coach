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

// CategorySpendingFilter holds the scoped inputs for category spending aggregation.
type CategorySpendingFilter struct {
	UserID            string
	Month             string
	MonthStart        time.Time
	MonthEndExclusive time.Time
}

// CategorySpendingItem is one category (or uncategorized) spend bucket.
type CategorySpendingItem struct {
	CategoryID       *string
	CategoryName     string
	CategoryColor    *string
	CategoryIcon     *string
	Amount           string
	TransactionCount int64
	Percentage       string
}

// CategorySpendingResult is the category spending breakdown for one user and month.
type CategorySpendingResult struct {
	Month        string
	TotalExpense string
	Items        []CategorySpendingItem
}
