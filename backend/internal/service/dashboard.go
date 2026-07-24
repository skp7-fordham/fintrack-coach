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
	GetCategorySpending(ctx context.Context, filter domain.CategorySpendingFilter) (*domain.CategorySpendingResult, error)
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
	userID, month, monthStart, monthEndExclusive, err := resolveUserAndMonth(query.UserID, query.Month)
	if err != nil {
		return nil, err
	}

	return s.repo.GetDashboardSummary(ctx, domain.DashboardSummaryFilter{
		UserID:            userID,
		Month:             month,
		MonthStart:        monthStart,
		MonthEndExclusive: monthEndExclusive,
	})
}

func (s *DashboardService) GetCategorySpending(
	ctx context.Context,
	query dto.CategorySpendingQuery,
) (*domain.CategorySpendingResult, error) {
	userID, month, monthStart, monthEndExclusive, err := resolveUserAndMonth(query.UserID, query.Month)
	if err != nil {
		return nil, err
	}

	return s.repo.GetCategorySpending(ctx, domain.CategorySpendingFilter{
		UserID:            userID,
		Month:             month,
		MonthStart:        monthStart,
		MonthEndExclusive: monthEndExclusive,
	})
}

func resolveUserAndMonth(rawUserID, rawMonth string) (userID, month string, monthStart, monthEndExclusive time.Time, err error) {
	userID = strings.TrimSpace(rawUserID)
	if !isValidUUID(userID) {
		return "", "", time.Time{}, time.Time{}, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}

	month = strings.TrimSpace(rawMonth)
	if month == "" {
		month = time.Now().UTC().Format("2006-01")
	}

	monthStart, monthEndExclusive, month, err = parseYearMonth(month)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	return userID, month, monthStart, monthEndExclusive, nil
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
