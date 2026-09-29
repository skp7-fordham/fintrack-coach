package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/auth"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/coach"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

type deleteConversationRepository struct {
	owners   map[string]string
	messages map[string][]domain.CoachMessage
}

func (r *deleteConversationRepository) CreateConversation(
	context.Context,
	domain.CreateCoachConversationInput,
) (*domain.CoachConversation, error) {
	return nil, nil
}

func (r *deleteConversationRepository) GetConversationByID(
	_ context.Context,
	userID, conversationID string,
) (*domain.CoachConversation, error) {
	if r.owners[conversationID] != userID {
		return nil, domain.ErrConversationNotFound
	}
	return &domain.CoachConversation{ID: conversationID, UserID: userID}, nil
}

func (r *deleteConversationRepository) ListConversations(
	context.Context,
	domain.ListCoachConversationsFilter,
) ([]domain.CoachConversation, int64, error) {
	return nil, 0, nil
}

func (r *deleteConversationRepository) TouchConversation(context.Context, string, string) error {
	return nil
}

func (r *deleteConversationRepository) AddMessage(
	context.Context,
	domain.AddCoachMessageInput,
) (*domain.CoachMessage, error) {
	return nil, nil
}

func (r *deleteConversationRepository) ListRecentVisibleMessages(
	context.Context,
	string,
	int,
) ([]domain.CoachMessage, error) {
	return nil, nil
}

func (r *deleteConversationRepository) ListVisibleMessages(
	context.Context,
	string,
	string,
) ([]domain.CoachMessage, error) {
	return nil, nil
}

func (r *deleteConversationRepository) DeleteConversation(
	_ context.Context,
	userID, conversationID string,
) error {
	if r.owners[conversationID] != userID {
		return domain.ErrConversationNotFound
	}
	delete(r.owners, conversationID)
	delete(r.messages, conversationID)
	return nil
}

func (r *deleteConversationRepository) ConsumeDemoAIMessage(
	context.Context,
	string,
	time.Time,
	int,
) (bool, error) {
	return true, nil
}

func TestDeleteConversationReturnsNoContentForDemoOwner(t *testing.T) {
	const (
		userID         = "11111111-1111-1111-1111-111111111111"
		conversationID = "22222222-2222-2222-2222-222222222222"
	)
	repo := &deleteConversationRepository{
		owners: map[string]string{conversationID: userID},
		messages: map[string][]domain.CoachMessage{
			conversationID: {{ID: "33333333-3333-3333-3333-333333333333"}},
		},
	}
	handler := newDeleteConversationHandler(repo)
	request := httptest.NewRequest(http.MethodDelete, "/coach/conversations/"+conversationID, nil)
	request.SetPathValue("id", conversationID)
	request = request.WithContext(auth.WithIdentity(request.Context(), userID, true))
	response := httptest.NewRecorder()

	handler.DeleteConversation(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if _, exists := repo.owners[conversationID]; exists {
		t.Fatal("conversation still exists after deletion")
	}
	if _, exists := repo.messages[conversationID]; exists {
		t.Fatal("conversation messages still exist after deletion")
	}
}

func TestDeleteConversationReturnsNotFoundForMissingOrForeignConversation(t *testing.T) {
	const (
		ownerID        = "11111111-1111-1111-1111-111111111111"
		otherUserID    = "22222222-2222-2222-2222-222222222222"
		conversationID = "33333333-3333-3333-3333-333333333333"
		missingID      = "44444444-4444-4444-4444-444444444444"
	)

	for name, testCase := range map[string]struct {
		conversationID string
		userID         string
	}{
		"foreign conversation": {conversationID: conversationID, userID: otherUserID},
		"missing conversation": {conversationID: missingID, userID: ownerID},
	} {
		t.Run(name, func(t *testing.T) {
			repo := &deleteConversationRepository{
				owners:   map[string]string{conversationID: ownerID},
				messages: map[string][]domain.CoachMessage{},
			}
			handler := newDeleteConversationHandler(repo)
			request := httptest.NewRequest(
				http.MethodDelete,
				"/coach/conversations/"+testCase.conversationID,
				nil,
			)
			request.SetPathValue("id", testCase.conversationID)
			request = request.WithContext(auth.WithIdentity(request.Context(), testCase.userID, false))
			response := httptest.NewRecorder()

			handler.DeleteConversation(response, request)

			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", response.Code)
			}
			if repo.owners[conversationID] != ownerID {
				t.Fatal("owner's conversation was deleted")
			}
		})
	}
}

func newDeleteConversationHandler(repo coach.Repository) *CoachHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewCoachHandler(coach.NewService(repo, nil, 5, logger), logger)
}
