package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

type CoachRepository struct {
	pool *pgxpool.Pool
}

func NewCoachRepository(pool *pgxpool.Pool) *CoachRepository {
	return &CoachRepository{pool: pool}
}

func (r *CoachRepository) CreateConversation(
	ctx context.Context,
	input domain.CreateCoachConversationInput,
) (*domain.CoachConversation, error) {
	const query = `
		INSERT INTO coach_conversations (user_id, title)
		VALUES ($1, $2)
		RETURNING
			id::text,
			user_id::text,
			title,
			created_at,
			updated_at
	`

	conversation, err := scanCoachConversation(r.pool.QueryRow(ctx, query, input.UserID, input.Title))
	if err != nil {
		return nil, fmt.Errorf("create coach conversation: %w", err)
	}
	return conversation, nil
}

func (r *CoachRepository) GetConversationByID(
	ctx context.Context,
	userID, conversationID string,
) (*domain.CoachConversation, error) {
	const query = `
		SELECT
			id::text,
			user_id::text,
			title,
			created_at,
			updated_at
		FROM coach_conversations
		WHERE id = $1 AND user_id = $2
	`

	conversation, err := scanCoachConversation(r.pool.QueryRow(ctx, query, conversationID, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConversationNotFound
		}
		return nil, fmt.Errorf("get coach conversation: %w", err)
	}
	return conversation, nil
}

func (r *CoachRepository) ListConversations(
	ctx context.Context,
	filter domain.ListCoachConversationsFilter,
) ([]domain.CoachConversation, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM coach_conversations
		WHERE user_id = $1
	`

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, filter.UserID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count coach conversations: %w", err)
	}

	const listQuery = `
		SELECT
			id::text,
			user_id::text,
			title,
			created_at,
			updated_at
		FROM coach_conversations
		WHERE user_id = $1
		ORDER BY updated_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`

	offset := (filter.Page - 1) * filter.PageSize
	rows, err := r.pool.Query(ctx, listQuery, filter.UserID, filter.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list coach conversations: %w", err)
	}
	defer rows.Close()

	conversations := make([]domain.CoachConversation, 0)
	for rows.Next() {
		conversation, err := scanCoachConversation(rows)
		if err != nil {
			return nil, 0, err
		}
		conversations = append(conversations, *conversation)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate coach conversations: %w", err)
	}

	return conversations, total, nil
}

func (r *CoachRepository) TouchConversation(ctx context.Context, userID, conversationID string) error {
	const query = `
		UPDATE coach_conversations
		SET updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`

	tag, err := r.pool.Exec(ctx, query, conversationID, userID)
	if err != nil {
		return fmt.Errorf("touch coach conversation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConversationNotFound
	}
	return nil
}

func (r *CoachRepository) AddMessage(
	ctx context.Context,
	input domain.AddCoachMessageInput,
) (*domain.CoachMessage, error) {
	const query = `
		INSERT INTO coach_messages (
			conversation_id,
			role,
			content,
			tool_name,
			tool_call_id
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id::text,
			conversation_id::text,
			role,
			content,
			tool_name,
			tool_call_id,
			created_at
	`

	message, err := scanCoachMessage(r.pool.QueryRow(
		ctx,
		query,
		input.ConversationID,
		input.Role,
		input.Content,
		input.ToolName,
		input.ToolCallID,
	))
	if err != nil {
		return nil, fmt.Errorf("add coach message: %w", err)
	}
	return message, nil
}

func (r *CoachRepository) ListRecentVisibleMessages(
	ctx context.Context,
	conversationID string,
	limit int,
) ([]domain.CoachMessage, error) {
	const query = `
		SELECT
			id::text,
			conversation_id::text,
			role,
			content,
			tool_name,
			tool_call_id,
			created_at
		FROM (
			SELECT
				id,
				conversation_id,
				role,
				content,
				tool_name,
				tool_call_id,
				created_at
			FROM coach_messages
			WHERE conversation_id = $1
			  AND role IN ('user', 'assistant')
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		) recent
		ORDER BY created_at ASC, id ASC
	`

	rows, err := r.pool.Query(ctx, query, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent coach messages: %w", err)
	}
	defer rows.Close()

	messages := make([]domain.CoachMessage, 0)
	for rows.Next() {
		message, err := scanCoachMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, *message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent coach messages: %w", err)
	}
	return messages, nil
}

func (r *CoachRepository) ListVisibleMessages(
	ctx context.Context,
	userID, conversationID string,
) ([]domain.CoachMessage, error) {
	const query = `
		SELECT
			m.id::text,
			m.conversation_id::text,
			m.role,
			m.content,
			m.tool_name,
			m.tool_call_id,
			m.created_at
		FROM coach_messages m
		INNER JOIN coach_conversations c ON c.id = m.conversation_id
		WHERE m.conversation_id = $1
		  AND c.user_id = $2
		  AND m.role IN ('user', 'assistant')
		ORDER BY m.created_at ASC, m.id ASC
	`

	rows, err := r.pool.Query(ctx, query, conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("list visible coach messages: %w", err)
	}
	defer rows.Close()

	messages := make([]domain.CoachMessage, 0)
	for rows.Next() {
		message, err := scanCoachMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, *message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate visible coach messages: %w", err)
	}
	return messages, nil
}

func (r *CoachRepository) DeleteConversation(ctx context.Context, userID, conversationID string) error {
	const query = `
		DELETE FROM coach_conversations
		WHERE id = $1 AND user_id = $2
	`

	tag, err := r.pool.Exec(ctx, query, conversationID, userID)
	if err != nil {
		return fmt.Errorf("delete coach conversation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConversationNotFound
	}
	return nil
}

type coachRowScanner interface {
	Scan(dest ...any) error
}

func scanCoachConversation(row coachRowScanner) (*domain.CoachConversation, error) {
	var conversation domain.CoachConversation
	if err := row.Scan(
		&conversation.ID,
		&conversation.UserID,
		&conversation.Title,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &conversation, nil
}

func scanCoachMessage(row coachRowScanner) (*domain.CoachMessage, error) {
	var message domain.CoachMessage
	if err := row.Scan(
		&message.ID,
		&message.ConversationID,
		&message.Role,
		&message.Content,
		&message.ToolName,
		&message.ToolCallID,
		&message.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &message, nil
}
