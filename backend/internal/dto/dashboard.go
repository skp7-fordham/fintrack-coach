package dto

// DashboardSummaryQuery holds raw query-string values for GET /dashboard/summary.
type DashboardSummaryQuery struct {
	UserID string
	Month  string
}
