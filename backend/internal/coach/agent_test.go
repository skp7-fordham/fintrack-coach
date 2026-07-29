package coach

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/aiprovider"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
)

type stubLLM struct {
	responses []aiprovider.GenerateResponse
	errs      []error
	calls     []aiprovider.GenerateRequest
}

func (s *stubLLM) Generate(ctx context.Context, req aiprovider.GenerateRequest) (aiprovider.GenerateResponse, error) {
	s.calls = append(s.calls, req)
	idx := len(s.calls) - 1
	if idx < len(s.errs) && s.errs[idx] != nil {
		return aiprovider.GenerateResponse{}, s.errs[idx]
	}
	if idx >= len(s.responses) {
		return aiprovider.GenerateResponse{}, errors.New("unexpected llm call")
	}
	return s.responses[idx], nil
}

type recordingTools struct {
	executed   []toolInvocation
	results    map[string]string
	execErrs   map[string]error
	lastUserID string
}

type toolInvocation struct {
	UserID  string
	Name    string
	RawArgs string
}

func (t *recordingTools) Definitions() []aiprovider.ToolDefinition {
	return []aiprovider.ToolDefinition{
		{
			Type: "function",
			Function: aiprovider.ToolFunctionSchema{
				Name:        toolGetDashboardSummary,
				Description: "summary",
				Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
			},
		},
		{
			Type: "function",
			Function: aiprovider.ToolFunctionSchema{
				Name:        toolListAccounts,
				Description: "accounts",
				Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
			},
		},
	}
}

func (t *recordingTools) Execute(ctx context.Context, userID, toolName, rawArgs string) (string, error) {
	t.lastUserID = userID
	t.executed = append(t.executed, toolInvocation{
		UserID:  userID,
		Name:    toolName,
		RawArgs: rawArgs,
	})
	if err, ok := t.execErrs[toolName]; ok {
		return "", err
	}
	if result, ok := t.results[toolName]; ok {
		return result, nil
	}
	if toolName != toolGetDashboardSummary && toolName != toolListAccounts {
		return "", errUnknownTool
	}
	return `{"ok":true}`, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAgentExecutesOneToolAndReturnsFinalResponse(t *testing.T) {
	llm := &stubLLM{
		responses: []aiprovider.GenerateResponse{
			{
				ToolCalls: []aiprovider.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Function: aiprovider.FunctionCall{
							Name:      toolGetDashboardSummary,
							Arguments: `{"month":"2026-07"}`,
						},
					},
				},
			},
			{Content: "Your net cash flow looks healthy this month."},
		},
	}
	tools := &recordingTools{
		results: map[string]string{
			toolGetDashboardSummary: `{"monthly_expense":"100.00"}`,
		},
	}
	agent := NewAgent(llm, tools, "test-model", 5, testLogger())

	result, err := agent.Run(context.Background(), "11111111-1111-1111-1111-111111111111", nil, "How much did I spend?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "Your net cash flow looks healthy this month." {
		t.Fatalf("unexpected content: %q", result.Content)
	}
	if len(result.ToolsUsed) != 1 || result.ToolsUsed[0] != toolGetDashboardSummary {
		t.Fatalf("unexpected tools used: %#v", result.ToolsUsed)
	}
	if len(tools.executed) != 1 {
		t.Fatalf("expected one tool execution, got %d", len(tools.executed))
	}
}

func TestAgentExecutesMultipleTools(t *testing.T) {
	llm := &stubLLM{
		responses: []aiprovider.GenerateResponse{
			{
				ToolCalls: []aiprovider.ToolCall{
					{
						ID:       "call_1",
						Type:     "function",
						Function: aiprovider.FunctionCall{Name: toolGetDashboardSummary, Arguments: `{}`},
					},
					{
						ID:       "call_2",
						Type:     "function",
						Function: aiprovider.FunctionCall{Name: toolListAccounts, Arguments: `{}`},
					},
				},
			},
			{Content: "Here is your summary and account breakdown."},
		},
	}
	tools := &recordingTools{results: map[string]string{
		toolGetDashboardSummary: `{"ok":true}`,
		toolListAccounts:        `{"accounts":[]}`,
	}}
	agent := NewAgent(llm, tools, "test-model", 5, testLogger())

	result, err := agent.Run(context.Background(), "11111111-1111-1111-1111-111111111111", nil, "Summarise my finances")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools.executed) != 2 {
		t.Fatalf("expected two tool executions, got %d", len(tools.executed))
	}
	if len(result.ToolsUsed) != 2 {
		t.Fatalf("expected two tools used, got %#v", result.ToolsUsed)
	}
}

func TestAgentRejectsUnknownToolSafely(t *testing.T) {
	llm := &stubLLM{
		responses: []aiprovider.GenerateResponse{
			{
				ToolCalls: []aiprovider.ToolCall{
					{
						ID:       "call_1",
						Type:     "function",
						Function: aiprovider.FunctionCall{Name: "drop_database", Arguments: `{}`},
					},
				},
			},
			{Content: "I could not use an unsupported operation."},
		},
	}
	tools := &recordingTools{}
	agent := NewAgent(llm, tools, "test-model", 5, testLogger())

	result, err := agent.Run(context.Background(), "11111111-1111-1111-1111-111111111111", nil, "Hack my data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content == "" {
		t.Fatal("expected final content")
	}
	if len(llm.calls) < 2 {
		t.Fatal("expected a follow-up llm call after tool error")
	}
	lastToolMsg := llm.calls[1].Messages[len(llm.calls[1].Messages)-1]
	if lastToolMsg.Role != "tool" || !strings.Contains(lastToolMsg.Content, "unknown tool") {
		t.Fatalf("expected unknown tool error payload, got %#v", lastToolMsg)
	}
}

func TestAgentRejectsMalformedToolArguments(t *testing.T) {
	tools := NewToolRegistry(nil, nil, nil, testLogger())
	_, err := tools.Execute(
		context.Background(),
		"11111111-1111-1111-1111-111111111111",
		toolGetMonthlyTrends,
		`{"months":"abc"}`,
	)
	if !errors.Is(err, errMalformedToolArgs) {
		t.Fatalf("expected malformed args error, got %v", err)
	}
}

func TestAgentToolIterationLimit(t *testing.T) {
	llm := &stubLLM{
		responses: []aiprovider.GenerateResponse{
			{ToolCalls: []aiprovider.ToolCall{{ID: "1", Type: "function", Function: aiprovider.FunctionCall{Name: toolListAccounts, Arguments: `{}`}}}},
			{ToolCalls: []aiprovider.ToolCall{{ID: "2", Type: "function", Function: aiprovider.FunctionCall{Name: toolListAccounts, Arguments: `{}`}}}},
			{ToolCalls: []aiprovider.ToolCall{{ID: "3", Type: "function", Function: aiprovider.FunctionCall{Name: toolListAccounts, Arguments: `{}`}}}},
		},
	}
	tools := &recordingTools{results: map[string]string{toolListAccounts: `{"accounts":[]}`}}
	agent := NewAgent(llm, tools, "test-model", 2, testLogger())

	_, err := agent.Run(context.Background(), "11111111-1111-1111-1111-111111111111", nil, "Keep checking")
	if !errors.Is(err, domain.ErrCoachToolLimitExceeded) {
		t.Fatalf("expected tool limit error, got %v", err)
	}
}

func TestAgentInjectsAuthenticatedUserID(t *testing.T) {
	userID := "22222222-2222-2222-2222-222222222222"
	llm := &stubLLM{
		responses: []aiprovider.GenerateResponse{
			{
				ToolCalls: []aiprovider.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Function: aiprovider.FunctionCall{
							Name:      toolListAccounts,
							Arguments: `{"user_id":"attacker-id"}`,
						},
					},
				},
			},
			{Content: "Done"},
		},
	}
	tools := &recordingTools{results: map[string]string{toolListAccounts: `{"accounts":[]}`}}
	agent := NewAgent(llm, tools, "test-model", 5, testLogger())

	if _, err := agent.Run(context.Background(), userID, nil, "List accounts"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tools.lastUserID != userID {
		t.Fatalf("expected injected user id %s, got %s", userID, tools.lastUserID)
	}
	if len(tools.executed) != 1 {
		t.Fatalf("expected one execution, got %d", len(tools.executed))
	}
}

func TestProviderFailureMapsSafely(t *testing.T) {
	llm := &stubLLM{
		errs: []error{errors.New("ai provider request failed: status 503")},
	}
	agent := NewAgent(llm, &recordingTools{}, "test-model", 5, testLogger())

	_, err := agent.Run(context.Background(), "11111111-1111-1111-1111-111111111111", nil, "Hello")
	if !errors.Is(err, domain.ErrCoachUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

type fakeCoachRepo struct {
	conversations map[string]*domain.CoachConversation
	messages      map[string][]domain.CoachMessage
}

func newFakeCoachRepo() *fakeCoachRepo {
	return &fakeCoachRepo{
		conversations: make(map[string]*domain.CoachConversation),
		messages:      make(map[string][]domain.CoachMessage),
	}
}

func (f *fakeCoachRepo) CreateConversation(ctx context.Context, input domain.CreateCoachConversationInput) (*domain.CoachConversation, error) {
	conversation := &domain.CoachConversation{
		ID:     "33333333-3333-3333-3333-333333333333",
		UserID: input.UserID,
		Title:  input.Title,
	}
	f.conversations[conversation.ID] = conversation
	return conversation, nil
}

func (f *fakeCoachRepo) GetConversationByID(ctx context.Context, userID, conversationID string) (*domain.CoachConversation, error) {
	conversation, ok := f.conversations[conversationID]
	if !ok || conversation.UserID != userID {
		return nil, domain.ErrConversationNotFound
	}
	copyConversation := *conversation
	return &copyConversation, nil
}

func (f *fakeCoachRepo) ListConversations(ctx context.Context, filter domain.ListCoachConversationsFilter) ([]domain.CoachConversation, int64, error) {
	return nil, 0, nil
}

func (f *fakeCoachRepo) TouchConversation(ctx context.Context, userID, conversationID string) error {
	_, err := f.GetConversationByID(ctx, userID, conversationID)
	return err
}

func (f *fakeCoachRepo) AddMessage(ctx context.Context, input domain.AddCoachMessageInput) (*domain.CoachMessage, error) {
	message := domain.CoachMessage{
		ID:             "44444444-4444-4444-4444-444444444444",
		ConversationID: input.ConversationID,
		Role:           input.Role,
		Content:        input.Content,
	}
	f.messages[input.ConversationID] = append(f.messages[input.ConversationID], message)
	return &message, nil
}

func (f *fakeCoachRepo) ListRecentVisibleMessages(ctx context.Context, conversationID string, limit int) ([]domain.CoachMessage, error) {
	return append([]domain.CoachMessage{}, f.messages[conversationID]...), nil
}

func (f *fakeCoachRepo) ListVisibleMessages(ctx context.Context, userID, conversationID string) ([]domain.CoachMessage, error) {
	if _, err := f.GetConversationByID(ctx, userID, conversationID); err != nil {
		return nil, err
	}
	return append([]domain.CoachMessage{}, f.messages[conversationID]...), nil
}

func (f *fakeCoachRepo) DeleteConversation(ctx context.Context, userID, conversationID string) error {
	if _, err := f.GetConversationByID(ctx, userID, conversationID); err != nil {
		return err
	}
	delete(f.conversations, conversationID)
	delete(f.messages, conversationID)
	return nil
}

func TestConversationBelongingToAnotherUserReturnsNotFound(t *testing.T) {
	repo := newFakeCoachRepo()
	ownerID := "11111111-1111-1111-1111-111111111111"
	otherID := "22222222-2222-2222-2222-222222222222"
	conversation, err := repo.CreateConversation(context.Background(), domain.CreateCoachConversationInput{
		UserID: ownerID,
		Title:  strPtr("Mine"),
	})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	svc := NewService(repo, NewAgent(&stubLLM{}, &recordingTools{}, "test-model", 5, testLogger()), testLogger())
	_, err = svc.GetConversation(context.Background(), otherID, conversation.ID)
	if !errors.Is(err, domain.ErrConversationNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}

	foreignID := conversation.ID
	_, err = svc.Chat(context.Background(), otherID, dto.CoachChatRequest{
		Message:        "Hello",
		ConversationID: &foreignID,
	})
	if !errors.Is(err, domain.ErrConversationNotFound) {
		t.Fatalf("expected not found on chat, got %v", err)
	}
}

func TestEmptyChatMessageReturnsValidationError(t *testing.T) {
	svc := NewService(newFakeCoachRepo(), NewAgent(&stubLLM{}, &recordingTools{}, "test-model", 5, testLogger()), testLogger())
	_, err := svc.Chat(context.Background(), "11111111-1111-1111-1111-111111111111", dto.CoachChatRequest{
		Message: "   ",
	})
	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if validationErr.Message != "message is required" {
		t.Fatalf("unexpected message: %s", validationErr.Message)
	}
}

func TestParseToolArgsIgnoresModelUserID(t *testing.T) {
	args, err := parseToolArgs(`{"user_id":"attacker","month":"2026-07"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	delete(args, "user_id")
	if _, ok := args["user_id"]; ok {
		t.Fatal("user_id should be removed before tool execution")
	}
	encoded, _ := json.Marshal(args)
	if strings.Contains(string(encoded), "attacker") {
		t.Fatal("attacker user id should not remain")
	}
}

func strPtr(value string) *string {
	return &value
}
