package coach

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
)

const (
	maxChatMessageLength = 2000
	maxConversationTitle = 80
	maxHistoryMessages   = 20
	defaultPage          = 1
	defaultPageSize      = 20
	maxPageSize          = 100
)

type Repository interface {
	CreateConversation(ctx context.Context, input domain.CreateCoachConversationInput) (*domain.CoachConversation, error)
	GetConversationByID(ctx context.Context, userID, conversationID string) (*domain.CoachConversation, error)
	ListConversations(ctx context.Context, filter domain.ListCoachConversationsFilter) ([]domain.CoachConversation, int64, error)
	TouchConversation(ctx context.Context, userID, conversationID string) error
	AddMessage(ctx context.Context, input domain.AddCoachMessageInput) (*domain.CoachMessage, error)
	ListRecentVisibleMessages(ctx context.Context, conversationID string, limit int) ([]domain.CoachMessage, error)
	ListVisibleMessages(ctx context.Context, userID, conversationID string) ([]domain.CoachMessage, error)
	DeleteConversation(ctx context.Context, userID, conversationID string) error
}

type Service struct {
	repo   Repository
	agent  *Agent
	logger *slog.Logger
}

type ListConversationsResult struct {
	Conversations []domain.CoachConversation
	Page          int
	PageSize      int
	TotalItems    int64
	TotalPages    int
}

type ConversationDetail struct {
	Conversation domain.CoachConversation
	Messages     []domain.CoachMessage
}

func NewService(repo Repository, agent *Agent, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		agent:  agent,
		logger: logger,
	}
}

func (s *Service) Chat(ctx context.Context, userID string, req dto.CoachChatRequest) (*domain.CoachChatResult, error) {
	start := time.Now()
	userID = strings.TrimSpace(userID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}

	message := strings.TrimSpace(req.Message)
	if message == "" {
		return nil, &domain.ValidationError{Message: "message is required"}
	}
	if utf8.RuneCountInString(message) > maxChatMessageLength {
		return nil, &domain.ValidationError{Message: "message must be at most 2000 characters"}
	}

	var conversation *domain.CoachConversation
	var err error

	if req.ConversationID != nil && strings.TrimSpace(*req.ConversationID) != "" {
		conversationID := strings.TrimSpace(*req.ConversationID)
		if !isValidUUID(conversationID) {
			return nil, &domain.ValidationError{Message: "conversation_id must be a valid UUID"}
		}
		conversation, err = s.repo.GetConversationByID(ctx, userID, conversationID)
		if err != nil {
			return nil, err
		}
	} else {
		title := buildConversationTitle(message)
		conversation, err = s.repo.CreateConversation(ctx, domain.CreateCoachConversationInput{
			UserID: userID,
			Title:  &title,
		})
		if err != nil {
			return nil, err
		}
	}

	if _, err := s.repo.AddMessage(ctx, domain.AddCoachMessageInput{
		ConversationID: conversation.ID,
		Role:           domain.CoachMessageRoleUser,
		Content:        message,
	}); err != nil {
		return nil, err
	}
	if err := s.repo.TouchConversation(ctx, userID, conversation.ID); err != nil {
		return nil, err
	}

	history, err := s.repo.ListRecentVisibleMessages(ctx, conversation.ID, maxHistoryMessages)
	if err != nil {
		return nil, err
	}
	// Exclude the just-added user message from history; agent appends it separately.
	if len(history) > 0 {
		last := history[len(history)-1]
		if last.Role == domain.CoachMessageRoleUser && last.Content == message {
			history = history[:len(history)-1]
		}
	}

	agentResult, err := s.agent.Run(ctx, userID, history, message)
	if err != nil {
		s.logger.Error(
			"coach chat failed",
			"user_id", userID,
			"conversation_id", conversation.ID,
			"duration_ms", time.Since(start).Milliseconds(),
			"err", err,
		)
		return nil, err
	}

	assistantMessage, err := s.repo.AddMessage(ctx, domain.AddCoachMessageInput{
		ConversationID: conversation.ID,
		Role:           domain.CoachMessageRoleAssistant,
		Content:        agentResult.Content,
	})
	if err != nil {
		return nil, err
	}
	if err := s.repo.TouchConversation(ctx, userID, conversation.ID); err != nil {
		return nil, err
	}

	toolsUsed := agentResult.ToolsUsed
	if toolsUsed == nil {
		toolsUsed = []string{}
	}

	s.logger.Info(
		"coach chat completed",
		"user_id", userID,
		"conversation_id", conversation.ID,
		"model", s.agent.model,
		"tool_call_count", len(toolsUsed),
		"duration_ms", time.Since(start).Milliseconds(),
		"status", "ok",
	)

	return &domain.CoachChatResult{
		ConversationID: conversation.ID,
		Message:        *assistantMessage,
		ToolsUsed:      toolsUsed,
	}, nil
}

func (s *Service) ListConversations(
	ctx context.Context,
	userID, pageRaw, pageSizeRaw string,
) (*ListConversationsResult, error) {
	userID = strings.TrimSpace(userID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}

	page, pageSize, err := parsePageParams(pageRaw, pageSizeRaw)
	if err != nil {
		return nil, err
	}

	conversations, total, err := s.repo.ListConversations(ctx, domain.ListCoachConversationsFilter{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	if conversations == nil {
		conversations = []domain.CoachConversation{}
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &ListConversationsResult{
		Conversations: conversations,
		Page:          page,
		PageSize:      pageSize,
		TotalItems:    total,
		TotalPages:    totalPages,
	}, nil
}

func (s *Service) GetConversation(ctx context.Context, userID, conversationID string) (*ConversationDetail, error) {
	userID = strings.TrimSpace(userID)
	conversationID = strings.TrimSpace(conversationID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	if !isValidUUID(conversationID) {
		return nil, &domain.ValidationError{Message: "conversation id must be a valid UUID"}
	}

	conversation, err := s.repo.GetConversationByID(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	messages, err := s.repo.ListVisibleMessages(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	if messages == nil {
		messages = []domain.CoachMessage{}
	}

	return &ConversationDetail{
		Conversation: *conversation,
		Messages:     messages,
	}, nil
}

func (s *Service) DeleteConversation(ctx context.Context, userID, conversationID string) error {
	userID = strings.TrimSpace(userID)
	conversationID = strings.TrimSpace(conversationID)
	if !isValidUUID(userID) {
		return &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	if !isValidUUID(conversationID) {
		return &domain.ValidationError{Message: "conversation id must be a valid UUID"}
	}
	return s.repo.DeleteConversation(ctx, userID, conversationID)
}

func buildConversationTitle(message string) string {
	title := strings.Join(strings.Fields(message), " ")
	if utf8.RuneCountInString(title) <= maxConversationTitle {
		return title
	}
	runes := []rune(title)
	return string(runes[:maxConversationTitle-3]) + "..."
}

func parsePageParams(pageRaw, pageSizeRaw string) (int, int, error) {
	page := defaultPage
	pageSize := defaultPageSize

	if value := strings.TrimSpace(pageRaw); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			return 0, 0, &domain.ValidationError{Message: "page must be a positive integer"}
		}
		page = parsed
	}
	if value := strings.TrimSpace(pageSizeRaw); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > maxPageSize {
			return 0, 0, &domain.ValidationError{
				Message: fmt.Sprintf("page_size must be an integer between 1 and %d", maxPageSize),
			}
		}
		pageSize = parsed
	}
	return page, pageSize, nil
}

func isValidUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, c := range value {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !isHex(c) {
				return false
			}
		}
	}
	return true
}

func isHex(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}
