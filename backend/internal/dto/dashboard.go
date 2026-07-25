package dto

// DashboardSummaryQuery holds raw query-string values for GET /dashboard/summary.
type DashboardSummaryQuery struct {
	Month string
}

// CategorySpendingQuery holds raw query-string values for GET /dashboard/category-spending.
type CategorySpendingQuery struct {
	Month string
}

// MonthlyTrendsQuery holds raw query-string values for GET /dashboard/monthly-trends.
type MonthlyTrendsQuery struct {
	Months string
}

// RecentTransactionsQuery holds raw query-string values for GET /dashboard/recent-transactions.
type RecentTransactionsQuery struct {
	Limit string
}
