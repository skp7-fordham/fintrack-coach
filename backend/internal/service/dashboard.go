package service

import (
	"context"
	"strings"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
)

type dashboardRepository interface {
	GetDashboardSummary(ctx context.Context, filter domain.DashboardSummaryFilter) (*domain.DashboardSummary, error)
}

type DashboardService struct {
	repo dashboardRepository
}

func NewDashboardService(repo dashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) GetSummary(
	ctx context.Context,
	query dto.DashboardSummaryQuery,
) (*domain.DashboardSummary, error) {
	filter, err := buildDashboardSummaryFilter(query)
	if err != nil {
		return nil, err
	}

	summary, err := s.repo.GetDashboardSummary(ctx, filter)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func buildDashboardSummaryFilter(query dto.DashboardSummaryQuery) (domain.DashboardSummaryFilter, error) {
	userID := strings.TrimSpace(query.UserID)
	if !isValidUUID(userID) {
		return domain.DashboardSummaryFilter{}, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}

	month := strings.TrimSpace(query.Month)
	if month == "" {
		month = time.Now().UTC().Format("2006-01")
	}

	monthStart, monthEndExclusive, normalizedMonth, err := parseYearMonth(month)
	if err != nil {
		return domain.DashboardSummaryFilter{}, err
	}

	return domain.DashboardSummaryFilter{
		UserID:            userID,
		Month:             normalizedMonth,
		MonthStart:        monthStart,
		MonthEndExclusive: monthEndExclusive,
	}, nil
}

func parseYearMonth(month string) (start, endExclusive time.Time, normalized string, err error) {
	if len(month) != 7 || month[4] != '-' {
		return time.Time{}, time.Time{}, "", &domain.ValidationError{Message: "month must use YYYY-MM"}
	}

	parsed, parseErr := time.Parse("2006-01", month)
	if parseErr != nil {
		return time.Time{}, time.Time{}, "", &domain.ValidationError{Message: "month must use YYYY-MM"}
	}

	start = time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
	endExclusive = start.AddDate(0, 1, 0)
	normalized = start.Format("2006-01")
	return start, endExclusive, normalized, nil
}
