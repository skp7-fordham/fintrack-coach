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
