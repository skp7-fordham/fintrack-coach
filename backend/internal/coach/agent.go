package coach

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/aiprovider"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

// LLM generates coach replies and optional tool calls.
type LLM interface {
	Generate(ctx context.Context, req aiprovider.GenerateRequest) (aiprovider.GenerateResponse, error)
}

// ToolExecutor validates and runs registered financial tools.
type ToolExecutor interface {
	Definitions() []aiprovider.ToolDefinition
	Execute(ctx context.Context, userID, toolName, rawArgs string) (string, error)
}

type Agent struct {
	llm           LLM
	tools         ToolExecutor
	model         string
	maxIterations int
	now           Clock
	logger        *slog.Logger
}

type AgentResult struct {
	Content   string
	ToolsUsed []string
}

func NewAgent(llm LLM, tools ToolExecutor, model string, maxIterations int, logger *slog.Logger) *Agent {
	return NewAgentWithClock(llm, tools, model, maxIterations, realUTCClock, logger)
}

func NewAgentWithClock(
	llm LLM,
	tools ToolExecutor,
	model string,
	maxIterations int,
	now Clock,
	logger *slog.Logger,
) *Agent {
	if now == nil {
		now = realUTCClock
	}
	return &Agent{
		llm:           llm,
		tools:         tools,
		model:         model,
		maxIterations: maxIterations,
		now:           now,
		logger:        logger,
	}
}

func (a *Agent) Run(ctx context.Context, userID string, history []domain.CoachMessage, userMessage string) (*AgentResult, error) {
	if a.llm == nil {
		return nil, domain.ErrCoachUnavailable
	}

	messages := make([]aiprovider.ChatMessage, 0, len(history)+2)
	messages = append(messages, aiprovider.ChatMessage{
		Role:    "system",
		Content: buildSystemPrompt(a.now()),
	})
	for _, msg := range history {
		if msg.Role != domain.CoachMessageRoleUser && msg.Role != domain.CoachMessageRoleAssistant {
			continue
		}
		messages = append(messages, aiprovider.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	messages = append(messages, aiprovider.ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	toolsUsed := make([]string, 0)
	seenTools := make(map[string]struct{})
	definitions := a.tools.Definitions()

	for iteration := 0; iteration < a.maxIterations; iteration++ {
		start := time.Now()
		response, err := a.llm.Generate(ctx, aiprovider.GenerateRequest{
			Model:    a.model,
			Messages: messages,
			Tools:    definitions,
		})
		a.logger.Info(
			"coach llm generate",
			"user_id", userID,
			"model", a.model,
			"iteration", iteration+1,
			"duration_ms", time.Since(start).Milliseconds(),
			"tool_call_count", len(response.ToolCalls),
			"ok", err == nil,
		)
		if err != nil {
			return nil, mapProviderError(err)
		}

		if len(response.ToolCalls) == 0 {
			content := strings.TrimSpace(response.Content)
			if content == "" {
				return nil, fmt.Errorf("%w", errEmptyAssistantAnswer)
			}
			return &AgentResult{
				Content:   content,
				ToolsUsed: toolsUsed,
			}, nil
		}

		messages = append(messages, aiprovider.ChatMessage{
			Role:      "assistant",
			Content:   response.Content,
			ToolCalls: response.ToolCalls,
		})

		for _, call := range response.ToolCalls {
			name := strings.TrimSpace(call.Function.Name)
			if name == "" {
				name = "unknown"
			}
			if _, ok := seenTools[name]; !ok {
				seenTools[name] = struct{}{}
				toolsUsed = append(toolsUsed, name)
			}

			toolResult, execErr := a.tools.Execute(ctx, userID, name, call.Function.Arguments)
			if execErr != nil {
				toolResult = marshalToolError(execErr)
			}

			messages = append(messages, aiprovider.ChatMessage{
				Role:       "tool",
				ToolCallID: call.ID,
				Name:       name,
				Content:    toolResult,
			})
		}
	}

	return nil, domain.ErrCoachToolLimitExceeded
}

func mapProviderError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return domain.ErrCoachUnavailable
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "api key") ||
		strings.Contains(msg, "status 401") ||
		strings.Contains(msg, "status 403") ||
		strings.Contains(msg, "status 429") ||
		strings.Contains(msg, "status 5") ||
		strings.Contains(msg, "request failed") ||
		strings.Contains(msg, "no choices") {
		return domain.ErrCoachUnavailable
	}
	return domain.ErrCoachUnavailable
}

func marshalToolError(err error) string {
	message := "tool execution failed"
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		message = validationErr.Message
	case errors.Is(err, errUnknownTool):
		message = "unknown tool"
	case errors.Is(err, errMalformedToolArgs):
		message = "malformed tool arguments"
	}
	payload, marshalErr := json.Marshal(map[string]string{"error": message})
	if marshalErr != nil {
		return `{"error":"tool execution failed"}`
	}
	return string(payload)
}
