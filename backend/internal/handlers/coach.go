package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/auth"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/coach"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
)

const maxCoachBodyBytes = 1 << 16

type CoachHandler struct {
	service *coach.Service
	logger  *slog.Logger
}

func NewCoachHandler(svc *coach.Service, logger *slog.Logger) *CoachHandler {
	return &CoachHandler{
		service: svc,
		logger:  logger,
	}
}

type coachChatResponse struct {
	Data coachChatData `json:"data"`
}

type coachChatData struct {
	ConversationID string           `json:"conversation_id"`
	Message        coachMessageData `json:"message"`
	ToolsUsed      []string         `json:"tools_used"`
}

type coachMessageData struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type coachConversationListResponse struct {
	Data       []coachConversationListItem `json:"data"`
	Pagination paginationResponse          `json:"pagination"`
}

type coachConversationListItem struct {
	ID        string  `json:"id"`
	Title     *string `json:"title"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type coachConversationDetailResponse struct {
	Data coachConversationDetailData `json:"data"`
}

type coachConversationDetailData struct {
	ID        string             `json:"id"`
	Title     *string            `json:"title"`
	Messages  []coachMessageData `json:"messages"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}

func (h *CoachHandler) Chat(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxCoachBodyBytes)

	var req dto.CoachChatRequest
	if err := decodeJSONStrict(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.Chat(r.Context(), userID, auth.IsDemoFromContext(r.Context()), req)
	if err != nil {
		h.writeCoachError(w, err, "failed to process coach chat")
		return
	}

	toolsUsed := result.ToolsUsed
	if toolsUsed == nil {
		toolsUsed = []string{}
	}

	writeJSON(w, http.StatusOK, coachChatResponse{
		Data: coachChatData{
			ConversationID: result.ConversationID,
			Message:        toCoachMessageData(result.Message),
			ToolsUsed:      toolsUsed,
		},
	})
}

func (h *CoachHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	result, err := h.service.ListConversations(
		r.Context(),
		userID,
		r.URL.Query().Get("page"),
		r.URL.Query().Get("page_size"),
	)
	if err != nil {
		h.writeCoachError(w, err, "failed to list coach conversations")
		return
	}

	data := make([]coachConversationListItem, 0, len(result.Conversations))
	for _, conversation := range result.Conversations {
		data = append(data, coachConversationListItem{
			ID:        conversation.ID,
			Title:     conversation.Title,
			CreatedAt: conversation.CreatedAt.UTC().Format(time.RFC3339Nano),
			UpdatedAt: conversation.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}

	writeJSON(w, http.StatusOK, coachConversationListResponse{
		Data: data,
		Pagination: paginationResponse{
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalItems: result.TotalItems,
			TotalPages: result.TotalPages,
		},
	})
}

func (h *CoachHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	conversationID := r.PathValue("id")
	detail, err := h.service.GetConversation(r.Context(), userID, conversationID)
	if err != nil {
		h.writeCoachError(w, err, "failed to get coach conversation")
		return
	}

	messages := make([]coachMessageData, 0, len(detail.Messages))
	for _, message := range detail.Messages {
		messages = append(messages, toCoachMessageData(message))
	}

	writeJSON(w, http.StatusOK, coachConversationDetailResponse{
		Data: coachConversationDetailData{
			ID:        detail.Conversation.ID,
			Title:     detail.Conversation.Title,
			Messages:  messages,
			CreatedAt: detail.Conversation.CreatedAt.UTC().Format(time.RFC3339Nano),
			UpdatedAt: detail.Conversation.UpdatedAt.UTC().Format(time.RFC3339Nano),
		},
	})
}

func (h *CoachHandler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	conversationID := r.PathValue("id")
	if err := h.service.DeleteConversation(r.Context(), userID, conversationID); err != nil {
		h.writeCoachError(w, err, "failed to delete coach conversation")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toCoachMessageData(message domain.CoachMessage) coachMessageData {
	return coachMessageData{
		ID:        message.ID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func (h *CoachHandler) writeCoachError(w http.ResponseWriter, err error, logMessage string) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: validationErr.Message})
	case errors.Is(err, domain.ErrConversationNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "conversation not found"})
	case errors.Is(err, domain.ErrCoachUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "financial coach is temporarily unavailable"})
	case errors.Is(err, domain.ErrCoachToolLimitExceeded):
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "financial coach is temporarily unavailable"})
	case errors.Is(err, domain.ErrDemoAILimitReached):
		writeJSON(w, http.StatusTooManyRequests, errorResponse{Error: "demo AI limit reached"})
	default:
		h.logger.Error(logMessage, "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
