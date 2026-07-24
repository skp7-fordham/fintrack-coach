package dto

// DashboardSummaryQuery holds raw query-string values for GET /dashboard/summary.
type DashboardSummaryQuery struct {
	UserID string
	Month  string
}

// CategorySpendingQuery holds raw query-string values for GET /dashboard/category-spending.
type CategorySpendingQuery struct {
	UserID string
	Month  string
}

// MonthlyTrendsQuery holds raw query-string values for GET /dashboard/monthly-trends.
type MonthlyTrendsQuery struct {
	UserID string
	Months string
}

// RecentTransactionsQuery holds raw query-string values for GET /dashboard/recent-transactions.
type RecentTransactionsQuery struct {
	UserID string
	Limit  string
}
