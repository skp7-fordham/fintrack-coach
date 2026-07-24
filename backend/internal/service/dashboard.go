package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
)

const (
	defaultTrendMonths = 6
	minTrendMonths     = 1
	maxTrendMonths     = 24

	defaultRecentLimit = 5
	minRecentLimit     = 1
	maxRecentLimit     = 20
)

type dashboardRepository interface {
	GetDashboardSummary(ctx context.Context, filter domain.DashboardSummaryFilter) (*domain.DashboardSummary, error)
	GetCategorySpending(ctx context.Context, filter domain.CategorySpendingFilter) (*domain.CategorySpendingResult, error)
	GetMonthlyTrends(ctx context.Context, filter domain.MonthlyTrendFilter) (*domain.MonthlyTrendResult, error)
	GetRecentTransactions(ctx context.Context, filter domain.RecentTransactionsFilter) (*domain.RecentTransactionsResult, error)
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

func (s *DashboardService) GetMonthlyTrends(
	ctx context.Context,
	query dto.MonthlyTrendsQuery,
) (*domain.MonthlyTrendResult, error) {
	userID, err := resolveUserID(query.UserID)
	if err != nil {
		return nil, err
	}

	months, err := parseBoundedInt(query.Months, defaultTrendMonths, minTrendMonths, maxTrendMonths, "months")
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	rangeStart := currentMonthStart.AddDate(0, -(months - 1), 0)
	rangeEndExclusive := currentMonthStart.AddDate(0, 1, 0)

	return s.repo.GetMonthlyTrends(ctx, domain.MonthlyTrendFilter{
		UserID:            userID,
		Months:            months,
		RangeStart:        rangeStart,
		CurrentMonthStart: currentMonthStart,
		RangeEndExclusive: rangeEndExclusive,
		FromMonth:         rangeStart.Format("2006-01"),
		ToMonth:           currentMonthStart.Format("2006-01"),
	})
}

func (s *DashboardService) GetRecentTransactions(
	ctx context.Context,
	query dto.RecentTransactionsQuery,
) (*domain.RecentTransactionsResult, error) {
	userID, err := resolveUserID(query.UserID)
	if err != nil {
		return nil, err
	}

	limit, err := parseBoundedInt(query.Limit, defaultRecentLimit, minRecentLimit, maxRecentLimit, "limit")
	if err != nil {
		return nil, err
	}

	return s.repo.GetRecentTransactions(ctx, domain.RecentTransactionsFilter{
		UserID: userID,
		Limit:  limit,
	})
}

func resolveUserAndMonth(rawUserID, rawMonth string) (userID, month string, monthStart, monthEndExclusive time.Time, err error) {
	userID, err = resolveUserID(rawUserID)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
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

func resolveUserID(rawUserID string) (string, error) {
	userID := strings.TrimSpace(rawUserID)
	if !isValidUUID(userID) {
		return "", &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	return userID, nil
}

func parseBoundedInt(raw string, defaultValue, minValue, maxValue int, field string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < minValue || parsed > maxValue {
		return 0, &domain.ValidationError{
			Message: fmt.Sprintf("%s must be an integer between %d and %d", field, minValue, maxValue),
		}
	}
	return parsed, nil
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
