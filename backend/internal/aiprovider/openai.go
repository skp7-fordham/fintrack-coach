package aiprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChatMessage is one OpenAI-compatible chat message.
type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall is a model-requested function call.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall carries the tool name and JSON argument payload.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolDefinition describes a function the model may call.
type ToolDefinition struct {
	Type     string             `json:"type"`
	Function ToolFunctionSchema `json:"function"`
}

// ToolFunctionSchema is the JSON-schema style tool description.
type ToolFunctionSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// GenerateRequest is the provider-agnostic coach generation input.
type GenerateRequest struct {
	Model    string
	Messages []ChatMessage
	Tools    []ToolDefinition
}

// GenerateResponse is the provider-agnostic coach generation output.
type GenerateResponse struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
}

// Client calls an OpenAI-compatible chat completions API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey, baseURL string, timeout time.Duration) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type chatCompletionRequest struct {
	Model    string           `json:"model"`
	Messages []ChatMessage    `json:"messages"`
	Tools    []ToolDefinition `json:"tools,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role      string     `json:"role"`
			Content   *string    `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// Generate performs one chat completion turn.
func (c *Client) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return GenerateResponse{}, fmt.Errorf("ai api key is not configured")
	}
	if strings.TrimSpace(req.Model) == "" {
		return GenerateResponse{}, fmt.Errorf("ai model is required")
	}

	body, err := json.Marshal(chatCompletionRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Tools:    req.Tools,
	})
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("create chat request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("ai provider request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("read ai provider response: %w", err)
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return GenerateResponse{}, fmt.Errorf("decode ai provider response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := "ai provider request failed"
		if parsed.Error != nil && parsed.Error.Message != "" {
			message = "ai provider request failed"
		}
		return GenerateResponse{}, fmt.Errorf("%s: status %d", message, resp.StatusCode)
	}

	if len(parsed.Choices) == 0 {
		return GenerateResponse{}, fmt.Errorf("ai provider returned no choices")
	}

	choice := parsed.Choices[0]
	content := ""
	if choice.Message.Content != nil {
		content = strings.TrimSpace(*choice.Message.Content)
	}

	return GenerateResponse{
		Content:      content,
		ToolCalls:    choice.Message.ToolCalls,
		FinishReason: choice.FinishReason,
	}, nil
}
