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

// MonthlyTrendFilter holds the scoped inputs for multi-month trend aggregation.
type MonthlyTrendFilter struct {
	UserID            string
	Months            int
	RangeStart        time.Time
	CurrentMonthStart time.Time
	RangeEndExclusive time.Time
	FromMonth         string
	ToMonth           string
}

// MonthlyTrendItem is one calendar month in a trends series.
type MonthlyTrendItem struct {
	Month            string
	Income           string
	Expense          string
	NetCashFlow      string
	TransactionCount int64
}

// MonthlyTrendResult is the multi-month trend series for one user.
type MonthlyTrendResult struct {
	Months    int
	FromMonth string
	ToMonth   string
	Items     []MonthlyTrendItem
}

// RecentTransactionsFilter holds inputs for the recent activity feed.
type RecentTransactionsFilter struct {
	UserID string
	Limit  int
}

// RecentTransactionItem is a dashboard-ready recent transaction row.
type RecentTransactionItem struct {
	ID                string
	AccountID         string
	AccountName       string
	CategoryID        *string
	CategoryName      string
	Description       string
	Merchant          *string
	Amount            string
	TransactionType   string
	TransactionStatus string
	TransactionDate   time.Time
	CreatedAt         time.Time
}

// RecentTransactionsResult is the recent activity list for one user.
type RecentTransactionsResult struct {
	Limit int
	Items []RecentTransactionItem
}
