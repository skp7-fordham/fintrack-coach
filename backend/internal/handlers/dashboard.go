package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/service"
)

type DashboardHandler struct {
	service *service.DashboardService
	logger  *slog.Logger
}

func NewDashboardHandler(svc *service.DashboardService, logger *slog.Logger) *DashboardHandler {
	return &DashboardHandler{
		service: svc,
		logger:  logger,
	}
}

type dashboardSummaryResponse struct {
	Data dashboardSummaryData `json:"data"`
}

type dashboardSummaryData struct {
	Month            string `json:"month"`
	TotalBalance     string `json:"total_balance"`
	MonthlyIncome    string `json:"monthly_income"`
	MonthlyExpense   string `json:"monthly_expense"`
	NetCashFlow      string `json:"net_cash_flow"`
	TransactionCount int64  `json:"transaction_count"`
}

type categorySpendingResponse struct {
	Data []categorySpendingItemData `json:"data"`
	Meta categorySpendingMeta       `json:"meta"`
}

type categorySpendingItemData struct {
	CategoryID       *string `json:"category_id"`
	CategoryName     string  `json:"category_name"`
	CategoryColor    *string `json:"category_color"`
	CategoryIcon     *string `json:"category_icon"`
	Amount           string  `json:"amount"`
	TransactionCount int64   `json:"transaction_count"`
	Percentage       string  `json:"percentage"`
}

type categorySpendingMeta struct {
	Month        string `json:"month"`
	TotalExpense string `json:"total_expense"`
}

type monthlyTrendsResponse struct {
	Data []monthlyTrendItemData `json:"data"`
	Meta monthlyTrendsMeta      `json:"meta"`
}

type monthlyTrendItemData struct {
	Month            string `json:"month"`
	Income           string `json:"income"`
	Expense          string `json:"expense"`
	NetCashFlow      string `json:"net_cash_flow"`
	TransactionCount int64  `json:"transaction_count"`
}

type monthlyTrendsMeta struct {
	Months    int    `json:"months"`
	FromMonth string `json:"from_month"`
	ToMonth   string `json:"to_month"`
}

type recentTransactionsResponse struct {
	Data []recentTransactionItemData `json:"data"`
	Meta recentTransactionsMeta      `json:"meta"`
}

type recentTransactionItemData struct {
	ID                string  `json:"id"`
	AccountID         string  `json:"account_id"`
	AccountName       string  `json:"account_name"`
	CategoryID        *string `json:"category_id"`
	CategoryName      string  `json:"category_name"`
	Description       string  `json:"description"`
	Merchant          *string `json:"merchant"`
	Amount            string  `json:"amount"`
	TransactionType   string  `json:"transaction_type"`
	TransactionStatus string  `json:"transaction_status"`
	TransactionDate   string  `json:"transaction_date"`
	CreatedAt         string  `json:"created_at"`
}

type recentTransactionsMeta struct {
	Limit int `json:"limit"`
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	query := dto.DashboardSummaryQuery{
		UserID: r.URL.Query().Get("user_id"),
		Month:  r.URL.Query().Get("month"),
	}

	summary, err := h.service.GetSummary(r.Context(), query)
	if err != nil {
		h.writeDashboardError(w, err, "failed to get dashboard summary")
		return
	}

	h.logger.Info(
		"dashboard summary retrieved",
		"user_id", strings.TrimSpace(query.UserID),
		"month", summary.Month,
		"transaction_count", summary.TransactionCount,
	)

	writeJSON(w, http.StatusOK, dashboardSummaryResponse{
		Data: dashboardSummaryData{
			Month:            summary.Month,
			TotalBalance:     summary.TotalBalance,
			MonthlyIncome:    summary.MonthlyIncome,
			MonthlyExpense:   summary.MonthlyExpense,
			NetCashFlow:      summary.NetCashFlow,
			TransactionCount: summary.TransactionCount,
		},
	})
}

func (h *DashboardHandler) CategorySpending(w http.ResponseWriter, r *http.Request) {
	query := dto.CategorySpendingQuery{
		UserID: r.URL.Query().Get("user_id"),
		Month:  r.URL.Query().Get("month"),
	}

	result, err := h.service.GetCategorySpending(r.Context(), query)
	if err != nil {
		h.writeDashboardError(w, err, "failed to get category spending")
		return
	}

	h.logger.Info(
		"category spending retrieved",
		"user_id", strings.TrimSpace(query.UserID),
		"month", result.Month,
		"category_count", len(result.Items),
	)

	data := make([]categorySpendingItemData, 0, len(result.Items))
	for _, item := range result.Items {
		data = append(data, categorySpendingItemData{
			CategoryID:       item.CategoryID,
			CategoryName:     item.CategoryName,
			CategoryColor:    item.CategoryColor,
			CategoryIcon:     item.CategoryIcon,
			Amount:           item.Amount,
			TransactionCount: item.TransactionCount,
			Percentage:       item.Percentage,
		})
	}

	writeJSON(w, http.StatusOK, categorySpendingResponse{
		Data: data,
		Meta: categorySpendingMeta{
			Month:        result.Month,
			TotalExpense: result.TotalExpense,
		},
	})
}

func (h *DashboardHandler) MonthlyTrends(w http.ResponseWriter, r *http.Request) {
	query := dto.MonthlyTrendsQuery{
		UserID: r.URL.Query().Get("user_id"),
		Months: r.URL.Query().Get("months"),
	}

	result, err := h.service.GetMonthlyTrends(r.Context(), query)
	if err != nil {
		h.writeDashboardError(w, err, "failed to get monthly trends")
		return
	}

	h.logger.Info(
		"monthly trends retrieved",
		"user_id", strings.TrimSpace(query.UserID),
		"months", result.Months,
		"row_count", len(result.Items),
	)

	data := make([]monthlyTrendItemData, 0, len(result.Items))
	for _, item := range result.Items {
		data = append(data, monthlyTrendItemData{
			Month:            item.Month,
			Income:           item.Income,
			Expense:          item.Expense,
			NetCashFlow:      item.NetCashFlow,
			TransactionCount: item.TransactionCount,
		})
	}

	writeJSON(w, http.StatusOK, monthlyTrendsResponse{
		Data: data,
		Meta: monthlyTrendsMeta{
			Months:    result.Months,
			FromMonth: result.FromMonth,
			ToMonth:   result.ToMonth,
		},
	})
}

func (h *DashboardHandler) RecentTransactions(w http.ResponseWriter, r *http.Request) {
	query := dto.RecentTransactionsQuery{
		UserID: r.URL.Query().Get("user_id"),
		Limit:  r.URL.Query().Get("limit"),
	}

	result, err := h.service.GetRecentTransactions(r.Context(), query)
	if err != nil {
		h.writeDashboardError(w, err, "failed to get recent transactions")
		return
	}

	h.logger.Info(
		"recent transactions retrieved",
		"user_id", strings.TrimSpace(query.UserID),
		"limit", result.Limit,
		"row_count", len(result.Items),
	)

	data := make([]recentTransactionItemData, 0, len(result.Items))
	for _, item := range result.Items {
		data = append(data, recentTransactionItemData{
			ID:                item.ID,
			AccountID:         item.AccountID,
			AccountName:       item.AccountName,
			CategoryID:        item.CategoryID,
			CategoryName:      item.CategoryName,
			Description:       item.Description,
			Merchant:          item.Merchant,
			Amount:            item.Amount,
			TransactionType:   item.TransactionType,
			TransactionStatus: item.TransactionStatus,
			TransactionDate:   item.TransactionDate.Format("2006-01-02"),
			CreatedAt:         item.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
	}

	writeJSON(w, http.StatusOK, recentTransactionsResponse{
		Data: data,
		Meta: recentTransactionsMeta{Limit: result.Limit},
	})
}

func (h *DashboardHandler) writeDashboardError(w http.ResponseWriter, err error, logMessage string) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: validationErr.Message})
	default:
		h.logger.Error(logMessage, "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
