package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

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

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	query := dto.DashboardSummaryQuery{
		UserID: r.URL.Query().Get("user_id"),
		Month:  r.URL.Query().Get("month"),
	}

	summary, err := h.service.GetSummary(r.Context(), query)
	if err != nil {
		h.writeSummaryError(w, err)
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

func (h *DashboardHandler) writeSummaryError(w http.ResponseWriter, err error) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: validationErr.Message})
	default:
		h.logger.Error("failed to get dashboard summary", "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
