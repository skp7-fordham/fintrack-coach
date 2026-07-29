package domain

import (
	"errors"
	"time"
)

const (
	CoachMessageRoleUser      = "user"
	CoachMessageRoleAssistant = "assistant"
	CoachMessageRoleTool      = "tool"
)

var (
	ErrConversationNotFound   = errors.New("conversation not found")
	ErrCoachUnavailable       = errors.New("financial coach is temporarily unavailable")
	ErrCoachToolLimitExceeded = errors.New("financial coach tool limit exceeded")
)

// CoachConversation is a persisted coach chat thread.
type CoachConversation struct {
	ID        string
	UserID    string
	Title     *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CoachMessage is one persisted message in a coach conversation.
type CoachMessage struct {
	ID             string
	ConversationID string
	Role           string
	Content        string
	ToolName       *string
	ToolCallID     *string
	CreatedAt      time.Time
}

type CreateCoachConversationInput struct {
	UserID string
	Title  *string
}

type AddCoachMessageInput struct {
	ConversationID string
	Role           string
	Content        string
	ToolName       *string
	ToolCallID     *string
}

type ListCoachConversationsFilter struct {
	UserID   string
	Page     int
	PageSize int
}

// CoachChatResult is the service result for POST /coach/chat.
type CoachChatResult struct {
	ConversationID string
	Message        CoachMessage
	ToolsUsed      []string
}
